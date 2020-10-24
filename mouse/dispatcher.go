package mouse

import (
	"bytes"
	"context"
	"errors"
	"log"
	"os"
	"sync"

	"github.com/badu/term"
	"github.com/badu/term/info"
	"github.com/badu/term/key"
)

// for readability
type channels []chan term.MouseEvent

// delete removes the element at index from channels.
// Note that this is the fastest version, which changes order of elements inside the slice
// Yes, this is repeated code, because avoiding use of interface{}
func (c *channels) delete(idx int) {
	(*c)[idx] = (*c)[len(*c)-1] // Copy last element to index i.
	(*c)[len(*c)-1] = nil       // Erase last element (write zero value).
	*c = (*c)[:len(*c)-1]       // Truncate slice.
}

type Finalizer func()

type Option func(core *eventDispatcher)

// WithTerminalInfo is mandatory for constructor, used by core
func WithTerminalInfo(ti *info.Term) Option {
	return func(d *eventDispatcher) {
		d.info = ti
	}
}

// WithFinalizer provides a way of calling a function upon dispatcher death
func WithFinalizer(c Finalizer) Option {
	return func(l *eventDispatcher) {
		l.finalizer = c
	}
}

// WithResizeDispatcher core self register for resize events
func WithResizeDispatcher(e term.ResizeDispatcher) Option {
	return func(l *eventDispatcher) {
		e.Register(l)
	}
}

// eventDispatcher is an implementation of mouse MouseDispatcher.
type eventDispatcher struct {
	sync.Mutex                       // guards other properties
	sync.Once                        // required for registering lifecycle goroutines exactly once
	info       *info.Term            //
	size       *term.Size            //
	wasBtn     bool                  //
	buttonDn   bool                  //
	inputCh    chan []byte           // channel for listening core.Engine inputs *os.File
	resizeCh   chan term.ResizeEvent // channel for listening resize events, so we can clip our coordinates
	died       chan struct{}         // this is a buffered channel of size one
	receivers  channels              // We need a slice of channels, on which our listeners will receive those events
	finalizer  Finalizer             // Yes, we have callback and we could reuse it, but we will affect readability doing so
	out        *os.File              // needed to pass hide/show mouse commands
}

// NewEventDispatcher ignites dispatcher and check for terminal info if mouse is supported.
// Also, any registration of term.ResizeEvent can be done via functional option.
func NewEventDispatcher(options ...Option) (term.MouseDispatcher, error) {
	res := &eventDispatcher{
		died:      make(chan struct{}),         // init of died channel, a buffered channel of exactly one
		receivers: make(channels, 0),           // channels that will receive any event is produced here
		inputCh:   make(chan []byte),           // channel for listening input, so we can build events
		resizeCh:  make(chan term.ResizeEvent), // channel for listening resize events, so we can clip mouse coordinates
	}

	for _, o := range options {
		o(res)
	}

	if res.info == nil {
		return nil, errors.New("creator needs to provide terminal info (hint : use WithTerminalInfo option)")
	}

	res.size = &term.Size{
		Width:  res.info.Width,
		Height: res.info.Height,
	}
	log.Printf("construct mouse dispatcher width = %03d, mouse height = %03d", res.size.Width, res.size.Height)
	if len(res.info.Mouse) == 0 {
		return nil, errors.New("terminal info reports NO mouse support (hint : creator should not enable mouse or create dispatcher)")
	}

	return res, nil
}

// LifeCycle implementation of term.MouseDispatcher interface, called from core
func (e *eventDispatcher) LifeCycle(ctx context.Context, out *os.File) {
	e.out = out
	// mount lifecycle - listens for chunks of []byte coming via inputCh, analyses them and builds mouse events
	e.lifeCycle(ctx)
}

// DyingChan implementation of term.Death interface, listened in core for waiting graceful shutdown
func (e *eventDispatcher) DyingChan() chan struct{} {
	return e.died
}

// InChan implementation of term.InputListener is used by core to send input from *os.File
func (e *eventDispatcher) InChan() chan []byte {
	return e.inputCh
}

// Register - implementation of term.MouseDispatcher interface - is registering receivers
func (e *eventDispatcher) Register(r term.MouseListener) {
	// check against double registration
	alreadyRegistered := false
	for _, ch := range e.receivers {
		// Two channel values are considered equal if they originated from the same make call (meaning they refer to the same channel value in memory).
		if ch == r.MouseListen() {
			alreadyRegistered = true
			break
		}
	}
	if alreadyRegistered {
		return
	}
	// we're fine, lets register it
	e.receivers = append(e.receivers, r.MouseListen())
	// mounting a go routine to listen bye-bye life
	go func() {
		// wait for death announcement
		<-r.DyingChan()
		// now lookup for that very channel and forget it
		for idx, ch := range e.receivers {
			// Two channel values are considered equal if they originated from the same make call (meaning they refer to the same channel value in memory).
			if ch == r.MouseListen() {
				e.receivers.delete(idx)
				break
			}
		}
	}()
}

// ResizeListen provides the channel for listening resize events
func (e *eventDispatcher) ResizeListen() chan term.ResizeEvent {
	return e.resizeCh
}

func (e *eventDispatcher) Enable() {
	e.info.PutEnableMouse(e.out)
}

func (e *eventDispatcher) Disable() {
	e.info.PutDisableMouse(e.out)
}

func (e *eventDispatcher) HasMouse() bool {
	return len(e.info.Mouse) != 0
}

func clip(x, y, w, h int) (int, int) {
	if x < 0 {
		x = 0
	}
	if y < 0 {
		y = 0
	}
	if x > w-1 {
		x = w - 1
	}
	if y > h-1 {
		y = h - 1
	}
	return x, y
}

// buildMouseEvent returns an event based on the supplied coordinates and button state.
// Note that the screen's mouse button state is updated based on the input to this function (i.e. it mutates the receiver).
func (e *eventDispatcher) buildMouseEvent(x, y, btn int) {
	// XTerm mouse events only report at most one button at a time, which may include a wheel button.
	// Wheel motion events are reported as single impulses, while other button events are reported as separate press & release events.
	button := ButtonNone
	mod := key.ModNone

	// Mouse wheel has bit 6 set, no release events.
	// It should be noted that wheel events are sometimes misdelivered as mouse button events during a click-drag, so we debounce these, considering them to be button press events unless we see an intervening release event.
	switch btn & 0x43 {
	case 0:
		button = Button1
		e.wasBtn = true
	case 1:
		button = Button3 // Note we prefer to treat right as button 2
		e.wasBtn = true
	case 2:
		button = Button2 // And the middle button as button 3
		e.wasBtn = true
	case 3:
		button = ButtonNone
		e.wasBtn = false
	case 0x40:
		if !e.wasBtn {
			button = WheelUp
		} else {
			button = Button1
		}
	case 0x41:
		if !e.wasBtn {
			button = WheelDown
		} else {
			button = Button2
		}
	}

	if btn&0x4 != 0 {
		mod |= key.ModShift
	}
	if btn&0x8 != 0 {
		mod |= key.ModAlt
	}
	if btn&0x10 != 0 {
		mod |= key.ModCtrl
	}

	// Some terminals will report mouse coordinates outside the screen, especially with click-drag events.
	// Clip the coordinates to the screen in that case.
	x, y = clip(x, y, e.size.Width, e.size.Height)
	ev := NewEvent(x, y, button, mod) // one event for everyone
	// send term.MouseEvent it to receivers
	for _, cons := range e.receivers {
		cons <- ev
	}
}

// scanInput reads input via channel
func (e *eventDispatcher) scanInput(buf *bytes.Buffer) error {
	e.Lock()
	defer e.Unlock()

	for {
		if len(buf.Bytes()) == 0 {
			buf.Reset()
			break
		}

		// mouse support already checked in the parent (... and constructor)
		if isComplete, err := e.readXTerm(buf); err != nil {
			if Debug {
				log.Printf("error reading mouse input xterm : %v", err)
			}
		} else if isComplete {
			continue
		}

		if isComplete, err := e.readSGR(buf); err != nil {
			if Debug {
				log.Printf("error reading mouse input xterm : %v", err)
			}
		} else if isComplete {
			continue
		}

		// well we have some partial data, wait until we get some more
		break
	}
	return nil
}

// readSGR attempts to locate an SGR mouse record at the start of the buffer.
// It returns true, true if it found one, and the associated bytes be removed from the buffer.
// It returns true, false if the buffer might contain such an event, but more bytes are necessary (partial match), and false, false if the content is definitely *not* an SGR mouse record.
func (e *eventDispatcher) readSGR(buf *bytes.Buffer) (bool, error) {
	b := buf.Bytes()

	var x, y, btn, state int
	dig := false
	neg := false
	motion := false
	i := 0
	val := 0

	for i = range b {
		switch b[i] {
		case '\x1b':
			if state != 0 {
				return false, nil
			}
			state = 1

		case '\x9b':
			if state != 0 {
				return false, nil
			}
			state = 2

		case '[':
			if state != 1 {
				return false, nil
			}
			state = 2

		case '<':
			if state != 2 {
				return false, nil
			}
			val = 0
			dig = false
			neg = false
			state = 3

		case '-':
			if state != 3 && state != 4 && state != 5 {
				return false, nil
			}
			if dig || neg {
				return false, nil
			}
			neg = true // stay in state

		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
			if state != 3 && state != 4 && state != 5 {
				return false, nil
			}
			val *= 10
			val += int(b[i] - '0')
			dig = true // stay in state

		case ';':
			if neg {
				val = -val
			}
			switch state {
			case 3:
				btn, val = val, 0
				neg, dig, state = false, false, 4
			case 4:
				x, val = val-1, 0
				neg, dig, state = false, false, 5
			default:
				return false, nil
			}

		case 'm', 'M':
			if state != 5 {
				return false, nil
			}
			if neg {
				val = -val
			}
			y = val - 1

			motion = (btn & 32) != 0
			btn &^= 32
			if b[i] == 'm' {
				// mouse release, clear all buttons
				btn |= 3
				btn &^= 0x40
				e.buttonDn = false
			} else if motion {
				// Some broken terminals appear to send mouse button one motion events, instead of encoding 35 (no buttons) into these events.
				// We resolve these by looking for a non-motion event first.
				if !e.buttonDn {
					btn |= 3
					btn &^= 0x40
				}
			} else {
				e.buttonDn = true
			}
			// consume the event bytes
			for i >= 0 {
				if _, err := buf.ReadByte(); err != nil {
					return false, err
				}
				i--
			}
			e.buildMouseEvent(x, y, btn)
			return true, nil
		}
	}

	// incomplete & inconclusive at this point
	return false, nil
}

// readXTerm is like readSGR, but it parses a legacy X11 mouse record.
func (e *eventDispatcher) readXTerm(buf *bytes.Buffer) (bool, error) {
	b := buf.Bytes()

	state := 0
	btn := 0
	x := 0
	y := 0
	for i := range b {
		switch state {
		case 0:
			switch b[i] {
			case '\x1b':
				state = 1
			case '\x9b':
				state = 2
			default:
				return false, nil
			}
		case 1:
			if b[i] != '[' {
				return false, nil
			}
			state = 2
		case 2:
			if b[i] != 'M' {
				return false, nil
			}
			state++
		case 3:
			btn = int(b[i])
			state++
		case 4:
			x = int(b[i]) - 32 - 1
			state++
		case 5:
			y = int(b[i]) - 32 - 1
			for i >= 0 {
				if _, err := buf.ReadByte(); err != nil {
					return false, err
				}
				i--
			}
			e.buildMouseEvent(x, y, btn)
			return true, nil
		}
	}
	return false, nil
}

// lifeCycle listens for context done or incoming input from *os.File
func (e *eventDispatcher) lifeCycle(ctx context.Context) {
	e.Once.Do(
		func() {
			// input listener
			go func() {
				buf := &bytes.Buffer{}
				for {
					select {
					case <-ctx.Done():
						e.endLifeCycle()
						return
					case ev := <-e.resizeCh:
						e.size = ev.Size()
						if Debug {
							log.Printf("resized : cols : %d lines : %d", e.size.Width, e.size.Height)
						}
					case chunk := <-e.inputCh:
						buf.Write(chunk)
						if err := e.scanInput(buf); err != nil {
							if Debug {
								log.Printf("error scanning input : %v", err)
							}
						}
					}
				}
			}()
		},
	)
}

// endOfLifecycle attempt to complete the lifecycle by shutting down gracefully
func (e *eventDispatcher) endLifeCycle() {
	// order matters, otherwise the finalizer won't get called
	if e.finalizer != nil {
		e.finalizer()
	}
	// notifying our death to a dispatcher (which listens in register)
	e.died <- struct{}{}
}
