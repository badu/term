package core

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log"
	"os"
	"runtime"
	"sync"

	"github.com/badu/term"
	"github.com/badu/term/color"
	enc "github.com/badu/term/encoding"
	"github.com/badu/term/info"
	_ "github.com/badu/term/info/base" // import the stock terminals
	"github.com/badu/term/key"
	"github.com/badu/term/mouse"
	"github.com/badu/term/style"
)

var (
	// ErrNoScreen indicates that no suitable screen could be found.
	// This may result from attempting to run on a platform where there is no support for either termios or console I/O (such as nacl), or from running in an environment where there is no access to a suitable console/terminal device.
	// (For example, running on without a controlling TTY or with no /dev/tty on POSIX platforms.)
	ErrNoScreen = errors.New("no suitable screen available")

	// ErrNoCharset indicates that the locale environment the program is not supported by the program, because no suitable encoding was found for it.
	// This problem never occurs if the environment is UTF-8 or UTF-16.
	ErrNoCharset = errors.New("character set not supported")
)

type Option func(core *core)

type Finalizer func()

// if the caller of the NewCore want
func WithFinalizer(c Finalizer) Option {
	return func(l *core) {
		l.finalizer = c
	}
}

// WithWinSizeBufferedChannelSize is a functional option to set the buffered channel size, used for resize events. Default is the runtime.NumCPU().
func WithWinSizeBufferedChannelSize(size int) Option {
	return func(c *core) {
		c.winSizeCh = make(chan os.Signal, size)
	}
}

// WithRunesFallback is a functional option to set a different runes fallback equivalence. See defaultRunesFallback for current defaults.
func WithRunesFallback(fallback map[rune]string) Option {
	return func(c *core) {
		if c.encoder == nil {
			if Debug {
				log.Printf("[core] encoder is NIL : should not happen!")
			}
			return
		}
		c.encoder.Lock()
		defer c.encoder.Unlock()
		c.encoder.fallback = make(map[rune]string)
		for k, v := range fallback {
			c.encoder.fallback[k] = v
		}
	}
}

// WithTrueColor is a functional option to disable true color, if needed. Just set the trueColor to "disable".
func WithTrueColor(trueColor string) Option {
	return func(c *core) {
		c.hasTrueColor = trueColor != "disable"
	}
}

// core represents a screen backed by a comm implementation.
type core struct {
	sync.Mutex                           // guards other properties
	sync.Once                            // required for registering lifecycle goroutines exactly once
	size            *term.Size           //
	comm            *info.Commander      // terminal Commander
	termIOSPrv      *termiosPrivate      // required by internalStart
	in              *os.File             // input, acquired in internalStart, released in internalShutdown
	out             *os.File             // output, acquired in internalStart, released in internalShutdown, used for displaying
	died            chan struct{}        // this is a buffered channel of size one
	winSizeCh       chan os.Signal       // listens for resize signals and transforms them into resize events in the dispatcher section
	mouseSwitch     chan bool            // listens for mouse enable/disable requests
	receivers       channels             // We need a slice of channels, on which our listeners will receive those events
	finalizer       Finalizer            // Yes, we have callback and we could reuse it, but we will affect readability doing so
	mouseDispatcher term.MouseDispatcher // mouse event dispatcher, exposes via term.Engine interface
	keyDispatcher   term.KeyDispatcher   // key event dispatcher, exposes via term.Engine interface
	encoder         *encoder             // used for encoding runes
	charset         string               // stores charset for getter
	style           term.Style           //
	cursorPosition  *term.Position       // the position of the cursor, if visible
	maximumPosition *term.Position       // the position of the cursor, outside the screen
	pixCancel       func()               // allows cancellation of listening to pixels changes
	cachedBG        color.Color          //
	cachedFG        color.Color          //
	cachedAttrs     style.Mask           //
	hasTrueColor    bool                 // as the name says
	canSetRGB       bool                 // true if len(comm.Term.SetFgRGB) > 0
	canSetBgFg      bool                 // true if len(comm.Term.SetFgBg) > 0
	canSetFg        bool                 // true if len(comm.Term.SetFg) > 0
	canSetBg        bool                 // true if len(comm.Term.SetBg) > 0
}

// NewCore returns a Engine that uses the stock TTY interface and POSIX termios, combined with a comm description taken from the $TERM environment variable.
// It returns an error if the terminal is not supported for any reason.
// For terminals that do not support dynamic resize events, the $LINES $COLUMNS environment variables can be set to the actual window size, otherwise defaults taken from the terminal database are used.
func NewCore(termEnv string, options ...Option) (term.Engine, error) {
	ti, err := info.LookupTerminfo(termEnv)
	if err != nil {
		ti, err = loadDynamicTerminfo(termEnv)
		if err != nil {
			return nil, err
		}
	}

	hasTrueColor := false
	if len(ti.SetFgBgRGB) > 0 || len(ti.SetFgRGB) > 0 || len(ti.SetBgRGB) > 0 {
		hasTrueColor = true
	}

	res := &core{
		comm:         info.NewCommander(ti),                  // terminal comm
		died:         make(chan struct{}),                    // init of died channel, a buffered channel of exactly one
		mouseSwitch:  make(chan bool, 1),                     // listens incoming requests from mouse
		receivers:    make(channels, 0),                      // init the receivers slice of channels which will register themselves for resizing events
		winSizeCh:    make(chan os.Signal, runtime.NumCPU()), // listening resize events (OS specific)
		style:        style.NewTermStyle(ti.Colors),
		charset:      getCharset(),
		hasTrueColor: hasTrueColor,
		canSetRGB:    len(ti.SetFgRGB) > 0,
		canSetBgFg:   len(ti.SetFgBg) > 0,
		canSetBg:     len(ti.SetBg) > 0,
		canSetFg:     len(ti.SetFg) > 0,
		cachedBG:     color.Default,
		cachedFG:     color.Default,
		cachedAttrs:  style.None,
	}

	info.RemoveAllInfos() // Commander was built, delete info map to free some RAM

	if e := enc.GetEncoding(res.charset); e != nil {
		res.encoder = newEncoder(e.NewEncoder())
		res.encoder.buildAlternateRunesMap(res.comm.AltChars, res.comm.EnterAcs, res.comm.ExitAcs)
		res.encoder.defaultRunesFallback()
	} else {
		return nil, ErrNoCharset
	}

	for _, o := range options {
		o(res)
	}

	// creating dispatchers for key and mouse
	res.mouseDispatcher, err = mouse.NewEventDispatcher(
		mouse.WithTerminalInfo(ti),
		mouse.WithResizeDispatcher(res),
		mouse.WithSwitchChannel(res.mouseSwitch),
	)
	if err != nil {
		if Debug {
			log.Printf("error creating mouse dispatcher : %v", err)
		}
		return nil, err
	}

	res.keyDispatcher, err = key.NewEventDispatcher(key.WithTerminalInfo(ti))
	if err != nil {
		if Debug {
			log.Printf("error creating key dispatcher : %v", err)
		}
		return nil, err
	}

	return res, nil
}

// Start implements the term.Engine interface, called from the composition to pass a cancellation context
func (c *core) Start(ctx context.Context) error {
	var err error
	c.Once.Do(func() {
		c.Lock()
		defer c.Unlock()

		if err = c.internalStart(); err != nil {
			if Debug {
				log.Printf("error while internal starting : %v", err)
			}
			return
		}

		c.lifeCycle(ctx) // mounting context cancel listener
		c.keyDispatcher.LifeCycle(ctx)
		if c.comm.HasMouse { // if we have mouse support
			c.mouseDispatcher.LifeCycle(ctx)
			c.mouseDispatcher.Enable()
		}
		if Debug {
			log.Println("[START] multiplexer mounted.")
		}

		c.comm.PutEnterCA(c.out)
		c.comm.PutHideCursor(c.out)
		c.comm.GoTo(c.out, c.maximumPosition.Hash()) // put cursor outside screen
		c.comm.PutEnableAcs(c.out)
		c.comm.PutClear(c.out)

		ev := &EventResize{size: c.size}   // create one event for everyone
		for _, cons := range c.receivers { // dispatch initial resize event, to inform listeners about width and height
			cons <- ev
		}
	})
	return err
}

// ResizeDispatcher implements the term.Engine interface, exposes so call to Register(r Receiver) method
func (c *core) ResizeDispatcher() term.ResizeDispatcher {
	c.Lock()
	defer c.Unlock()

	return c
}

// KeyDispatcher implements the term.Engine interface, exposes so call to Register(r Receiver) method
func (c *core) KeyDispatcher() term.KeyDispatcher {
	c.Lock()
	defer c.Unlock()

	return c.keyDispatcher
}

// MouseDispatcher implements the term.Engine interface, exposes so call to Register(r Receiver) method
func (c *core) MouseDispatcher() term.MouseDispatcher {
	c.Lock()
	defer c.Unlock()

	return c.mouseDispatcher
}

// CanDisplay - checks if a rune can be displayed, implementation of term.Engine interface
func (c *core) CanDisplay(r rune, checkFallbacks bool) bool {
	return c.encoder.canDisplay(r, checkFallbacks)
}

// CharacterSet exposes current charset, implementation for term.Engine interface
func (c *core) CharacterSet() string {
	c.Lock()
	defer c.Unlock()

	return c.charset
}

// SetRuneFallback replaces a rune with a fallback
func (c *core) SetRuneFallback(orig rune, fallback string) {
	c.encoder.setRuneFallback(orig, fallback)
}

// UnsetRuneFallback forgets a replaced rune fallback
func (c *core) UnsetRuneFallback(orig rune) {
	c.encoder.unsetRuneFallback(orig)
}

// Size returns the current size of the terminal window
func (c *core) Size() *term.Size {
	c.Lock()
	defer c.Unlock()

	return c.size
}

// Colors returns the number of color of the current terminal
func (c *core) NumColors() int {
	c.Lock()
	defer c.Unlock()

	if c.hasTrueColor {
		return 1 << 24
	}
	return c.comm.Colors
}

// Out returns the output file
func (c *core) Out() *os.File {
	c.Lock()
	defer c.Unlock()

	return c.out
}

// HasTrueColor
func (c *core) HasTrueColor() bool {
	c.Lock()
	defer c.Unlock()

	return c.hasTrueColor
}

// ActivePixels gathers changing channels and listens them inside a goroutine, so we can automatically draw changes and immediately draws all the pixels
func (c *core) ActivePixels(pixels []term.PixelGetter) {
	c.Lock()
	if c.pixCancel != nil {
		c.pixCancel() // shutdownPixel previous context, so we exit "main" goroutine
	}

	var ctx context.Context
	ctx, c.pixCancel = context.WithCancel(context.Background()) // create a new context, to allow cancellation

	shutdownPixel := newCancellationPixel() // create one cancellation pixel, it will be used for sending shutdown message to all goroutines below

	for _, pixel := range pixels {
		// mount a goroutine for each pixel. The exit mechanism is a convention: a pixel that has -1,-1 coordinates
		go func(out *os.File, pix term.PixelGetter) {
			for msg := range pix.DrawCh() { // listen incoming messages over the pixel draw request channel
				if msg.PositionHash() == term.MinusOneMinusOne { // check if this is the cancellation pixel, we're exiting the goroutine
					return
				}
				go func(o *os.File, p term.PixelGetter) { // running in a separate goroutine, because it blocks reading new messages
					c.Lock()
					defer c.Unlock()
					c.drawPixels(o, p)
				}(out, msg)
			}
		}(c.out, pixel)
		// for each pixel, mounting a kill switch, which will write the shutdown message when the context is done
		go func(pixCh chan term.PixelGetter, done <-chan struct{}, shutdownPix term.PixelGetter) {
			<-done               // blocking wait for done
			pixCh <- shutdownPix // write shutdownPixel pixel, so go routine above exits
		}(pixel.DrawCh(), ctx.Done(), shutdownPixel) // observe that all parameters are passed to the goroutine
	}
	c.Unlock() // Redraw locks it again
	c.Redraw(pixels)
	if Debug {
		log.Printf("[core] %d pixels were drawn [%03d x %03d]", len(pixels), c.size.Width, c.size.Height)
	}
}

// Redraw immediately draws all the pixels
// Yes, the pixels needs to be the same as in ActivePixels, but we don't check this
func (c *core) Redraw(cells []term.PixelGetter) {
	c.Lock()
	defer c.Unlock()
	buf := bytes.NewBuffer(nil)
	c.drawPixels(buf, cells...) // we use buffering, since we're redrawing everything

	if _, err := buf.WriteTo(c.out); err != nil { // writing buffer content to out
		if Debug {
			log.Printf("error writing to out : " + err.Error())
		}
	}
}

// Cursor returns the current cursor position
func (c *core) Cursor() *term.Position {
	c.Lock()
	defer c.Unlock()

	return c.cursorPosition
}

// ShowCursor displays the cursor at the indicated coordinates
func (c *core) ShowCursor(where *term.Position) {
	c.Lock()
	defer c.Unlock()

	if where.Hash() > term.MinusOneMinusOne || where.Hash() > c.maximumPosition.Hash() {
		// does not update cursor position
		if c.comm.HasHideCursor {
			c.comm.PutHideCursor(c.out)
			return
		}
		// No way to hide cursor, stick it at bottom right of screen
		c.comm.GoTo(c.out, c.maximumPosition.Hash())
		return
	}
	c.cursorPosition = where
	c.comm.GoTo(c.out, c.cursorPosition.Hash())
	c.comm.PutShowCursor(c.out)
}

// HideCursor hides the cursor from the screen
func (c *core) HideCursor() {
	c.Lock()
	defer c.Unlock()

	c.cursorPosition = nil
	// does not update cursor position
	if c.comm.HasHideCursor {
		log.Println("has hide cursor")
		c.comm.PutHideCursor(c.out)
		return
	}
	log.Println("cannot hide cursor : moving it outside of screen")
	// No way to hide cursor, stick it at bottom right of screen
	c.comm.GoTo(c.out, c.maximumPosition.Hash())
}

// Clear
func (c *core) Clear() {
	c.Lock()
	defer c.Unlock()
	c.comm.PutClear(c.out)
}

// Style
func (c *core) Style() term.Style {
	return c.style
}

// resize remembers the width and height of the terminal. if a shutdown flag is set, the pixels listeners are forgot
func (c *core) resize(w, h int, shutdown bool) {
	if shutdown && c.pixCancel != nil {
		c.pixCancel() // cancel context if we're shutting down and cancellation was declared
	}
	if c.size != nil && (c.size.Height == h && c.size.Width == w) {
		return
	}
	c.size = &term.Size{Width: w, Height: h}
	mp := term.NewPosition(w, h)
	c.maximumPosition = mp
	c.comm.MakeGoToCache(c.size, term.Hash)
}

// drawPixels - locked inside caller function
func (c *core) drawPixels(w io.Writer, pixels ...term.PixelGetter) {
	for _, pixel := range pixels {
		c.comm.GoTo(w, pixel.PositionHash()) // first we go to
		fg, bg, attrs := pixel.Style()       // read pixel colors and attributes
		if fg == c.cachedFG && bg == c.cachedBG && c.cachedAttrs == attrs {
			goto cachedStyle // if the previous pixel had the same attributes and colors, we jump to displaying runes
		}

		if c.comm.Colors > 0 {
			c.comm.PutAttrOff(w) // about to send colors

			if fg == color.Reset || bg == color.Reset {
				c.comm.PutResetFgBg(w)
			}

			if c.hasTrueColor && c.canSetRGB { // we can use SetFgRGB
				if color.IsRGB(fg) && color.IsRGB(bg) { // both are RGB
					c.comm.WriteBothColors(w, fg, bg, false)
					goto colorDone
				}

				// not both are RGB
				if color.IsRGB(fg) {
					c.comm.WriteColor(w, fg, true, false)
					fg = color.Default //  resets cache
				}

				if color.IsRGB(bg) {
					c.comm.WriteColor(w, bg, false, false)
					bg = color.Default // resets cache
				}
			}

			if color.Valid(fg) {
				fg = c.style.FindColor(fg) // attempt to find the color from comm.Term colors
			}

			if color.Valid(bg) {
				bg = c.style.FindColor(bg) // same as above
			}

			if color.Valid(fg) && color.Valid(bg) && c.canSetBgFg {
				c.comm.WriteBothColors(w, fg, bg, true)
				goto colorDone
			}

			if color.Valid(fg) && c.canSetFg {
				c.comm.WriteColor(w, fg, true, true)
			}

			if color.Valid(bg) && c.canSetBg {
				c.comm.WriteColor(w, bg, false, true)
			}
		}

	colorDone:

		if attrs&style.Bold != 0 {
			c.comm.PutBold(w)
		}
		if attrs&style.Underline != 0 {
			c.comm.PutUnderline(w)
		}
		if attrs&style.Reverse != 0 {
			c.comm.PutReverse(w)
		}
		if attrs&style.Blink != 0 {
			c.comm.PutBlink(w)
		}
		if attrs&style.Dim != 0 {
			c.comm.PutDim(w)
		}
		if attrs&style.Italic != 0 {
			c.comm.PutItalic(w)
		}
		if attrs&style.StrikeThrough != 0 {
			c.comm.PutStrikeThrough(w)
		}

		// cache for speeding up display same pixels (like a bunch of black background with white text)
		c.cachedAttrs = attrs
		c.cachedBG = bg
		c.cachedFG = fg

	cachedStyle:

		runes := make([]byte, 0, 6)
		runes = c.encoder.encodeRune(pixel.Rune(), runes)
		if pixel.HasUnicode() {
			uni := pixel.Unicode()
			for _, r := range *uni {
				runes = c.encoder.encodeRune(r, runes)
			}
		}

		// TODO : implement width checking for too wide to fit or chars not being able to display
		// str := string(runes)
		// if pixel.Width() > 1 && str == "?" {
		// No FullWidth character support
		// str = "? "
		// }
		// if pixel.Position().X > c.size.Width-pixel.Width() {
		// if Debug {
		//	log.Printf("too wide to fit : %d [%d]", pixel.Width(), c.size.Width)
		// }
		// str = " " // too wide to fit; emit a single space instead
		// }

		if _, err := w.Write(runes); err != nil {
			if Debug {
				log.Printf("error writing to io : " + err.Error())
			}
		}
	}
}
