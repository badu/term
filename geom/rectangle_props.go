package geom

import (
	"github.com/badu/term"
	"github.com/badu/term/color"
	"github.com/badu/term/style"
)

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

// IsVisible returns true if this object is visible, false otherwise
func (r *Rectangle) Visible() bool {
	return !r.hidden
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

// MinSize returns the specified minimum size, if set, or nil otherwise
func (r *Rectangle) MinSize() *term.Size {
	return r.min
}

// AtLeft indicates alignment on the left edge.
func (r *Rectangle) AtLeft() bool {
	return r.aligned == style.Left
}

// AtHorizontalCenter indicates horizontally centered.
func (r *Rectangle) AtHorizontalCenter() bool {
	return r.aligned == style.HCenter
}

// AtRight indicates alignment on the right edge.
func (r *Rectangle) AtRight() bool {
	return r.aligned == style.Right
}

// AtTop indicates alignment on the top edge.
func (r *Rectangle) AtTop() bool {
	return r.aligned == style.Top
}

// AtVerticalCenter indicates vertically centered.
func (r *Rectangle) AtVerticalCenter() bool {
	return r.aligned == style.VCenter
}

// AtBottom indicates alignment on the bottom edge.
func (r *Rectangle) AtBottom() bool {
	return r.aligned == style.Bottom
}

// AtBegin indicates alignment at the top left corner.
func (r *Rectangle) AtBegin() bool {
	return r.aligned == style.Begin
}

// AtEnd indicates alignment at the bottom right corner.
func (r *Rectangle) AtEnd() bool {
	return r.aligned == style.End
}

// AtMiddle indicates full centering.
func (r *Rectangle) AtMiddle() bool {
	return r.aligned == style.Middle
}
