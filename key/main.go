package key

import (
	"fmt"

	"github.com/badu/term"
)

// KeyEvent represents a key press.
// Usually this is a key press followed by a key release, but since terminal programs don't have a way to report key release events, we usually get just one event.
// If a key is held down then the terminal may synthesize repeated key presses at some predefined rate.
// We have no control over that, nor visibility into it.
//
// In some cases, we can have a modifier key, such as ModAlt, that can be generated with a key press.
// (This usually is represented by having the high bit set, or in some cases, by sending an ESC prior to the rune.)
//
// If the value of Key() is Rune, then the actual key value will be available with the Rune() method.
// This will be the case for most keys.
// In most situations, the modifiers will not be set.
// For example, if the rune is 'A', this will be reported without the ModShift bit set, since really can't tell if the Shift key was pressed (it might have been CAPSLOCK, or a terminal that only can send capitals, or keyboard with separate capital letters from lower case letters).
//
// Generally, terminal applications have far less visibility into keyboard activity than graphical applications.
// Hence, they should avoid depending overly much on availability of modifiers, or the availability of any specific keys.

type event struct {
	mod term.ModMask
	key term.Key
	r   rune
}

// Rune returns the rune corresponding to the key press, if it makes sense.
// The result is only defined if the value of Key() is Rune.
func (ev *event) Rune() rune {
	return ev.r
}

// Key returns a virtual key code.
// We use this to identify specific key codes, such as Enter, etc.
// Most control and function keys are reported with unique Key values.
// Normal alphanumeric and punctuation keys will generally return Rune here; the specific key can be further decoded using the Rune() function.
func (ev *event) Key() term.Key {
	return ev.key
}

// Modifiers returns the modifiers that were present with the key press.
// Note that not all platforms and terminals support this equally well, and some cases we will not not know for sure.
// Hence, applications should avoid using this in most circumstances.
func (ev *event) Modifiers() term.ModMask {
	return ev.mod
}

const (
	Shift = "Shift"
	Alt   = "Alt"
	Ctrl  = "Ctrl"
	Meta  = "Meta"

	EnterStr          = "Enter"
	BackspaceStr      = "Backspace"
	TabStr            = "Tab"
	BackTabStr        = "Backtab"
	EscStr            = "Esc"
	Backspace2Str     = "Backspace2"
	DeleteStr         = "Delete"
	InsertStr         = "Insert"
	UpStr             = "Up"
	DownStr           = "Down"
	LeftStr           = "Left"
	RightStr          = "Right"
	HomeStr           = "Home"
	EndStr            = "End"
	UpLeftStr         = "UpLeft"
	UpRightStr        = "UpRight"
	DownLeftStr       = "DownLeft"
	DownRightStr      = "DownRight"
	CenterStr         = "Center"
	PgDnStr           = "PgDn"
	PgUpStr           = "PgUp"
	ClearStr          = "Clear"
	ExitStr           = "Exit"
	CancelStr         = "Cancel"
	PauseStr          = "Pause"
	PrintStr          = "Print"
	F1Str             = "F1"
	F2Str             = "F2"
	F3Str             = "F3"
	F4Str             = "F4"
	F5Str             = "F5"
	F6Str             = "F6"
	F7Str             = "F7"
	F8Str             = "F8"
	F9Str             = "F9"
	F10Str            = "F10"
	F11Str            = "F11"
	F12Str            = "F12"
	CtrlAStr          = "Ctrl-A"
	CtrlBStr          = "Ctrl-B"
	CtrlCStr          = "Ctrl-C"
	CtrlDStr          = "Ctrl-D"
	CtrlEStr          = "Ctrl-E"
	CtrlFStr          = "Ctrl-F"
	CtrlGStr          = "Ctrl-G"
	CtrlJStr          = "Ctrl-J"
	CtrlKStr          = "Ctrl-K"
	CtrlLStr          = "Ctrl-L"
	CtrlNStr          = "Ctrl-N"
	CtrlOStr          = "Ctrl-O"
	CtrlPStr          = "Ctrl-P"
	CtrlQStr          = "Ctrl-Q"
	CtrlRStr          = "Ctrl-R"
	CtrlSStr          = "Ctrl-S"
	CtrlTStr          = "Ctrl-T"
	CtrlUStr          = "Ctrl-U"
	CtrlVStr          = "Ctrl-V"
	CtrlWStr          = "Ctrl-W"
	CtrlXStr          = "Ctrl-X"
	CtrlYStr          = "Ctrl-Y"
	CtrlZStr          = "Ctrl-Z"
	CtrlSpaceStr      = "Ctrl-Space"
	CtrlUnderscoreStr = "Ctrl-_"
	CtrlRightSqStr    = "Ctrl-]"
	CtrlBackslashStr  = "Ctrl-\\"
	CtrlCaratStr      = "Ctrl-^"
)

func (ev *event) ModName() string {
	switch ev.mod {
	case ModShift:
		return "SHIFT"
	case ModCtrl:
		return "CTRL"
	case ModAlt:
		return "ALT"
	case ModMeta:
		return "META"
	case ModNone:
		return "NONE"
	}
	return ""
}

// Name returns a printable value or the key stroke.
// This can be used when printing the event, for example.
func (ev *event) Name() string {
	var (
		s       string
		hasCtrl bool
	)

	switch ev.key {
	case Enter:
		s = EnterStr
	case Backspace:
		s = BackspaceStr
	case Tab:
		s = TabStr
	case BackTab:
		s = BackTabStr
	case Esc:
		s = EscStr
	case Backspace2:
		s = Backspace2Str
	case Delete:
		s = DeleteStr
	case Insert:
		s = InsertStr
	case Up:
		s = UpStr
	case Down:
		s = DownStr
	case Left:
		s = LeftStr
	case Right:
		s = RightStr
	case Home:
		s = HomeStr
	case End:
		s = EndStr
	case UpLeft:
		s = UpLeftStr
	case UpRight:
		s = UpRightStr
	case DownLeft:
		s = DownLeftStr
	case DownRight:
		s = DownRightStr
	case Center:
		s = CenterStr
	case PgDn:
		s = PgDnStr
	case PgUp:
		s = PgUpStr
	case Clear:
		s = ClearStr
	case Exit:
		s = ExitStr
	case Cancel:
		s = CancelStr
	case Pause:
		s = PauseStr
	case Print:
		s = PrintStr
	case F1:
		s = F1Str
	case F2:
		s = F2Str
	case F3:
		s = F3Str
	case F4:
		s = F4Str
	case F5:
		s = F5Str
	case F6:
		s = F6Str
	case F7:
		s = F7Str
	case F8:
		s = F8Str
	case F9:
		s = F9Str
	case F10:
		s = F10Str
	case F11:
		s = F11Str
	case F12:
		s = F12Str
	case CtrlA:
		s = CtrlAStr
		hasCtrl = true
	case CtrlB:
		s = CtrlBStr
		hasCtrl = true
	case CtrlC:
		s = CtrlCStr
		hasCtrl = true
	case CtrlD:
		s = CtrlDStr
		hasCtrl = true
	case CtrlE:
		s = CtrlEStr
		hasCtrl = true
	case CtrlF:
		s = CtrlFStr
		hasCtrl = true
	case CtrlG:
		s = CtrlGStr
		hasCtrl = true
	case CtrlJ:
		s = CtrlJStr
		hasCtrl = true
	case CtrlK:
		s = CtrlKStr
		hasCtrl = true
	case CtrlL:
		s = CtrlLStr
		hasCtrl = true
	case CtrlN:
		s = CtrlNStr
		hasCtrl = true
	case CtrlO:
		s = CtrlOStr
		hasCtrl = true
	case CtrlP:
		s = CtrlPStr
		hasCtrl = true
	case CtrlQ:
		s = CtrlQStr
		hasCtrl = true
	case CtrlR:
		s = CtrlRStr
		hasCtrl = true
	case CtrlS:
		s = CtrlSStr
		hasCtrl = true
	case CtrlT:
		s = CtrlTStr
		hasCtrl = true
	case CtrlU:
		s = CtrlUStr
		hasCtrl = true
	case CtrlV:
		s = CtrlVStr
		hasCtrl = true
	case CtrlW:
		s = CtrlWStr
		hasCtrl = true
	case CtrlX:
		s = CtrlXStr
		hasCtrl = true
	case CtrlY:
		s = CtrlYStr
		hasCtrl = true
	case CtrlZ:
		s = CtrlZStr
		hasCtrl = true
	case CtrlSpace:
		s = CtrlSpaceStr
		hasCtrl = true
	case CtrlUnderscore:
		s = CtrlUnderscoreStr
		hasCtrl = true
	case CtrlRightSq:
		s = CtrlRightSqStr
		hasCtrl = true
	case CtrlBackslash:
		s = CtrlBackslashStr
		hasCtrl = true
	case CtrlCarat:
		s = CtrlCaratStr
		hasCtrl = true
	case Rune:
		s = "Rune[" + string(ev.r) + "]"
	default:
		s = fmt.Sprintf("Key[%d,%d]", ev.key, int(ev.r))
	}

	// order matters, so we can display Ctrl+Alt+Shift+Meta+Whatever
	if ev.mod&ModCtrl != 0 && !hasCtrl {
		s = Ctrl + "+" + s
	}
	if ev.mod&ModAlt != 0 {
		s = Alt + "+" + s
	}
	if ev.mod&ModShift != 0 {
		s = Shift + "+" + s
	}
	if ev.mod&ModMeta != 0 {
		s = Meta + "+" + s
	}

	return s
}

// NewEvent attempts to create a suitable event.
// It parses the various ASCII control sequences if Rune is passed for Key, but if the caller has more precise information it should set that specifically.
// Callers that aren't sure about modifier state (most) should just pass ModNone.
func NewEvent(k term.Key, ch rune, mod term.ModMask) term.KeyEvent {
	if k == Rune && (ch < ' ' || ch == 0x7f) {
		// Turn specials into proper key codes. This is for control characters and the DEL.
		k = term.Key(ch)
		if mod == ModNone && ch < ' ' {
			switch term.Key(ch) {
			case Backspace, Tab, Esc, Enter:
				// these keys are directly type-able without CTRL
			default:
				// most likely entered with a CTRL key press
				mod = ModCtrl
			}
		}
	}
	return &event{key: k, r: ch, mod: mod}
}

// These are the modifiers keys that can be sent either with a key press, or a mouse event.
// Note that as of now, due to the confusion associated with Meta, and the lack of support for it on many/most platforms, the current implementations never use it.
// Instead, they use ModAlt, even for events that could possibly have been distinguished from ModAlt.
const (
	ModShift term.ModMask = 1 << iota
	ModCtrl
	ModAlt
	ModMeta
	ModNone term.ModMask = 0
)

// This is the list of named keys.
// Rune is special however, in that it is a place holder key indicating that a printable character was sent.
// The actual value of the rune will be transported in the Rune of the associated KeyEvent.
const (
	Rune term.Key = iota + 256
	Up
	Down
	Right
	Left
	UpLeft
	UpRight
	DownLeft
	DownRight
	Center
	PgUp
	PgDn
	Home
	End
	Insert
	Delete
	Help
	Exit
	Clear
	Cancel
	Print
	Pause
	BackTab
	F1
	F2
	F3
	F4
	F5
	F6
	F7
	F8
	F9
	F10
	F11
	F12
)

// These are the control keys.  Note that they overlap with other keys, perhaps.
// For example, CtrlH is the same as Backspace.
const (
	CtrlSpace term.Key = iota
	CtrlA
	CtrlB
	CtrlC
	CtrlD
	CtrlE
	CtrlF
	CtrlG
	CtrlH
	CtrlI
	CtrlJ
	CtrlK
	CtrlL
	CtrlM
	CtrlN
	CtrlO
	CtrlP
	CtrlQ
	CtrlR
	CtrlS
	CtrlT
	CtrlU
	CtrlV
	CtrlW
	CtrlX
	CtrlY
	CtrlZ
	CtrlLeftSq // Escape
	CtrlBackslash
	CtrlRightSq
	CtrlCarat
	CtrlUnderscore
)

// Special values - these are fixed in an attempt to make it more likely that aliases will encode the same way.

// These are the defined ASCII values for key codes.
// They generally match with KeyCtrl values.
const (
	NUL term.Key = iota
	SOH
	STX
	ETX
	EOT
	ENQ
	ACK
	BEL
	BS
	TAB
	LF
	VT
	FF
	CR
	SO
	SI
	DLE
	DC1
	DC2
	DC3
	DC4
	NAK
	SYN
	ETB
	CAN
	EM
	SUB
	ESC
	FS
	GS
	RS
	US
	DEL term.Key = 0x7F
)

// These keys are aliases for other names.
const (
	Backspace  = BS
	Tab        = TAB
	Esc        = ESC
	Escape     = ESC
	Enter      = CR
	Backspace2 = DEL
)

// Code represents a combination of a key code and modifiers.
type Code struct {
	Key term.Key
	Mod term.ModMask
}
