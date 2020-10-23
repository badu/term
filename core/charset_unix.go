// +build !windows,!nacl,!plan9

package core

import (
	"os"
	"strings"
)

func getCharset() string {
	// Determine the character set. This can help us later.
	// Per POSIX, we search for LC_ALL first, then LC_CTYPE, and finally LANG.  First one set wins.
	locale := ""
	if locale = os.Getenv("LC_ALL"); locale == "" {
		if locale = os.Getenv("LC_CTYPE"); locale == "" {
			locale = os.Getenv("LANG")
		}
	}
	if locale == "POSIX" || locale == "C" {
		return "US-ASCII"
	}
	if i := strings.IndexRune(locale, '@'); i >= 0 {
		locale = locale[:i]
	}
	if i := strings.IndexRune(locale, '.'); i >= 0 {
		locale = locale[i+1:]
	} else {
		// Default assumption, and on Linux we can see LC_ALL without a character set, which we assume implies UTF-8.
		return "UTF-8"
	}
	// TODO: in the future, add support for aliases
	return locale
}
