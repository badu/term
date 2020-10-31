package geom

import (
	"image"
	imageColor "image/color"
	"math"
	"sync"

	"github.com/badu/term"
	"github.com/badu/term/color"
)

// LinearGradient defines a Gradient travelling straight at a given angle.
// The only supported values for the angle are `0.0` (vertical) and `90.0` (horizontal), currently.
type LinearGradient struct {
	sync.RWMutex                //
	position     *term.Position // The current position of the Rectangle
	size         *term.Size     // The current size of the Rectangle
	min          *term.Size     // The minimum size this object can be
	StartColor   color.Color    // The beginning color of the gradient
	EndColor     color.Color    // The end color of the gradient
	Angle        float64        // The angle of the gradient (0/180 for vertical; 90/270 for horizontal)
	Hidden       bool           // Is this object currently hidden
}

// Generate calculates an image of the gradient with the specified width and height.
func (l *LinearGradient) Generate(iw, ih int) image.Image {
	w, h := float64(iw), float64(ih)
	var generator func(x, y float64) float64
	switch l.Angle {
	case 90: // horizontal flipped
		generator = func(x, _ float64) float64 {
			return (w - x) / w
		}
	case 270: // horizontal
		generator = func(x, _ float64) float64 {
			return x / w
		}
	case 45: // diagonal negative flipped
		generator = func(x, y float64) float64 {
			return math.Abs((w+h)-(x+h-y)) / math.Abs(w+h)
		}
	case 225: // diagonal negative
		generator = func(x, y float64) float64 {
			return math.Abs(x+h-y) / math.Abs(w+h)
		}
	case 135: // diagonal positive flipped
		generator = func(x, y float64) float64 {
			return math.Abs((w+h)-(x+y)) / math.Abs(w+h)
		}
	case 315: // diagonal positive
		generator = func(x, y float64) float64 {
			return math.Abs(x+y) / math.Abs(w+h)
		}
	case 180: // vertical flipped
		generator = func(_, y float64) float64 {
			return (h - y) / h
		}
	default: // vertical
		generator = func(_, y float64) float64 {
			return y / h
		}
	}
	return computeGradient(generator, iw, ih, l.StartColor, l.EndColor)
}

// CurrentSize returns the current size of this rectangle object
func (l *LinearGradient) Size() *term.Size {
	l.RLock()
	defer l.RUnlock()

	return l.size
}

// Resize sets a new size for the rectangle object
func (l *LinearGradient) Resize(size *term.Size) {
	l.Lock()
	defer l.Unlock()

	l.size = size
}

// CurrentPosition gets the current position of this rectangle object, relative to its children / canvas
func (l *LinearGradient) Position() *term.Position {
	l.RLock()
	defer l.RUnlock()

	return l.position
}

// Move the rectangle object to a new position, relative to its children / canvas
func (l *LinearGradient) Move(pos *term.Position) {
	l.Lock()
	defer l.Unlock()

	l.position = pos
}

// MinSize returns the specified minimum size, if set, or {1, 1} otherwise
func (l *LinearGradient) MinSize() *term.Size {
	l.RLock()
	defer l.RUnlock()

	if l.min.Width == 0 && l.min.Height == 0 {
		return term.NewSize(1, 1)
	}

	return l.min
}

// SetMinSize specifies the smallest size this object should be
func (l *LinearGradient) SetMinSize(size *term.Size) {
	l.Lock()
	defer l.Unlock()

	l.min = size
}

// IsVisible returns true if this object is visible, false otherwise
func (l *LinearGradient) Visible() bool {
	l.RLock()
	defer l.RUnlock()

	return !l.Hidden
}

// Show will set this object to be visible
func (l *LinearGradient) Show() {
	l.Lock()
	defer l.Unlock()

	l.Hidden = false
}

// Hide will set this object to not be visible
func (l *LinearGradient) Hide() {
	l.Lock()
	defer l.Unlock()

	l.Hidden = true
}

// RadialGradient defines a Gradient travelling radially from a center point outward.
type RadialGradient struct {
	sync.RWMutex                 //
	position      *term.Position // The current position of the Rectangle
	size          *term.Size     // The current size of the Rectangle
	min           *term.Size     // The minimum size this object can be
	StartColor    color.Color    // The beginning color of the gradient
	EndColor      color.Color    // The end color of the gradient
	CenterOffsetX float64        // The offset of the center for generation of the gradient. This is not a DP measure but relates to the width/height. A value of 0.5 would move the center by the half width/height.
	CenterOffsetY float64        //
	Hidden        bool           // Is this object currently hidden
}

// Generate calculates an image of the gradient with the specified width and height.
func (r *RadialGradient) Generate(iw, ih int) image.Image {
	w, h := float64(iw), float64(ih)
	// define center plus offset
	centerX := w/2 + w*r.CenterOffsetX
	centerY := h/2 + h*r.CenterOffsetY

	// handle negative offsets
	var a, b float64
	if r.CenterOffsetX < 0 {
		a = w - centerX
	} else {
		a = centerX
	}
	if r.CenterOffsetY < 0 {
		b = h - centerY
	} else {
		b = centerY
	}

	generator := func(x, y float64) float64 {
		// calculate distance from center for gradient multiplier
		dx, dy := centerX-x, centerY-y
		da := math.Sqrt(dx*dx + dy*dy*a*a/b/b)
		if da > a {
			return 1
		}
		return da / a
	}
	return computeGradient(generator, iw, ih, r.StartColor, r.EndColor)
}

func calculatePixel(d float64, startColor, endColor color.Color) imageColor.Color {
	// fetch RGBA values
	aR, aG, aB := color.ToRGB(startColor)
	bR, bG, bB := color.ToRGB(endColor)

	// Get difference
	dR := float64(bR) - float64(aR)
	dG := float64(bG) - float64(aG)
	dB := float64(bB) - float64(aB)

	// Apply gradations
	pixel := &imageColor.RGBA64{
		R: uint16(float64(aR) + d*dR),
		B: uint16(float64(aB) + d*dB),
		G: uint16(float64(aG) + d*dG),
	}

	return pixel
}

func computeGradient(generator func(x, y float64) float64, w, h int, startColor, endColor color.Color) image.Image {
	img := image.NewNRGBA(image.Rect(0, 0, w, h))

	for x := 0; x < w; x++ {
		for y := 0; y < h; y++ {
			distance := generator(float64(x)+0.5, float64(y)+0.5)
			img.Set(x, y, calculatePixel(distance, startColor, endColor))
		}
	}
	return img
}

// NewHorizontalGradient creates a new horizontally travelling linear gradient.
// The start color will be at the left of the gradient and the end color will be at the right.
func NewHorizontalGradient(start, end color.Color) *LinearGradient {
	g := &LinearGradient{StartColor: start, EndColor: end}
	g.Angle = 270
	return g
}

// NewLinearGradient creates a linear gradient at a the specified angle.
// The angle parameter is the degree angle along which the gradient is calculated.
// A NewHorizontalGradient uses 270 degrees and NewVerticalGradient is 0 degrees.
func NewLinearGradient(start, end color.Color, angle float64) *LinearGradient {
	g := &LinearGradient{StartColor: start, EndColor: end}
	g.Angle = angle
	return g
}

// NewRadialGradient creates a new radial gradient.
func NewRadialGradient(start, end color.Color) *RadialGradient {
	return &RadialGradient{StartColor: start, EndColor: end}
}

// NewVerticalGradient creates a new vertically travelling linear gradient.
// The start color will be at the top of the gradient and the end color will be at the bottom.
func NewVerticalGradient(start color.Color, end color.Color) *LinearGradient {
	return &LinearGradient{StartColor: start, EndColor: end}
}

// CurrentSize returns the current size of this rectangle object
func (r *RadialGradient) Size() *term.Size {
	r.RLock()
	defer r.RUnlock()

	return r.size
}

// Resize sets a new size for the rectangle object
func (r *RadialGradient) Resize(size *term.Size) {
	r.Lock()
	defer r.Unlock()

	r.size = size
}

// CurrentPosition gets the current position of this rectangle object, relative to its children / canvas
func (r *RadialGradient) Position() *term.Position {
	r.RLock()
	defer r.RUnlock()

	return r.position
}

// Move the rectangle object to a new position, relative to its children / canvas
func (r *RadialGradient) Move(pos *term.Position) {
	r.Lock()
	defer r.Unlock()

	r.position = pos
}

// MinSize returns the specified minimum size, if set, or {1, 1} otherwise
func (r *RadialGradient) MinSize() *term.Size {
	r.RLock()
	defer r.RUnlock()

	if r.min.Width == 0 && r.min.Height == 0 {
		return term.NewSize(1, 1)
	}

	return r.min
}

// SetMinSize specifies the smallest size this object should be
func (r *RadialGradient) SetMinSize(size *term.Size) {
	r.Lock()
	defer r.Unlock()

	r.min = size
}

// IsVisible returns true if this object is visible, false otherwise
func (r *RadialGradient) Visible() bool {
	r.RLock()
	defer r.RUnlock()

	return !r.Hidden
}

// Show will set this object to be visible
func (r *RadialGradient) Show() {
	r.Lock()
	defer r.Unlock()

	r.Hidden = false
}

// Hide will set this object to not be visible
func (r *RadialGradient) Hide() {
	r.Lock()
	defer r.Unlock()

	r.Hidden = true
}
