package core

import (
	"bytes"
	"context"
	"errors"
	"image"
	"io"
	"log"
	"os"
	"runtime"
	"sync"
	"unicode/utf8"

	"golang.org/x/text/encoding"
	"golang.org/x/text/transform"

	"github.com/badu/term"
	"github.com/badu/term/color"
	enc "github.com/badu/term/encoding"
	"github.com/badu/term/info"
	"github.com/badu/term/key"
	"github.com/badu/term/mouse"
	"github.com/badu/term/style"

	_ "github.com/badu/term/info/base" // import the stock terminals
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
		c.fallback = make(map[rune]string)
		for k, v := range fallback {
			c.fallback[k] = v
		}
	}
}

// WithTrueColor is a functional option to disable true color, if needed. Just set the trueColor to "disable".
func WithTrueColor(trueColor string) Option {
	return func(c *core) {
		c.hasTrueColor = trueColor != "disable"
	}
}

// core represents a screen backed by a info implementation.
type core struct {
	sync.Mutex                                   // guards other properties
	sync.Once                                    // required for registering lifecycle goroutines exactly once
	size             *term.Size                  //
	info             *info.Term                  // terminal info
	termIOSPrv       *termiosPrivate             // required by internalStart
	in               *os.File                    // input, acquired in internalStart, released in internalShutdown
	out              *os.File                    // output, acquired in internalStart, released in internalShutdown, used for displaying
	died             chan struct{}               // this is a buffered channel of size one
	winSizeCh        chan os.Signal              // listens for resize signals and transforms them into resize events in the dispatcher section
	receivers        channels                    // We need a slice of channels, on which our listeners will receive those events
	finalizer        Finalizer                   // Yes, we have callback and we could reuse it, but we will affect readability doing so
	mouseDispatcher  term.MouseDispatcher        // mouse event dispatcher, exposes via term.Engine interface
	keyDispatcher    term.KeyDispatcher          // key event dispatcher, exposes via term.Engine interface
	encoder          transform.Transformer       // used for encoding runes
	charset          string                      // stores charset for getter
	fallback         map[rune]string             // runes fallback
	altChars         map[rune]string             // alternative runes
	buf              bytes.Buffer                // buffer, works in conjunction with useDrawBuffering (might be removed, due to the nature of pixels)
	colors           map[color.Color]color.Color // TODO : move this to styles ?
	palette          []color.Color               // TODO : move this to styles ?
	cursorPosition   *image.Point                // the position of the cursor, if visible
	pixCancel        func()                      // allows cancellation of listening to pixels changes
	hasTrueColor     bool                        // as the name says
	useDrawBuffering bool                        // true if we are collecting writes to buf instead of sending directly to out
}

// NewCore returns a Engine that uses the stock TTY interface and POSIX termios, combined with a info description taken from the $TERM environment variable.
// It returns an error if the terminal is not supported for any reason.
// For terminals that do not support dynamic resize events, the $LINES $COLUMNS environment variables can be set to the actual window size, otherwise defaults taken from the terminal database are used.
func NewCore(termEnv string, options ...Option) (term.Engine, error) {
	ti, err := info.LookupTerminfo(termEnv)
	if err != nil {
		ti, err = loadDynamicTerminfo(termEnv)
		if err != nil {
			return nil, err
		}
		info.AddTerminfo(ti)
	}

	hasTrueColor := false
	if len(ti.SetFgBgRGB) > 0 || len(ti.SetFgRGB) > 0 || len(ti.SetBgRGB) > 0 {
		hasTrueColor = true
	}

	res := &core{
		info:         ti,                                     // terminal info
		died:         make(chan struct{}),                    // init of died channel, a buffered channel of exactly one
		receivers:    make(channels, 0),                      // init the receivers slice of channels which will register themselves for resizing events
		winSizeCh:    make(chan os.Signal, runtime.NumCPU()), // listening resize events (OS specific)
		colors:       make(map[color.Color]color.Color),
		palette:      make([]color.Color, ti.Colors),
		charset:      getCharset(),
		hasTrueColor: hasTrueColor,
	}

	if e := enc.GetEncoding(res.charset); e != nil {
		res.encoder = e.NewEncoder()
	} else {
		return nil, ErrNoCharset
	}

	buildAlternateRunesMap(res)
	defaultRunesFallback(res)

	for i := 0; i < ti.Colors; i++ {
		res.palette[i] = color.Color(i) | color.Valid
		res.colors[color.Color(i)|color.Valid] = color.Color(i) | color.Valid // identity map for our builtin colors
	}

	for _, o := range options {
		o(res)
	}

	finalizer := func() {
		if Debug {
			log.Println("[core] : component died")
		}
	}
	// creating dispatchers for key and mouse
	res.mouseDispatcher, err = mouse.NewEventDispatcher(mouse.WithTerminalInfo(ti), mouse.WithResizeDispatcher(res), mouse.WithFinalizer(finalizer))
	if err != nil {
		if Debug {
			log.Printf("error creating mouse dispatcher : %v", err)
		}
		return nil, err
	}

	res.keyDispatcher, err = key.NewEventDispatcher(key.WithTerminalInfo(ti), key.WithFinalizer(finalizer))
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

		c.info.Init(c.out, c.size)

		c.lifeCycle(ctx) // mounting context cancel listener
		c.keyDispatcher.LifeCycle(ctx)
		if len(c.info.Mouse) > 0 { // if we have mouse support
			c.mouseDispatcher.LifeCycle(ctx, c.out)
			c.mouseDispatcher.Enable()
		}
		if Debug {
			log.Println("[START] multiplexer mounted.")
		}

		c.info.PutEnterCA(c.out)
		c.info.PutHideCursor(c.out)
		c.info.GoTo(c.out, c.size.Width, c.size.Height) // put cursor outside screen
		c.info.PutEnableAcs(c.out)
		c.info.PutClear(c.out)

		ev := EventResize{size: *c.size}   // create one event for everyone
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
	c.Lock()
	defer c.Unlock()

	if enco := c.encoder; enco != nil {
		nb := make([]byte, 6)
		ob := make([]byte, 6)
		num := utf8.EncodeRune(ob, r)

		enco.Reset()
		dst, _, err := enco.Transform(nb, ob[:num], true)
		if dst != 0 && err == nil && nb[0] != '\x1A' {
			return true
		}
	}
	// Terminal fallbacks always permitted, since we assume they are basically nearly perfect renditions.
	if _, ok := c.altChars[r]; ok {
		return true
	}
	if !checkFallbacks {
		return false
	}
	if _, ok := c.fallback[r]; ok {
		return true
	}
	return false
}

// CharacterSet exposes current charset, implementation for term.Engine interface
func (c *core) CharacterSet() string {
	c.Lock()
	defer c.Unlock()

	return c.charset
}

// SetRuneFallback replaces a rune with a fallback
func (c *core) SetRuneFallback(orig rune, fallback string) {
	c.Lock()
	defer c.Unlock()

	c.fallback[orig] = fallback
}

// UnsetRuneFallback forgets a replaced rune fallback
func (c *core) UnsetRuneFallback(orig rune) {
	c.Lock()
	defer c.Unlock()

	delete(c.fallback, orig)
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
	return c.info.Colors
}

// Term returns the underlying term
func (c *core) Term() *info.Term {
	c.Lock()
	defer c.Unlock()

	return c.info
}

// encodeRune appends a buffer with encoded runes
func (c *core) encodeRune(r rune, buf []byte) []byte {
	nb := make([]byte, 6)
	ob := make([]byte, 6)
	num := utf8.EncodeRune(ob, r)
	ob = ob[:num]
	dst := 0
	var err error
	if enco := c.encoder; enco != nil {
		enco.Reset()
		dst, _, err = enco.Transform(nb, ob, true)
	}
	if err != nil || dst == 0 || nb[0] == encoding.ASCIISub {
		// Combining characters are elided
		if len(buf) == 0 {
			if acs, ok := c.altChars[r]; ok {
				buf = append(buf, []byte(acs)...)
			} else if fb, ok := c.fallback[r]; ok {
				buf = append(buf, []byte(fb)...)
			} else {
				buf = append(buf, '?')
			}
		}
	} else {
		buf = append(buf, nb[:dst]...)
	}

	return buf
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

// Palette
func (c *core) Palette() []color.Color {
	c.Lock()
	defer c.Unlock()

	return c.palette
}

// Colors
func (c *core) Colors() map[color.Color]color.Color {
	c.Lock()
	defer c.Unlock()

	return c.colors
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
		go func(pix term.PixelGetter) {
			for msg := range pix.DrawCh() { // listen incoming messages over the pixel draw request channel
				if msg.Position().X == -1 && msg.Position().Y == -1 { // check if this is the cancellation pixel, we're exiting the goroutine
					return
				}
				go func(p term.PixelGetter) { // running in a separate goroutine, because it blocks reading new messages
					c.Lock()
					c.drawPixel(p)
					c.Unlock()
				}(msg)
			}
		}(pixel)
		// for each pixel, mounting a kill switch, which will write the shutdown message when the context is done
		go func(pixCh chan term.PixelGetter, done <-chan struct{}, shutdownPix term.PixelGetter) {
			<-done               // blocking wait for done
			pixCh <- shutdownPix // write shutdownPixel pixel, so go routine above exits
		}(pixel.DrawCh(), ctx.Done(), shutdownPixel) // observe that all parameters are passed to the goroutine
	}
	c.Unlock() // redraw locks it again
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

	c.useDrawBuffering = true // we use buffering, since we're redrawing everything
	c.drawPixel(cells...)
	c.useDrawBuffering = false // finished buffering

	if _, err := c.buf.WriteTo(c.out); err != nil { // writing buffer content to out
		if Debug {
			log.Printf("error writing to out : " + err.Error())
		}
	}
	c.buf.Reset() // reset the buffer for the next time
}

// Cursor returns the current cursor position
func (c *core) Cursor() *image.Point {
	c.Lock()
	defer c.Unlock()

	return c.cursorPosition
}

// ShowCursor displays the cursor at the indicated coordinates
func (c *core) ShowCursor(where *image.Point) {
	c.Lock()
	defer c.Unlock()

	if where.X < 0 || where.Y < 0 || where.X >= c.size.Width || where.Y >= c.size.Height {
		// does not update cursor position
		if len(c.info.HideCursor) > 0 {
			c.info.PutHideCursor(c.out)
		}
		// No way to hide cursor, stick it at bottom right of screen
		c.info.GoTo(c.out, c.size.Width, c.size.Height) // Y is column, X is row
		return
	}
	c.cursorPosition = where
	c.info.GoTo(c.out, c.cursorPosition.Y, c.cursorPosition.X) // Y is column, X is row
	c.info.PutShowCursor(c.out)
}

// HideCursor hides the cursor from the screen
func (c *core) HideCursor() {
	c.Lock()
	defer c.Unlock()

	c.cursorPosition = nil
	// does not update cursor position
	if len(c.info.HideCursor) > 0 {
		c.info.PutHideCursor(c.out)
	}
	// No way to hide cursor, stick it at bottom right of screen
	c.info.GoTo(c.out, c.size.Width, c.size.Height) // Y is column, X is row
}

// Clear
func (c *core) Clear() {
	c.Lock()
	defer c.Unlock()
	c.info.PutClear(c.out)
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
}

// drawPixel - locked inside caller function
func (c *core) drawPixel(pixels ...term.PixelGetter) {
	var w io.Writer
	if c.useDrawBuffering {
		w = &c.buf
	} else {
		w = c.out
	}
	for _, pixel := range pixels {
		c.info.GoTo(w, pixel.Position().Y, pixel.Position().X) // Y is column, X is row

		c.info.PutAttrOff(w)

		if c.info.Colors != 0 {
			// putting background and foreground colors
			fg := pixel.FgCol()
			bg := pixel.BgCol()
			if fg == color.Reset || bg == color.Reset {
				c.info.PutResetFgBg(w)
				goto colorDone // TODO : check correct?
			}

			if c.hasTrueColor {
				if len(c.info.SetFgRGB) > 0 { // we can use SetFgRGB
					if fg.IsRGB() && bg.IsRGB() { // both are RGB
						c.info.WriteTrueColors(w, fg, bg)
						goto colorDone
					}
					// not both are RGB
					if fg.IsRGB() {
						c.info.WriteColor(w, fg, true)
						fg = color.Default
					}

					if bg.IsRGB() {
						c.info.WriteColor(w, bg, false)
						bg = color.Default
					}
				}
			}

			if fg.Valid() {
				if v, ok := c.colors[fg]; ok {
					fg = v
				} else {
					v = color.FindColor(fg, c.palette)
					c.colors[fg] = v
					fg = v
				}
			}

			if bg.Valid() {
				if v, ok := c.colors[bg]; ok {
					bg = v
				} else {
					v = color.FindColor(bg, c.palette)
					c.colors[bg] = v
					bg = v
				}
			}

			if fg.Valid() && bg.Valid() && len(c.info.SetFgBg) > 0 {
				def := c.info.TParam(c.info.SetFgBg, int(fg&0xff), int(bg&0xff))
				if err := c.info.WriteString(w, def); err != nil {
					if Debug {
						log.Printf("[core-sendFgBg] error writing string : %v", err)
					}
				}
				goto colorDone
			}

			if fg.Valid() && len(c.info.SetFg) > 0 {
				def := c.info.TParam(c.info.SetFg, int(fg&0xff))
				if err := c.info.WriteString(w, def); err != nil {
					if Debug {
						log.Printf("[core-sendFgBg] error writing string : %v", err)
					}
				}
			}

			if bg.Valid() && len(c.info.SetBg) > 0 {
				def := c.info.TParam(c.info.SetBg, int(bg&0xff))
				if err := c.info.WriteString(w, def); err != nil {
					if Debug {
						log.Printf("[core-sendFgBg] error writing string : %v", err)
					}
				}
			}
		}
	colorDone:

		if pixel.Attrs() != style.None { // it's not just *normal* text
			attrs := pixel.Attrs()
			if attrs&style.Bold != 0 {
				c.info.PutBold(w)
			}
			if attrs&style.Underline != 0 {
				c.info.PutUnderline(w)
			}
			if attrs&style.Reverse != 0 {
				c.info.PutReverse(w)
			}
			if attrs&style.Blink != 0 {
				c.info.PutBlink(w)
			}
			if attrs&style.Dim != 0 {
				c.info.PutDim(w)
			}
			if attrs&style.Italic != 0 {
				c.info.PutItalic(w)
			}
			if attrs&style.StrikeThrough != 0 {
				c.info.PutStrikeThrough(w)
			}
		}

		buf := make([]byte, 0, 6)
		buf = c.encodeRune(pixel.Rune(), buf)
		if pixel.HasUnicode() {
			uni := pixel.Unicode()
			for _, r := range *uni {
				buf = c.encodeRune(r, buf)
			}
		}

		// TODO : implement width checking
		// str := string(buf)
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

		if c.useDrawBuffering {
			if _, err := c.buf.Write(buf); err != nil {
				if Debug {
					log.Printf("error writing to io : " + err.Error())
				}
			}
		} else {
			if _, err := c.out.Write(buf); err != nil {
				if Debug {
					log.Printf("error writing to io : " + err.Error())
				}
			}
		}
	}

	if c.cursorPosition == nil { // check if we were displaying cursor
		if len(c.info.HideCursor) > 0 {
			c.info.PutHideCursor(c.out)
		}
		// There is no way to hide cursor, put it at bottom right of screen
		c.info.GoTo(c.out, c.size.Width, c.size.Height) // Y is column, X is row
		return
	}
	// ok, we had a cursor before : restore cursor position
	c.info.GoTo(c.out, c.cursorPosition.Y, c.cursorPosition.X) // Y is column, X is row
	c.info.PutShowCursor(c.out)
}
