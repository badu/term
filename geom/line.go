package geom

import (
	"context"

	"github.com/badu/term"
)

type LineOption func(l *Line)

func WithPositions(from, to *term.Position) LineOption {
	return func(l *Line) {
		l.topLeft = from
		l.bottomRight = to
		l.acquirePositions()
	}
}

func (l *Line) acquirePositions() {
	if l.topLeft == nil || l.bottomRight == nil {
		return
	}
	l.positions = make([]term.Position, 0)

	leftPoint, rightPoint := l.topLeft, l.bottomRight
	if leftPoint.Row > rightPoint.Row {
		leftPoint, rightPoint = rightPoint, leftPoint
	}

	rowDist := term.Abs(leftPoint.Row - rightPoint.Row)
	colDist := term.Abs(leftPoint.Column - rightPoint.Column)
	slope := float64(colDist) / float64(rowDist)
	slopeSign := 1
	if rightPoint.Column < leftPoint.Column {
		slopeSign = -1
	}

	targetCol := float64(leftPoint.Column)
	currCol := leftPoint.Column
	for row := leftPoint.Row; row < rightPoint.Row; row++ {
		l.positions = append(l.positions, term.NewPosition(currCol, row))
		targetCol += slope * float64(slopeSign)
		for currCol != int(targetCol) {
			l.positions = append(l.positions, term.NewPosition(currCol, row))
			currCol += slopeSign
		}
	}
}

// WithTopLeft
func WithTopLeft(pos *term.Position) LineOption {
	return func(l *Line) {
		l.topLeft = pos
		l.acquirePositions()
	}
}

// WithBottomRight
func WithBottomRight(pos *term.Position) LineOption {
	return func(l *Line) {
		l.bottomRight = pos
		l.acquirePositions()
	}
}

// Line describes a colored line primitive.
// Lines are special as they can have a negative width or height to indicate
// an inverse slope (i.e. slope up vs down).
type Line struct {
	positions   []term.Position //
	topLeft     *term.Position  // The current top-left position of the Line
	bottomRight *term.Position  // The current bottom right position of the Line
	hidden      bool            // Is this Line currently hidden
}

// Size returns the current size of bounding box for this line object
func (l *Line) Size() term.Size {
	return term.NewSize(term.Abs(l.bottomRight.Row-l.topLeft.Row), term.Abs(l.bottomRight.Column-l.topLeft.Column))
}

// Resize sets a new bottom-right position for the line object and it will then be refreshed.
func (l *Line) Resize(size term.Size) {
	newPos := term.NewPosition(l.topLeft.Column+size.Height, l.topLeft.Row+size.Width)
	l.bottomRight = &newPos
	l.acquirePositions()
}

// Position gets the current top-left position of this line object, relative to its parent / canvas
func (l *Line) Position() term.Position {
	return term.NewPosition(term.Min(l.topLeft.Column, l.bottomRight.Column), term.Min(l.topLeft.Row, l.bottomRight.Row))
}

// Move the line object to a new position, relative to its parent / canvas
func (l *Line) Move(pos *term.Position) {
	size := l.Size()
	l.topLeft = pos
	newPos := term.NewPosition(l.topLeft.Column+size.Height, l.topLeft.Row+size.Width)
	l.bottomRight = &newPos
}

// MinSize for a Line simply returns Size{1, 1} as there is no
// explicit content
func (l *Line) MinSize() term.Size {
	return term.NewSize(1, 1)
}

// Visible returns true if this line// Show will set this circle to be visible is visible, false otherwise
func (l *Line) Visible() bool {
	return !l.hidden
}

// Show will set this line to be visible
func (l *Line) Show() {
	l.hidden = false
}

// Hide will set this line to not be visible
func (l *Line) Hide() {
	l.hidden = true
}

// NewLine returns a new Line instance
func NewLine(ctx context.Context, opts ...LineOption) *Line {
	res := &Line{}
	for _, opt := range opts {
		opt(res)
	}
	return res
}
