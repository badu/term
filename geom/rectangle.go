package geom

import (
	"context"
	"errors"
	"log"
	"sync"

	"github.com/badu/term"
	"github.com/badu/term/color"
	"github.com/badu/term/style"
)

var (
	rectCounter = 0
	mu          sync.Mutex
)

// we cannot compare rectangles, for this reason we're using this incremental id
func getNextRectId() int {
	mu.Lock()
	defer mu.Unlock()
	rectCounter++
	return rectCounter
}

// RectangleOption
type RectangleOption func(r *Rectangle)

// WithBackgroundColor
func WithBackgroundColor(c color.Color) RectangleOption {
	return func(r *Rectangle) {
		r.st.Bg = c
	}
}

// WithForegroundColor
func WithForegroundColor(c color.Color) RectangleOption {
	return func(r *Rectangle) {
		r.st.Fg = c
	}
}

// WithTopCorner
func WithTopCorner(col, row int) RectangleOption {
	return func(r *Rectangle) {
		r.topCorner = term.NewPosition(col, row)
	}
}

// WithBottomCorner
func WithBottomCorner(col, row int) RectangleOption {
	return func(r *Rectangle) {
		r.bottomCorner = term.NewPosition(col, row)
	}
}

// WithAlignment
func WithAlignment(a style.Alignment) RectangleOption {
	return func(r *Rectangle) {
		r.aligned = a
	}
}

// WithOrientation
func WithOrientation(o style.Orientation) RectangleOption {
	return func(r *Rectangle) {
		r.orientation = o
	}
}

// WithMinSize
func WithMinSize(size *term.Size) RectangleOption {
	return func(r *Rectangle) {
		r.min = size
	}
}

// WithChildren
func WithChildren(children ...*Rectangle) RectangleOption {
	return func(r *Rectangle) {
		for _, child := range children {
			if child.st.Bg == color.Default {
				child.st.Bg = r.st.Bg // inherit background color
			}
			if child.st.Fg == color.Default {
				child.st.Fg = r.st.Fg // inherit foreground color
			}
			wasAlreadyAddedToChildren := false
			if child.width != nil {
				r.children = append(r.children, child) // mark it as resizable in this parent
				wasAlreadyAddedToChildren = true
			}
			if child.height != nil {
				if !wasAlreadyAddedToChildren {
					r.children = append(r.children, child) // mark it as resizable in this parent
				}
			}
		}
	}
}

// AsLine
func AsLine(line, startColumn, endColumn int) RectangleOption {
	return func(r *Rectangle) {
		r.topCorner = term.NewPosition(startColumn, line)
		r.bottomCorner = term.NewPosition(endColumn, line)
	}
}

// AsColumn
func AsColumn(column, startRow, endRow int) RectangleOption {
	return func(r *Rectangle) {
		r.topCorner = term.NewPosition(column, startRow)
		r.bottomCorner = term.NewPosition(column, endRow)
	}
}

// WithRowColAndSize
func WithRowColAndSize(row, column, numRows, numCols int) RectangleOption {
	return func(r *Rectangle) {
		r.topCorner = term.NewPosition(column, row)
		r.bottomCorner = term.NewPosition(column+numCols, row+numRows)
	}
}

// WithAcquisitionChan is provided by the caller, so we can "ask" for pixels
func WithAcquisitionChan(pixelCh chan term.Position) RectangleOption {
	return func(r *Rectangle) {
		r.pixelAskCh = pixelCh
	}
}

// WithReleasingChan is provided by the caller, so we can "free" pixels
func WithReleasingChan(relePixelCh chan term.Position) RectangleOption {
	return func(r *Rectangle) {
		r.pixelReleaseCh = relePixelCh
	}
}

// Rectangle describes a colored rectangle primitive
type Rectangle struct {
	id             int                //
	rows           [][]px             // when organized by rows
	cols           [][]px             // when organized by cols
	children       []*Rectangle       // children Rectangles, used for variable sizing // TODO : mount death listener for children
	topCorner      *term.Position     // The current top corner of the Rectangle
	bottomCorner   *term.Position     // The current top corner of the Rectangle
	aligned        style.Alignment    // alignment. Default is style.Begin (topCorner)
	orientation    style.Orientation  // orientation dictates pixel slices above (rows or cols). Default orientation is style.Vertical
	st             style.Style        //
	died           chan struct{}      // Channel for killing (context.Done)
	pixelAskCh     chan term.Position // Channel for asking pixels
	pixelReleaseCh chan term.Position // Channel for releasing pixels
	pixelReceiveCh chan px            // Channel for receiving pixels
	width          *int               // width in percents, pointer indicates is optional
	height         *int               // height in percents, pointer indicates is optional
	min            *term.Size         // The minimum size this object can be. Note that this can exist in the same time with width/height in percents // TODO : check new size (%) is not smaller than min allowed size
	hidden         bool               // Is this object currently hidden
}

// TODO : thinking maybe this should be a private constructor. Ask Page to give you a Rectangle and it will give it already populated and ready to use. For now (testing purposes), I'll leave it as it is.
// NewRectangle returns a new Rectangle instance
func NewRectangle(ctx context.Context, opts ...RectangleOption) (*Rectangle, error) {
	defStyle := style.NewStyle(style.WithBg(color.Default), style.WithFg(color.Default), style.WithAttrs(style.None))
	res := &Rectangle{
		id:             getNextRectId(), // TODO : use some form of hash, maybe ???
		orientation:    style.Vertical,  // default orientation
		aligned:        style.Begin,     // aligned top left
		st:             *defStyle,
		pixelAskCh:     make(chan term.Position),
		pixelReleaseCh: make(chan term.Position),
		pixelReceiveCh: make(chan px),
		died:           make(chan struct{}),
		topCorner:      term.NewPosition(-1, -1), // default
		bottomCorner:   term.NewPosition(-1, -1),
	}
	for _, opt := range opts {
		opt(res)
	}
	if res.OffScreen() {
		return nil, errors.New("rectangle lacks valid coordinates")
	}
	if res.pixelAskCh == nil {
		return nil, errors.New("acquisition channel is mandatory")
	}
	rectSize := res.Size() // init rows/cols
	if rectSize.Width <= 0 || rectSize.Height <= 0 {
		return nil, errors.New("invalid rectangle size")
	}
	log.Printf("%03d rows %03d colums", rectSize.Height, rectSize.Width)
	switch res.orientation {
	case style.Vertical:
		log.Println("Vertical (ROWS)")
		res.rows = make([][]px, rectSize.Height)
		for row := 0; row < rectSize.Height; row++ {
			res.rows[row] = make([]px, rectSize.Width)
		}
	case style.Horizontal:
		log.Println("Horizontal (COLUMNS)")
		res.cols = make([][]px, rectSize.Width)
		for col := 0; col < rectSize.Width; col++ {
			res.cols[col] = make([]px, rectSize.Height)
		}
	default:
		return nil, errors.New("orientation must be horizontal (rows) or vertical (columns)")
	}
	// all ok, listening for context.Done to exit
	go func() {
		for {
			select {
			case <-ctx.Done():
				res.releasePositions()
				res.died <- struct{}{} // tell the Page that we've died, to free up pixels
				return
			case pix := <-res.pixelReceiveCh:
				res.registerPixel(pix)
			}
		}
	}()
	return res, nil
}

// OffScreen returns true if coordinates are not set
func (r *Rectangle) OffScreen() bool {
	return r.topCorner.Row < 0 || r.topCorner.Column < 0 || r.bottomCorner.Row < 0 || r.bottomCorner.Column < 0
}

// Id is used by Page to identify owners and operate them
func (r *Rectangle) Id() int {
	return r.id
}

// PixelAskChan - used by Page to deliver pixels
func (r *Rectangle) PixelAskChan() chan term.Position {
	return r.pixelAskCh
}

// PixelReleaseChan - use by Page to change owner of the pixel
func (r *Rectangle) PixelReleaseChan() chan term.Position {
	return r.pixelReleaseCh
}

// PixelReceiverChan - used by Rectangle to get pixels from Page
func (r *Rectangle) PixelReceiverChan() chan px {
	return r.pixelReceiveCh
}

// CurrentSize returns the current size of this rectangle object
func (r *Rectangle) Size() *term.Size {
	return term.NewSize(term.Width(r.topCorner, r.bottomCorner), term.Height(r.topCorner, r.bottomCorner))
}

// Width returns r's width.
func (r *Rectangle) Width() int {
	return term.Width(r.topCorner, r.bottomCorner)
}

// Height returns r's height.
func (r *Rectangle) Height() int {
	return term.Height(r.topCorner, r.bottomCorner)
}

// Center of a Rectangle
func (r *Rectangle) Center() *term.Position {
	return term.Center(r.topCorner, r.bottomCorner)
}

// HasPerfectCenter - allows the caller of Center not to fool themselves (e.g. center is shifted towards top right)
func (r *Rectangle) HasPerfectCenter() bool {
	return term.Width(r.topCorner, r.bottomCorner)%2 == 1 && term.Height(r.topCorner, r.bottomCorner)%2 == 1
}

// Row returns the row of pixels at index (absolute, starting with zero)
// if Rectangle is vertical oriented (columns), returns nothing (shame on caller)
func (r *Rectangle) Row(index int) []px {
	if r.orientation == style.Horizontal {
		log.Println("bad call to Rectangle.Row : orientation is not horizontal")
		return nil
	}
	if index <= 0 {
		log.Println("bad call to Rectangle.Row : bad index")
		return nil
	}
	if index-1 > len(r.rows) {
		log.Println("bad call to Rectangle.Row : index outside number of rows")
		return nil
	}
	return r.rows[index-1]
}

// Column returns the column of pixels at index (absolute, starting with zero)
// if Rectangle is horizontal oriented (rows), returns nothing (shame on caller)
func (r *Rectangle) Column(index int) []px {
	if r.orientation == style.Vertical {
		log.Println("bad call to Rectangle.Column : orientation is not vertical")
		return nil
	}
	if index <= 0 {
		log.Println("bad call to Rectangle.Column : bad index")
		return nil
	}
	if index-1 > len(r.cols) {
		log.Println("bad call to Rectangle.Column : index outside number of columns")
		return nil
	}
	return r.cols[index-1]
}

// Move the rectangle object to a new position, relative to its children / canvas
func (r *Rectangle) Move(pos *term.Position) {
	r.topCorner = pos
	r.acquirePositions()
}

// MinSize returns the specified minimum size, if set, or nil otherwise
func (r *Rectangle) MinSize() *term.Size {
	return r.min
}

// SetMinSize specifies the smallest size this object should be
func (r *Rectangle) SetMinSize(size *term.Size) {
	r.min = size
}

// IsVisible returns true if this object is visible, false otherwise
func (r *Rectangle) Visible() bool {
	return !r.hidden
}

// Show will set this object to be visible
func (r *Rectangle) Show() {
	r.hidden = false
	r.acquirePositions()
}

// Hide will set this object to not be visible
func (r *Rectangle) Hide() {
	r.hidden = true
	r.releasePositions()
}

// Bg
func (r *Rectangle) Bg() color.Color {
	return r.st.Bg
}

// Fg
func (r *Rectangle) Fg() color.Color {
	return r.st.Fg
}

// Top
func (r *Rectangle) Top() *term.Position {
	return r.topCorner
}

// Bottom
func (r *Rectangle) Bottom() *term.Position {
	return r.bottomCorner
}

// Empty returns if the Rectangle is zero rows and zero columns
func (r *Rectangle) Empty() bool {
	return r.topCorner.Row >= r.bottomCorner.Row || r.topCorner.Column >= r.bottomCorner.Column
}

// Union returns the smallest rectangle that contains both r and s.
func (r *Rectangle) Union(s *Rectangle) *Rectangle {
	if r.Empty() {
		return s
	}
	if s.Empty() {
		return r
	}
	if r.topCorner.Row > s.topCorner.Row {
		r.topCorner.Row = s.topCorner.Row
	}
	if r.topCorner.Column > s.topCorner.Column {
		r.topCorner.Column = s.topCorner.Column
	}
	if r.bottomCorner.Row < s.bottomCorner.Row {
		r.bottomCorner.Row = s.bottomCorner.Row
	}
	if r.bottomCorner.Column < s.bottomCorner.Column {
		r.bottomCorner.Column = s.bottomCorner.Column
	}
	return r
}

// Intersect returns the largest rectangle contained by both r and s. If the
// two rectangles do not overlap then the zero rectangle will be returned.
func (r *Rectangle) Intersect(s *Rectangle) *Rectangle {
	if r.topCorner.Column < s.topCorner.Column {
		r.topCorner.Column = s.topCorner.Column
	}
	if r.topCorner.Row < s.topCorner.Row {
		r.topCorner.Row = s.topCorner.Row
	}
	if r.bottomCorner.Column > s.bottomCorner.Column {
		r.bottomCorner.Column = s.bottomCorner.Column
	}
	if r.bottomCorner.Row > s.bottomCorner.Row {
		r.bottomCorner.Row = s.bottomCorner.Row
	}
	// Letting r0 and s0 be the values of r and s at the time that the method is called, this next line is equivalent to:
	// if max(r0.topCorner.Column, s0.topCorner.Column) >= min(r0.bottomCorner.Column, s0.bottomCorner.Column) || likewiseForRow { etc }
	if r.Empty() {
		return &Rectangle{}
	}
	return r
}

// Overlaps reports whether r and s have a non-empty intersection.
func (r *Rectangle) Overlaps(s *Rectangle) bool {
	return !r.Empty() && !s.Empty() && r.topCorner.Column < s.bottomCorner.Column && s.topCorner.Column < r.bottomCorner.Column && r.topCorner.Row < s.bottomCorner.Row && s.topCorner.Row < r.bottomCorner.Row
}

// In reports whether every point in r is in s.
func (r *Rectangle) In(s *Rectangle) bool {
	if r.Empty() {
		return true
	}
	// Note that r.bottomCorner is an exclusive bound for r, so that r.In(s)/ does not require that r.bottomCorner.In(s).
	return s.topCorner.Column <= r.topCorner.Column && r.bottomCorner.Column <= s.bottomCorner.Column && s.topCorner.Row <= r.topCorner.Row && r.bottomCorner.Row <= s.bottomCorner.Row
}

// Inset returns the rectangle r inset by n, which may be negative. If either
// of r's dimensions is less than 2*n then an empty rectangle near the center
// of r will be returned.
func (r *Rectangle) Inset(n int) *Rectangle {
	if r.Width() < 2*n {
		r.topCorner.Column = (r.topCorner.Column + r.bottomCorner.Column) / 2
		r.bottomCorner.Column = r.topCorner.Column
	} else {
		r.topCorner.Column += n
		r.bottomCorner.Column -= n
	}
	if r.Height() < 2*n {
		r.topCorner.Row = (r.topCorner.Row + r.bottomCorner.Row) / 2
		r.bottomCorner.Row = r.topCorner.Row
	} else {
		r.topCorner.Row += n
		r.bottomCorner.Row -= n
	}
	return r
}

// Resize on a rectangle updates the new size of this object.
// If it has a stroke width this will cause it to Refresh.
func (r *Rectangle) Resize(size *term.Size) {
	hasWidthPercent := r.width != nil
	hasHeightPercent := r.height != nil
	if hasWidthPercent || hasHeightPercent {
		widthDiff := 0
		//ps := r.children.Size()
		if hasWidthPercent {
			//	newWidth := ps.Width * (*r.width) / 100 // 30% from 100 pxs
			//	widthDiff = newWidth - r.bottomCorner.Column
			//	r.bottomCorner.Column = newWidth
		}
		heightDiff := 0
		if hasHeightPercent {
			//	newHeight := ps.Height * (*r.height) / 100
			//	heightDiff = newHeight - r.bottomCorner.Row
			//	r.bottomCorner.Row = newHeight
		}
		if !r.hidden {
			// TODO : acquire new pixels
			_ = widthDiff  // might be negative => release pixels
			_ = heightDiff // might be negative => release pixels
		}
	}
}

func (r *Rectangle) acquirePositions() {
	rowDist := term.Abs(r.bottomCorner.Row - r.topCorner.Row)
	colDist := term.Abs(r.bottomCorner.Column - r.topCorner.Column)
	minCol := term.Min(r.bottomCorner.Column, r.topCorner.Column)
	minRow := term.Min(r.bottomCorner.Row, r.topCorner.Row)
	for col := minCol; col < colDist; col++ {
		for row := minRow; row < rowDist; row++ {
			r.pixelAskCh <- term.Position{Row: row, Column: col}
		}
	}
}

func (r *Rectangle) releasePositions() {
	rowDist := term.Abs(r.bottomCorner.Row - r.topCorner.Row)
	colDist := term.Abs(r.bottomCorner.Column - r.topCorner.Column)
	minCol := term.Min(r.bottomCorner.Column, r.topCorner.Column)
	minRow := term.Min(r.bottomCorner.Row, r.topCorner.Row)
	for col := minCol; col < colDist; col++ {
		for row := minRow; row < rowDist; row++ {
			r.pixelReleaseCh <- term.Position{Row: row, Column: col}
		}
	}
}

// registerPixel a pixel was received from a Page
func (r *Rectangle) registerPixel(pixel px) {

}
