package geom

import (
	"context"
	"sync"

	"github.com/badu/term"
	"github.com/badu/term/color"
	"github.com/badu/term/style"
)

// RectangleOption
type RectangleOption func(r *Rectangle)

// WithBackgroundColor
func WithBackgroundColor(c color.Color) RectangleOption {
	return func(r *Rectangle) {
		r.bg = c
	}
}

// WithForegroundColor
func WithForegroundColor(c color.Color) RectangleOption {
	return func(r *Rectangle) {
		r.fg = c
	}
}

// WithTopCorner
func WithTopCorner(pos term.Position) RectangleOption {
	return func(r *Rectangle) {
		r.topCorner = pos
	}
}

// WithBottomCorner
func WithBottomCorner(pos term.Position) RectangleOption {
	return func(r *Rectangle) {
		r.bottomCorner = pos
	}
}

// WithAlignment
func WithAlignment(a style.Alignment) RectangleOption {
	return func(r *Rectangle) {
		r.aligned = a
	}
}

func WithMinSize(size *term.Size) RectangleOption {
	return func(r *Rectangle) {
		r.min = size
	}
}

// WithVariableSizeInParent
func WithVariableSizeInParent(parent *Rectangle, widthPercent, heightPercent int) RectangleOption {
	return func(r *Rectangle) {
		r.parent = parent
		if widthPercent > 0 {
			r.width = &widthPercent // only positive values set, so we allow variable width with fixed height and vice versa
		}
		if heightPercent > 0 {
			r.height = &heightPercent
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

// Rectangle describes a colored rectangle primitive
type Rectangle struct {
	sync.RWMutex                   // guards properties
	parent       *Rectangle        // parent Rectangle, used for variable sizing
	width        *int              // width in percents, pointer indicates is optional
	height       *int              // height in percents, pointer indicates is optional
	positions    [][]term.Position // pixels references [col][row]
	min          *term.Size        // The minimum size this object can be
	topCorner    term.Position     // The current top corner of the Rectangle
	bottomCorner term.Position     // The current top corner of the Rectangle
	aligned      style.Alignment   // alignment
	fg           color.Color       // The rectangle fill color
	bg           color.Color       // The rectangle stroke color
	hidden       bool              // Is this object currently hidden
}

// Resize on a rectangle updates the new size of this object.
// If it has a stroke width this will cause it to Refresh.
func (r *Rectangle) Resize(size *term.Size) {
	r.Lock()
	defer r.Unlock()
	hasWidthPercent := r.width != nil
	hasHeightPercent := r.height != nil
	if hasWidthPercent || hasHeightPercent {
		widthDiff := 0
		ps := r.parent.Size()
		if hasWidthPercent {
			newWidth := ps.Width * (*r.width) / 100 // 30% from 100 pxs
			widthDiff = newWidth - r.bottomCorner.Column
			r.bottomCorner.Column = newWidth
		}
		heightDiff := 0
		if hasHeightPercent {
			newHeight := ps.Height * (*r.height) / 100
			heightDiff = newHeight - r.bottomCorner.Row
			r.bottomCorner.Row = newHeight
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
	r.positions = make([][]term.Position, rowDist)
	for col := minCol; col < colDist; col++ {
		r.positions[col] = make([]term.Position, colDist)
		for row := minRow; row < rowDist; row++ {
			r.positions[col][row] = term.Position{}
		}
	}
}

func (r *Rectangle) releasePositions() {
	r.positions = nil
}

// NewRectangle returns a new Rectangle instance
func NewRectangle(ctx context.Context, opts ...RectangleOption) *Rectangle {
	res := &Rectangle{}
	for _, opt := range opts {
		opt(res)
	}
	return res
}

// CurrentSize returns the current size of this rectangle object
func (r *Rectangle) Size() term.Size {
	r.RLock()
	defer r.RUnlock()
	return term.NewSize(term.Abs(r.bottomCorner.Row-r.topCorner.Row), term.Abs(r.bottomCorner.Column-r.topCorner.Column))
}

// CurrentPosition gets the current position of this rectangle object, relative to its parent / canvas
func (r *Rectangle) Position() term.Position {
	r.RLock()
	defer r.RUnlock()
	minCol := term.Min(r.bottomCorner.Column, r.topCorner.Column)
	minRow := term.Min(r.bottomCorner.Row, r.topCorner.Row)
	return term.Position{Row: minRow, Column: minCol}
}

// Move the rectangle object to a new position, relative to its parent / canvas
func (r *Rectangle) Move(pos term.Position) {
	r.Lock()
	defer r.Unlock()
	r.topCorner = pos
	r.acquirePositions()
}

// MinSize returns the specified minimum size, if set, or nil otherwise
func (r *Rectangle) MinSize() *term.Size {
	r.RLock()
	defer r.RUnlock()

	return r.min
}

// SetMinSize specifies the smallest size this object should be
func (r *Rectangle) SetMinSize(size *term.Size) {
	r.Lock()
	defer r.Unlock()
	r.min = size
}

// IsVisible returns true if this object is visible, false otherwise
func (r *Rectangle) Visible() bool {
	r.RLock()
	defer r.RUnlock()
	return !r.hidden
}

// Show will set this object to be visible
func (r *Rectangle) Show() {
	r.Lock()
	defer r.Unlock()
	r.hidden = false
	r.acquirePositions()
}

// Hide will set this object to not be visible
func (r *Rectangle) Hide() {
	r.Lock()
	defer r.Unlock()
	r.hidden = true
	r.releasePositions()
}
