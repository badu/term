package encoding

import (
	"strings"
	"sync"
	"unicode/utf8"

	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/encoding/korean"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/encoding/traditionalchinese"
)

var (
	encodings        = make(map[string]encoding.Encoding)
	encodingLk       sync.Mutex
	encodingFallback Fallback = FallbackFail
)

// RegisterEncoding may be called by the application to register an encoding.
// The presence of additional encodings will facilitate application usage with terminal environments where the I/O subsystem does not support Unicode.
// Windows systems use Unicode natively, and do not need any of the encoding subsystem when using Windows Console screens.
//
// Please see the Go documentation for golang.org/x/text/encoding -- most of the common ones exist already as stock variables.
// For example, ISO8859-15 can be registered using the following code:
//
//   import "golang.org/x/text/encoding/charmap"
//   ...
//   RegisterEncoding("ISO8859-15", charmap.ISO8859_15)
//
// Aliases can be registered as well, for example "8859-15" could be an alias for "ISO8859-15".
//
// For POSIX systems, the term package will check the environment variables LC_ALL, LC_CTYPE,  and LANG (in that order) to determine the character set.
// These are expected to have the following pattern:
//
//	 $language[.$codeset[@$variant]
//
// We extract only the $codeset part, which will usually be something like UTF-8 or ISO8859-15 or KOI8-R.
// Note that if the locale is either "POSIX" or "C", then we assume US-ASCII (the POSIX 'portable character set' and assume all other characters are somehow invalid.)
//
// Modern POSIX systems and terminal emulators may use UTF-8, and for those systems, this API is also unnecessary.
// For example, Darwin (MacOS X) and modern Linux running modern xterm generally will out of the box without any of this.
// Use of UTF-8 is recommended when possible, as it saves quite a lot processing overhead.
//
// Note that some encodings are quite large (for example GB18030 which is a superset of Unicode) and so the application size can be expected to increase quite a bit as each encoding is added.
// The East Asian encodings have been seen to add 100-200K per encoding to the application size.
//
func RegisterEncoding(charset string, enc encoding.Encoding) {
	encodingLk.Lock()
	charset = strings.ToLower(charset)
	encodings[charset] = enc
	encodingLk.Unlock()
}

// Fallback describes how the system behaves when the locale requires a character set that we do not support.
// The system always supports UTF-8 and US-ASCII.
// On Windows consoles, UTF-16LE is also supported automatically.
// Other character sets must be added using the RegisterEncoding API.
// (A large group of nearly all of them can be added using the RegisterAll function in the encoding sub package.)
type Fallback int

const (
	FallbackFail  = iota // FallbackFail behavior causes GetEncoding to fail when it cannot find an encoding.
	FallbackASCII        // FallbackASCII behavior causes GetEncoding to fall back to a 7-bit ASCII encoding, if no other encoding can be found.
	FallbackUTF8         // FallbackUTF8 behavior causes GetEncoding to assume UTF8 can pass unmodified upon failure. Note that this behavior is not recommended, unless you are sure your terminal can cope  with real UTF8 sequences.
)

// SetEncodingFallback changes the behavior of GetEncoding when a suitable encoding is not found.
// The default is FallbackFail, which causes GetEncoding to simply return nil.
func SetEncodingFallback(fb Fallback) {
	// caller forgot to call register, no problem
	if len(encodings) == 0 {
		Register()
	}
	encodingLk.Lock()
	encodingFallback = fb
	encodingLk.Unlock()
}

// GetEncoding is used by Screen implementors who want to locate an encoding for the given character set name.
// Note that this will return nil for either the Unicode (UTF-8) or ASCII encodings, since we don't use encodings for them but instead have our own native methods.
func GetEncoding(charset string) encoding.Encoding {
	charset = strings.ToLower(charset)
	// caller forgot to call register, no problem
	if len(encodings) == 0 {
		Register()
	}
	encodingLk.Lock()
	defer encodingLk.Unlock()
	if enc, ok := encodings[charset]; ok {
		return enc
	}
	switch encodingFallback {
	case FallbackASCII:
		return encodings["ascii"]
	case FallbackUTF8:
		return encoding.Nop
	}
	return nil
}

type validUtf8 struct{}

func (validUtf8) NewDecoder() *encoding.Decoder {
	return &encoding.Decoder{Transformer: encoding.UTF8Validator}
}

func (validUtf8) NewEncoder() *encoding.Encoder {
	return &encoding.Encoder{Transformer: encoding.UTF8Validator}
}

// Register registers all known encodings.  This is a short-cut to
// add full character set support to your program.  Note that this can
// add several megabytes to your program's size, because some of the encodings
// are rather large (particularly those from East Asia.)
func Register() {
	// UTF8 is an encoding for UTF-8.
	// All it does is verify that the UTF-8 in is valid.
	// The main reason for its existence is that it will detect and report ErrSrcShort or ErrDstShort, whereas the Nop encoding just passes every byte, blithely.
	var UTF8 encoding.Encoding = validUtf8{}
	// We always support UTF-8 and ASCII.
	encodings["utf-8"] = UTF8
	encodings["utf8"] = UTF8

	// ASCII represents the 7-bit US-ASCII scheme.
	// It decodes directly to UTF-8 without change, as all ASCII values are legal UTF-8.
	// Unicode values less than 128 (i.e. 7 bits) map 1:1 with ASCII.
	// It encodes runes outside of that to 0x1A, the ASCII substitution character.
	amap := make(map[byte]rune)
	for i := 128; i <= 255; i++ {
		amap[byte(i)] = utf8.RuneError
	}

	ASCII := &CharMap{Map: amap}
	ASCII.Init()

	encodings["us-ascii"] = ASCII
	encodings["ascii"] = ASCII
	encodings["iso646"] = ASCII

	// ISO8859_1 represents the 8-bit ISO8859-1 scheme.
	// It decodes directly to UTF-8 without change, as all ISO8859-1 values are legal UTF-8.
	// Unicode values less than 256 (i.e. 8 bits) map 1:1 with 8859-1.
	// It encodes runes outside of that to 0x1A, the ASCII substitution character.
	Iso88591 := &CharMap{}
	// 8859-1 is the 8-bit identity map for Unicode.
	Iso88591.Init()
	// We supply latin1 and latin5, because Go doesn't
	RegisterEncoding("ISO8859-1", Iso88591)

	// ISO8859_9 represents the 8-bit ISO8859-9 scheme.
	Iso88599 := &CharMap{Map: map[byte]rune{
		0xD0: 'Ğ',
		0xDD: 'İ',
		0xDE: 'Ş',
		0xF0: 'ğ',
		0xFD: 'ı',
		0xFE: 'ş',
	}}
	Iso88599.Init()

	RegisterEncoding("ISO8859-9", Iso88599)

	RegisterEncoding("ISO8859-10", charmap.ISO8859_10)
	RegisterEncoding("ISO8859-13", charmap.ISO8859_13)
	RegisterEncoding("ISO8859-14", charmap.ISO8859_14)
	RegisterEncoding("ISO8859-15", charmap.ISO8859_15)
	RegisterEncoding("ISO8859-16", charmap.ISO8859_16)
	RegisterEncoding("ISO8859-2", charmap.ISO8859_2)
	RegisterEncoding("ISO8859-3", charmap.ISO8859_3)
	RegisterEncoding("ISO8859-4", charmap.ISO8859_4)
	RegisterEncoding("ISO8859-5", charmap.ISO8859_5)
	RegisterEncoding("ISO8859-6", charmap.ISO8859_6)
	RegisterEncoding("ISO8859-7", charmap.ISO8859_7)
	RegisterEncoding("ISO8859-8", charmap.ISO8859_8)
	RegisterEncoding("KOI8-R", charmap.KOI8R)
	RegisterEncoding("KOI8-U", charmap.KOI8U)

	// Asian stuff
	RegisterEncoding("EUC-JP", japanese.EUCJP)
	RegisterEncoding("SHIFT_JIS", japanese.ShiftJIS)
	RegisterEncoding("ISO2022JP", japanese.ISO2022JP)

	RegisterEncoding("EUC-KR", korean.EUCKR)

	RegisterEncoding("GB18030", simplifiedchinese.GB18030)
	RegisterEncoding("GB2312", simplifiedchinese.HZGB2312)
	RegisterEncoding("GBK", simplifiedchinese.GBK)

	RegisterEncoding("Big5", traditionalchinese.Big5)

	// Common aliases
	aliases := map[string]string{
		"8859-1":      "ISO8859-1",
		"ISO-8859-1":  "ISO8859-1",
		"8859-13":     "ISO8859-13",
		"ISO-8859-13": "ISO8859-13",
		"8859-14":     "ISO8859-14",
		"ISO-8859-14": "ISO8859-14",
		"8859-15":     "ISO8859-15",
		"ISO-8859-15": "ISO8859-15",
		"8859-16":     "ISO8859-16",
		"ISO-8859-16": "ISO8859-16",
		"8859-2":      "ISO8859-2",
		"ISO-8859-2":  "ISO8859-2",
		"8859-3":      "ISO8859-3",
		"ISO-8859-3":  "ISO8859-3",
		"8859-4":      "ISO8859-4",
		"ISO-8859-4":  "ISO8859-4",
		"8859-5":      "ISO8859-5",
		"ISO-8859-5":  "ISO8859-5",
		"8859-6":      "ISO8859-6",
		"ISO-8859-6":  "ISO8859-6",
		"8859-7":      "ISO8859-7",
		"ISO-8859-7":  "ISO8859-7",
		"8859-8":      "ISO8859-8",
		"ISO-8859-8":  "ISO8859-8",
		"8859-9":      "ISO8859-9",
		"ISO-8859-9":  "ISO8859-9",

		"SJIS":        "Shift_JIS",
		"EUCJP":       "EUC-JP",
		"2022-JP":     "ISO2022JP",
		"ISO-2022-JP": "ISO2022JP",

		"EUCKR": "EUC-KR",

		// ISO646 isn't quite exactly ASCII, but the 1991 IRV (international reference version) is so.
		// This helps some older systems that may use "646" for POSIX locales.
		"646":    "US-ASCII",
		"ISO646": "US-ASCII",

		// Other names for UTF-8
		"UTF8": "UTF-8",
	}
	for n, v := range aliases {
		if enc := GetEncoding(v); enc != nil {
			RegisterEncoding(n, enc)
		}
	}
}
