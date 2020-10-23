package style

import (
	"sync"

	"github.com/badu/term/color"
	"github.com/badu/term/info"
)

type TermStyle struct {
	sync.Mutex                             // guards other properties
	colors     map[color.Color]color.Color //
	palette    []color.Color               //
}

func NewTermStyle(ti *info.Term) *TermStyle {
	res := TermStyle{
		colors:  make(map[color.Color]color.Color),
		palette: make([]color.Color, ti.Colors),
	}
	for i := 0; i < ti.Colors; i++ {
		res.palette[i] = color.Color(i) | color.Valid
		res.colors[color.Color(i)|color.Valid] = color.Color(i) | color.Valid // identity map for our builtin colors
	}
	return &res
}

// Palette
func (s *TermStyle) Palette() []color.Color {
	s.Lock()
	defer s.Unlock()

	return s.palette
}

// Colors
func (s *TermStyle) Colors() map[color.Color]color.Color {
	s.Lock()
	defer s.Unlock()

	return s.colors
}

// FindColor
func (s *TermStyle) FindColor(c color.Color) color.Color {
	if v, ok := s.colors[c]; ok {
		return v
	}
	v := color.FindColor(c, s.palette)
	s.colors[c] = v
	return v
}
