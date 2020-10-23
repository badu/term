// +build windows

package core

import (
	"github.com/badu/term"
)

// NewConsoleScreen returns a console based screen.  This platform
// doesn't have support for any, so it returns nil and a suitable error.
func NewConsoleScreen() (term.Engine, error) {
	return nil, ErrNoScreen
}
