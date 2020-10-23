package color

import (
	"math"
)

// FindColor attempts to find a given color, or the best match possible for it,
// from the palette given.  This is an expensive operation, so results should
// be cached by the caller.
func FindColor(c Color, palette []Color) Color {
	match := Default
	dist := float64(0)
	r, g, b := c.RGB()
	c1 := RGB{
		R: float64(r) / 255.0,
		G: float64(g) / 255.0,
		B: float64(b) / 255.0,
	}
	for _, d := range palette {
		r, g, b = d.RGB()
		c2 := RGB{
			R: float64(r) / 255.0,
			G: float64(g) / 255.0,
			B: float64(b) / 255.0,
		}
		// CIE94 is more accurate, but really-really expensive.
		nd := DistanceCIE76(c1, c2)
		if math.IsNaN(nd) {
			nd = math.Inf(1)
		}
		if match == Default || nd < dist {
			match = d
			dist = nd
		}
	}
	return match
}
