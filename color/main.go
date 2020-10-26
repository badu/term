package color

import (
	"strconv"
	"strings"
)

// Color represents a color.  The low numeric values are the same as used by ECMA-48, and beyond that XTerm.
// A 24-bit RGB value may be used by adding in the IsRGB flag.
// For Color names we use the W3C approved color names.
//
// We use a 64-bit integer to allow future expansion if we want to add an 8-bit alpha, while still leaving us some room for extra options.
//
// Note that on various terminals colors may be approximated however, or not supported at all.
// If no suitable representation for a color is known, the library will simply not set any color, deferring to whatever default attributes the terminal uses.
type Color uint64

const (
	Default Color = 0       // Default is used to leave the Color unchanged from whatever system or terminal default may exist. It's also the zero value.
	valid   Color = 1 << 32 // ColorIsValid is used to indicate the color value is actually/ valid (initialized). This is useful to permit the zero value to be treated as the default.
	isRGB   Color = 1 << 33 // IsRGB is used to indicate that the numeric value is not a known color constant, but rather an RGB value. The lower order 3 bytes are RGB.
	Special Color = 1 << 34 // Special is a flag used to indicate that the values have special meaning, and live outside of the color space(s).
)

const (
	ValidConst = valid
)

// Note that the order of these options is important -- it follows the
// definitions used by ECMA and XTerm.  Hence any further named colors
// must begin at a value not less than 256.
const (
	Black = valid + iota
	Maroon
	Green
	Olive
	Navy
	Purple
	Teal
	Silver
	Gray
	Red
	Lime
	Yellow
	Blue
	Fuchsia
	Aqua
	White
	Noname16
	Noname17
	Noname18
	Noname19
	Noname20
	Noname21
	Noname22
	Noname23
	Noname24
	Noname25
	Noname26
	Noname27
	Noname28
	Noname29
	Noname30
	Noname31
	Noname32
	Noname33
	Noname34
	Noname35
	Noname36
	Noname37
	Noname38
	Noname39
	Noname40
	Noname41
	Noname42
	Noname43
	Noname44
	Noname45
	Noname46
	Noname47
	Noname48
	Noname49
	Noname50
	Noname51
	Noname52
	Noname53
	Noname54
	Noname55
	Noname56
	Noname57
	Noname58
	Noname59
	Noname60
	Noname61
	Noname62
	Noname63
	Noname64
	Noname65
	Noname66
	Noname67
	Noname68
	Noname69
	Noname70
	Noname71
	Noname72
	Noname73
	Noname74
	Noname75
	Noname76
	Noname77
	Noname78
	Noname79
	Noname80
	Noname81
	Noname82
	Noname83
	Noname84
	Noname85
	Noname86
	Noname87
	Noname88
	Noname89
	Noname90
	Noname91
	Noname92
	Noname93
	Noname94
	Noname95
	Noname96
	Noname97
	Noname98
	Noname99
	Noname100
	Noname101
	Noname102
	Noname103
	Noname104
	Noname105
	Noname106
	Noname107
	Noname108
	Noname109
	Noname110
	Noname111
	Noname112
	Noname113
	Noname114
	Noname115
	Noname116
	Noname117
	Noname118
	Noname119
	Noname120
	Noname121
	Noname122
	Noname123
	Noname124
	Noname125
	Noname126
	Noname127
	Noname128
	Noname129
	Noname130
	Noname131
	Noname132
	Noname133
	Noname134
	Noname135
	Noname136
	Noname137
	Noname138
	Noname139
	Noname140
	Noname141
	Noname142
	Noname143
	Noname144
	Noname145
	Noname146
	Noname147
	Noname148
	Noname149
	Noname150
	Noname151
	Noname152
	Noname153
	Noname154
	Noname155
	Noname156
	Noname157
	Noname158
	Noname159
	Noname160
	Noname161
	Noname162
	Noname163
	Noname164
	Noname165
	Noname166
	Noname167
	Noname168
	Noname169
	Noname170
	Noname171
	Noname172
	Noname173
	Noname174
	Noname175
	Noname176
	Noname177
	Noname178
	Noname179
	Noname180
	Noname181
	Noname182
	Noname183
	Noname184
	Noname185
	Noname186
	Noname187
	Noname188
	Noname189
	Noname190
	Noname191
	Noname192
	Noname193
	Noname194
	Noname195
	Noname196
	Noname197
	Noname198
	Noname199
	Noname200
	Noname201
	Noname202
	Noname203
	Noname204
	Noname205
	Noname206
	Noname207
	Noname208
	Noname209
	Noname210
	Noname211
	Noname212
	Noname213
	Noname214
	Noname215
	Noname216
	Noname217
	Noname218
	Noname219
	Noname220
	Noname221
	Noname222
	Noname223
	Noname224
	Noname225
	Noname226
	Noname227
	Noname228
	Noname229
	Noname230
	Noname231
	Noname232
	Noname233
	Noname234
	Noname235
	Noname236
	Noname237
	Noname238
	Noname239
	Noname240
	Noname241
	Noname242
	Noname243
	Noname244
	Noname245
	Noname246
	Noname247
	Noname248
	Noname249
	Noname250
	Noname251
	Noname252
	Noname253
	Noname254
	Noname255
	AliceBlue
	AntiqueWhite
	AquaMarine
	Azure
	Beige
	Bisque
	BlanchedAlmond
	BlueViolet
	Brown
	BurlyWood
	CadetBlue
	Chartreuse
	Chocolate
	Coral
	CornflowerBlue
	CornSilk
	Crimson
	DarkBlue
	DarkCyan
	DarkGoldenrod
	DarkGray
	DarkGreen
	DarkKhaki
	DarkMagenta
	DarkOliveGreen
	DarkOrange
	DarkOrchid
	DarkRed
	DarkSalmon
	DarkSeaGreen
	DarkSlateBlue
	DarkSlateGray
	DarkTurquoise
	DarkViolet
	DeepPink
	DeepSkyBlue
	DimGray
	DodgerBlue
	FireBrick
	FloralWhite
	ForestGreen
	GainsBoro
	GhostWhite
	Gold
	Goldenrod
	GreenYellow
	Honeydew
	HotPink
	IndianRed
	Indigo
	Ivory
	Khaki
	Lavender
	LavenderBlush
	LawnGreen
	LemonChiffon
	LightBlue
	LightCoral
	LightCyan
	LightGoldenrodYellow
	LightGray
	LightGreen
	LightPink
	LightSalmon
	LightSeaGreen
	LightSkyBlue
	LightSlateGray
	LightSteelBlue
	LightYellow
	LimeGreen
	Linen
	MediumAquamarine
	MediumBlue
	MediumOrchid
	MediumPurple
	MediumSeaGreen
	MediumSlateBlue
	MediumSpringGreen
	MediumTurquoise
	MediumVioletRed
	MidnightBlue
	MintCream
	MistyRose
	Moccasin
	NavajoWhite
	OldLace
	OliveDrab
	Orange
	OrangeRed
	Orchid
	PaleGoldenrod
	PaleGreen
	PaleTurquoise
	PaleVioletRed
	PapayaWhip
	PeachPuff
	Peru
	Pink
	Plum
	PowderBlue
	RebeccaPurple
	RosyBrown
	RoyalBlue
	SaddleBrown
	Salmon
	SandyBrown
	SeaGreen
	Seashell
	Sienna
	SkyBlue
	SlateBlue
	SlateGray
	Snow
	SpringGreen
	SteelBlue
	Tan
	Thistle
	Tomato
	Turquoise
	Violet
	Wheat
	WhiteSmoke
	YellowGreen
)

// Special colors.
const (
	// Reset is used to indicate that the color should use the vanilla terminal colors. (Basically go back to the defaults.)
	Reset = Special | iota
)

// NewHexColor returns a color using the given 24-bit RGB value.
func NewHexColor(v int32) Color {
	return isRGB | Color(v) | valid
}

// NewRGBColor returns a new color with the given red, green, and blue values.
// Each value must be represented in the range 0-255.
func NewRGBColor(r, g, b int32) Color {
	return NewHexColor(((r & 0xFF) << 16) | ((g & 0xFF) << 8) | (b & 0xFF))
}

// NewRGBAColor makes a color from imageColor ("image/color")
func NewRGBAColor(r, g, b, a uint32) Color {
	ai := int32(a >> 8)
	ri := (int32(r>>8) & ai) << 16
	gi := (int32(g>>8) & ai) << 8
	bi := int32(b>>8) & ai
	return NewHexColor(ri | gi | bi)
}

// Light returns the amount of light in a color.
func Light(c Color) float64 {
	v := Hex(c)
	if v < 0 {
		return 0.0
	}
	_, _, l := ToHCL(RGB{R: float64((v>>16)&0xFF) / 255, G: float64((v>>8)&0xFF) / 255, B: float64(v&0xFF) / 255})
	return l
}

// TrueColor returns the true color (RGB) version of the provided color.
// This is useful for ensuring color accuracy when using named colors.
// This will override terminal theme colors.
func TrueColor(c Color) Color {
	if c&valid == 0 {
		return Default
	}
	if c&isRGB != 0 {
		return c
	}
	return Color(Hex(c)) | isRGB | valid
}

// ToRGB returns the red, green, and blue components of the color, with each component represented as a value 0-255.
// In the event that the color cannot be broken up (not set usually), -1 is returned for each value.
func ToRGB(c Color) (int, int, int) {
	v := Hex(c)
	if v < 0 {
		return -1, -1, -1
	}
	return (int(v) >> 16) & 0xFF, (int(v) >> 8) & 0xFF, int(v) & 0xFF
}

// Valid indicates the color is a valid value (has been set).
func Valid(c Color) bool {
	return c&valid != 0
}

// IsRGB is true if the color is an RGB specific value.
func IsRGB(c Color) bool {
	return c&(valid|isRGB) == (valid | isRGB)
}

// NewRGBFromHex returns the color's hexadecimal RGB 24-bit value with each component consisting of a single byte, ala R << 16 | G << 8 | B.
// If the color is unknown or unset, -1 is returned.
func Hex(c Color) int32 {
	if c&valid == 0 {
		return -1
	}
	if c&isRGB != 0 {
		return int32(c) & 0xFFFFFF
	}

	switch c {
	case Black:
		return 0x000000
	case Maroon:
		return 0x800000
	case Green:
		return 0x008000
	case Olive:
		return 0x808000
	case Navy:
		return 0x000080
	case Purple:
		return 0x800080
	case Teal:
		return 0x008080
	case Silver:
		return 0xC0C0C0
	case Gray:
		return 0x808080
	case Red:
		return 0xFF0000
	case Lime:
		return 0x00FF00
	case Yellow:
		return 0xFFFF00
	case Blue:
		return 0x0000FF
	case Fuchsia:
		return 0xFF00FF
	case Aqua:
		return 0x00FFFF
	case White:
		return 0xFFFFFF
	case Noname16:
		return 0x000000 // black
	case Noname17:
		return 0x00005F
	case Noname18:
		return 0x000087
	case Noname19:
		return 0x0000AF
	case Noname20:
		return 0x0000D7
	case Noname21:
		return 0x0000FF // blue
	case Noname22:
		return 0x005F00
	case Noname23:
		return 0x005F5F
	case Noname24:
		return 0x005F87
	case Noname25:
		return 0x005FAF
	case Noname26:
		return 0x005FD7
	case Noname27:
		return 0x005FFF
	case Noname28:
		return 0x008700
	case Noname29:
		return 0x00875F
	case Noname30:
		return 0x008787
	case Noname31:
		return 0x0087Af
	case Noname32:
		return 0x0087D7
	case Noname33:
		return 0x0087FF
	case Noname34:
		return 0x00AF00
	case Noname35:
		return 0x00AF5F
	case Noname36:
		return 0x00AF87
	case Noname37:
		return 0x00AFAF
	case Noname38:
		return 0x00AFD7
	case Noname39:
		return 0x00AFFF
	case Noname40:
		return 0x00D700
	case Noname41:
		return 0x00D75F
	case Noname42:
		return 0x00D787
	case Noname43:
		return 0x00D7AF
	case Noname44:
		return 0x00D7D7
	case Noname45:
		return 0x00D7FF
	case Noname46:
		return 0x00FF00 // lime
	case Noname47:
		return 0x00FF5F
	case Noname48:
		return 0x00FF87
	case Noname49:
		return 0x00FFAF
	case Noname50:
		return 0x00FFd7
	case Noname51:
		return 0x00FFFF // aqua
	case Noname52:
		return 0x5F0000
	case Noname53:
		return 0x5F005F
	case Noname54:
		return 0x5F0087
	case Noname55:
		return 0x5F00AF
	case Noname56:
		return 0x5F00D7
	case Noname57:
		return 0x5F00FF
	case Noname58:
		return 0x5F5F00
	case Noname59:
		return 0x5F5F5F
	case Noname60:
		return 0x5F5F87
	case Noname61:
		return 0x5F5FAF
	case Noname62:
		return 0x5F5FD7
	case Noname63:
		return 0x5F5FFF
	case Noname64:
		return 0x5F8700
	case Noname65:
		return 0x5F875F
	case Noname66:
		return 0x5F8787
	case Noname67:
		return 0x5F87AF
	case Noname68:
		return 0x5F87D7
	case Noname69:
		return 0x5F87FF
	case Noname70:
		return 0x5FAF00
	case Noname71:
		return 0x5FAF5F
	case Noname72:
		return 0x5FAF87
	case Noname73:
		return 0x5FAFAF
	case Noname74:
		return 0x5FAFD7
	case Noname75:
		return 0x5FAFFF
	case Noname76:
		return 0x5FD700
	case Noname77:
		return 0x5FD75F
	case Noname78:
		return 0x5FD787
	case Noname79:
		return 0x5FD7AF
	case Noname80:
		return 0x5FD7D7
	case Noname81:
		return 0x5FD7FF
	case Noname82:
		return 0x5FFF00
	case Noname83:
		return 0x5FFF5F
	case Noname84:
		return 0x5FFF87
	case Noname85:
		return 0x5FFFAF
	case Noname86:
		return 0x5FFFD7
	case Noname87:
		return 0x5FFFFF
	case Noname88:
		return 0x870000
	case Noname89:
		return 0x87005F
	case Noname90:
		return 0x870087
	case Noname91:
		return 0x8700AF
	case Noname92:
		return 0x8700D7
	case Noname93:
		return 0x8700FF
	case Noname94:
		return 0x875F00
	case Noname95:
		return 0x875F5F
	case Noname96:
		return 0x875F87
	case Noname97:
		return 0x875FAF
	case Noname98:
		return 0x875FD7
	case Noname99:
		return 0x875FFF
	case Noname100:
		return 0x878700
	case Noname101:
		return 0x87875F
	case Noname102:
		return 0x878787
	case Noname103:
		return 0x8787AF
	case Noname104:
		return 0x8787D7
	case Noname105:
		return 0x8787FF
	case Noname106:
		return 0x87AF00
	case Noname107:
		return 0x87AF5F
	case Noname108:
		return 0x87AF87
	case Noname109:
		return 0x87AFAF
	case Noname110:
		return 0x87AFD7
	case Noname111:
		return 0x87AFFF
	case Noname112:
		return 0x87D700
	case Noname113:
		return 0x87D75F
	case Noname114:
		return 0x87D787
	case Noname115:
		return 0x87D7AF
	case Noname116:
		return 0x87D7D7
	case Noname117:
		return 0x87D7FF
	case Noname118:
		return 0x87FF00
	case Noname119:
		return 0x87FF5F
	case Noname120:
		return 0x87FF87
	case Noname121:
		return 0x87FFAF
	case Noname122:
		return 0x87FFD7
	case Noname123:
		return 0x87FFFF
	case Noname124:
		return 0xAF0000
	case Noname125:
		return 0xAF005F
	case Noname126:
		return 0xAF0087
	case Noname127:
		return 0xAF00AF
	case Noname128:
		return 0xAF00D7
	case Noname129:
		return 0xAF00FF
	case Noname130:
		return 0xAF5F00
	case Noname131:
		return 0xAF5F5F
	case Noname132:
		return 0xAF5F87
	case Noname133:
		return 0xAF5FAF
	case Noname134:
		return 0xAF5FD7
	case Noname135:
		return 0xAF5FFF
	case Noname136:
		return 0xAF8700
	case Noname137:
		return 0xAF875F
	case Noname138:
		return 0xAF8787
	case Noname139:
		return 0xAF87AF
	case Noname140:
		return 0xAF87D7
	case Noname141:
		return 0xAF87FF
	case Noname142:
		return 0xAFAF00
	case Noname143:
		return 0xAFAF5F
	case Noname144:
		return 0xAFAF87
	case Noname145:
		return 0xAFAFAF
	case Noname146:
		return 0xAFAFD7
	case Noname147:
		return 0xAFAFFF
	case Noname148:
		return 0xAFD700
	case Noname149:
		return 0xAFD75F
	case Noname150:
		return 0xAFD787
	case Noname151:
		return 0xAFD7AF
	case Noname152:
		return 0xAFD7D7
	case Noname153:
		return 0xAFD7FF
	case Noname154:
		return 0xAFFF00
	case Noname155:
		return 0xAFFF5F
	case Noname156:
		return 0xAFFF87
	case Noname157:
		return 0xAFFFAF
	case Noname158:
		return 0xAFFFD7
	case Noname159:
		return 0xAFFFFF
	case Noname160:
		return 0xD70000
	case Noname161:
		return 0xD7005F
	case Noname162:
		return 0xD70087
	case Noname163:
		return 0xD700AF
	case Noname164:
		return 0xD700D7
	case Noname165:
		return 0xD700FF
	case Noname166:
		return 0xD75F00
	case Noname167:
		return 0xD75F5F
	case Noname168:
		return 0xD75F87
	case Noname169:
		return 0xD75FAF
	case Noname170:
		return 0xD75FD7
	case Noname171:
		return 0xD75FFF
	case Noname172:
		return 0xD78700
	case Noname173:
		return 0xD7875F
	case Noname174:
		return 0xD78787
	case Noname175:
		return 0xD787AF
	case Noname176:
		return 0xD787D7
	case Noname177:
		return 0xD787FF
	case Noname178:
		return 0xD7AF00
	case Noname179:
		return 0xD7AF5F
	case Noname180:
		return 0xD7AF87
	case Noname181:
		return 0xD7AFAF
	case Noname182:
		return 0xD7AFD7
	case Noname183:
		return 0xD7AFFF
	case Noname184:
		return 0xD7D700
	case Noname185:
		return 0xD7D75F
	case Noname186:
		return 0xD7D787
	case Noname187:
		return 0xD7D7AF
	case Noname188:
		return 0xD7D7D7
	case Noname189:
		return 0xD7D7FF
	case Noname190:
		return 0xD7FF00
	case Noname191:
		return 0xD7FF5F
	case Noname192:
		return 0xD7FF87
	case Noname193:
		return 0xD7FFAF
	case Noname194:
		return 0xD7FFD7
	case Noname195:
		return 0xD7FFFF
	case Noname196:
		return 0xFF0000 // red
	case Noname197:
		return 0xFF005F
	case Noname198:
		return 0xFF0087
	case Noname199:
		return 0xFF00AF
	case Noname200:
		return 0xFF00D7
	case Noname201:
		return 0xFF00FF // fuchsia
	case Noname202:
		return 0xFF5F00
	case Noname203:
		return 0xFF5F5F
	case Noname204:
		return 0xFF5F87
	case Noname205:
		return 0xFF5FAF
	case Noname206:
		return 0xFF5FD7
	case Noname207:
		return 0xFF5FFF
	case Noname208:
		return 0xFF8700
	case Noname209:
		return 0xFF875F
	case Noname210:
		return 0xFF8787
	case Noname211:
		return 0xFF87AF
	case Noname212:
		return 0xFF87D7
	case Noname213:
		return 0xFF87FF
	case Noname214:
		return 0xFFAF00
	case Noname215:
		return 0xFFAF5F
	case Noname216:
		return 0xFFAF87
	case Noname217:
		return 0xFFAFAF
	case Noname218:
		return 0xFFAFD7
	case Noname219:
		return 0xFFAFFF
	case Noname220:
		return 0xFFD700
	case Noname221:
		return 0xFFD75F
	case Noname222:
		return 0xFFD787
	case Noname223:
		return 0xFFD7AF
	case Noname224:
		return 0xFFD7D7
	case Noname225:
		return 0xFFD7FF
	case Noname226:
		return 0xFFFF00 // yellow
	case Noname227:
		return 0xFFFF5F
	case Noname228:
		return 0xFFFF87
	case Noname229:
		return 0xFFFFAF
	case Noname230:
		return 0xFFFFD7
	case Noname231:
		return 0xFFFFFF // white
	case Noname232:
		return 0x080808
	case Noname233:
		return 0x121212
	case Noname234:
		return 0x1C1C1C
	case Noname235:
		return 0x262626
	case Noname236:
		return 0x303030
	case Noname237:
		return 0x3A3A3A
	case Noname238:
		return 0x444444
	case Noname239:
		return 0x4E4E4E
	case Noname240:
		return 0x585858
	case Noname241:
		return 0x626262
	case Noname242:
		return 0x6C6C6C
	case Noname243:
		return 0x767676
	case Noname244:
		return 0x808080 // grey
	case Noname245:
		return 0x8A8A8A
	case Noname246:
		return 0x949494
	case Noname247:
		return 0x9E9E9E
	case Noname248:
		return 0xA8A8A8
	case Noname249:
		return 0xB2B2B2
	case Noname250:
		return 0xBCBCBC
	case Noname251:
		return 0xC6C6C6
	case Noname252:
		return 0xD0D0D0
	case Noname253:
		return 0xDADADA
	case Noname254:
		return 0xE4E4E4
	case Noname255:
		return 0xEEEEEE
	case AliceBlue:
		return 0xF0F8FF
	case AntiqueWhite:
		return 0xFAEBD7
	case AquaMarine:
		return 0x7FFFD4
	case Azure:
		return 0xF0FFFF
	case Beige:
		return 0xF5F5DC
	case Bisque:
		return 0xFFE4C4
	case BlanchedAlmond:
		return 0xFFEBCD
	case BlueViolet:
		return 0x8A2BE2
	case Brown:
		return 0xA52A2A
	case BurlyWood:
		return 0xDEB887
	case CadetBlue:
		return 0x5F9EA0
	case Chartreuse:
		return 0x7FFF00
	case Chocolate:
		return 0xD2691E
	case Coral:
		return 0xFF7F50
	case CornflowerBlue:
		return 0x6495ED
	case CornSilk:
		return 0xFFF8DC
	case Crimson:
		return 0xDC143C
	case DarkBlue:
		return 0x00008B
	case DarkCyan:
		return 0x008B8B
	case DarkGoldenrod:
		return 0xB8860B
	case DarkGray:
		return 0xA9A9A9
	case DarkGreen:
		return 0x006400
	case DarkKhaki:
		return 0xBDB76B
	case DarkMagenta:
		return 0x8B008B
	case DarkOliveGreen:
		return 0x556B2F
	case DarkOrange:
		return 0xFF8C00
	case DarkOrchid:
		return 0x9932CC
	case DarkRed:
		return 0x8B0000
	case DarkSalmon:
		return 0xE9967A
	case DarkSeaGreen:
		return 0x8FBC8F
	case DarkSlateBlue:
		return 0x483D8B
	case DarkSlateGray:
		return 0x2F4F4F
	case DarkTurquoise:
		return 0x00CED1
	case DarkViolet:
		return 0x9400D3
	case DeepPink:
		return 0xFF1493
	case DeepSkyBlue:
		return 0x00BFFF
	case DimGray:
		return 0x696969
	case DodgerBlue:
		return 0x1E90FF
	case FireBrick:
		return 0xB22222
	case FloralWhite:
		return 0xFFFAF0
	case ForestGreen:
		return 0x228B22
	case GainsBoro:
		return 0xDCDCDC
	case GhostWhite:
		return 0xF8F8FF
	case Gold:
		return 0xFFD700
	case Goldenrod:
		return 0xDAA520
	case GreenYellow:
		return 0xADFF2F
	case Honeydew:
		return 0xF0FFF0
	case HotPink:
		return 0xFF69B4
	case IndianRed:
		return 0xCD5C5C
	case Indigo:
		return 0x4B0082
	case Ivory:
		return 0xFFFFF0
	case Khaki:
		return 0xF0E68C
	case Lavender:
		return 0xE6E6FA
	case LavenderBlush:
		return 0xFFF0F5
	case LawnGreen:
		return 0x7CFC00
	case LemonChiffon:
		return 0xFFFACD
	case LightBlue:
		return 0xADD8E6
	case LightCoral:
		return 0xF08080
	case LightCyan:
		return 0xE0FFFF
	case LightGoldenrodYellow:
		return 0xFAFAD2
	case LightGray:
		return 0xD3D3D3
	case LightGreen:
		return 0x90EE90
	case LightPink:
		return 0xFFB6C1
	case LightSalmon:
		return 0xFFA07A
	case LightSeaGreen:
		return 0x20B2AA
	case LightSkyBlue:
		return 0x87CEFA
	case LightSlateGray:
		return 0x778899
	case LightSteelBlue:
		return 0xB0C4DE
	case LightYellow:
		return 0xFFFFE0
	case LimeGreen:
		return 0x32CD32
	case Linen:
		return 0xFAF0E6
	case MediumAquamarine:
		return 0x66CDAA
	case MediumBlue:
		return 0x0000CD
	case MediumOrchid:
		return 0xBA55D3
	case MediumPurple:
		return 0x9370DB
	case MediumSeaGreen:
		return 0x3CB371
	case MediumSlateBlue:
		return 0x7B68EE
	case MediumSpringGreen:
		return 0x00FA9A
	case MediumTurquoise:
		return 0x48D1CC
	case MediumVioletRed:
		return 0xC71585
	case MidnightBlue:
		return 0x191970
	case MintCream:
		return 0xF5FFFA
	case MistyRose:
		return 0xFFE4E1
	case Moccasin:
		return 0xFFE4B5
	case NavajoWhite:
		return 0xFFDEAD
	case OldLace:
		return 0xFDF5E6
	case OliveDrab:
		return 0x6B8E23
	case Orange:
		return 0xFFA500
	case OrangeRed:
		return 0xFF4500
	case Orchid:
		return 0xDA70D6
	case PaleGoldenrod:
		return 0xEEE8AA
	case PaleGreen:
		return 0x98FB98
	case PaleTurquoise:
		return 0xAFEEEE
	case PaleVioletRed:
		return 0xDB7093
	case PapayaWhip:
		return 0xFFEFD5
	case PeachPuff:
		return 0xFFDAB9
	case Peru:
		return 0xCD853F
	case Pink:
		return 0xFFC0CB
	case Plum:
		return 0xDDA0DD
	case PowderBlue:
		return 0xB0E0E6
	case RebeccaPurple:
		return 0x663399
	case RosyBrown:
		return 0xBC8F8F
	case RoyalBlue:
		return 0x4169E1
	case SaddleBrown:
		return 0x8B4513
	case Salmon:
		return 0xFA8072
	case SandyBrown:
		return 0xF4A460
	case SeaGreen:
		return 0x2E8B57
	case Seashell:
		return 0xFFF5EE
	case Sienna:
		return 0xA0522D
	case SkyBlue:
		return 0x87CEEB
	case SlateBlue:
		return 0x6A5ACD
	case SlateGray:
		return 0x708090
	case Snow:
		return 0xFFFAFA
	case SpringGreen:
		return 0x00FF7F
	case SteelBlue:
		return 0x4682B4
	case Tan:
		return 0xD2B48C
	case Thistle:
		return 0xD8BFD8
	case Tomato:
		return 0xFF6347
	case Turquoise:
		return 0x40E0D0
	case Violet:
		return 0xEE82EE
	case Wheat:
		return 0xF5DEB3
	case WhiteSmoke:
		return 0xF5F5F5
	case YellowGreen:
		return 0x9ACD32
	}

	return -1
}

// Name - should have been Stringer implementation, but it's not needed for now
func Name(c Color) string {
	switch c {
	case Black:
		return "black"
	case Maroon:
		return "maroon"
	case Green:
		return "green"
	case Olive:
		return "olive"
	case Navy:
		return "navy"
	case Purple:
		return "purple"
	case Teal:
		return "teal"
	case Silver:
		return "silver"
	case Gray:
		return "gray"
	case Red:
		return "red"
	case Lime:
		return "lime"
	case Yellow:
		return "yellow"
	case Blue:
		return "blue"
	case Fuchsia:
		return "fuchsia"
	case Aqua:
		return "aqua"
	case White:
		return "white"
	case AliceBlue:
		return "aliceblue"
	case AntiqueWhite:
		return "antiquewhite"
	case AquaMarine:
		return "aquamarine"
	case Azure:
		return "azure"
	case Beige:
		return "beige"
	case Bisque:
		return "bisque"
	case BlanchedAlmond:
		return "blanchedalmond"
	case BlueViolet:
		return "blueviolet"
	case Brown:
		return "brown"
	case BurlyWood:
		return "burlywood"
	case CadetBlue:
		return "cadetblue"
	case Chartreuse:
		return "chartreuse"
	case Chocolate:
		return "chocolate"
	case Coral:
		return "coral"
	case CornflowerBlue:
		return "cornflowerblue"
	case CornSilk:
		return "cornsilk"
	case Crimson:
		return "crimson"
	case DarkBlue:
		return "darkblue"
	case DarkCyan:
		return "darkcyan"
	case DarkGoldenrod:
		return "darkgoldenrod"
	case DarkGray:
		return "darkgray"
	case DarkGreen:
		return "darkgreen"
	case DarkKhaki:
		return "darkkhaki"
	case DarkMagenta:
		return "darkmagenta"
	case DarkOliveGreen:
		return "darkolivegreen"
	case DarkOrange:
		return "darkorange"
	case DarkOrchid:
		return "darkorchid"
	case DarkRed:
		return "darkred"
	case DarkSalmon:
		return "darksalmon"
	case DarkSeaGreen:
		return "darkseagreen"
	case DarkSlateBlue:
		return "darkslateblue"
	case DarkSlateGray:
		return "darkslategray"
	case DarkTurquoise:
		return "darkturquoise"
	case DarkViolet:
		return "darkviolet"
	case DeepPink:
		return "deeppink"
	case DeepSkyBlue:
		return "deepskyblue"
	case DimGray:
		return "dimgray"
	case DodgerBlue:
		return "dodgerblue"
	case FireBrick:
		return "firebrick"
	case FloralWhite:
		return "floralwhite"
	case ForestGreen:
		return "forestgreen"
	case GainsBoro:
		return "gainsboro"
	case GhostWhite:
		return "ghostwhite"
	case Gold:
		return "gold"
	case Goldenrod:
		return "goldenrod"
	case GreenYellow:
		return "greenyellow"
	case Honeydew:
		return "honeydew"
	case HotPink:
		return "hotpink"
	case IndianRed:
		return "indianred"
	case Indigo:
		return "indigo"
	case Ivory:
		return "ivory"
	case Khaki:
		return "khaki"
	case Lavender:
		return "lavender"
	case LavenderBlush:
		return "lavenderblush"
	case LawnGreen:
		return "lawngreen"
	case LemonChiffon:
		return "lemonchiffon"
	case LightBlue:
		return "lightblue"
	case LightCoral:
		return "lightcoral"
	case LightCyan:
		return "lightcyan"
	case LightGoldenrodYellow:
		return "lightgoldenrodyellow"
	case LightGray:
		return "lightgray"
	case LightGreen:
		return "lightgreen"
	case LightPink:
		return "lightpink"
	case LightSalmon:
		return "lightsalmon"
	case LightSeaGreen:
		return "lightseagreen"
	case LightSkyBlue:
		return "lightskyblue"
	case LightSlateGray:
		return "lightslategray"
	case LightSteelBlue:
		return "lightsteelblue"
	case LightYellow:
		return "lightyellow"
	case LimeGreen:
		return "limegreen"
	case Linen:
		return "linen"
	case MediumAquamarine:
		return "mediumaquamarine"
	case MediumBlue:
		return "mediumblue"
	case MediumOrchid:
		return "mediumorchid"
	case MediumPurple:
		return "mediumpurple"
	case MediumSeaGreen:
		return "mediumseagreen"
	case MediumSlateBlue:
		return "mediumslateblue"
	case MediumSpringGreen:
		return "mediumspringgreen"
	case MediumTurquoise:
		return "mediumturquoise"
	case MediumVioletRed:
		return "mediumvioletred"
	case MidnightBlue:
		return "midnightblue"
	case MintCream:
		return "mintcream"
	case MistyRose:
		return "mistyrose"
	case Moccasin:
		return "moccasin"
	case NavajoWhite:
		return "navajowhite"
	case OldLace:
		return "oldlace"
	case OliveDrab:
		return "olivedrab"
	case Orange:
		return "orange"
	case OrangeRed:
		return "orangered"
	case Orchid:
		return "orchid"
	case PaleGoldenrod:
		return "palegoldenrod"
	case PaleGreen:
		return "palegreen"
	case PaleTurquoise:
		return "paleturquoise"
	case PaleVioletRed:
		return "palevioletred"
	case PapayaWhip:
		return "papayawhip"
	case PeachPuff:
		return "peachpuff"
	case Peru:
		return "peru"
	case Pink:
		return "pink"
	case Plum:
		return "plum"
	case PowderBlue:
		return "powderblue"
	case RebeccaPurple:
		return "rebeccapurple"
	case RosyBrown:
		return "rosybrown"
	case RoyalBlue:
		return "royalblue"
	case SaddleBrown:
		return "saddlebrown"
	case Salmon:
		return "salmon"
	case SandyBrown:
		return "sandybrown"
	case SeaGreen:
		return "seagreen"
	case Seashell:
		return "seashell"
	case Sienna:
		return "sienna"
	case SkyBlue:
		return "skyblue"
	case SlateBlue:
		return "slateblue"
	case SlateGray:
		return "slategray"
	case Snow:
		return "snow"
	case SpringGreen:
		return "springgreen"
	case SteelBlue:
		return "steelblue"
	case Tan:
		return "tan"
	case Thistle:
		return "thistle"
	case Tomato:
		return "tomato"
	case Turquoise:
		return "turquoise"
	case Violet:
		return "violet"
	case Wheat:
		return "wheat"
	case WhiteSmoke:
		return "whitesmoke"
	case YellowGreen:
		return "yellowgreen"
	default:
		return "noname"
	}
}

// NewColor creates a Color from a color name (W3C name).
// A hex value may be supplied as a string in the format "#ffffff".
func NewColor(name string) Color {
	switch strings.ToLower(name) {
	case "black":
		return Black
	case "maroon":
		return Maroon
	case "green":
		return Green
	case "olive":
		return Olive
	case "navy":
		return Navy
	case "purple":
		return Purple
	case "teal":
		return Teal
	case "silver":
		return Silver
	case "gray":
		return Gray
	case "red":
		return Red
	case "lime":
		return Lime
	case "yellow":
		return Yellow
	case "blue":
		return Blue
	case "fuchsia":
		return Fuchsia
	case "aqua":
		return Aqua
	case "white":
		return White
	case "aliceblue":
		return AliceBlue
	case "antiquewhite":
		return AntiqueWhite
	case "aquamarine":
		return AquaMarine
	case "azure":
		return Azure
	case "beige":
		return Beige
	case "bisque":
		return Bisque
	case "blanchedalmond":
		return BlanchedAlmond
	case "blueviolet":
		return BlueViolet
	case "brown":
		return Brown
	case "burlywood":
		return BurlyWood
	case "cadetblue":
		return CadetBlue
	case "chartreuse":
		return Chartreuse
	case "chocolate":
		return Chocolate
	case "coral":
		return Coral
	case "cornflowerblue":
		return CornflowerBlue
	case "cornsilk":
		return CornSilk
	case "crimson":
		return Crimson
	case "darkblue":
		return DarkBlue
	case "darkcyan":
		return DarkCyan
	case "darkgoldenrod":
		return DarkGoldenrod
	case "darkgray":
		return DarkGray
	case "darkgreen":
		return DarkGreen
	case "darkkhaki":
		return DarkKhaki
	case "darkmagenta":
		return DarkMagenta
	case "darkolivegreen":
		return DarkOliveGreen
	case "darkorange":
		return DarkOrange
	case "darkorchid":
		return DarkOrchid
	case "darkred":
		return DarkRed
	case "darksalmon":
		return DarkSalmon
	case "darkseagreen":
		return DarkSeaGreen
	case "darkslateblue":
		return DarkSlateBlue
	case "darkslategray":
		return DarkSlateGray
	case "darkturquoise":
		return DarkTurquoise
	case "darkviolet":
		return DarkViolet
	case "deeppink":
		return DeepPink
	case "deepskyblue":
		return DeepSkyBlue
	case "dimgray":
		return DimGray
	case "dodgerblue":
		return DodgerBlue
	case "firebrick":
		return FireBrick
	case "floralwhite":
		return FloralWhite
	case "forestgreen":
		return ForestGreen
	case "gainsboro":
		return GainsBoro
	case "ghostwhite":
		return GhostWhite
	case "gold":
		return Gold
	case "goldenrod":
		return Goldenrod
	case "greenyellow":
		return GreenYellow
	case "honeydew":
		return Honeydew
	case "hotpink":
		return HotPink
	case "indianred":
		return IndianRed
	case "indigo":
		return Indigo
	case "ivory":
		return Ivory
	case "khaki":
		return Khaki
	case "lavender":
		return Lavender
	case "lavenderblush":
		return LavenderBlush
	case "lawngreen":
		return LawnGreen
	case "lemonchiffon":
		return LemonChiffon
	case "lightblue":
		return LightBlue
	case "lightcoral":
		return LightCoral
	case "lightcyan":
		return LightCyan
	case "lightgoldenrodyellow":
		return LightGoldenrodYellow
	case "lightgray":
		return LightGray
	case "lightgreen":
		return LightGreen
	case "lightpink":
		return LightPink
	case "lightsalmon":
		return LightSalmon
	case "lightseagreen":
		return LightSeaGreen
	case "lightskyblue":
		return LightSkyBlue
	case "lightslategray":
		return LightSlateGray
	case "lightsteelblue":
		return LightSteelBlue
	case "lightyellow":
		return LightYellow
	case "limegreen":
		return LimeGreen
	case "linen":
		return Linen
	case "mediumaquamarine":
		return MediumAquamarine
	case "mediumblue":
		return MediumBlue
	case "mediumorchid":
		return MediumOrchid
	case "mediumpurple":
		return MediumPurple
	case "mediumseagreen":
		return MediumSeaGreen
	case "mediumslateblue":
		return MediumSlateBlue
	case "mediumspringgreen":
		return MediumSpringGreen
	case "mediumturquoise":
		return MediumTurquoise
	case "mediumvioletred":
		return MediumVioletRed
	case "midnightblue":
		return MidnightBlue
	case "mintcream":
		return MintCream
	case "mistyrose":
		return MistyRose
	case "moccasin":
		return Moccasin
	case "navajowhite":
		return NavajoWhite
	case "oldlace":
		return OldLace
	case "olivedrab":
		return OliveDrab
	case "orange":
		return Orange
	case "orangered":
		return OrangeRed
	case "orchid":
		return Orchid
	case "palegoldenrod":
		return PaleGoldenrod
	case "palegreen":
		return PaleGreen
	case "paleturquoise":
		return PaleTurquoise
	case "palevioletred":
		return PaleVioletRed
	case "papayawhip":
		return PapayaWhip
	case "peachpuff":
		return PeachPuff
	case "peru":
		return Peru
	case "pink":
		return Pink
	case "plum":
		return Plum
	case "powderblue":
		return PowderBlue
	case "rebeccapurple":
		return RebeccaPurple
	case "rosybrown":
		return RosyBrown
	case "royalblue":
		return RoyalBlue
	case "saddlebrown":
		return SaddleBrown
	case "salmon":
		return Salmon
	case "sandybrown":
		return SandyBrown
	case "seagreen":
		return SeaGreen
	case "seashell":
		return Seashell
	case "sienna":
		return Sienna
	case "skyblue":
		return SkyBlue
	case "slateblue":
		return SlateBlue
	case "slategray":
		return SlateGray
	case "snow":
		return Snow
	case "springgreen":
		return SpringGreen
	case "steelblue":
		return SteelBlue
	case "tan":
		return Tan
	case "thistle":
		return Thistle
	case "tomato":
		return Tomato
	case "turquoise":
		return Turquoise
	case "violet":
		return Violet
	case "wheat":
		return Wheat
	case "whitesmoke":
		return WhiteSmoke
	case "yellowgreen":
		return YellowGreen
	case "grey":
		return Gray
	case "dimgrey":
		return DimGray
	case "darkgrey":
		return DarkGray
	case "darkslategrey":
		return DarkSlateGray
	case "lightgrey":
		return LightGray
	case "lightslategrey":
		return LightSlateGray
	case "slategrey":
		return SlateGray
	default:
		if len(name) == 7 && name[0] == '#' {
			if v, e := strconv.ParseInt(name[1:], 16, 32); e == nil {
				return NewHexColor(int32(v))
			}
		}
		return Default
	}
}

// PaletteColor creates a color based on the palette index.
func PaletteColor(index int) Color {
	return Color(index) | valid
}
