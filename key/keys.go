package key

import (
	"strings"

	"github.com/badu/term"
	enc "github.com/badu/term/encoding"
	"github.com/badu/term/info"
)

func prepareKeyMod(c *eventDispatcher, k term.Key, mod term.ModMask, val string) {
	if len(val) == 0 {
		return
	}
	// Do not override codes that already exist
	if _, exist := c.keyCodes[val]; !exist {
		c.keyExist[k] = struct{}{}
		c.keyCodes[val] = &Code{Key: k, Mod: mod}
	}
}

func prepareKeyModReplace(c *eventDispatcher, k term.Key, replace term.Key, mod term.ModMask, val string) {
	if val != "" {
		// Do not override codes that already exist
		if old, exist := c.keyCodes[val]; !exist || old.Key == replace {
			c.keyExist[k] = struct{}{}
			c.keyCodes[val] = &Code{Key: k, Mod: mod}
		}
	}
}

func prepareKeyModXTerm(c *eventDispatcher, k term.Key, val string) {
	if strings.HasPrefix(val, "\x1b[") && strings.HasSuffix(val, "~") {
		// Drop the trailing ~
		val = val[:len(val)-1]
		// These suffixes are calculated assuming Xterm style modifier suffixes.
		// Please see https://invisible-island.net/xterm/ctlseqs/ctlseqs.pdf for more information (specifically "PC-Style Function Keys").
		prepareKeyModReplace(c, k, k+12, ModShift, val+";2~")
		prepareKeyModReplace(c, k, k+48, ModAlt, val+";3~")
		prepareKeyModReplace(c, k, k+60, ModAlt|ModShift, val+";4~")
		prepareKeyModReplace(c, k, k+24, ModCtrl, val+";5~")
		prepareKeyModReplace(c, k, k+36, ModCtrl|ModShift, val+";6~")
		prepareKeyMod(c, k, ModAlt|ModCtrl, val+";7~")
		prepareKeyMod(c, k, ModShift|ModAlt|ModCtrl, val+";8~")
		prepareKeyMod(c, k, ModMeta, val+";9~")
		prepareKeyMod(c, k, ModMeta|ModShift, val+";10~")
		prepareKeyMod(c, k, ModMeta|ModAlt, val+";11~")
		prepareKeyMod(c, k, ModMeta|ModAlt|ModShift, val+";12~")
		prepareKeyMod(c, k, ModMeta|ModCtrl, val+";13~")
		prepareKeyMod(c, k, ModMeta|ModCtrl|ModShift, val+";14~")
		prepareKeyMod(c, k, ModMeta|ModCtrl|ModAlt, val+";15~")
		prepareKeyMod(c, k, ModMeta|ModCtrl|ModAlt|ModShift, val+";16~")
		return
	}
	if strings.HasPrefix(val, "\x1bO") && len(val) == 3 {
		val = val[2:]
		prepareKeyModReplace(c, k, k+12, ModShift, "\x1b[1;2"+val)
		prepareKeyModReplace(c, k, k+48, ModAlt, "\x1b[1;3"+val)
		prepareKeyModReplace(c, k, k+24, ModCtrl, "\x1b[1;5"+val)
		prepareKeyModReplace(c, k, k+36, ModCtrl|ModShift, "\x1b[1;6"+val)
		prepareKeyModReplace(c, k, k+60, ModAlt|ModShift, "\x1b[1;4"+val)
		prepareKeyMod(c, k, ModAlt|ModCtrl, "\x1b[1;7"+val)
		prepareKeyMod(c, k, ModShift|ModAlt|ModCtrl, "\x1b[1;8"+val)
		prepareKeyMod(c, k, ModMeta, "\x1b[1;9"+val)
		prepareKeyMod(c, k, ModMeta|ModShift, "\x1b[1;10"+val)
		prepareKeyMod(c, k, ModMeta|ModAlt, "\x1b[1;11"+val)
		prepareKeyMod(c, k, ModMeta|ModAlt|ModShift, "\x1b[1;12"+val)
		prepareKeyMod(c, k, ModMeta|ModCtrl, "\x1b[1;13"+val)
		prepareKeyMod(c, k, ModMeta|ModCtrl|ModShift, "\x1b[1;14"+val)
		prepareKeyMod(c, k, ModMeta|ModCtrl|ModAlt, "\x1b[1;15"+val)
		prepareKeyMod(c, k, ModMeta|ModCtrl|ModAlt|ModShift, "\x1b[1;16"+val)
	}
}

func prepareXtermModifiers(c *eventDispatcher) {
	if c.info.Modifiers != info.XTerm {
		return
	}
	prepareKeyModXTerm(c, Right, c.info.KeyRight)
	prepareKeyModXTerm(c, Left, c.info.KeyLeft)
	prepareKeyModXTerm(c, Up, c.info.KeyUp)
	prepareKeyModXTerm(c, Down, c.info.KeyDown)
	prepareKeyModXTerm(c, Insert, c.info.KeyInsert)
	prepareKeyModXTerm(c, Delete, c.info.KeyDelete)
	prepareKeyModXTerm(c, PgUp, c.info.KeyPgUp)
	prepareKeyModXTerm(c, PgDn, c.info.KeyPgDn)
	prepareKeyModXTerm(c, Home, c.info.KeyHome)
	prepareKeyModXTerm(c, End, c.info.KeyEnd)
	prepareKeyModXTerm(c, F1, c.info.KeyF1)
	prepareKeyModXTerm(c, F2, c.info.KeyF2)
	prepareKeyModXTerm(c, F3, c.info.KeyF3)
	prepareKeyModXTerm(c, F4, c.info.KeyF4)
	prepareKeyModXTerm(c, F5, c.info.KeyF5)
	prepareKeyModXTerm(c, F6, c.info.KeyF6)
	prepareKeyModXTerm(c, F7, c.info.KeyF7)
	prepareKeyModXTerm(c, F8, c.info.KeyF8)
	prepareKeyModXTerm(c, F9, c.info.KeyF9)
	prepareKeyModXTerm(c, F10, c.info.KeyF10)
	prepareKeyModXTerm(c, F11, c.info.KeyF11)
	prepareKeyModXTerm(c, F12, c.info.KeyF12)
}

func prepareKey(c *eventDispatcher, k term.Key, val string) {
	prepareKeyMod(c, k, ModNone, val)
}

func prepareKeys(c *eventDispatcher) {
	prepareKey(c, Backspace, c.info.KeyBackspace)
	prepareKey(c, F1, c.info.KeyF1)
	prepareKey(c, F2, c.info.KeyF2)
	prepareKey(c, F3, c.info.KeyF3)
	prepareKey(c, F4, c.info.KeyF4)
	prepareKey(c, F5, c.info.KeyF5)
	prepareKey(c, F6, c.info.KeyF6)
	prepareKey(c, F7, c.info.KeyF7)
	prepareKey(c, F8, c.info.KeyF8)
	prepareKey(c, F9, c.info.KeyF9)
	prepareKey(c, F10, c.info.KeyF10)
	prepareKey(c, F11, c.info.KeyF11)
	prepareKey(c, F12, c.info.KeyF12)
	prepareKey(c, Insert, c.info.KeyInsert)
	prepareKey(c, Delete, c.info.KeyDelete)
	prepareKey(c, Home, c.info.KeyHome)
	prepareKey(c, End, c.info.KeyEnd)
	prepareKey(c, Up, c.info.KeyUp)
	prepareKey(c, Down, c.info.KeyDown)
	prepareKey(c, Left, c.info.KeyLeft)
	prepareKey(c, Right, c.info.KeyRight)
	prepareKey(c, PgUp, c.info.KeyPgUp)
	prepareKey(c, PgDn, c.info.KeyPgDn)
	prepareKey(c, Help, c.info.KeyHelp)
	prepareKey(c, Print, c.info.KeyPrint)
	prepareKey(c, Cancel, c.info.KeyCancel)
	prepareKey(c, Exit, c.info.KeyExit)
	prepareKey(c, BackTab, c.info.KeyBacktab)

	prepareKeyMod(c, Right, ModShift, c.info.KeyShfRight)
	prepareKeyMod(c, Left, ModShift, c.info.KeyShfLeft)
	prepareKeyMod(c, Up, ModShift, c.info.KeyShfUp)
	prepareKeyMod(c, Down, ModShift, c.info.KeyShfDown)
	prepareKeyMod(c, Home, ModShift, c.info.KeyShfHome)
	prepareKeyMod(c, End, ModShift, c.info.KeyShfEnd)
	prepareKeyMod(c, PgUp, ModShift, c.info.KeyShfPgUp)
	prepareKeyMod(c, PgDn, ModShift, c.info.KeyShfPgDn)

	prepareKeyMod(c, Right, ModCtrl, c.info.KeyCtrlRight)
	prepareKeyMod(c, Left, ModCtrl, c.info.KeyCtrlLeft)
	prepareKeyMod(c, Up, ModCtrl, c.info.KeyCtrlUp)
	prepareKeyMod(c, Down, ModCtrl, c.info.KeyCtrlDown)
	prepareKeyMod(c, Home, ModCtrl, c.info.KeyCtrlHome)
	prepareKeyMod(c, End, ModCtrl, c.info.KeyCtrlEnd)

	// Sadly, xterm handling of keyCodes is somewhat erratic.
	// In particular, different codes are sent depending on application mode is in use or not, and the entries for many of these are simply absent from info on many systems.
	// So we insert a number of escape sequences if they are not already used, in order to have the widest correct usage.
	// Note that prepareKey will not inject codes if the escape sequence is already known.
	// We also only do this for terminals that have the application mode present.

	// Cursor mode
	if len(c.info.EnterKeypad) > 0 {
		prepareKey(c, Up, "\x1b[A")
		prepareKey(c, Down, "\x1b[B")
		prepareKey(c, Right, "\x1b[C")
		prepareKey(c, Left, "\x1b[D")
		prepareKey(c, End, "\x1b[F")
		prepareKey(c, Home, "\x1b[H")
		prepareKey(c, Delete, "\x1b[3~")
		prepareKey(c, Home, "\x1b[1~")
		prepareKey(c, End, "\x1b[4~")
		prepareKey(c, PgUp, "\x1b[5~")
		prepareKey(c, PgDn, "\x1b[6~")

		// Application mode
		prepareKey(c, Up, "\x1bOA")
		prepareKey(c, Down, "\x1bOB")
		prepareKey(c, Right, "\x1bOC")
		prepareKey(c, Left, "\x1bOD")
		prepareKey(c, Home, "\x1bOH")
	}

	prepareXtermModifiers(c)

startOver:
	// Add key mappings for control keys.
	for i := 0; i < enc.Space; i++ {
		// Do not insert direct key codes for ambiguous keys.
		// For example, ESC is used for lots of other keys, so when parsing this we don't want to fast path handling of it, but instead wait a bit before parsing it as in isolation.
		for esc := range c.keyCodes {
			if []byte(esc)[0] == byte(i) {
				continue startOver
			}
		}

		c.keyExist[term.Key(i)] = struct{}{}

		mod := ModCtrl
		switch term.Key(i) {
		case BS, TAB, ESC, CR:
			// directly typeable - no control sequence
			mod = ModNone
		}
		c.keyCodes[string(rune(i))] = &Code{Key: term.Key(i), Mod: mod}
	}
}
