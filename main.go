package term

import (
	"context"
	"image"
	"os"

	"github.com/badu/term/color"
	"github.com/badu/term/style"
)

// Death is provided by both dispatchers and receivers
type Death interface {
	DyingChan() chan struct{}
}

// ResizeEvent is an interface that implemented by core
type ResizeEvent interface {
	Size() Size
}

// ResizeListener is for listeners that must implement this interface
type ResizeListener interface {
	Death
	ResizeListen() chan ResizeEvent
}

// InputListener is implemented by MouseDispatcher and KeyDispatcher
type InputListener interface {
	InChan() chan []byte
}

// ResizeDispatcher is implemented by core engine
type ResizeDispatcher interface {
	Death
	Register(r ResizeListener)
}

// MouseDispatcher is implemented in mouse package
type MouseDispatcher interface {
	ResizeListener
	InputListener
	Register(r MouseListener)
	LifeCycle(ctx context.Context, out *os.File)
	Enable()
	Disable()
}

// MouseListener is implemented by listeners of mouse events
type MouseListener interface {
	Death
	MouseListen() chan MouseEvent
}

// MouseEvent is an interface that is implemented by mouse package
type MouseEvent interface {
	Buttons() ButtonMask
	Modifiers() ModMask
	Position() (int, int)
	ButtonNames() string
	ModName() string
}

// ButtonMask is a mask of mouse buttons and wheel events.
// Mouse button presses are normally delivered as both press and release events.
// Mouse wheel events are normally just single impulse events.
// Windows supports up to eight separate buttons plus all four wheel directions, but XTerm can only support mouse buttons 1-3 and wheel up/down.
// Its not unheard of for terminals to support only one or two buttons (think Macs).
// Old terminals, and true emulations (such as vt100) won't support mice at all, of course.
type ButtonMask int16

// ModMask is a mask of modifier keys.
// Note that it will not always be possible to report modifier keys.
type ModMask int16

type KeyEvent interface {
	Rune() rune
	Key() Key
	Modifiers() ModMask
	Name() string
	ModName() string
}

// KeyListener must be implementers of KeyEvent
type KeyListener interface {
	Death
	KeyListen() chan KeyEvent
}

// we hide our implementation behind this interface
type KeyDispatcher interface {
	Death
	InputListener
	Register(r KeyListener)
	LifeCycle(ctx context.Context)
	HasKey(k Key) bool
}

// Key is a generic value for representing keys, and especially special keys (function keys, cursor movement keys, etc.)
// For normal keys, like  ASCII letters, we use Rune, and then expect the application to inspect the Rune() member of the KeyEvent.
type Key int16

// Engine is the interface of the core
type Engine interface {
	Death                                        // returns the chan that creator needs to be listen for graceful shutdown
	Start(ctx context.Context) error             // returns error if we cannot start
	ResizeDispatcher() ResizeDispatcher          // returns the event dispatcher, so listeners can call Register(r Receiver) method
	KeyDispatcher() KeyDispatcher                // returns the event dispatcher, so listeners can call Register(r Receiver) method
	MouseDispatcher() MouseDispatcher            // returns the event dispatcher, so listeners can call Register(r Receiver) method
	CanDisplay(r rune, checkFallbacks bool) bool // checks if a rune can be displayed
	CharacterSet() string                        // getter for current charset
	SetRuneFallback(orig rune, fallback string)  // sets a fallback for a rune
	UnsetRuneFallback(orig rune)                 // forgets fallback for a rune
	NumColors() int                              // returns the number of colors of the current display
	Size() *Size                                 // returns the size of the current display
	HasTrueColor() bool                          // returns true if can display true color
	Style() *style.TermStyle                     // returns the terminal styles and palette
	ActivePixels(pixels []PixelGetter)           // registers the active pixels, forgetting the old ones. This behaviour should be found in Pages
	Redraw(pixels []PixelGetter)                 // does a buffered redraw of the screen (TODO : should not be used)
	ShowCursor(where *image.Point)               // shows the cursor at the indicated position
	HideCursor()                                 // hides the cursor
	Cursor() *image.Point                        // returns the cursor current position
	Clear()                                      // cleans the screen
}

type Unicode []rune

// PixelGetter is the complete interface (both setter and getter)
type PixelGetter interface {
	DrawCh() chan PixelGetter // channel which tells core that pixel needs drawing
	Position() *image.Point   // position (x,y) where is pixel is placed
	BgCol() color.Color       // background color, if any
	FgCol() color.Color       // foreground color, if any
	Attrs() style.Mask        // the attributes, if any (bold, italic, etc)
	Rune() rune               // the rune pixel contains
	Width() int               // the width of the rune, usually is 1, but unicode adds more to it
	HasUnicode() bool         // helper for not reading unicode everytime
	Unicode() *Unicode        // if unicode, it adds to the rune
}

// PixelSetter is the complete interface (both setter and getter)
type PixelSetter interface {
	Set(r rune, fg, bg color.Color)                             // sets both colors and rune so we don't do three calls
	SetFgBg(fg, bg color.Color)                                 // sets both colors, so we don't do two calls (usually only if the colors have changed)
	SetForeground(c color.Color)                                // as name says
	SetBackground(c color.Color)                                // as name says
	SetAttrs(m style.Mask)                                      // as name says
	SetUnicode(u Unicode)                                       // as name says
	SetRune(r rune)                                             // sets the content of the pixel and triggers redrawing
	SetAll(bg, fg color.Color, m style.Mask, r rune, u Unicode) // in case we need to call all changes, without placing too many calls
}

type Pixel interface {
	PixelGetter
	PixelSetter
}
