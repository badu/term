package style

import (
	"github.com/badu/term/color"
)

// Style represents a complete text style, including both foreground and background color.  We encode it in a 64-bit int for efficiency.
// The coding is (MSB): <7b flags><1b><24b fgcolor><7b attr><1b><24b bgcolor>.
// The <1b> is set true to indicate that the color is an RGB color, rather than a named index.
//
// This gives 24bit color options, if it ever becomes truly necessary.
// However, applications must not rely on this encoding.
//
// Note that not all terminals can display all colors or attributes, and many might have specific incompatibilities between specific attributes and color combinations.
//
// The intention is to extend styles to support palette-ing, in which case some flag bit(s) would be set, and the foreground and background colors would be replaced with a palette number and palette index.
//
// To use Style, just declare a variable of its type.
type Style struct {
	Fg    color.Color
	Bg    color.Color
	Attrs Mask
}

type Option func(s *Style)

// WithAttrs
func WithAttrs(mask Mask) Option {
	return func(s *Style) {
		s.Attrs = mask
	}
}

// WithFg
func WithFg(c color.Color) Option {
	return func(s *Style) {
		s.Fg = c
	}
}

// WithBg
func WithBg(c color.Color) Option {
	return func(s *Style) {
		s.Bg = c
	}
}

// WithBold
func WithBold(on bool) Option {
	return func(s *Style) {
		s.mergeAttrs(Bold, on)
	}
}

// WithBlink
func WithBlink(on bool) Option {
	return func(s *Style) {
		s.mergeAttrs(Blink, on)
	}
}

// WithDim
func WithDim(on bool) Option {
	return func(s *Style) {
		s.mergeAttrs(Dim, on)
	}
}

// WithItalic
func WithItalic(on bool) Option {
	return func(s *Style) {
		s.mergeAttrs(Italic, on)
	}
}

// WithReverse
func WithReverse(on bool) Option {
	return func(s *Style) {
		s.mergeAttrs(Reverse, on)
	}
}

// WithUnderline
func WithUnderline(on bool) Option {
	return func(s *Style) {
		s.mergeAttrs(Underline, on)
	}
}

// WithStrikeThrough
func WithStrikeThrough(on bool) Option {
	return func(s *Style) {
		s.mergeAttrs(StrikeThrough, on)
	}
}

// FromStyle
func FromStyle(clone *Style) Option {
	return func(s *Style) {
		s.Bg = clone.Bg
		s.Fg = clone.Bg
		s.Attrs = clone.Attrs
	}
}

func (s Style) IsValid() bool {
	return s.Attrs != Invalid
}

func NewStyle(opts ...Option) *Style {
	res := &Style{}
	for _, opt := range opts {
		opt(res)
	}
	return res
}

// Expand breaks a style up, returning the foreground, background, and other attributes.
func (s Style) Expand() (color.Color, color.Color, Mask) {
	return s.Fg, s.Bg, s.Attrs
}

func (s *Style) mergeAttrs(attrs Mask, on bool) {
	if on {
		s.Attrs = s.Attrs | attrs
		return
	}
	attrs = s.Attrs &^ attrs
}

func (s *Style) merge(other *Style, except color.Color) {
	s.mergeAttrs(other.Attrs, true)
	if other.Fg != except {
		s.Fg = other.Fg
	}
	if other.Bg != except {
		s.Bg = other.Bg
	}
}

// TODO : test
func (s *Style) Merge(other *Style) {
	s.merge(other, color.Default)
}

// Normal returns the style with all attributes disabled.
func (s *Style) Normal() Style {
	return Style{
		Fg: s.Fg,
		Bg: s.Bg,
	}
}
