package color

import (
	"testing"
)

func TestColorValues(t *testing.T) {
	var values = []struct {
		color Color
		hex   int32
	}{
		{Red, 0x00FF0000},
		{Green, 0x00008000},
		{Lime, 0x0000FF00},
		{Blue, 0x000000FF},
		{Black, 0x00000000},
		{White, 0x00FFFFFF},
		{Silver, 0x00C0C0C0},
	}

	for _, tc := range values {
		if tc.color.Hex() != tc.hex {
			t.Errorf("Color: %x != %x", tc.color.Hex(), tc.hex)
		}
	}
}

func TestColorFitting(t *testing.T) {
	var pal []Color
	for i := 0; i < 255; i++ {
		pal = append(pal, PaletteColor(i))
	}

	// Exact color fitting on ANSI colors
	for i := 0; i < 7; i++ {
		if FindColor(PaletteColor(i), pal[:8]) != PaletteColor(i) {
			t.Errorf("color ANSI fit fail at %d", i)
		}
	}
	// Grey is closest to Silver
	if FindColor(PaletteColor(8), pal[:8]) != PaletteColor(7) {
		t.Errorf("grey does not fit to silver")
	}
	// Color fitting of upper 8 colors.
	for i := 9; i < 16; i++ {
		if FindColor(PaletteColor(i), pal[:8]) != PaletteColor(i%8) {
			t.Errorf("color fit fail at %d", i)
		}
	}
	// Imperfect fit
	if FindColor(OrangeRed, pal[:16]) != Red ||
		FindColor(AliceBlue, pal[:16]) != White ||
		FindColor(Pink, pal) != Noname217 ||
		FindColor(Sienna, pal) != Noname173 ||
		FindColor(NewColor("#00FD00"), pal) != Lime {
		t.Errorf("imperfect color fit")
	}

}

func TestColorNameLookup(t *testing.T) {
	var values = []struct {
		name  string
		color Color
		rgb   bool
	}{
		{"#FF0000", Red, true},
		{"black", Black, false},
		{"orange", Orange, false},
		{"door", Default, false},
	}
	for _, v := range values {
		c := NewColor(v.name)
		if c.Hex() != v.color.Hex() {
			t.Errorf("wrong color for %v: %v", v.name, c.Hex())
		}
		if v.rgb {
			if c&IsRGB == 0 {
				t.Errorf("color should have RGB")
			}
		} else {
			if c&IsRGB != 0 {
				t.Errorf("named color should not be RGB")
			}
		}

		if TrueColor(c).Hex() != v.color.Hex() {
			t.Errorf("trueColor did not match")
		}
	}
}

func TestColorRGB(t *testing.T) {
	r, g, b := NewColor("#112233").RGB()
	if r != 0x11 || g != 0x22 || b != 0x33 {
		t.Errorf("RGB wrong value (%x, %x, %x)", r, g, b)
	}
}
