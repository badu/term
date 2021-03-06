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

// WithCore
func WithCore(engine term.Engine) RectangleOption {
	return func(r *Rectangle) {
		r.engine = engine
	}
}

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

// WithWidthAndHeight
func WithWidthAndHeight(widthPercent, heightPercent int) RectangleOption {
	return func(r *Rectangle) {
		r.width = &widthPercent
		r.height = &heightPercent
	}
}

// WithWidth
func WithWidth(percent int) RectangleOption {
	return func(r *Rectangle) {
		r.width = &percent
	}
}

// WithHeight
func WithHeight(percent int) RectangleOption {
	return func(r *Rectangle) {
		r.height = &percent
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

// Screen
type Screen = Rectangle

// Window
type Window = Rectangle

// Rectangle describes a colored rectangle primitive
type Rectangle struct {
	root
	id             int                   //
	children       []*Rectangle          // children Rectangles, used for variable sizing // TODO : mount death listener for children
	aligned        style.Alignment       // alignment. Default is style.Begin (topCorner)
	st             style.Style           // rectangle style, for inheritance
	died           chan struct{}         // Channel for killing (context.Done)
	pixelAskCh     chan term.Position    // Channel for asking pixels
	pixelReleaseCh chan term.Position    // Channel for releasing pixels
	pixelReceiveCh chan px               // Channel for receiving pixels
	resizeCh       chan term.ResizeEvent // channel for listening resize events, so we can clip our coordinates
	width          *int                  // width in percents, pointer indicates is optional
	height         *int                  // height in percents, pointer indicates is optional
	min            *term.Size            // The minimum size this object can be. Note that this can exist in the same time with width/height in percents // TODO : check new size (%) is not smaller than min allowed size
	engine         term.Engine           // listen resize events
	hidden         bool                  // Is this object currently hidden
}

// TODO : thinking maybe this should be a private constructor. Ask Page to give you a Rectangle and it will give it already populated and ready to use. For now (testing purposes), I'll leave it as it is.
// NewRectangle returns a new Rectangle instance
func NewRectangle(ctx context.Context, opts ...RectangleOption) (*Rectangle, error) {
	defStyle := style.NewStyle(style.WithBg(color.Default), style.WithFg(color.Default), style.WithAttrs(style.None))
	r := &Rectangle{
		id:             getNextRectId(),          // id for equality comparison
		aligned:        style.Begin,              // aligned top-left-corner
		st:             *defStyle,                // default rectangle style (TODO : if default style is being used, no inheritance of style happens)
		pixelReleaseCh: make(chan term.Position), //
		pixelReceiveCh: make(chan px),            //
		died:           make(chan struct{}),      // death announcement channel
		root: root{
			orientation:  style.Vertical,           // default orientation
			topCorner:    term.NewPosition(-1, -1), // by default, rectangle is nowhere
			bottomCorner: term.NewPosition(-1, -1), // by default, rectangle is nowhere
		},
	}

	for _, opt := range opts {
		opt(r)
	}

	if r.pixelAskCh == nil {
		return nil, errors.New("acquisition channel is mandatory")
	}
	if r.min == nil && r.Invalid() {
		return nil, errors.New("declared rectangle is outside the screen")
	}

	// all ok, listening for context.Done to exit
	go func() {
		for {
			select {
			case <-ctx.Done():
				log.Println("context is done : releasing pixels and die")
				r.releasePositions()
				close(r.died) // notifying our death to a dispatcher (which listens in register)
				return
			case pix := <-r.pixelReceiveCh:
				r.registerPixel(pix)
			}
		}
	}()
	return r, nil
}

// DyingChan implementation of term.Death interface, listened in core for waiting graceful shutdown
func (r *Rectangle) DyingChan() chan struct{} {
	return r.died
}

// ResizeListen
func (r *Rectangle) ResizeListen() chan term.ResizeEvent {
	return r.resizeCh
}

// Invalid returns true if coordinates are NOT set
func (r *Rectangle) Invalid() bool {
	return r.topCorner.Row < 0 || r.topCorner.Column < 0 || r.bottomCorner.Row < 0 || r.bottomCorner.Column < 0
}

// Move the rectangle object to a new position, relative to its children / canvas
func (r *Rectangle) Move(pos *term.Position) {
	r.topCorner = pos
	r.acquirePositions()
}

// SetMinSize specifies the smallest size this object should be
func (r *Rectangle) SetMinSize(size *term.Size) {
	r.min = size
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

// calculatedWidth
func (r *Rectangle) calculatedWidth(parentWidth int) int {
	w := *r.width
	return parentWidth * w / 100
}

// calculatedHeight
func (r *Rectangle) calculatedHeight(parentHeight int) int {
	h := *r.height
	return parentHeight * h / 100
}

// SetChildren - general convention that all siblings are registered together so we can perform calculations of positions and invalidate recursively the children rectangles
func (r *Rectangle) SetChildren(children ...*Rectangle) {
	lastRow := 0
	lastColumn := 0
	for _, child := range children {
		if r.HasRows() {
			// variable sized rectangle (both width and height)
			if child.width != nil && child.height != nil {
				// with percent of the width and height
				child.topCorner.Row = lastRow
				child.topCorner.Column = lastColumn
				child.bottomCorner.Row = child.calculatedHeight(r.bottomCorner.Row)
				child.bottomCorner.Column = child.calculatedWidth(r.bottomCorner.Column)
				lastColumn = child.bottomCorner.Column
				lastRow = child.bottomCorner.Row
				log.Printf("parent at %04d,%04d", lastColumn, lastRow)
				log.Printf("child [%04d,%04d->%04d,%04d]", child.topCorner.Column, child.topCorner.Row, child.bottomCorner.Column, child.bottomCorner.Row)
				child.invalidateSize()
				continue
			}

			// fixed size rectangle
			if child.width == nil && child.height == nil {
				child.topCorner.Row = lastRow
				child.topCorner.Column = lastColumn
				child.bottomCorner.Row = r.bottomCorner.Row
				child.bottomCorner.Column = r.bottomCorner.Column
				child.invalidateSize()
				continue
			}

			// variable width rectangle
			if child.width != nil {
				child.bottomCorner.Column = child.calculatedWidth(r.bottomCorner.Column)
				lastColumn = child.bottomCorner.Column
			} else {
				// with 100% width
				child.bottomCorner.Column = r.bottomCorner.Column
				lastColumn = child.bottomCorner.Column
			}

			// variable height rectangle
			if child.height != nil {
				// with percent of the height
				child.bottomCorner.Row = child.calculatedHeight(r.bottomCorner.Row)
				lastRow = child.bottomCorner.Row
			} else {
				// with 100% height
				child.bottomCorner.Row = r.bottomCorner.Row
				lastRow = child.bottomCorner.Row
			}

		} else if r.HasColumns() {
			// columns orientation
			if child.width == nil && child.height == nil {
				// fixed size rectangle
			} else {
				// variable sized rectangle
				if child.width != nil {
					// with percent of the width
				} else {
					// with 100% width
				}
				if child.height != nil {
					// with percent of the height
				} else {
					// with 100% height
				}
			}
		} else {
			log.Println("bad call to Rectangle.SetChildren : orientation is not set (should never happen)")
		}
		r.children = append(r.children, child) // TODO : mount death listener and remove child when shutdown
		child.invalidateSize()
	}
}

func (r *Rectangle) invalidateSize() {
	rectSize := r.Size()
	log.Printf("%03d rows %03d columns", rectSize.Rows, rectSize.Columns)
	switch r.orientation {
	case style.Vertical:
		log.Println("Vertical (ROWS)")
		r.rows = make(PixelsMatrix, rectSize.Rows)
		for row := 0; row < rectSize.Rows; row++ {
			r.rows[row] = make(Pixels, rectSize.Columns)
		}
	case style.Horizontal:
		log.Println("Horizontal (COLUMNS)")
		r.cols = make(PixelsMatrix, rectSize.Columns)
		for col := 0; col < rectSize.Columns; col++ {
			r.cols[col] = make(Pixels, rectSize.Rows)
		}
	default:
		// return nil, errors.New("orientation must be horizontal (rows) or vertical (columns)")
	}
}

func (r *Rectangle) childRemoved() {

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
