package color

import (
	"image/color"
	"math"
	"math/rand"
	"reflect"
	"strings"
	"testing"
)

var benchResult float64 // Dummy for benchmarks to avoid optimization

// Checks whether the relative error is below eps
func almosteqEps(v1, v2, eps float64) bool {
	if math.Abs(v1) > delta {
		return math.Abs((v1-v2)/v1) < eps
	}
	return true
}

// Checks whether the relative error is below the 8bit RGB delta, which should be good enough.
const delta = 1.0 / 256.0

func almosteq(v1, v2 float64) bool {
	return almosteqEps(v1, v2, delta)
}

// Note: the XYZ, L*a*b*, etc. are using D65 white and D50 white if postfixed by "50".
// See http://www.brucelindbloom.com/index.html?ColorCalcHelp.html
// For d50 white, no "adaptation" and the sRGB model are used in colorful
// HCL values form http://www.easyrgb.com/index.php?X=CALC and missing ones hand-computed from lab ones
var vals = []struct {
	c      RGB
	hsl    [3]float64
	hsv    [3]float64
	hex    string
	xyz    [3]float64
	xyy    [3]float64
	lab    [3]float64
	lab50  [3]float64
	luv    [3]float64
	luv50  [3]float64
	hcl    [3]float64
	hcl50  [3]float64
	rgba   [4]uint32
	rgb255 [3]uint8
}{
	{RGB{1.0, 1.0, 1.0}, [3]float64{0.0, 0.0, 1.00}, [3]float64{0.0, 0.0, 1.0}, "#ffffff", [3]float64{0.950470, 1.000000, 1.088830}, [3]float64{0.312727, 0.329023, 1.000000}, [3]float64{1.000000, 0.000000, 0.000000}, [3]float64{1.000000, -0.023881, -0.193622}, [3]float64{1.00000, 0.00000, 0.00000}, [3]float64{1.00000, -0.14716, -0.25658}, [3]float64{0.0000, 0.000000, 1.000000}, [3]float64{262.9688, 0.195089, 1.000000}, [4]uint32{65535, 65535, 65535, 65535}, [3]uint8{255, 255, 255}},
	{RGB{0.5, 1.0, 1.0}, [3]float64{180.0, 1.0, 0.75}, [3]float64{180.0, 0.5, 1.0}, "#80ffff", [3]float64{0.626296, 0.832848, 1.073634}, [3]float64{0.247276, 0.328828, 0.832848}, [3]float64{0.931390, -0.353319, -0.108946}, [3]float64{0.931390, -0.374100, -0.301663}, [3]float64{0.93139, -0.53909, -0.11630}, [3]float64{0.93139, -0.67615, -0.35528}, [3]float64{197.1371, 0.369735, 0.931390}, [3]float64{218.8817, 0.480574, 0.931390}, [4]uint32{32768, 65535, 65535, 65535}, [3]uint8{128, 255, 255}},
	{RGB{1.0, 0.5, 1.0}, [3]float64{300.0, 1.0, 0.75}, [3]float64{300.0, 0.5, 1.0}, "#ff80ff", [3]float64{0.669430, 0.437920, 0.995150}, [3]float64{0.318397, 0.208285, 0.437920}, [3]float64{0.720892, 0.651673, -0.422133}, [3]float64{0.720892, 0.630425, -0.610035}, [3]float64{0.72089, 0.60047, -0.77626}, [3]float64{0.72089, 0.49438, -0.96123}, [3]float64{327.0661, 0.776450, 0.720892}, [3]float64{315.9417, 0.877257, 0.720892}, [4]uint32{65535, 32768, 65535, 65535}, [3]uint8{255, 128, 255}},
	{RGB{1.0, 1.0, 0.5}, [3]float64{60.0, 1.0, 0.75}, [3]float64{60.0, 0.5, 1.0}, "#ffff80", [3]float64{0.808654, 0.943273, 0.341930}, [3]float64{0.386203, 0.450496, 0.943273}, [3]float64{0.977637, -0.165795, 0.602017}, [3]float64{0.977637, -0.188424, 0.470410}, [3]float64{0.97764, 0.05759, 0.79816}, [3]float64{0.97764, -0.08628, 0.54731}, [3]float64{105.3975, 0.624430, 0.977637}, [3]float64{111.8287, 0.506743, 0.977637}, [4]uint32{65535, 65535, 32768, 65535}, [3]uint8{255, 255, 128}},
	{RGB{0.5, 0.5, 1.0}, [3]float64{240.0, 1.0, 0.75}, [3]float64{240.0, 0.5, 1.0}, "#8080ff", [3]float64{0.345256, 0.270768, 0.979954}, [3]float64{0.216329, 0.169656, 0.270768}, [3]float64{0.590453, 0.332846, -0.637099}, [3]float64{0.590453, 0.315806, -0.824040}, [3]float64{0.59045, -0.07568, -1.04877}, [3]float64{0.59045, -0.16257, -1.20027}, [3]float64{297.5843, 0.718805, 0.590453}, [3]float64{290.9689, 0.882482, 0.590453}, [4]uint32{32768, 32768, 65535, 65535}, [3]uint8{128, 128, 255}},
	{RGB{1.0, 0.5, 0.5}, [3]float64{0.0, 1.0, 0.75}, [3]float64{0.0, 0.5, 1.0}, "#ff8080", [3]float64{0.527613, 0.381193, 0.248250}, [3]float64{0.455996, 0.329451, 0.381193}, [3]float64{0.681085, 0.483884, 0.228328}, [3]float64{0.681085, 0.464258, 0.110043}, [3]float64{0.68108, 0.92148, 0.19879}, [3]float64{0.68108, 0.82125, 0.02404}, [3]float64{25.2610, 0.535049, 0.681085}, [3]float64{13.3347, 0.477121, 0.681085}, [4]uint32{65535, 32768, 32768, 65535}, [3]uint8{255, 128, 128}},
	{RGB{0.5, 1.0, 0.5}, [3]float64{120.0, 1.0, 0.75}, [3]float64{120.0, 0.5, 1.0}, "#80ff80", [3]float64{0.484480, 0.776121, 0.326734}, [3]float64{0.305216, 0.488946, 0.776121}, [3]float64{0.906026, -0.600870, 0.498993}, [3]float64{0.906026, -0.619946, 0.369365}, [3]float64{0.90603, -0.58869, 0.76102}, [3]float64{0.90603, -0.72202, 0.52855}, [3]float64{140.2920, 0.781050, 0.906026}, [3]float64{149.2134, 0.721640, 0.906026}, [4]uint32{32768, 65535, 32768, 65535}, [3]uint8{128, 255, 128}},
	{RGB{0.5, 0.5, 0.5}, [3]float64{0.0, 0.0, 0.50}, [3]float64{0.0, 0.0, 0.5}, "#808080", [3]float64{0.203440, 0.214041, 0.233054}, [3]float64{0.312727, 0.329023, 0.214041}, [3]float64{0.533890, 0.000000, 0.000000}, [3]float64{0.533890, -0.014285, -0.115821}, [3]float64{0.53389, 0.00000, 0.00000}, [3]float64{0.53389, -0.07857, -0.13699}, [3]float64{0.0000, 0.000000, 0.533890}, [3]float64{262.9688, 0.116699, 0.533890}, [4]uint32{32768, 32768, 32768, 65535}, [3]uint8{128, 128, 128}},
	{RGB{0.0, 1.0, 1.0}, [3]float64{180.0, 1.0, 0.50}, [3]float64{180.0, 1.0, 1.0}, "#00ffff", [3]float64{0.538014, 0.787327, 1.069496}, [3]float64{0.224656, 0.328760, 0.787327}, [3]float64{0.911132, -0.480875, -0.141312}, [3]float64{0.911132, -0.500630, -0.333781}, [3]float64{0.91113, -0.70477, -0.15204}, [3]float64{0.91113, -0.83886, -0.38582}, [3]float64{196.3762, 0.501209, 0.911132}, [3]float64{213.6923, 0.601698, 0.911132}, [4]uint32{0, 65535, 65535, 65535}, [3]uint8{0, 255, 255}},
	{RGB{1.0, 0.0, 1.0}, [3]float64{300.0, 1.0, 0.50}, [3]float64{300.0, 1.0, 1.0}, "#ff00ff", [3]float64{0.592894, 0.284848, 0.969638}, [3]float64{0.320938, 0.154190, 0.284848}, [3]float64{0.603242, 0.982343, -0.608249}, [3]float64{0.603242, 0.961939, -0.794531}, [3]float64{0.60324, 0.84071, -1.08683}, [3]float64{0.60324, 0.75194, -1.24161}, [3]float64{328.2350, 1.155407, 0.603242}, [3]float64{320.4444, 1.247640, 0.603242}, [4]uint32{65535, 0, 65535, 65535}, [3]uint8{255, 0, 255}},
	{RGB{1.0, 1.0, 0.0}, [3]float64{60.0, 1.0, 0.50}, [3]float64{60.0, 1.0, 1.0}, "#ffff00", [3]float64{0.770033, 0.927825, 0.138526}, [3]float64{0.419320, 0.505246, 0.927825}, [3]float64{0.971393, -0.215537, 0.944780}, [3]float64{0.971393, -0.237800, 0.847398}, [3]float64{0.97139, 0.07706, 1.06787}, [3]float64{0.97139, -0.06590, 0.81862}, [3]float64{102.8512, 0.969054, 0.971393}, [3]float64{105.6754, 0.880131, 0.971393}, [4]uint32{65535, 65535, 0, 65535}, [3]uint8{255, 255, 0}},
	{RGB{0.0, 0.0, 1.0}, [3]float64{240.0, 1.0, 0.50}, [3]float64{240.0, 1.0, 1.0}, "#0000ff", [3]float64{0.180437, 0.072175, 0.950304}, [3]float64{0.150000, 0.060000, 0.072175}, [3]float64{0.322970, 0.791875, -1.078602}, [3]float64{0.322970, 0.778150, -1.263638}, [3]float64{0.32297, -0.09405, -1.30342}, [3]float64{0.32297, -0.14158, -1.38629}, [3]float64{306.2849, 1.338076, 0.322970}, [3]float64{301.6248, 1.484014, 0.322970}, [4]uint32{0, 0, 65535, 65535}, [3]uint8{0, 0, 255}},
	{RGB{0.0, 1.0, 0.0}, [3]float64{120.0, 1.0, 0.50}, [3]float64{120.0, 1.0, 1.0}, "#00ff00", [3]float64{0.357576, 0.715152, 0.119192}, [3]float64{0.300000, 0.600000, 0.715152}, [3]float64{0.877347, -0.861827, 0.831793}, [3]float64{0.877347, -0.879067, 0.739170}, [3]float64{0.87735, -0.83078, 1.07398}, [3]float64{0.87735, -0.95989, 0.84887}, [3]float64{136.0160, 1.197759, 0.877347}, [3]float64{139.9409, 1.148534, 0.877347}, [4]uint32{0, 65535, 0, 65535}, [3]uint8{0, 255, 0}},
	{RGB{1.0, 0.0, 0.0}, [3]float64{0.0, 1.0, 0.50}, [3]float64{0.0, 1.0, 1.0}, "#ff0000", [3]float64{0.412456, 0.212673, 0.019334}, [3]float64{0.640000, 0.330000, 0.212673}, [3]float64{0.532408, 0.800925, 0.672032}, [3]float64{0.532408, 0.782845, 0.621518}, [3]float64{0.53241, 1.75015, 0.37756}, [3]float64{0.53241, 1.67180, 0.24096}, [3]float64{39.9990, 1.045518, 0.532408}, [3]float64{38.4469, 0.999566, 0.532408}, [4]uint32{65535, 0, 0, 65535}, [3]uint8{255, 0, 0}},
	{RGB{0.0, 0.0, 0.0}, [3]float64{0.0, 0.0, 0.00}, [3]float64{0.0, 0.0, 0.0}, "#000000", [3]float64{0.000000, 0.000000, 0.000000}, [3]float64{0.312727, 0.329023, 0.000000}, [3]float64{0.000000, 0.000000, 0.000000}, [3]float64{0.000000, 0.000000, 0.000000}, [3]float64{0.00000, 0.00000, 0.00000}, [3]float64{0.00000, 0.00000, 0.00000}, [3]float64{0.0000, 0.000000, 0.000000}, [3]float64{0.0000, 0.000000, 0.000000}, [4]uint32{0, 0, 0, 65535}, [3]uint8{0, 0, 0}},
}

// For testing short-hex values, since the above contains colors which don't
// have corresponding short hexes.
var shorthexvals = []struct {
	c   RGB
	hex string
}{
	{RGB{1.0, 1.0, 1.0}, "#fff"},
	{RGB{0.6, 1.0, 1.0}, "#9ff"},
	{RGB{1.0, 0.6, 1.0}, "#f9f"},
	{RGB{1.0, 1.0, 0.6}, "#ff9"},
	{RGB{0.6, 0.6, 1.0}, "#99f"},
	{RGB{1.0, 0.6, 0.6}, "#f99"},
	{RGB{0.6, 1.0, 0.6}, "#9f9"},
	{RGB{0.6, 0.6, 0.6}, "#999"},
	{RGB{0.0, 1.0, 1.0}, "#0ff"},
	{RGB{1.0, 0.0, 1.0}, "#f0f"},
	{RGB{1.0, 1.0, 0.0}, "#ff0"},
	{RGB{0.0, 0.0, 1.0}, "#00f"},
	{RGB{0.0, 1.0, 0.0}, "#0f0"},
	{RGB{1.0, 0.0, 0.0}, "#f00"},
	{RGB{0.0, 0.0, 0.0}, "#000"},
}

/// RGBA ///
////////////

func TestRGBAConversion(t *testing.T) {
	for i, tt := range vals {
		r, g, b, a := tt.c.RGBA()
		if r != tt.rgba[0] || g != tt.rgba[1] || b != tt.rgba[2] || a != tt.rgba[3] {
			t.Errorf("%v. %v.RGBA() => (%v), want %v (delta %v)", i, tt.c, []uint32{r, g, b, a}, tt.rgba, delta)
		}
	}
}

/// RGB255 ///
////////////

func TestRGB255Conversion(t *testing.T) {
	for i, tt := range vals {
		r, g, b := RGB255(tt.c)
		if r != tt.rgb255[0] || g != tt.rgb255[1] || b != tt.rgb255[2] {
			t.Errorf("%v. %v.RGB255() => (%v), want %v (delta %v)", i, tt.c, []uint8{r, g, b}, tt.rgb255, delta)
		}
	}
}

/// ToHSV ///
///////////

func TestHsvCreation(t *testing.T) {
	for i, tt := range vals {
		c := NewRGBFromHSV(tt.hsv[0], tt.hsv[1], tt.hsv[2])
		if !AlmostEqualRGB(c, tt.c) {
			t.Errorf("%v. ToHSV(%v) => (%v), want %v (delta %v)", i, tt.hsv, c, tt.c, delta)
		}
	}
}

func TestHsvConversion(t *testing.T) {
	for i, tt := range vals {
		h, s, v := ToHSV(tt.c)
		if !almosteq(h, tt.hsv[0]) || !almosteq(s, tt.hsv[1]) || !almosteq(v, tt.hsv[2]) {
			t.Errorf("%v. %v.ToHSV() => (%v), want %v (delta %v)", i, tt.c, []float64{h, s, v}, tt.hsv, delta)
		}
	}
}

/// HSL ///
///////////

func TestHslCreation(t *testing.T) {
	for i, tt := range vals {
		c := NewRGBFromHSL(tt.hsl[0], tt.hsl[1], tt.hsl[2])
		if !AlmostEqualRGB(c, tt.c) {
			t.Errorf("%v. ToHSL(%v) => (%v), want %v (delta %v)", i, tt.hsl, c, tt.c, delta)
		}
	}
}

func TestHslConversion(t *testing.T) {
	for i, tt := range vals {
		h, s, l := ToHSL(tt.c)
		if !almosteq(h, tt.hsl[0]) || !almosteq(s, tt.hsl[1]) || !almosteq(l, tt.hsl[2]) {
			t.Errorf("%v. %v.ToHSL() => (%v), want %v (delta %v)", i, tt.c, []float64{h, s, l}, tt.hsl, delta)
		}
	}
}

/// String ///
///////////

func TestHexCreation(t *testing.T) {
	for i, tt := range vals {
		c, err := NewRGBFromHex(tt.hex)
		if err != nil || !AlmostEqualRGB(c, tt.c) {
			t.Errorf("%v. String(%v) => (%v), want %v (delta %v)", i, tt.hex, c, tt.c, delta)
		}
	}
}

func TestHEXCreation(t *testing.T) {
	for i, tt := range vals {
		c, err := NewRGBFromHex(strings.ToUpper(tt.hex))
		if err != nil || !AlmostEqualRGB(c, tt.c) {
			t.Errorf("%v. HEX(%v) => (%v), want %v (delta %v)", i, strings.ToUpper(tt.hex), c, tt.c, delta)
		}
	}
}

func TestShortHexCreation(t *testing.T) {
	for i, tt := range shorthexvals {
		c, err := NewRGBFromHex(tt.hex)
		if err != nil || !AlmostEqualRGB(c, tt.c) {
			t.Errorf("%v. String(%v) => (%v), want %v (delta %v)", i, tt.hex, c, tt.c, delta)
		}
	}
}

func TestShortHEXCreation(t *testing.T) {
	for i, tt := range shorthexvals {
		c, err := NewRGBFromHex(strings.ToUpper(tt.hex))
		if err != nil || !AlmostEqualRGB(c, tt.c) {
			t.Errorf("%v. String(%v) => (%v), want %v (delta %v)", i, strings.ToUpper(tt.hex), c, tt.c, delta)
		}
	}
}

func TestHexConversion(t *testing.T) {
	for i, tt := range vals {
		hex := tt.c.String()
		if hex != tt.hex {
			t.Errorf("%v. %v.String() => (%v), want %v (delta %v)", i, tt.c, hex, tt.hex, delta)
		}
	}
}

/// Linear ///
//////////////

// NewRGBFromLinearRGB itself is implicitly tested by XYZ conversions below (they use it).
// So what we do here is just test that the NewRGBFromFastLinearRGB approximation is "good enough"
func TestFastLinearRgb(t *testing.T) {
	const eps = 6.0 / 255.0 // We want that "within 6 RGB values total" is "good enough".

	for r := 0.0; r < 256.0; r++ {
		for g := 0.0; g < 256.0; g++ {
			for b := 0.0; b < 256.0; b++ {
				c := RGB{r / 255.0, g / 255.0, b / 255.0}
				rWant, gWant, bWant := ToLinearRGB(c)
				rGot, gGot, bGot := ToFastLinearRGB(c)
				dr, dg, db := math.Abs(rWant-rGot), math.Abs(gWant-gGot), math.Abs(bWant-bGot)
				if dr+dg+db > eps {
					t.Errorf("ToFastLinearRGB not precise enough for %v: differences are (%v, %v, %v), allowed total difference is %v", c, dr, dg, db, eps)
					return
				}

				cWant := NewRGBFromLinearRGB(r/255.0, g/255.0, b/255.0)
				cGot := NewRGBFromFastLinearRGB(r/255.0, g/255.0, b/255.0)
				dr, dg, db = math.Abs(cWant.R-cGot.R), math.Abs(cWant.G-cGot.G), math.Abs(cWant.B-cGot.B)
				if dr+dg+db > eps {
					t.Errorf("ToFastLinearRGB not precise enough for (%v, %v, %v): differences are (%v, %v, %v), allowed total difference is %v", r, g, b, dr, dg, db, eps)
					return
				}
			}
		}
	}
}

// Also include some benchmarks to make sure the `Fast` versions are actually significantly faster!
// (Sounds silly, but the original ones weren't!)

func BenchmarkColorToLinear(bench *testing.B) {
	var r, g, b float64
	for n := 0; n < bench.N; n++ {
		r, g, b = ToLinearRGB(RGB{rand.Float64(), rand.Float64(), rand.Float64()})
	}
	benchResult = r + g + b
}

func BenchmarkFastColorToLinear(bench *testing.B) {
	var r, g, b float64
	for n := 0; n < bench.N; n++ {
		r, g, b = ToFastLinearRGB(RGB{rand.Float64(), rand.Float64(), rand.Float64()})
	}
	benchResult = r + g + b
}

func BenchmarkLinearToColor(bench *testing.B) {
	var c RGB
	for n := 0; n < bench.N; n++ {
		c = NewRGBFromLinearRGB(rand.Float64(), rand.Float64(), rand.Float64())
	}
	benchResult = c.R + c.G + c.B
}

func BenchmarkFastLinearToColor(bench *testing.B) {
	var c RGB
	for n := 0; n < bench.N; n++ {
		c = NewRGBFromFastLinearRGB(rand.Float64(), rand.Float64(), rand.Float64())
	}
	benchResult = c.R + c.G + c.B
}

/// XYZ ///
///////////
func TestXyzCreation(t *testing.T) {
	for i, tt := range vals {
		c := FromXYZ(tt.xyz[0], tt.xyz[1], tt.xyz[2])
		if !AlmostEqualRGB(c, tt.c) {
			t.Errorf("%v. FromXYZ(%v) => (%v), want %v (delta %v)", i, tt.xyz, c, tt.c, delta)
		}
	}
}

func TestXyzConversion(t *testing.T) {
	for i, tt := range vals {
		x, y, z := ToXYZ(tt.c)
		if !almosteq(x, tt.xyz[0]) || !almosteq(y, tt.xyz[1]) || !almosteq(z, tt.xyz[2]) {
			t.Errorf("%v. %v.FromXYZ() => (%v), want %v (delta %v)", i, tt.c, [3]float64{x, y, z}, tt.xyz, delta)
		}
	}
}

/// xyY ///
///////////
func TestXyyCreation(t *testing.T) {
	for i, tt := range vals {
		c := FromXYY(tt.xyy[0], tt.xyy[1], tt.xyy[2])
		if !AlmostEqualRGB(c, tt.c) {
			t.Errorf("%v. ToXYY(%v) => (%v), want %v (delta %v)", i, tt.xyy, c, tt.c, delta)
		}
	}
}

func TestXyyConversion(t *testing.T) {
	for i, tt := range vals {
		x, y, Y := ToXYY(tt.c)
		if !almosteq(x, tt.xyy[0]) || !almosteq(y, tt.xyy[1]) || !almosteq(Y, tt.xyy[2]) {
			t.Errorf("%v. %v.ToXYY() => (%v), want %v (delta %v)", i, tt.c, [3]float64{x, y, Y}, tt.xyy, delta)
		}
	}
}

/// L*a*b* ///
//////////////
func TestLabCreation(t *testing.T) {
	for i, tt := range vals {
		c := Lab(tt.lab[0], tt.lab[1], tt.lab[2])
		if !AlmostEqualRGB(c, tt.c) {
			t.Errorf("%v. ToLab(%v) => (%v), want %v (delta %v)", i, tt.lab, c, tt.c, delta)
		}
	}
}

func TestLabConversion(t *testing.T) {
	for i, tt := range vals {
		l, a, b := ToLab(tt.c)
		if !almosteq(l, tt.lab[0]) || !almosteq(a, tt.lab[1]) || !almosteq(b, tt.lab[2]) {
			t.Errorf("%v. %v.ToLab() => (%v), want %v (delta %v)", i, tt.c, [3]float64{l, a, b}, tt.lab, delta)
		}
	}
}

func TestLabWhiteRefCreation(t *testing.T) {
	var D50 = [3]float64{0.96422, 1.00000, 0.82521}

	for i, tt := range vals {
		c := GenLabWhiteRef(tt.lab50[0], tt.lab50[1], tt.lab50[2], D50)
		if !AlmostEqualRGB(c, tt.c) {
			t.Errorf("%v. LabWhiteRef(%v, D50) => (%v), want %v (delta %v)", i, tt.lab50, c, tt.c, delta)
		}
	}
}

func TestLabWhiteRefConversion(t *testing.T) {
	var D50 = [3]float64{0.96422, 1.00000, 0.82521}

	for i, tt := range vals {
		l, a, b := LabWhiteRef(tt.c, D50)
		if !almosteq(l, tt.lab50[0]) || !almosteq(a, tt.lab50[1]) || !almosteq(b, tt.lab50[2]) {
			t.Errorf("%v. %v.LabWhiteRef(D50) => (%v), want %v (delta %v)", i, tt.c, [3]float64{l, a, b}, tt.lab50, delta)
		}
	}
}

/// L*u*v* ///
//////////////
func TestLuvCreation(t *testing.T) {
	for i, tt := range vals {
		c := NewRGBFromLuv(tt.luv[0], tt.luv[1], tt.luv[2])
		if !AlmostEqualRGB(c, tt.c) {
			t.Errorf("%v. ToLuv(%v) => (%v), want %v (delta %v)", i, tt.luv, c, tt.c, delta)
		}
	}
}

func TestLuvConversion(t *testing.T) {
	for i, tt := range vals {
		l, u, v := ToLuv(tt.c)
		if !almosteq(l, tt.luv[0]) || !almosteq(u, tt.luv[1]) || !almosteq(v, tt.luv[2]) {
			t.Errorf("%v. %v.ToLuv() => (%v), want %v (delta %v)", i, tt.c, [3]float64{l, u, v}, tt.luv, delta)
		}
	}
}

func TestLuvWhiteRefCreation(t *testing.T) {
	var D50 = [3]float64{0.96422, 1.00000, 0.82521}

	for i, tt := range vals {
		c := NewRGBFromLuvWhiteRef(tt.luv50[0], tt.luv50[1], tt.luv50[2], D50)
		if !AlmostEqualRGB(c, tt.c) {
			t.Errorf("%v. NewRGBFromLuvWhiteRef(%v, D50) => (%v), want %v (delta %v)", i, tt.luv50, c, tt.c, delta)
		}
	}
}

func TestLuvWhiteRefConversion(t *testing.T) {
	var D50 = [3]float64{0.96422, 1.00000, 0.82521}

	for i, tt := range vals {
		l, u, v := ToLuvWhiteRef(tt.c, D50)
		if !almosteq(l, tt.luv50[0]) || !almosteq(u, tt.luv50[1]) || !almosteq(v, tt.luv50[2]) {
			t.Errorf("%v. %v.NewRGBFromLuvWhiteRef(D50) => (%v), want %v (delta %v)", i, tt.c, [3]float64{l, u, v}, tt.luv50, delta)
		}
	}
}

/// HCL ///
///////////
// CIE-L*a*b* in polar coordinates.
func TestHclCreation(t *testing.T) {
	for i, tt := range vals {
		c := NewRGBFromHCL(tt.hcl[0], tt.hcl[1], tt.hcl[2])
		if !AlmostEqualRGB(c, tt.c) {
			t.Errorf("%v. ToHCL(%v) => (%v), want %v (delta %v)", i, tt.hcl, c, tt.c, delta)
		}
	}
}

func TestHclConversion(t *testing.T) {
	for i, tt := range vals {
		h, c, l := ToHCL(tt.c)
		if !almosteq(h, tt.hcl[0]) || !almosteq(c, tt.hcl[1]) || !almosteq(l, tt.hcl[2]) {
			t.Errorf("%v. %v.ToHCL() => (%v), want %v (delta %v)", i, tt.c, [3]float64{h, c, l}, tt.hcl, delta)
		}
	}
}

func TestHclWhiteRefCreation(t *testing.T) {
	var D50 = [3]float64{0.96422, 1.00000, 0.82521}

	for i, tt := range vals {
		c := NewRGBFromHCLWhite(tt.hcl50[0], tt.hcl50[1], tt.hcl50[2], D50)
		if !AlmostEqualRGB(c, tt.c) {
			t.Errorf("%v. HCLWhiteRef(%v, D50) => (%v), want %v (delta %v)", i, tt.hcl50, c, tt.c, delta)
		}
	}
}

func TestHclWhiteRefConversion(t *testing.T) {
	var D50 = [3]float64{0.96422, 1.00000, 0.82521}

	for i, tt := range vals {
		h, c, l := HCLWhiteRef(tt.c, D50)
		if !almosteq(h, tt.hcl50[0]) || !almosteq(c, tt.hcl50[1]) || !almosteq(l, tt.hcl50[2]) {
			t.Errorf("%v. %v.HCLWhiteRef(D50) => (%v), want %v (delta %v)", i, tt.c, [3]float64{h, c, l}, tt.hcl50, delta)
		}
	}
}

/// Test distances ///
//////////////////////

// Ground-truth from http://www.brucelindbloom.com/index.html?ColorDifferenceCalcHelp.html
var dists = []struct {
	c1  RGB
	c2  RGB
	d76 float64 // That's also dLab
	d94 float64
	d00 float64
}{
	{RGB{1.0, 1.0, 1.0}, RGB{1.0, 1.0, 1.0}, 0.0, 0.0, 0.0},
	{RGB{0.0, 0.0, 0.0}, RGB{0.0, 0.0, 0.0}, 0.0, 0.0, 0.0},

	// Just pairs of values of the table way above.
	{Lab(1.000000, 0.000000, 0.000000), Lab(0.931390, -0.353319, -0.108946), 0.37604638, 0.37604638, 0.23528129},
	{Lab(0.720892, 0.651673, -0.422133), Lab(0.977637, -0.165795, 0.602017), 1.33531088, 0.65466377, 0.75175896},
	{Lab(0.590453, 0.332846, -0.637099), Lab(0.681085, 0.483884, 0.228328), 0.88317072, 0.42541075, 0.37688153},
	{Lab(0.906026, -0.600870, 0.498993), Lab(0.533890, 0.000000, 0.000000), 0.86517280, 0.41038323, 0.39960503},
	{Lab(0.911132, -0.480875, -0.141312), Lab(0.603242, 0.982343, -0.608249), 1.56647162, 0.87431457, 0.57983482},
	{Lab(0.971393, -0.215537, 0.944780), Lab(0.322970, 0.791875, -1.078602), 2.35146891, 1.11858192, 1.03426977},
	{Lab(0.877347, -0.861827, 0.831793), Lab(0.532408, 0.800925, 0.672032), 1.70565338, 0.68800270, 0.86608245},
}

func TestLabDistance(t *testing.T) {
	for i, tt := range dists {
		d := DistanceCIE76(tt.c1, tt.c2)
		if !almosteq(d, tt.d76) {
			t.Errorf("%v. %v.DistanceCIE76(%v) => (%v), want %v (delta %v)", i, tt.c1, tt.c2, d, tt.d76, delta)
		}
	}
}

func TestCIE94Distance(t *testing.T) {
	for i, tt := range dists {
		d := DistanceCIE94(tt.c1, tt.c2)
		if !almosteq(d, tt.d94) {
			t.Errorf("%v. %v.DistanceCIE94(%v) => (%v), want %v (delta %v)", i, tt.c1, tt.c2, d, tt.d94, delta)
		}
	}
}

func TestCIEDE2000Distance(t *testing.T) {
	for i, tt := range dists {
		d := DistanceCIEDE2000(tt.c1, tt.c2)
		if !almosteq(d, tt.d00) {
			t.Errorf("%v. %v.DistanceCIEDE2000(%v) => (%v), want %v (delta %v)", i, tt.c1, tt.c2, d, tt.d00, delta)
		}
	}
}

/// Test utilities ///
//////////////////////

func TestClamp(t *testing.T) {
	cOrig := RGB{1.1, -0.1, 0.5}
	cWant := RGB{1.0, 0.0, 0.5}
	if NewRGBFromClamped(cOrig) != cWant {
		t.Errorf("%v.NewRGBFromClamped() => %v, want %v", cOrig, NewRGBFromClamped(cOrig), cWant)
	}
}

func TestMakeColor(t *testing.T) {
	cOrigNrgba := color.NRGBA{R: 123, G: 45, B: 67, A: 255}
	cOurs, ok := NewRGB(cOrigNrgba)
	r, g, b := RGB255(cOurs)
	if r != 123 || g != 45 || b != 67 || !ok {
		t.Errorf("NRGBA->Colorful->RGB255 error: %v became (%v, %v, %v, %t)", cOrigNrgba, r, g, b, ok)
	}

	cOrigNrgba64 := color.NRGBA64{R: 123 << 8, G: 45 << 8, B: 67 << 8, A: 0xffff}
	cOurs, ok = NewRGB(cOrigNrgba64)
	r, g, b = RGB255(cOurs)
	if r != 123 || g != 45 || b != 67 || !ok {
		t.Errorf("NRGBA64->Colorful->RGB255 error: %v became (%v, %v, %v, %t)", cOrigNrgba64, r, g, b, ok)
	}

	cOrigGray := color.Gray{Y: 123}
	cOurs, ok = NewRGB(cOrigGray)
	r, g, b = RGB255(cOurs)
	if r != 123 || g != 123 || b != 123 || !ok {
		t.Errorf("Gray->Colorful->RGB255 error: %v became (%v, %v, %v, %t)", cOrigGray, r, g, b, ok)
	}

	cOrigGray16 := color.Gray16{Y: 123 << 8}
	cOurs, ok = NewRGB(cOrigGray16)
	r, g, b = RGB255(cOurs)
	if r != 123 || g != 123 || b != 123 || !ok {
		t.Errorf("Gray16->Colorful->RGB255 error: %v became (%v, %v, %v, %t)", cOrigGray16, r, g, b, ok)
	}

	cOrigRgba := color.RGBA{R: 255, G: 255, B: 255}
	cOurs, ok = NewRGB(cOrigRgba)
	r, g, b = RGB255(cOurs)
	if r != 0 || g != 0 || b != 0 || ok {
		t.Errorf("RGBA->Colorful->RGB255 error: %v became (%v, %v, %v, %t)", cOrigRgba, r, g, b, ok)
	}
}

/// Issues raised on github ///
///////////////////////////////

// https://github.com/badu/term/colorful/issues/11
func TestIssue11(t *testing.T) {
	c1hex := "#1a1a46"
	c2hex := "#666666"

	c1, _ := NewRGBFromHex(c1hex)
	c2, _ := NewRGBFromHex(c2hex)

	blend := NewRGBFromBlendHSV(c1, c2, 0).String()
	if blend != c1hex {
		t.Errorf("Issue11: %v --ToHSV-> %v = %v, want %v", c1hex, c2hex, blend, c1hex)
	}
	blend = NewRGBFromBlendHSV(c1, c2, 1).String()
	if blend != c2hex {
		t.Errorf("Issue11: %v --ToHSV-> %v = %v, want %v", c1hex, c2hex, blend, c2hex)
	}

	blend = NewRGBBlendLuv(c1, c2, 0).String()
	if blend != c1hex {
		t.Errorf("Issue11: %v --ToLuv-> %v = %v, want %v", c1hex, c2hex, blend, c1hex)
	}
	blend = NewRGBBlendLuv(c1, c2, 1).String()
	if blend != c2hex {
		t.Errorf("Issue11: %v --ToLuv-> %v = %v, want %v", c1hex, c2hex, blend, c2hex)
	}

	blend = NewRGBFromBlendRGB(c1, c2, 0).String()
	if blend != c1hex {
		t.Errorf("Issue11: %v --Rgb-> %v = %v, want %v", c1hex, c2hex, blend, c1hex)
	}
	blend = NewRGBFromBlendRGB(c1, c2, 1).String()
	if blend != c2hex {
		t.Errorf("Issue11: %v --Rgb-> %v = %v, want %v", c1hex, c2hex, blend, c2hex)
	}

	blend = NewRGBFromBlendLab(c1, c2, 0).String()
	if blend != c1hex {
		t.Errorf("Issue11: %v --ToLab-> %v = %v, want %v", c1hex, c2hex, blend, c1hex)
	}
	blend = NewRGBFromBlendLab(c1, c2, 1).String()
	if blend != c2hex {
		t.Errorf("Issue11: %v --ToLab-> %v = %v, want %v", c1hex, c2hex, blend, c2hex)
	}

	blend = NewRGBFromBlendHCL(c1, c2, 0).String()
	if blend != c1hex {
		t.Errorf("Issue11: %v --ToHCL-> %v = %v, want %v", c1hex, c2hex, blend, c1hex)
	}
	blend = NewRGBFromBlendHCL(c1, c2, 1).String()
	if blend != c2hex {
		t.Errorf("Issue11: %v --ToHCL-> %v = %v, want %v", c1hex, c2hex, blend, c2hex)
	}
}

// For testing angular interpolation internal function
// NOTE: They are being tested in both directions.
var anglevals = []struct {
	a0 float64
	a1 float64
	t  float64
	at float64
}{
	{0.0, 1.0, 0.0, 0.0},
	{0.0, 1.0, 0.25, 0.25},
	{0.0, 1.0, 0.5, 0.5},
	{0.0, 1.0, 1.0, 1.0},
	{0.0, 90.0, 0.0, 0.0},
	{0.0, 90.0, 0.25, 22.5},
	{0.0, 90.0, 0.5, 45.0},
	{0.0, 90.0, 1.0, 90.0},
	{0.0, 178.0, 0.0, 0.0}, // Exact 0-180 is ambiguous.
	{0.0, 178.0, 0.25, 44.5},
	{0.0, 178.0, 0.5, 89.0},
	{0.0, 178.0, 1.0, 178.0},
	{0.0, 182.0, 0.0, 0.0}, // Exact 0-180 is ambiguous.
	{0.0, 182.0, 0.25, 315.5},
	{0.0, 182.0, 0.5, 271.0},
	{0.0, 182.0, 1.0, 182.0},
	{0.0, 270.0, 0.0, 0.0},
	{0.0, 270.0, 0.25, 337.5},
	{0.0, 270.0, 0.5, 315.0},
	{0.0, 270.0, 1.0, 270.0},
	{0.0, 359.0, 0.0, 0.0},
	{0.0, 359.0, 0.25, 359.75},
	{0.0, 359.0, 0.5, 359.5},
	{0.0, 359.0, 1.0, 359.0},
}

func TestInterpolation(t *testing.T) {
	// Forward
	for i, tt := range anglevals {
		res := interpBetwAng(tt.a0, tt.a1, tt.t)
		if !almosteqEps(res, tt.at, 1e-15) {
			t.Errorf("%v. interpBetwAng(%v, %v, %v) => (%v), want %v", i, tt.a0, tt.a1, tt.t, res, tt.at)
		}
	}
	// Backward
	for i, tt := range anglevals {
		res := interpBetwAng(tt.a1, tt.a0, 1.0-tt.t)
		if !almosteqEps(res, tt.at, 1e-15) {
			t.Errorf("%v. interpBetwAng(%v, %v, %v) => (%v), want %v", i, tt.a1, tt.a0, 1.0-tt.t, res, tt.at)
		}
	}
}

func TestHexColor(t *testing.T) {
	for _, tc := range []struct {
		hc RGB
		s  string
	}{
		{RGB{R: 0, G: 0, B: 0}, "#000000"},
		{RGB{R: 1, G: 0, B: 0}, "#ff0000"},
		{RGB{R: 0, G: 1, B: 0}, "#00ff00"},
		{RGB{R: 0, G: 0, B: 1}, "#0000ff"},
		{RGB{R: 1, G: 1, B: 1}, "#ffffff"},
	} {
		var gotHC RGB
		if err := gotHC.Scan(tc.s); err != nil {
			t.Errorf("_.Scan(%q) == %v, want <nil>", tc.s, err)
		}
		if !reflect.DeepEqual(gotHC, tc.hc) {
			t.Errorf("_.Scan(%q) wrote %v, want %v", tc.s, gotHC, tc.hc)
		}
		if gotValue, err := tc.hc.Value(); err != nil || !reflect.DeepEqual(gotValue, tc.s) {
			t.Errorf("%v.Value() == %v, %v, want %v, <nil>", tc.hc, gotValue, err, tc.s)
		}
	}
}
