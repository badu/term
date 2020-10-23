package base

import (
	// The following imports just register themselves --
	// thse are the terminal types we aggregate in this package.
	_ "github.com/badu/term/info/a/ansi"
	_ "github.com/badu/term/info/v/vt100"
	_ "github.com/badu/term/info/v/vt102"
	_ "github.com/badu/term/info/v/vt220"
	_ "github.com/badu/term/info/x/xterm"
)
