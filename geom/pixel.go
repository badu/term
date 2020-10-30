package geom

import (
	"errors"
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

// WithPosition is required - row is X, column is Y
func WithPosition(pos term.Position) PixelOption {
	return func(p *px) {
		pos.UpdateHash()
		p.pos = pos
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
	pos           term.Position         // required, for each pixel. default to {-1,-1} and validated in the constructor
	drawCh        chan term.PixelGetter // required, triggers core.drawPixel via setters
	fgCol         color.Color           // optional, defaults to color.Default
	bgCol         color.Color           // optional, defaults to color.Default
	attrs         style.Mask            // optional, defaults to style.None
	content       rune                  // optional, defaults to encoding.Space
	unicode       *term.Unicode         // optional, no default (don't waste memory)
	width         int                   // defaults to 1 if encoding.Space
	wasRegistered bool                  // flag, which is set by the draw channel getter, so we don't write to that channel until we've been asked for it
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

// Width - usually 1
func (p *px) Width() int {
	return p.width
}

// Position - returns the position of the pixel
func (p *px) PositionHash() int {
	return p.pos.Hash()
}

// DrawCh - usually called from core, to register listening for changes
func (p *px) DrawCh() chan term.PixelGetter {
	if !p.wasRegistered {
		p.wasRegistered = true // we assume that the engine is the one which asked about draw channel, so we're marking ourselves as ready to dispatch changes
	}
	return p.drawCh
}

// SetFgBg
func (p *px) SetFgBg(fg, bg color.Color) {
	if p.bgCol == bg && p.fgCol == fg {
		return
	}
	p.bgCol = bg
	p.fgCol = fg
	if p.wasRegistered {
		p.drawCh <- p // if not registered, it will cause blocking
	}
}

// Set - sets rune, background and foreground
func (p *px) Set(r rune, fg, bg color.Color) {
	if p.bgCol == bg && p.fgCol == fg && p.content == r {
		return
	}
	p.content = r
	p.bgCol = bg
	p.fgCol = fg
	p.width = 1
	if p.wasRegistered {
		p.drawCh <- p // if not registered, it will cause blocking
	}
}

// SetAll - set all possible properties and dispatches changes
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
	if p.wasRegistered {
		p.drawCh <- p // if not registered, it will cause blocking
	}
}

// SetAttrs - sets mask and dispatches changes
func (p *px) SetAttrs(m style.Mask) {
	if p.attrs == m {
		return
	}
	p.attrs = m
	if p.wasRegistered {
		p.drawCh <- p // if not registered, it will cause blocking
	}
}

// SetBackground - sets background color and dispatches changes
func (p *px) SetBackground(c color.Color) {
	if p.bgCol == c {
		return
	}
	p.bgCol = c
	if p.wasRegistered {
		p.drawCh <- p // if not registered, it will cause blocking
	}
}

// SetForeground - sets foreground color and dispatches changes
func (p *px) SetForeground(c color.Color) {
	if p.fgCol == c {
		return
	}
	p.fgCol = c
	if p.wasRegistered {
		p.drawCh <- p // if not registered, it will cause blocking
	}
}

// SetRune - sets rune and dispatches changes
func (p *px) SetRune(r rune) {
	if p.content == r {
		return
	}
	p.content = r
	p.width = 1
	if p.wasRegistered {
		p.drawCh <- p // if not registered, it will cause blocking
	}
}

// SetUnicode - sets unicode and dispatches changes
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
	if p.wasRegistered {
		p.drawCh <- p // if not registered, it will cause blocking
	}
}

// NewPixel constructs a term.Pixel implementation, to be used as both term.Pixel and term.PixelSetter interfaces
// Note : the reason for which we're using image.Point relies on the functionality it provides regarding image package, e.g : image.Point.In(r image.Rectangle)
// Another note, important : the background and foreground needs to be defaulted to color.Default because engine performs extra steps otherwise. Search core.drawPixel method for `positions "needed" from geom.Pixel`
func NewPixel(opts ...PixelOption) (term.Pixel, error) {
	// because composition components will "own" a set of pixels, it's not a good idea to to cache our GoTo []byte here
	defPos := term.Position{Row: -1, Column: -1}
	res := &px{
		pos:     defPos,
		drawCh:  make(chan term.PixelGetter),
		bgCol:   color.Default,  // default has color Default
		fgCol:   color.Default,  // default has color Default
		attrs:   style.None,     // default has no style
		content: encoding.Space, // it's just a space char
	}
	// apply functional options
	for _, opt := range opts {
		opt(res)
	}
	// validate point is NOT -1,-1
	// yes, it's an or condition, so we always do positive
	if res.pos.Row == -1 && res.pos.Column == -1 {
		return nil, errors.New("coordinates -1 for Row AND Column have a special meaning")
	}
	return res, nil
}

// internal constructor, used mainly by Page
func newPixel(col, row int) px {
	defPos := term.Position{Row: row, Column: col}
	res := px{
		pos:     defPos,
		drawCh:  make(chan term.PixelGetter),
		bgCol:   color.Default,  // default has color Default
		fgCol:   color.Default,  // default has color Default
		attrs:   style.None,     // default has no style
		content: encoding.Space, // it's just a space char
	}
	return res
}
