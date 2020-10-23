package encoding

import (
	"sync"
	"unicode/utf8"

	"golang.org/x/text/encoding"
	"golang.org/x/text/transform"
)

// CharMap is a structure for setting up encodings for 8-bit character sets, for transforming between UTF8 and that other character set.
// It has some ideas borrowed from golang.org/x/text/encoding/charmap, but it uses a different implementation.
// This implementation uses maps, and supports user-defined maps.
//
// We do assume that a character map has a reasonable substitution character, and that valid encodings are stable (exactly a 1:1 map) and stateless
// (that is there is no shift character or anything like that.)
// Hence this approach will not work for many East Asian character sets.
//
// Measurement shows little or no measurable difference in the performance of the two approaches.
// The difference was down to a couple of nsec/op, and no consistent pattern as to which ran faster.
// With the conversion to UTF-8 the code takes about 25 nsec/op.
// The conversion in the reverse direction takes about 100 nsec/op.
// The larger cost for conversion from UTF-8 is most likely due to the need to convert the UTF-8 byte stream to a rune before conversion.
//
type CharMap struct {
	transform.NopResetter
	bytes map[rune]byte
	runes [256][]byte
	once  sync.Once

	// The map between bytes and runes.  To indicate that a specific
	// byte value is invalid for a character set, use the rune
	// utf8.RuneError.  Values that are absent from this map will
	// be assumed to have the identity mapping -- that is the default
	// is to assume ISO8859-1, where all 8-bit characters have the same
	// numeric value as their Unicode runes.  (Not to be confused with
	// the UTF-8 values, which *will* be different for non-ASCII runes.)
	//
	// If no values less than RuneSelf are changed (or have non-identity
	// mappings), then the character set is assumed to be an ASCII
	// superset, and certain assumptions and optimizations become
	// available for ASCII bytes.
	Map map[byte]rune

	// The ReplacementChar is the byte value to use for substitution.
	// It should normally be ASCIISub for ASCII encodings.  This may be
	// unset (left to zero) for mappings that are strictly ASCII supersets.
	// In that case ASCIISub will be assumed instead.
	ReplacementChar byte
}

type cmapDecoder struct {
	transform.NopResetter
	runes [256][]byte
}

type cmapEncoder struct {
	transform.NopResetter
	bytes   map[rune]byte
	replace byte
}

// Init initializes internal values of a character map.  This should
// be done early, to minimize the cost of allocation of transforms
// later.  It is not strictly necessary however, as the allocation
// functions will arrange to call it if it has not already been done.
func (c *CharMap) Init() {
	c.once.Do(c.initialize)
}

func (c *CharMap) initialize() {
	c.bytes = make(map[rune]byte)
	ascii := true

	for i := 0; i < 256; i++ {
		r, ok := c.Map[byte(i)]
		if !ok {
			r = rune(i)
		}
		if r < 128 && r != rune(i) {
			ascii = false
		}
		if r != utf8.RuneError {
			c.bytes[r] = byte(i)
		}
		utf := make([]byte, utf8.RuneLen(r))
		utf8.EncodeRune(utf, r)
		c.runes[i] = utf
	}
	if ascii && c.ReplacementChar == '\x00' {
		c.ReplacementChar = encoding.ASCIISub
	}
}

// NewDecoder returns a Decoder the converts from the 8-bit character set to UTF-8.
// Unknown mappings, if any, are mapped to '\uFFFD'.
func (c *CharMap) NewDecoder() *encoding.Decoder {
	c.Init()
	return &encoding.Decoder{Transformer: &cmapDecoder{runes: c.runes}}
}

// NewEncoder returns a Transformer that converts from UTF8 to the 8-bit character set.
// Unknown mappings are mapped to 0x1A.
func (c *CharMap) NewEncoder() *encoding.Encoder {
	c.Init()
	return &encoding.Encoder{
		Transformer: &cmapEncoder{
			bytes:   c.bytes,
			replace: c.ReplacementChar,
		},
	}
}

func (d *cmapDecoder) Transform(dst, src []byte, atEOF bool) (int, int, error) {
	var (
		err        error
		cDst, cSrc int
	)

	for _, c := range src {
		b := d.runes[c]
		l := len(b)

		if cDst+l > len(dst) {
			err = transform.ErrShortDst
			break
		}
		for i := 0; i < l; i++ {
			dst[cDst] = b[i]
			cDst++
		}
		cSrc++
	}
	return cDst, cSrc, err
}

func (d *cmapEncoder) Transform(dst, src []byte, atEOF bool) (int, int, error) {
	var (
		err        error
		cDst, cSrc int
	)
	for cSrc < len(src) {
		if cDst >= len(dst) {
			err = transform.ErrShortDst
			break
		}

		r, sz := utf8.DecodeRune(src[cSrc:])
		if r == utf8.RuneError && sz == 1 {
			// If its inconclusive due to insufficient data in in the source, report it
			if !atEOF && !utf8.FullRune(src[cSrc:]) {
				err = transform.ErrShortSrc
				break
			}
		}

		if c, ok := d.bytes[r]; ok {
			dst[cDst] = c
		} else {
			dst[cDst] = d.replace
		}
		cSrc += sz
		cDst++
	}

	return cDst, cSrc, err
}
