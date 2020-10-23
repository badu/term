package core

import (
	enc "github.com/badu/term/encoding"
)

func defaultRunesFallback(c *core) {
	c.fallback = make(map[rune]string)
	defaultFallback := map[rune]string{
		enc.Sterling: "f",
		enc.DArrow:   "v",
		enc.LArrow:   "<",
		enc.RArrow:   ">",
		enc.UArrow:   "^",
		enc.Bullet:   "o",
		enc.Board:    "#",
		enc.CkBoard:  ":",
		enc.Degree:   "\\",
		enc.Diamond:  "+",
		enc.GEqual:   ">",
		enc.Pi:       "*",
		enc.HLine:    "-",
		enc.Lantern:  "#",
		enc.Plus:     "+",
		enc.LEqual:   "<",
		enc.LLCorner: "+",
		enc.LRCorner: "+",
		enc.NEqual:   "!",
		enc.PlMinus:  "#",
		enc.S1:       "~",
		enc.S3:       "-",
		enc.S7:       "-",
		enc.S9:       "_",
		enc.Block:    "#",
		enc.TTee:     "+",
		enc.RTee:     "+",
		enc.LTee:     "+",
		enc.BTee:     "+",
		enc.ULCorner: "+",
		enc.URCorner: "+",
		enc.VLine:    "|",
	}
	for k, v := range defaultFallback {
		c.fallback[k] = v
	}
}

// buildAlternateRunesMap builds a map of characters that we translate from Unicode to alternate character encodings.
// To do this, we use the standard VT100 ACS maps.
// This is only done if the terminal lacks support for Unicode; we always prefer to emit Unicode glyphs when we are able.
func buildAlternateRunesMap(c *core) {
	altChars := c.info.AltChars
	c.altChars = make(map[rune]string)
	for len(altChars) > 2 {
		src := altChars[0]
		dest := string(altChars[1])

		// vtACSNames is a map of bytes defined by info that are used in the terminals Alternate Character Set to represent other glyphs.
		// For example, the upper left corner of the box drawing set can be displayed by printing "l" while in the alternate character set.
		// Its not quite that simple, since the "l" is the info name, and it may be necessary to use a different character based on the terminal implementation (or the terminal may lack support for this altogether).
		// See buildAlternateRunesMap below for detail.
		if r, ok := (map[byte]rune{
			'+': enc.RArrow,
			',': enc.LArrow,
			'-': enc.UArrow,
			'.': enc.DArrow,
			'0': enc.Block,
			'`': enc.Diamond,
			'a': enc.CkBoard,
			'b': '␉', // VT100, Not defined by info
			'c': '␌', // VT100, Not defined by info
			'd': '␋', // VT100, Not defined by info
			'e': '␊', // VT100, Not defined by info
			'f': enc.Degree,
			'g': enc.PlMinus,
			'h': enc.Board,
			'i': enc.Lantern,
			'j': enc.LRCorner,
			'k': enc.URCorner,
			'l': enc.ULCorner,
			'm': enc.LLCorner,
			'n': enc.Plus,
			'o': enc.S1,
			'p': enc.S3,
			'q': enc.HLine,
			'r': enc.S7,
			's': enc.S9,
			't': enc.LTee,
			'u': enc.RTee,
			'v': enc.BTee,
			'w': enc.TTee,
			'x': enc.VLine,
			'y': enc.LEqual,
			'z': enc.GEqual,
			'{': enc.Pi,
			'|': enc.NEqual,
			'}': enc.Sterling,
			'~': enc.Bullet,
		})[src]; ok {
			c.altChars[r] = c.info.EnterAcs + dest + c.info.ExitAcs
		}
		altChars = altChars[2:]
	}
}
