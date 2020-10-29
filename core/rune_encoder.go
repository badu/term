package core

import (
	"sync"
	"unicode/utf8"

	enc "github.com/badu/term/encoding"
	"golang.org/x/text/encoding"
	"golang.org/x/text/transform"
)

type encoder struct {
	sync.Mutex
	transform.Transformer
	cachedEncodedRunes map[rune][]byte // cached encoded runes
	fallback           map[rune]string // runes fallback
	altChars           map[rune]string // alternative runes
}

func newEncoder(parent *encoding.Encoder) *encoder {
	res := encoder{
		Transformer:        parent,
		cachedEncodedRunes: make(map[rune][]byte),
	}
	return &res
}

// encodeRune appends a buffer with encoded runes
func (c *encoder) encodeRune(r rune, buf []byte) []byte {
	c.Lock()
	defer c.Unlock()
	if cache, ok := c.cachedEncodedRunes[r]; ok {
		buf = append(buf, cache...)
		return buf
	}
	nb := make([]byte, 6)
	ob := make([]byte, 6)
	num := utf8.EncodeRune(ob, r)
	ob = ob[:num]
	dst := 0
	var err error
	if enco := c; enco != nil {
		enco.Reset()
		dst, _, err = enco.Transform(nb, ob, true)
	}
	if err != nil || dst == 0 || nb[0] == encoding.ASCIISub {
		// Combining characters are elided
		if len(buf) == 0 {
			if acs, ok := c.altChars[r]; ok {
				buf = append(buf, []byte(acs)...)
			} else if fb, ok := c.fallback[r]; ok {
				buf = append(buf, []byte(fb)...)
			} else {
				buf = append(buf, '?')
			}
		}
	} else {
		buf = append(buf, nb[:dst]...)
	}
	cache := make([]byte, 6)
	copy(cache, buf)
	c.cachedEncodedRunes[r] = cache
	return buf
}

func (c *encoder) defaultRunesFallback() {
	c.Lock()
	defer c.Unlock()
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
func (c *encoder) buildAlternateRunesMap(AltChars, EnterAcs, ExitAcs string) {
	c.Lock()
	defer c.Unlock()
	c.altChars = make(map[rune]string)
	for len(AltChars) > 2 {
		src := AltChars[0]
		dest := string(AltChars[1])

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
			c.altChars[r] = EnterAcs + dest + ExitAcs
		}
		AltChars = AltChars[2:]
	}
}

// setRuneFallback replaces a rune with a fallback
func (c *encoder) setRuneFallback(orig rune, fallback string) {
	c.Lock()
	defer c.Unlock()
	c.fallback[orig] = fallback
}

// unsetRuneFallback forgets a replaced rune fallback
func (c *encoder) unsetRuneFallback(orig rune) {
	c.Lock()
	defer c.Unlock()
	delete(c.fallback, orig)
}

// canDisplay - checks if a rune can be displayed, implementation of term.Engine interface
func (c *encoder) canDisplay(r rune, checkFallbacks bool) bool {
	c.Lock()
	defer c.Unlock()

	nb := make([]byte, 6)
	ob := make([]byte, 6)
	num := utf8.EncodeRune(ob, r)

	c.Reset()
	dst, _, err := c.Transform(nb, ob[:num], true)
	if dst != 0 && err == nil && nb[0] != '\x1A' {
		return true
	}

	// Terminal fallbacks always permitted, since we assume they are basically nearly perfect renditions.
	if _, ok := c.altChars[r]; ok {
		return true
	}
	if !checkFallbacks {
		return false
	}
	if _, ok := c.fallback[r]; ok {
		return true
	}
	return false
}
