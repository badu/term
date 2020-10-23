package encoding

import (
	"bytes"
	"testing"
	"unicode/utf8"

	"golang.org/x/text/encoding"
)

func verifyMap(t *testing.T, enc encoding.Encoding, b byte, r rune) {
	verifyFromUTF(t, enc, b, r)
	verifyToUTF(t, enc, b, r)
}

func verifyFromUTF(t *testing.T, enc encoding.Encoding, b byte, r rune) {

	encoder := enc.NewEncoder()

	out := make([]byte, 6)
	utf := make([]byte, utf8.RuneLen(r))
	utf8.EncodeRune(utf, r)

	ndst, nsrc, err := encoder.Transform(out, utf, true)
	if err != nil {
		t.Errorf("Transform failed: %v", err)
	}
	if nsrc != len(utf) {
		t.Errorf("Length of source incorrect: %d != %d", nsrc, len(utf))
	}
	if ndst != 1 {
		t.Errorf("Dest length (%d) != 1", ndst)
	}
	if b != out[0] {
		t.Errorf("From UTF incorrect map %v != %v", b, out[0])
	}
}

func verifyToUTF(t *testing.T, enc encoding.Encoding, b byte, r rune) {
	decoder := enc.NewDecoder()

	out := make([]byte, 6)
	nat := []byte{b}
	utf := make([]byte, utf8.RuneLen(r))
	utf8.EncodeRune(utf, r)

	ndst, nsrc, err := decoder.Transform(out, nat, true)
	if err != nil {
		t.Errorf("Transform failed: %v", err)
	}
	if nsrc != 1 {
		t.Errorf("Src length (%d) != 1", nsrc)
	}
	if !bytes.Equal(utf, out[:ndst]) {
		t.Errorf("UTF expected %v, but got %v for %x\n", utf, out, b)
	}
}
