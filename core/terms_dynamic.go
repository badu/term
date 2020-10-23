// +build !term_minimal,!nacl,!js,!zos,!plan9,!windows,!android

package core

import (
	// This imports a dynamic version of the terminal database, which is built using infocmp.
	// This relies on a working installation of infocmp (typically supplied with ncurses).
	// We only do this for systems likely to have that -- i.e. UNIX based hosts.
	// We also don't support Android here, because you really don't want to run external programs there.
	// Generally the android terminals will be automatically included anyway.
	"github.com/badu/term/info"
	"github.com/badu/term/info/dynamic"
)

func loadDynamicTerminfo(term string) (*info.Term, error) {
	ti, _, e := dynamic.LoadTerminfo(term)
	if e != nil {
		return nil, e
	}
	return ti, nil
}
