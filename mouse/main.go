package mouse

import (
	"github.com/badu/term"
	"github.com/badu/term/key"
)

// These are the actual button values.
// Note that term version 1.x reversed buttons two and three on *nix based terminals.
// We use button 1 as the primary, and button 2 as the secondary, and button 3 (which is often missing) as the middle.
const (
	Button1         term.ButtonMask = 1 << iota // Usually the left (primary) mouse button.
	Button2                                     // Usually the right (secondary) mouse button.
	Button3                                     // Usually the middle mouse button.
	Button4                                     // Often a side button (thumb/next).
	Button5                                     // Often a side button (thumb/prev).
	Button6                                     // Seems like a scorpio to me
	Button7                                     // Well, it's a 7 legs spider not a mouse
	Button8                                     // Not it's an octopus
	WheelUp                                     // Wheel motion up/away from user.
	WheelDown                                   // Wheel motion down/towards user.
	WheelLeft                                   // Wheel motion to left.
	WheelRight                                  // Wheel motion to right.
	ButtonNone      term.ButtonMask = 0         // No button or wheel events.
	ButtonPrimary                   = Button1   // Alias
	ButtonSecondary                 = Button2   // Alias
	ButtonMiddle                    = Button3   // Alias
)

// MouseEvent is a mouse event.
// It is sent on either mouse up or mouse down events.
// It is also sent on mouse motion events - if the terminal supports it.
// We make every effort to ensure that mouse release events are delivered.
// Hence, click drag can be identified by a motion event with the mouse down, without any intervening button release.
// On some terminals only the initiating press and terminating release event will be delivered.
//
// Mouse wheel events, when reported, may appear on their own as individual impulses; that is, there will normally not be a release event delivered
// for mouse wheel movements.
//
// Most terminals cannot report the state of more than one button at a time -- and some cannot report motion events unless a button is pressed.
//
// Applications can inspect the time between events to resolve double or triple clicks.
type event struct {
	btn term.ButtonMask
	mod term.ModMask
	x   int
	y   int
}

// Buttons returns the list of buttons that were pressed or wheel motions.
func (ev *event) Buttons() term.ButtonMask {
	return ev.btn
}

// Modifiers returns a list of keyboard modifiers that were pressed with the mouse button(s).
func (ev *event) Modifiers() term.ModMask {
	return ev.mod
}

// Position returns the mouse position in character cells.
// The origin 0, 0 is at the upper left corner.
func (ev *event) Position() (int, int) {
	return ev.x, ev.y
}

// ButtonNames returns buttons as string
func (ev *event) ButtonNames() string {
	switch ev.btn {
	case Button1:
		return "PRIMARY"
	case Button2:
		return "SECONDARY"
	case Button3:
		return "MIDDLE"
	case Button4:
		return "THUMB_NEXT"
	case Button5:
		return "THUMB_PREV"
	case Button6:
		return "SIXTH"
	case Button7:
		return "SEVENTH"
	case Button8:
		return "EIGHT"
	case WheelUp:
		return "WHEEL_UP"
	case WheelDown:
		return "WHEEL_DOWN"
	case WheelLeft:
		return "WHEEL_LEFT"
	case WheelRight:
		return "WHEEL_RIGHT"
	case ButtonNone:
		return "NONE"
	}
	return "NONE"
}

// ModName returns the modifier's name
func (ev *event) ModName() string {
	switch ev.mod {
	case key.ModShift:
		return "SHIFT"
	case key.ModCtrl:
		return "CTRL"
	case key.ModAlt:
		return "ALT"
	case key.ModMeta:
		return "META"
	case key.ModNone:
		return "NONE"
	}
	return "none"
}

// NewEvent is used to create a new mouse event.
// Applications shouldn't need to use this; its mostly for screen implementors.
func NewEvent(x, y int, btn term.ButtonMask, mod term.ModMask) term.MouseEvent {
	return &event{x: x, y: y, btn: btn, mod: mod}
}

// TODO : implement me for below functionality
type IOnLeftButton interface {
	OnLeftButton()
}

type IOnRightButton interface {
	OnRightButton()
}

type IOnMiddleButton interface {
	OnMiddleButton()
}

func CheckInterfaces(target interface{}) {
	if _, ok := target.(IOnLeftButton); ok {

	}
	if _, ok := target.(IOnRightButton); ok {

	}
	if _, ok := target.(IOnMiddleButton); ok {

	}
}
