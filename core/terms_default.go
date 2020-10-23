// +build !term_minimal

package core

import (
	// This imports the default terminal entries.  To disable, use the term_minimal build tag.
	_ "github.com/badu/term/info/extended"
)
