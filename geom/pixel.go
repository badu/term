package geom

import (
	"errors"
	"image"
	"log"
	"unicode/utf8"

	"github.com/badu/term"
	"github.com/badu/term/color"
	"github.com/badu/term/encoding"
	"github.com/badu/term/style"
)

// PixelOption for functional options
type PixelOption func(p *px)

// WithBackground is optional
func WithBackground(c color.Color) PixelOption {
	return func(p *px) {
		p.bgCol = c
	}
}

// WithForeground is optional
func WithForeground(c color.Color) PixelOption {
	return func(p *px) {
		p.fgCol = c
	}
}

// WithPoint is required - row is X, column is Y
func WithPoint(point image.Point) PixelOption {
	return func(p *px) {
		p.Point = point
	}
}

// WithRune is optional
func WithRune(r rune) PixelOption {
	return func(p *px) {
		p.content = r
	}
}

// WithUnicode is optional
func WithUnicode(u term.Unicode) PixelOption {
	return func(p *px) {
		p.unicode = &u
	}
}

// WithAttrs is optional
func WithAttrs(m style.Mask) PixelOption {
	return func(p *px) {
		p.attrs = m
	}
}

type px struct {
	image.Point                       // required, for each pixel. default to {-1,-1} and validated in the constructor, row is X, column is Y
	drawCh      chan term.PixelGetter // required, triggers core.drawPixel via setters
	fgCol       color.Color           // optional, defaults to color.Default
	bgCol       color.Color           // optional, defaults to color.Default
	attrs       style.Mask            // optional, defaults to style.None
	content     rune                  // optional, defaults to encoding.Space
	unicode     *term.Unicode         // optional, no default (don't waste memory)
	width       int                   // defaults to 1 if encoding.Space
}

// BgCol
func (p *px) BgCol() color.Color {
	return p.bgCol
}

// FgCol
func (p *px) FgCol() color.Color {
	return p.fgCol
}

// HasUnicode
func (p *px) HasUnicode() bool {
	return p.unicode != nil
}

// Unicode
func (p *px) Unicode() *term.Unicode {
	return p.unicode
}

// Run
func (p *px) Rune() rune {
	return p.content
}

// Attrs
func (p *px) Attrs() style.Mask {
	return p.attrs
}

// Size
func (p *px) Width() int {
	return p.width
}

// Position
func (p *px) Position() *image.Point {
	return &p.Point
}

// DrawCh
func (p px) DrawCh() chan term.PixelGetter {
	return p.drawCh
}

// SetFgBg
func (p *px) SetFgBg(fg, bg color.Color) {
	if p.bgCol == bg && p.fgCol == fg {
		return
	}
	p.bgCol = bg
	p.fgCol = fg
	p.drawCh <- p
}

// Set
func (p *px) Set(r rune, fg, bg color.Color) {
	if p.bgCol == bg && p.fgCol == fg && p.content == r {
		return
	}
	p.content = r
	p.bgCol = bg
	p.fgCol = fg
	p.width = 1
	p.drawCh <- p
}

// SetAll
func (p *px) SetAll(bg, fg color.Color, m style.Mask, r rune, u term.Unicode) {
	var currUnicode term.Unicode
	if p.unicode != nil {
		currUnicode = *p.unicode
	}
	// validate and compare
	equal := len(currUnicode) == len(u)
	for idx, r := range u {
		if !utf8.ValidRune(r) {
			if Debug {
				log.Printf("error : invalid rune provided : %#v", r)
			}
			return
		}
		// only if equal len, we're comparing rune by rune
		if equal && len(currUnicode) > idx && currUnicode[idx] != r {
			// never enter here again, just continue validating runes
			equal = false
		}
	}
	if p.bgCol == bg && p.fgCol == fg && p.content == r && equal {
		return
	}
	p.bgCol = bg
	p.fgCol = fg
	p.content = r
	p.unicode = &u
	p.attrs = m
	p.drawCh <- p
}

// SetAttrs
func (p *px) SetAttrs(m style.Mask) {
	if p.attrs == m {
		return
	}
	p.attrs = m
	p.drawCh <- p
}

// SetBackground
func (p *px) SetBackground(c color.Color) {
	if p.bgCol == c {
		return
	}
	p.bgCol = c
	p.drawCh <- p
}

// SetForeground
func (p *px) SetForeground(c color.Color) {
	if p.fgCol == c {
		return
	}
	p.fgCol = c
	p.drawCh <- p
}

// SetRune
func (p *px) SetRune(r rune) {
	if p.content == r {
		return
	}
	p.content = r
	p.width = 1
	p.drawCh <- p
}

// SetUnicode
func (p *px) SetUnicode(u term.Unicode) {
	var currUnicode term.Unicode
	if p.unicode != nil {
		currUnicode = *p.unicode
	}
	// validate and compare
	equal := len(currUnicode) == len(u)
	newSize := 0
	for idx, r := range u {
		if !utf8.ValidRune(r) {
			if Debug {
				log.Printf("error : invalid rune provided : %#v", r)
			}
			return
		}
		// only if equal len, we're comparing rune by rune
		if equal && len(currUnicode) > idx && currUnicode[idx] != r {
			// never enter here again, just continue validating runes
			equal = false
		}
		newSize += utf8.RuneLen(r)
	}
	if equal {
		return
	}
	p.width = newSize + 1 // +1 because of the rune
	p.unicode = &u
	p.drawCh <- p
}

// NewPixel constructs a term.Pixel implementation, to be used as both term.Pixel and term.PixelSetter interfaces
// Note : the reason for which we're using image.Point relies on the functionality it provides regarding image package, e.g : image.Point.In(r image.Rectangle)
func NewPixel(opts ...PixelOption) (term.Pixel, error) {
	// because composition components will "own" a set of pixels, it's not a good idea to to cache our GoTo []byte here
	res := &px{
		Point:   image.Point{X: -1, Y: -1},
		drawCh:  make(chan term.PixelGetter),
		bgCol:   color.Reset,    // default has no color
		fgCol:   color.Reset,    // default has no color
		attrs:   style.None,     // default has no style
		content: encoding.Space, // it's just a space char
	}
	// apply functional options
	for _, opt := range opts {
		opt(res)
	}
	// validate point is NOT -1,-1
	// yes, it's an or condition, so we always do positive
	if res.Point.X == -1 || res.Point.Y == -1 {
		return nil, errors.New("coordinates -1 for X or Y have a special meaning")
	}
	return res, nil
}
