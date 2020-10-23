// The colorful package provides all kinds of functions for working with colors.
package color

import (
	"database/sql/driver"
	"fmt"
	"image/color"
	"math"
	"reflect"
)

// A color is stored internally using sRGB (standard RGB) values in the range 0-1
type RGB struct {
	R float64
	G float64
	B float64
}

// Constructs a color.RGB from something implementing color.Color
func NewRGB(col color.Color) (RGB, bool) {
	r, g, b, alpha := col.RGBA()
	if alpha == 0 {
		return RGB{0, 0, 0}, false
	}

	// Since color.RGB is alpha pre-multiplied, we need to divide the
	// RGB values by alpha again in order to get back the original RGB.
	r *= 0xffff
	r /= alpha
	g *= 0xffff
	g /= alpha
	b *= 0xffff
	b /= alpha

	return RGB{float64(r) / 65535.0, float64(g) / 65535.0, float64(b) / 65535.0}, true
}

// Might come in handy sometimes to reduce boilerplate code.
func RGB255(c RGB) (r, g, b uint8) {
	return uint8(c.R*255.0 + 0.5), uint8(c.G*255.0 + 0.5), uint8(c.B*255.0 + 0.5)
}

// This is the tolerance used when comparing colors using AlmostEqualRGB.
const Delta = 1.0 / 255.0

func clamp01(v float64) float64 {
	return math.Max(0.0, math.Min(v, 1.0))
}

// Returns Clamps the color into valid range, clamping each value to [0..1]
// If the color is valid already, this is a no-op.
func NewRGBFromClamped(c RGB) RGB {
	return RGB{R: clamp01(c.R), G: clamp01(c.G), B: clamp01(c.B)}
}

func sq(v float64) float64 {
	return v * v
}

func cub(v float64) float64 {
	return v * v * v
}

// DistanceRGB computes the distance between two colors in RGB space.
// This is not a good measure! Rather do it in Lab space.
func DistanceRGB(c1, c2 RGB) float64 {
	return math.Sqrt(sq(c1.R-c2.R) + sq(c1.G-c2.G) + sq(c1.B-c2.B))
}

// Check for equality between colors within the tolerance Delta (1/255).
func AlmostEqualRGB(c1, c2 RGB) bool {
	return math.Abs(c1.R-c2.R)+math.Abs(c1.G-c2.G)+math.Abs(c1.B-c2.B) < 3.0*Delta
}

// Use NewRGBFromBlendLab, NewRGBBlendLuv or NewRGBFromBlendHCL.
func NewRGBFromBlendRGB(c1, c2 RGB, t float64) RGB {
	return RGB{c1.R + t*(c2.R-c1.R), c1.G + t*(c2.G-c1.G), c1.B + t*(c2.B-c1.B)}
}

// Utility used by Hxx color-spaces for interpolating between two angles in [0,360].
func interpBetwAng(a0, a1, t float64) float64 {
	// Based on the answer here: http://stackoverflow.com/a/14498790/2366315
	// With potential proof that it works here: http://math.stackexchange.com/a/2144499
	delta := math.Mod(math.Mod(a1-a0, 360.0)+540, 360.0) - 180.0
	return math.Mod(a0+t*delta+360.0, 360.0)
}

/// ToHSV ///
///////////
// From http://en.wikipedia.org/wiki/HSL_and_HSV
// Note that h is in [0..360] and s,v in [0..1]

// NewRGBFromHSV returns the Hue [0..360], Saturation and Value [0..1] of the color.
func ToHSV(c RGB) (h, s, v float64) {
	min := math.Min(math.Min(c.R, c.G), c.B)
	v = math.Max(math.Max(c.R, c.G), c.B)
	C := v - min

	s = 0.0
	if v != 0.0 {
		s = C / v
	}

	h = 0.0 // We use 0 instead of undefined as in wp.
	if min != v {
		if v == c.R {
			h = math.Mod((c.G-c.B)/C, 6.0)
		}
		if v == c.G {
			h = (c.B-c.R)/C + 2.0
		}
		if v == c.B {
			h = (c.R-c.G)/C + 4.0
		}
		h *= 60.0
		if h < 0.0 {
			h += 360.0
		}
	}
	return
}

// NewRGBFromHSV creates a new RGB given a Hue in [0..360], a Saturation and a Value in [0..1]
func NewRGBFromHSV(H, S, V float64) RGB {
	Hp := H / 60.0
	C := V * S
	X := C * (1.0 - math.Abs(math.Mod(Hp, 2.0)-1.0))

	m := V - C
	r, g, b := 0.0, 0.0, 0.0

	switch {
	case 0.0 <= Hp && Hp < 1.0:
		r = C
		g = X
	case 1.0 <= Hp && Hp < 2.0:
		r = X
		g = C
	case 2.0 <= Hp && Hp < 3.0:
		g = C
		b = X
	case 3.0 <= Hp && Hp < 4.0:
		g = X
		b = C
	case 4.0 <= Hp && Hp < 5.0:
		r = X
		b = C
	case 5.0 <= Hp && Hp < 6.0:
		r = C
		b = X
	}

	return RGB{m + r, m + g, m + b}
}

// Use for NewRGBFromBlendLab, NewRGBBlendLuv or NewRGBFromBlendHCL.
func NewRGBFromBlendHSV(c1, c2 RGB, t float64) RGB {
	h1, s1, v1 := ToHSV(c1)
	h2, s2, v2 := ToHSV(c2)
	// We know that h are both in [0..360]
	return NewRGBFromHSV(interpBetwAng(h1, h2, t), s1+t*(s2-s1), v1+t*(v2-v1))
}

/// HSL ///
///////////

// NewRGBFromHSL returns the Hue [0..360], Saturation [0..1], and Luminance (lightness) [0..1] of the color.
func ToHSL(c RGB) (h, s, l float64) {
	min := math.Min(math.Min(c.R, c.G), c.B)
	max := math.Max(math.Max(c.R, c.G), c.B)

	l = (max + min) / 2

	if min == max {
		s = 0
		h = 0
	} else {
		if l < 0.5 {
			s = (max - min) / (max + min)
		} else {
			s = (max - min) / (2.0 - max - min)
		}

		if max == c.R {
			h = (c.G - c.B) / (max - min)
		} else if max == c.G {
			h = 2.0 + (c.B-c.R)/(max-min)
		} else {
			h = 4.0 + (c.R-c.G)/(max-min)
		}

		h *= 60

		if h < 0 {
			h += 360
		}
	}

	return
}

// NewRGBFromHSL creates a new RGB given a Hue in [0..360], a Saturation [0..1], and a Luminance (lightness) in [0..1]
func NewRGBFromHSL(h, s, l float64) RGB {
	if s == 0 {
		return RGB{l, l, l}
	}

	var r, g, b float64
	var t1 float64
	var t2 float64
	var tr float64
	var tg float64
	var tb float64

	if l < 0.5 {
		t1 = l * (1.0 + s)
	} else {
		t1 = l + s - l*s
	}

	t2 = 2*l - t1
	h /= 360
	tr = h + 1.0/3.0
	tg = h
	tb = h - 1.0/3.0

	if tr < 0 {
		tr++
	}
	if tr > 1 {
		tr--
	}
	if tg < 0 {
		tg++
	}
	if tg > 1 {
		tg--
	}
	if tb < 0 {
		tb++
	}
	if tb > 1 {
		tb--
	}

	// Red
	if 6*tr < 1 {
		r = t2 + (t1-t2)*6*tr
	} else if 2*tr < 1 {
		r = t1
	} else if 3*tr < 2 {
		r = t2 + (t1-t2)*(2.0/3.0-tr)*6
	} else {
		r = t2
	}

	// Green
	if 6*tg < 1 {
		g = t2 + (t1-t2)*6*tg
	} else if 2*tg < 1 {
		g = t1
	} else if 3*tg < 2 {
		g = t2 + (t1-t2)*(2.0/3.0-tg)*6
	} else {
		g = t2
	}

	// Blue
	if 6*tb < 1 {
		b = t2 + (t1-t2)*6*tb
	} else if 2*tb < 1 {
		b = t1
	} else if 3*tb < 2 {
		b = t2 + (t1-t2)*(2.0/3.0-tb)*6
	} else {
		b = t2
	}

	return RGB{r, g, b}
}

// NewRGBFromHex parses a "html" hex color-string, either in the 3 "#f0c" or 6 "#ff1034" digits form.
func NewRGBFromHex(color string) (RGB, error) {
	format := "#%02x%02x%02x"
	factor := 1.0 / 255.0
	if len(color) == 4 {
		format = "#%1x%1x%1x"
		factor = 1.0 / 15.0
	}

	var r, g, b uint8
	n, err := fmt.Sscanf(color, format, &r, &g, &b)
	if err != nil {
		return RGB{}, err
	}
	if n != 3 {
		return RGB{}, fmt.Errorf("color: %v is not a hex-color", color)
	}

	return RGB{float64(r) * factor, float64(g) * factor, float64(b) * factor}, nil
}

/// Linear ///
//////////////
// http://www.sjbrown.co.uk/2004/05/14/gamma-correct-rendering/
// http://www.brucelindbloom.com/Eqn_RGB_to_XYZ.html

func linearize(v float64) float64 {
	if v <= 0.04045 {
		return v / 12.92
	}
	return math.Pow((v+0.055)/1.055, 2.4)
}

// NewRGBFromLinearRGB converts the color into the linear RGB space (see http://www.sjbrown.co.uk/2004/05/14/gamma-correct-rendering/).
func ToLinearRGB(c RGB) (r, g, b float64) {
	return linearize(c.R), linearize(c.G), linearize(c.B)
}

// A much faster and still quite precise linearization using a 6th-order Taylor approximation.
// See the accompanying Jupyter notebook for derivation of the constants.
func linearizeFast(v float64) float64 {
	v1 := v - 0.5
	v2 := v1 * v1
	v3 := v2 * v1
	v4 := v2 * v2
	//v5 := v3*v2
	return -0.248750514614486 + 0.925583310193438*v + 1.16740237321695*v2 + 0.280457026598666*v3 - 0.0757991963780179*v4 //+ 0.0437040411548932*v5
}

// NewRGBFromFastLinearRGB is much faster than and almost as accurate as NewRGBFromLinearRGB.
// they only produce good results for valid colors r,g,b in [0,1].
func ToFastLinearRGB(c RGB) (r, g, b float64) {
	r = linearizeFast(c.R)
	g = linearizeFast(c.G)
	b = linearizeFast(c.B)
	return
}

func deLinearize(v float64) float64 {
	if v <= 0.0031308 {
		return 12.92 * v
	}
	return 1.055*math.Pow(v, 1.0/2.4) - 0.055
}

// NewRGBFromLinearRGB creates an sRGB color out of the given linear RGB color (see http://www.sjbrown.co.uk/2004/05/14/gamma-correct-rendering/).
func NewRGBFromLinearRGB(r, g, b float64) RGB {
	return RGB{deLinearize(r), deLinearize(g), deLinearize(b)}
}

func deLinearizeFast(v float64) float64 {
	// This function (fractional root) is much harder to linearize, so we need to split.
	if v > 0.2 {
		v1 := v - 0.6
		v2 := v1 * v1
		v3 := v2 * v1
		v4 := v2 * v2
		v5 := v3 * v2
		return 0.442430344268235 + 0.592178981271708*v - 0.287864782562636*v2 + 0.253214392068985*v3 - 0.272557158129811*v4 + 0.325554383321718*v5
	} else if v > 0.03 {
		v1 := v - 0.115
		v2 := v1 * v1
		v3 := v2 * v1
		v4 := v2 * v2
		v5 := v3 * v2
		return 0.194915592891669 + 1.55227076330229*v - 3.93691860257828*v2 + 18.0679839248761*v3 - 101.468750302746*v4 + 632.341487393927*v5
	}
	v1 := v - 0.015
	v2 := v1 * v1
	v3 := v2 * v1
	v4 := v2 * v2
	v5 := v3 * v2
	// low-end is highly nonlinear.
	return 0.0519565234928877 + 5.09316778537561*v - 99.0338180489702*v2 + 3484.52322764895*v3 - 150028.083412663*v4 + 7168008.42971613*v5
}

// NewRGBFromFastLinearRGB is much faster than and almost as accurate as NewRGBFromLinearRGB.
// BUT it is important to NOTE that they only produce good results for valid inputs r,g,b in [0,1].
func NewRGBFromFastLinearRGB(r, g, b float64) RGB {
	return RGB{deLinearizeFast(r), deLinearizeFast(g), deLinearizeFast(b)}
}

// XYZToLinearRGB converts from CIE XYZ-space to Linear RGB space.
func XYZToLinearRGB(x, y, z float64) (r, g, b float64) {
	return 3.2404542*x - 1.5371385*y - 0.4985314*z, -0.9692660*x + 1.8760108*y + 0.0415560*z, 0.0556434*x - 0.2040259*y + 1.0572252*z
}

func LinearRGBToXYZ(r, g, b float64) (x, y, z float64) {
	x = 0.4124564*r + 0.3575761*g + 0.1804375*b
	y = 0.2126729*r + 0.7151522*g + 0.0721750*b
	z = 0.0193339*r + 0.1191920*g + 0.9503041*b
	return
}

/// XYZ ///
///////////
// http://www.sjbrown.co.uk/2004/05/14/gamma-correct-rendering/

func ToXYZ(c RGB) (x, y, z float64) {
	return LinearRGBToXYZ(ToLinearRGB(c))
}

func FromXYZ(x, y, z float64) RGB {
	return NewRGBFromLinearRGB(XYZToLinearRGB(x, y, z))
}

/// xyY ///
///////////
// http://www.brucelindbloom.com/Eqn_XYZ_to_xyY.html

// Well, the name is bad, since it's xyY but Golang needs me to start with a
// capital letter to make the method public.
func XYZToXYY(X, Y, Z float64) (x, y, Yout float64) {
	// This is the default reference white point.
	var D65 = [3]float64{0.95047, 1.00000, 1.08883}
	return XYZToXYYWhiteRef(X, Y, Z, D65)
}

func XYZToXYYWhiteRef(X, Y, Z float64, wref [3]float64) (x, y, Yout float64) {
	Yout = Y
	N := X + Y + Z
	if math.Abs(N) < 1e-14 {
		// When we have black, Bruce Lindbloom recommends to use
		// the reference white's chromacity for x and y.
		x = wref[0] / (wref[0] + wref[1] + wref[2])
		y = wref[1] / (wref[0] + wref[1] + wref[2])
	} else {
		x = X / N
		y = Y / N
	}
	return
}

func XYYToXYZ(x, y, Y float64) (X, Yout, Z float64) {
	Yout = Y

	if -1e-14 < y && y < 1e-14 {
		X = 0.0
		Z = 0.0
	} else {
		X = Y / y * x
		Z = Y / y * (1.0 - x - y)
	}

	return
}

// Converts the given color to CIE xyY space using D65 as reference white.
// (Note that the reference white is only used for black input.)
// x, y and Y are in [0..1]
func ToXYY(c RGB) (x, y, Y float64) {
	return XYZToXYY(ToXYZ(c))
}

// Converts the given color to CIE xyY space, taking into account
// a given reference white. (i.e. the monitor's white)
// (Note that the reference white is only used for black input.)
// x, y and Y are in [0..1]
func ToXYYWhiteRef(c RGB, wref [3]float64) (x, y, Y float64) {
	X, Y2, Z := ToXYZ(c)
	return XYZToXYYWhiteRef(X, Y2, Z, wref)
}

// Generates a color by using data given in CIE xyY space.
// x, y and Y are in [0..1]
func FromXYY(x, y, Y float64) RGB {
	return FromXYZ(XYYToXYZ(x, y, Y))
}

/// L*a*b* ///
//////////////
// http://en.wikipedia.org/wiki/Lab_color_space#CIELAB-CIEXYZ_conversions
// For L*a*b*, we need to L*a*b*<->XYZ->RGB and the first one is device dependent.

func FromLab(t float64) float64 {
	if t > 6.0/29.0*6.0/29.0*6.0/29.0 {
		return math.Cbrt(t)
	}
	return t/3.0*29.0/6.0*29.0/6.0 + 4.0/29.0
}

func XYZToLab(x, y, z float64) (l, a, b float64) {
	// Use D65 white as reference point by default.
	// http://www.fredmiranda.com/forum/topic/1035332
	// http://en.wikipedia.org/wiki/Standard_illuminant
	var D65 = [3]float64{0.95047, 1.00000, 1.08883}
	return XYZToLabWhiteRef(x, y, z, D65)
}

func XYZToLabWhiteRef(x, y, z float64, wref [3]float64) (l, a, b float64) {
	fy := FromLab(y / wref[1])
	l = 1.16*fy - 0.16
	a = 5.0 * (FromLab(x/wref[0]) - fy)
	b = 2.0 * (fy - FromLab(z/wref[2]))
	return
}

func labFInv(t float64) float64 {
	if t > 6.0/29.0 {
		return t * t * t
	}
	return 3.0 * 6.0 / 29.0 * 6.0 / 29.0 * (t - 4.0/29.0)
}

func LabToXYZ(l, a, b float64) (x, y, z float64) {
	// D65 white
	var D65 = [3]float64{0.95047, 1.00000, 1.08883}
	return LabToXYZWhiteRef(l, a, b, D65)
}

func LabToXYZWhiteRef(l, a, b float64, wref [3]float64) (x, y, z float64) {
	l2 := (l + 0.16) / 1.16
	x = wref[0] * labFInv(l2+a/5.0)
	y = wref[1] * labFInv(l2)
	z = wref[2] * labFInv(l2-b/2.0)
	return
}

// Converts the given color to CIE L*a*b* space using D65 as reference white.
func ToLab(c RGB) (l, a, b float64) {
	return XYZToLab(ToXYZ(c))
}

// Converts the given color to CIE L*a*b* space, taking into account
// a given reference white. (i.e. the monitor's white)
func LabWhiteRef(c RGB, wref [3]float64) (l, a, b float64) {
	x, y, z := ToXYZ(c)
	return XYZToLabWhiteRef(x, y, z, wref)
}

// Generates a color by using data given in CIE L*a*b* space using D65 as reference white.
// WARNING: many combinations of `l`, `a`, and `b` values do not have corresponding
//          valid RGB values, check the FAQ in the README if you're unsure.
func Lab(l, a, b float64) RGB {
	return FromXYZ(LabToXYZ(l, a, b))
}

// Generates a color by using data given in CIE L*a*b* space, taking
// into account a given reference white. (i.e. the monitor's white)
func GenLabWhiteRef(l, a, b float64, wref [3]float64) RGB {
	return FromXYZ(LabToXYZWhiteRef(l, a, b, wref))
}

// DistanceLab is a good measure of visual similarity between two colors!
// A result of 0 would mean identical colors, while a result of 1 or higher
// means the colors differ a lot.
func DistanceLab(c1, c2 RGB) float64 {
	l1, a1, b1 := ToLab(c1)
	l2, a2, b2 := ToLab(c2)
	return math.Sqrt(sq(l1-l2) + sq(a1-a2) + sq(b1-b2))
}

// That's actually the same, but I don't want to break code.
func DistanceCIE76(c1, c2 RGB) float64 {
	return DistanceLab(c1, c2)
}

// Uses the CIE94 formula to calculate color distance. More accurate than
// DistanceLab, but also more work.
func DistanceCIE94(c1, c2 RGB) float64 {
	l1, a1, b1 := ToLab(c1)
	l2, a2, b2 := ToLab(c2)

	// NOTE: Since all those formulas expect L,a,b values 100x larger than we
	//       have them in this library, we either need to adjust all constants
	//       in the formula, or convert the ranges of L,a,b before, and then
	//       scale the distances down again. The latter is less error-prone.
	l1, a1, b1 = l1*100.0, a1*100.0, b1*100.0
	l2, a2, b2 = l2*100.0, a2*100.0, b2*100.0

	kl := 1.0 // 2.0 for textiles
	kc := 1.0
	kh := 1.0
	k1 := 0.045 // 0.048 for textiles
	k2 := 0.015 // 0.014 for textiles.

	deltaL := l1 - l2
	sqrtC1 := math.Sqrt(sq(a1) + sq(b1))
	sqrtC2 := math.Sqrt(sq(a2) + sq(b2))
	deltaCab := sqrtC1 - sqrtC2

	// Not taking Sqrt here for stability, and it's unnecessary.
	deltaHab2 := sq(a1-a2) + sq(b1-b2) - sq(deltaCab)
	sl := 1.0
	sc := 1.0 + k1*sqrtC1
	sh := 1.0 + k2*sqrtC1

	vL2 := sq(deltaL / (kl * sl))
	vC2 := sq(deltaCab / (kc * sc))
	vH2 := deltaHab2 / sq(kh*sh)

	return math.Sqrt(vL2+vC2+vH2) * 0.01 // See above.
}

// DistanceCIEDE2000 uses the Delta E 2000 formula to calculate color
// distance. It is more expensive but more accurate than both DistanceLab
// and DistanceCIE94.
func DistanceCIEDE2000(c1, c2 RGB) float64 {
	return DistanceCIEDE2000klch(c1, c2, 1.0, 1.0, 1.0)
}

// DistanceCIEDE2000klch uses the Delta E 2000 formula with custom values
// for the weighting factors kL, kC, and kH.
func DistanceCIEDE2000klch(c1, c2 RGB, kl, kc, kh float64) float64 {
	l1, a1, b1 := ToLab(c1)
	l2, a2, b2 := ToLab(c2)

	// As with CIE94, we scale up the ranges of L,a,b beforehand and scale
	// them down again afterwards.
	l1, a1, b1 = l1*100.0, a1*100.0, b1*100.0
	l2, a2, b2 = l2*100.0, a2*100.0, b2*100.0

	cab1 := math.Sqrt(sq(a1) + sq(b1))
	cab2 := math.Sqrt(sq(a2) + sq(b2))
	cabmean := (cab1 + cab2) / 2

	g := 0.5 * (1 - math.Sqrt(math.Pow(cabmean, 7)/(math.Pow(cabmean, 7)+math.Pow(25, 7))))
	ap1 := (1 + g) * a1
	ap2 := (1 + g) * a2
	cp1 := math.Sqrt(sq(ap1) + sq(b1))
	cp2 := math.Sqrt(sq(ap2) + sq(b2))

	hp1 := 0.0
	if b1 != ap1 || ap1 != 0 {
		hp1 = math.Atan2(b1, ap1)
		if hp1 < 0 {
			hp1 += math.Pi * 2
		}
		hp1 *= 180 / math.Pi
	}
	hp2 := 0.0
	if b2 != ap2 || ap2 != 0 {
		hp2 = math.Atan2(b2, ap2)
		if hp2 < 0 {
			hp2 += math.Pi * 2
		}
		hp2 *= 180 / math.Pi
	}

	deltaLp := l2 - l1
	deltaCp := cp2 - cp1
	dhp := 0.0
	cpProduct := cp1 * cp2
	if cpProduct != 0 {
		dhp = hp2 - hp1
		if dhp > 180 {
			dhp -= 360
		} else if dhp < -180 {
			dhp += 360
		}
	}
	deltaHp := 2 * math.Sqrt(cpProduct) * math.Sin(dhp/2*math.Pi/180)

	lpmean := (l1 + l2) / 2
	cpmean := (cp1 + cp2) / 2
	hpmean := hp1 + hp2
	if cpProduct != 0 {
		hpmean /= 2
		if math.Abs(hp1-hp2) > 180 {
			if hp1+hp2 < 360 {
				hpmean += 180
			} else {
				hpmean -= 180
			}
		}
	}

	t := 1 - 0.17*math.Cos((hpmean-30)*math.Pi/180) + 0.24*math.Cos(2*hpmean*math.Pi/180) + 0.32*math.Cos((3*hpmean+6)*math.Pi/180) - 0.2*math.Cos((4*hpmean-63)*math.Pi/180)
	deltaTheta := 30 * math.Exp(-sq((hpmean-275)/25))
	rc := 2 * math.Sqrt(math.Pow(cpmean, 7)/(math.Pow(cpmean, 7)+math.Pow(25, 7)))
	sl := 1 + (0.015*sq(lpmean-50))/math.Sqrt(20+sq(lpmean-50))
	sc := 1 + 0.045*cpmean
	sh := 1 + 0.015*cpmean*t
	rt := -math.Sin(2*deltaTheta*math.Pi/180) * rc

	return math.Sqrt(sq(deltaLp/(kl*sl))+sq(deltaCp/(kc*sc))+sq(deltaHp/(kh*sh))+rt*(deltaCp/(kc*sc))*(deltaHp/(kh*sh))) * 0.01
}

// NewRGBFromBlendLab blends two colors in the L*a*b* color-space, which should result in a smoother blend.
// t == 0 results in c1, t == 1 results in c2
func NewRGBFromBlendLab(c1, c2 RGB, t float64) RGB {
	l1, a1, b1 := ToLab(c1)
	l2, a2, b2 := ToLab(c2)
	return Lab(l1+t*(l2-l1),
		a1+t*(a2-a1),
		b1+t*(b2-b1))
}

/// L*u*v* ///
//////////////
// http://en.wikipedia.org/wiki/CIELUV#XYZ_.E2.86.92_CIELUV_and_CIELUV_.E2.86.92_XYZ_conversions
// For L*u*v*, we need to L*u*v*<->XYZ<->RGB and the first one is device dependent.

func XYZToLuv(x, y, z float64) (l, a, b float64) {
	// Use D65 white as reference point by default.
	// http://www.fredmiranda.com/forum/topic/1035332
	// http://en.wikipedia.org/wiki/Standard_illuminant
	var D65 = [3]float64{0.95047, 1.00000, 1.08883}
	return XYZToLuvWhiteRef(x, y, z, D65)
}

func XYZToLuvWhiteRef(x, y, z float64, wref [3]float64) (l, u, v float64) {
	if y/wref[1] <= 6.0/29.0*6.0/29.0*6.0/29.0 {
		l = y / wref[1] * 29.0 / 3.0 * 29.0 / 3.0 * 29.0 / 3.0
	} else {
		l = 1.16*math.Cbrt(y/wref[1]) - 0.16
	}
	ubis, vbis := xyzToUv(x, y, z)
	un, vn := xyzToUv(wref[0], wref[1], wref[2])
	u = 13.0 * l * (ubis - un)
	v = 13.0 * l * (vbis - vn)
	return
}

// For this part, we do as R's graphics.hcl does, not as wikipedia does.
// Or is it the same?
func xyzToUv(x, y, z float64) (u, v float64) {
	denom := x + 15.0*y + 3.0*z
	if denom == 0.0 {
		u, v = 0.0, 0.0
	} else {
		u = 4.0 * x / denom
		v = 9.0 * y / denom
	}
	return
}

func LuvToXYZ(l, u, v float64) (x, y, z float64) {
	// D65 white
	var D65 = [3]float64{0.95047, 1.00000, 1.08883}
	return LuvToXYZWhiteRef(l, u, v, D65)
}

func LuvToXYZWhiteRef(l, u, v float64, wref [3]float64) (x, y, z float64) {
	//y = wref[1] * labFInv((l + 0.16) / 1.16)
	if l <= 0.08 {
		y = wref[1] * l * 100.0 * 3.0 / 29.0 * 3.0 / 29.0 * 3.0 / 29.0
	} else {
		y = wref[1] * cub((l+0.16)/1.16)
	}
	un, vn := xyzToUv(wref[0], wref[1], wref[2])
	if l != 0.0 {
		ubis := u/(13.0*l) + un
		vbis := v/(13.0*l) + vn
		x = y * 9.0 * ubis / (4.0 * vbis)
		z = y * (12.0 - 3.0*ubis - 20.0*vbis) / (4.0 * vbis)
	} else {
		x, y = 0.0, 0.0
	}
	return
}

// Converts the given color to CIE L*u*v* space using D65 as reference white.
// L* is in [0..1] and both u* and v* are in about [-1..1]
func ToLuv(c RGB) (l, u, v float64) {
	return XYZToLuv(ToXYZ(c))
}

// Converts the given color to CIE L*u*v* space, taking into account
// a given reference white. (i.e. the monitor's white)
// L* is in [0..1] and both u* and v* are in about [-1..1]
func ToLuvWhiteRef(c RGB, wref [3]float64) (l, u, v float64) {
	x, y, z := ToXYZ(c)
	return XYZToLuvWhiteRef(x, y, z, wref)
}

// Generates a color by using data given in CIE L*u*v* space using D65 as reference white.
// L* is in [0..1] and both u* and v* are in about [-1..1]
// WARNING: many combinations of `l`, `a`, and `b` values do not have corresponding
//          valid RGB values, check the FAQ in the README if you're unsure.
func NewRGBFromLuv(l, u, v float64) RGB {
	return FromXYZ(LuvToXYZ(l, u, v))
}

// Generates a color by using data given in CIE L*u*v* space, taking
// into account a given reference white. (i.e. the monitor's white)
// L* is in [0..1] and both u* and v* are in about [-1..1]
func NewRGBFromLuvWhiteRef(l, u, v float64, wref [3]float64) RGB {
	return FromXYZ(LuvToXYZWhiteRef(l, u, v, wref))
}

// DistanceLuv is a good measure of visual similarity between two colors!
// A result of 0 would mean identical colors, while a result of 1 or higher
// means the colors differ a lot.
func DistanceLuv(c1, c2 RGB) float64 {
	l1, u1, v1 := ToLuv(c1)
	l2, u2, v2 := ToLuv(c2)
	return math.Sqrt(sq(l1-l2) + sq(u1-u2) + sq(v1-v2))
}

// NewRGBBlendLuv blends two colors in the CIE-L*u*v* color-space, which should result in a smoother blend.
// t == 0 results in c1, t == 1 results in c2
func NewRGBBlendLuv(c1, c2 RGB, t float64) RGB {
	l1, u1, v1 := ToLuv(c1)
	l2, u2, v2 := ToLuv(c2)
	return NewRGBFromLuv(l1+t*(l2-l1),
		u1+t*(u2-u1),
		v1+t*(v2-v1))
}

/// HCL ///
///////////
// HCL is nothing else than L*a*b* in cylindrical coordinates!
// (this was wrong on English wikipedia, I fixed it, let's hope the fix stays.)
// But it is widely popular since it is a "correct ToHSV"
// http://www.hunterlab.com/appnotes/an09_96a.pdf

// Converts the given color to HCL space using D65 as reference white.
// H values are in [0..360], C and L values are in [0..1] although C can overshoot 1.0
func ToHCL(c RGB) (h, co, l float64) {
	// This is the default reference white point.
	var D65 = [3]float64{0.95047, 1.00000, 1.08883}
	return HCLWhiteRef(c, D65)
}

func LabToHCL(L, a, b float64) (h, c, l float64) {
	// floating point workaround necessary if a ~= b and both are very small (i.e. almost zero).
	if math.Abs(b-a) > 1e-4 && math.Abs(a) > 1e-4 {
		h = math.Mod(57.29577951308232087721*math.Atan2(b, a)+360.0, 360.0) // Rad2Deg
	} else {
		h = 0.0
	}
	c = math.Sqrt(sq(a) + sq(b))
	l = L
	return
}

// Converts the given color to HCL space, taking into account
// a given reference white. (i.e. the monitor's white)
// H values are in [0..360], C and L values are in [0..1]
func HCLWhiteRef(c RGB, wref [3]float64) (h, co, l float64) {
	L, a, b := LabWhiteRef(c, wref)
	return LabToHCL(L, a, b)
}

// Generates a color by using data given in HCL space using D65 as reference white.
// H values are in [0..360], C and L values are in [0..1]
// WARNING: many combinations of `l`, `a`, and `b` values do not have corresponding
//          valid RGB values, check the FAQ in the README if you're unsure.
func NewRGBFromHCL(h, c, l float64) RGB {
	// This is the default reference white point.
	var D65 = [3]float64{0.95047, 1.00000, 1.08883}
	return NewRGBFromHCLWhite(h, c, l, D65)
}

func HCLToLab(h, c, l float64) (L, a, b float64) {
	H := 0.01745329251994329576 * h // Deg2Rad
	a = c * math.Cos(H)
	b = c * math.Sin(H)
	L = l
	return
}

// Generates a color by using data given in HCL space, taking
// into account a given reference white. (i.e. the monitor's white)
// H values are in [0..360], C and L values are in [0..1]
func NewRGBFromHCLWhite(h, c, l float64, wref [3]float64) RGB {
	L, a, b := HCLToLab(h, c, l)
	return GenLabWhiteRef(L, a, b, wref)
}

// NewRGBFromBlendHCL blends two colors in the CIE-L*C*h° color-space, which should result in a smoother blend.
// t == 0 results in c1, t == 1 results in c2
func NewRGBFromBlendHCL(col1, col2 RGB, t float64) RGB {
	h1, c1, l1 := ToHCL(col1)
	h2, c2, l2 := ToHCL(col2)

	// We know that h are both in [0..360]
	return NewRGBFromHCL(interpBetwAng(h1, h2, t), c1+t*(c2-c1), l1+t*(l2-l1))
}

type errUnsupportedType struct {
	got  interface{}
	want reflect.Type
}

// Implement the Go color.Color interface.
func (r RGB) RGBA() (red, g, b, alpha uint32) {
	return uint32(r.R*65535.0 + 0.5), uint32(r.G*65535.0 + 0.5), uint32(r.B*65535.0 + 0.5), 0xFFFF
}

// Checks whether the color exists in RGB space, i.e. all values are in [0..1]
func (r *RGB) IsValid() bool {
	return 0.0 <= r.R && r.R <= 1.0 && 0.0 <= r.G && r.G <= 1.0 && 0.0 <= r.B && r.B <= 1.0
}

func (r *RGB) Scan(value interface{}) error {
	s, ok := value.(string)
	if !ok {
		return errUnsupportedType{got: reflect.TypeOf(value), want: reflect.TypeOf("")}
	}
	c, err := NewRGBFromHex(s)
	if err != nil {
		return err
	}
	*r = c
	return nil
}

func (r *RGB) Value() (driver.Value, error) {
	return (*r).String(), nil
}

// String returns the hex "html" representation of the color, as in #ff0080.
func (r RGB) String() string {
	// Add 0.5 for rounding
	return fmt.Sprintf("#%02x%02x%02x", uint8(r.R*255.0+0.5), uint8(r.G*255.0+0.5), uint8(r.B*255.0+0.5))
}

func (e errUnsupportedType) Error() string {
	return fmt.Sprintf("unsupported type: got %v, want a %s", e.got, e.want)
}
