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
	attrs Mask
}

// DefaultStyle represents a default style, based upon the context.
// It is the zero value.
var DefaultStyle Style

// NewInvalidStyle is just an arbitrary invalid style used internally.
func NewInvalidStyle() Style {
	return Style{attrs: Invalid}
}

func (s Style) Attrs(mask Mask) Style {
	return Style{
		Fg:    s.Fg,
		Bg:    s.Bg,
		attrs: mask,
	}
}

// Foreground returns a new style based on s, with the foreground color set as requested.
// Default can be used to select the global default.
func (s Style) Foreground(c color.Color) Style {
	return Style{
		Fg:    c,
		Bg:    s.Bg,
		attrs: s.attrs,
	}
}

// Background returns a new style based on s, with the background color set as requested.
// Default can be used to select the global default.
func (s Style) Background(c color.Color) Style {
	return Style{
		Fg:    s.Fg,
		Bg:    c,
		attrs: s.attrs,
	}
}

// Expand breaks a style up, returning the foreground, background, and other attributes.
func (s Style) Expand() (color.Color, color.Color, Mask) {
	return s.Fg, s.Bg, s.attrs
}

func (s Style) setAttrs(attrs Mask, on bool) Style {
	if on {
		return Style{
			Fg:    s.Fg,
			Bg:    s.Bg,
			attrs: s.attrs | attrs,
		}
	}
	return Style{
		Fg:    s.Fg,
		Bg:    s.Bg,
		attrs: s.attrs &^ attrs,
	}
}

// Normal returns the style with all attributes disabled.
func (s Style) Normal() Style {
	return Style{
		Fg: s.Fg,
		Bg: s.Bg,
	}
}

// Bold returns a new style based on s, with the bold attribute set as requested.
func (s Style) Bold(on bool) Style {
	return s.setAttrs(Bold, on)
}

// Blink returns a new style based on s, with the blink attribute set as requested.
func (s Style) Blink(on bool) Style {
	return s.setAttrs(Blink, on)
}

// Dim returns a new style based on s, with the dim attribute set as requested.
func (s Style) Dim(on bool) Style {
	return s.setAttrs(Dim, on)
}

// Italic returns a new style based on s, with the italic attribute set as requested.
func (s Style) Italic(on bool) Style {
	return s.setAttrs(Italic, on)
}

// Reverse returns a new style based on s, with the reverse attribute set as requested.
// (Reverse usually changes the foreground and background colors.)
func (s Style) Reverse(on bool) Style {
	return s.setAttrs(Reverse, on)
}

// Underline returns a new style based on s, with the underline attribute set as requested.
func (s Style) Underline(on bool) Style {
	return s.setAttrs(Underline, on)
}

// StrikeThrough sets strikethrough mode.
func (s Style) StrikeThrough(on bool) Style {
	return s.setAttrs(StrikeThrough, on)
}

func (s Style) IsValid() bool {
	return s.attrs != Invalid
}
