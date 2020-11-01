package core

import (
	"github.com/badu/term"
	"github.com/badu/term/color"
	"github.com/badu/term/style"
)

// cancellationPixel is just a term.Pixel implementation that allows to exit the term.Pixel listener goroutine
type cancellationPixel struct{}

func (s *cancellationPixel) DrawCh() chan term.PixelGetter { return nil }
func (s *cancellationPixel) Style() (color.Color, color.Color, style.Mask) {
	return color.Default, color.Default, style.None
}
func (s *cancellationPixel) HasUnicode() bool       { return false }
func (s *cancellationPixel) Unicode() *term.Unicode { return nil }
func (s *cancellationPixel) Rune() rune             { return ' ' }
func (s *cancellationPixel) Width() int             { return 0 }
func (s *cancellationPixel) Activate()              {}
func (s *cancellationPixel) Deactivate()            {}
func (s *cancellationPixel) PositionHash() int      { return term.MinusOneMinusOne }

// by convention, having coordinates -1,-1 and sent to each pixel listening goroutine, so it shuts down
func newCancellationPixel() term.PixelGetter {
	return &cancellationPixel{}
}
