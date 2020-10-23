// +build term_minimal nacl js zos plan9 windows android

package core

import (
	"errors"

	"github.com/badu/term/info"
)

func loadDynamicTerminfo(_ string) (*info.Term, error) {
	return nil, errors.New("terminal type unsupported")
}
