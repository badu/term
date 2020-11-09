package key

import (
	"bytes"
	"context"
	"log"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/badu/term"
	enc "github.com/badu/term/encoding"
	"github.com/badu/term/info"
	"golang.org/x/text/transform"
)

const (
	defaultDuration = time.Millisecond * 50
)

// Option
type Option func(core *eventDispatcher)

// Finalizer
type Finalizer func()

// for readability
type channels []chan term.KeyEvent

// delete removes the element at index from channels.
// Note that this is the fastest version, which changes order of elements inside the slice
// Yes, this is repeated code, because avoiding use of interface{}
func (c *channels) delete(idx int) {
	(*c)[idx] = (*c)[len(*c)-1] // Copy last element to index i.
	(*c)[len(*c)-1] = nil       // Erase last element (write zero value).
	*c = (*c)[:len(*c)-1]       // Truncate slice.
}

// eventDispatcher
type eventDispatcher struct {
	sync.Mutex                             // guards other properties
	sync.Once                              // required for registering lifecycle goroutines exactly once
	inputCh          chan []byte           // channel for listening core.Engine inputs *os.File
	keyExist         map[term.Key]struct{} //
	keyCodes         map[string]*Code      //
	keyTimer         *time.Timer           //
	keyTimerDuration time.Duration         //
	keyExpire        time.Time             //
	decoder          transform.Transformer //
	died             chan struct{}         // this is a buffered channel of size one
	receivers        channels              // a slice of channels, on which our listeners will receive those events
	finalizer        Finalizer             // if a finalizer is provided, it will be called before shutdown
	ctx              context.Context       //
	escaped          bool                  //
}

// WithFinalizer provides a way of calling a function upon dispatcher death
func WithFinalizer(c Finalizer) Option {
	return func(l *eventDispatcher) {
		l.finalizer = c
	}
}

// WithKeyTimerInterval is a functional option to set the interval between key reads, for those of you brave enough to go slower or faster. Default is 50 milliseconds.
func WithKeyTimerInterval(duration time.Duration) Option {
	return func(c *eventDispatcher) {
		c.keyTimer = time.NewTimer(duration)
		c.keyTimerDuration = duration
	}
}

// WithTerminalInfo is mandatory for the composition, provided by core
func WithTerminalInfo(ti *info.Term) Option {
	return func(d *eventDispatcher) {
		prepareKeys(d, ti) // prepare "known" keys
	}
}

func NewEventDispatcher(opts ...Option) (term.KeyDispatcher, error) {
	res := &eventDispatcher{
		keyExist:         make(map[term.Key]struct{}),    // holds information about known keys
		keyCodes:         make(map[string]*Code),         // holds information about known keys
		keyTimer:         time.NewTimer(defaultDuration), //
		keyTimerDuration: defaultDuration,                //
		inputCh:          make(chan []byte),              // init of the channel which receives inputs from *os.File
		died:             make(chan struct{}),            // init of died channel, a buffered channel of exactly one
		receivers:        make(channels, 0),
	}

	for _, o := range opts {
		o(res)
	}

	return res, nil
}

// LifeCycle implementation of term.KeyDispatcher, called from core
func (d *eventDispatcher) LifeCycle(ctx context.Context) {
	d.ctx = ctx
	// mount lifecycle - listens for chunks of []byte coming via inputCh, analyses them and builds key events
	d.lifeCycle()
}

// DyingChan implementation of term.Death interface, listened in core for waiting graceful shutdown
func (d *eventDispatcher) DyingChan() chan struct{} {
	return d.died
}

// InChan implementation of term.InputListener is used by core to send input from *os.File
func (d *eventDispatcher) InChan() chan []byte {
	return d.inputCh
}

// Register is registering receivers
func (d *eventDispatcher) Register(r term.KeyListener) {
	if d.ctx == nil {
		if Debug {
			log.Fatal("context not set : cannot listen context.Done()")
		}
		return
	}
	// check against double registration
	alreadyRegistered := false
	for _, ch := range d.receivers {
		// Two channel values are considered equal if they originated from the same make call (meaning they refer to the same channel value in memory).
		if ch == r.KeyListen() {
			alreadyRegistered = true
			break
		}
	}
	if alreadyRegistered {
		return
	}
	// we're fine, lets register it
	d.receivers = append(d.receivers, r.KeyListen())
	// mounting a go routine to listen bye-bye life
	go func() {
		// wait for death announcement
		<-r.DyingChan()
		// now lookup for that very channel and forget it
		for idx, ch := range d.receivers {
			// Two channel values are considered equal if they originated from the same make call (meaning they refer to the same channel value in memory).
			if ch == r.KeyListen() {
				d.receivers.delete(idx)
				break
			}
		}
	}()
}

// HasKey - implementation of term.KeyDispatcher interface
func (d *eventDispatcher) HasKey(k term.Key) bool {
	if k == Rune {
		return true
	}
	_, ok := d.keyExist[k]
	return ok
}

// readFuncKey checks for function key and dispatches event via channels
func (d *eventDispatcher) readFuncKey(buf *bytes.Buffer) (bool, bool, error) {
	b := buf.Bytes()
	partial := false
	for kk, kv := range d.keyCodes {
		esc := []byte(kk)
		if (len(esc) == 1) && (esc[0] == '\x1b') {
			continue
		}
		if bytes.HasPrefix(b, esc) {
			// matched
			var r rune
			if len(esc) == 1 {
				r = rune(b[0])
			}
			mod := kv.Mod
			if d.escaped {
				mod |= ModAlt
				d.escaped = false
			}
			ev := NewEvent(kv.Key, r, mod) // one event for everyone
			for _, cons := range d.receivers {
				cons <- ev
			}
			for i := 0; i < len(esc); i++ {
				if _, err := buf.ReadByte(); err != nil {
					return false, false, err
				}
			}
			return true, true, nil
		}
		if bytes.HasPrefix(esc, b) {
			partial = true
		}
	}
	return partial, false, nil
}

// readRuneKey checks for rune key and dispatches event via channels
func (d *eventDispatcher) readRuneKey(buf *bytes.Buffer) (bool, bool, error) {
	b := buf.Bytes()
	if b[0] >= enc.Space && b[0] <= 0x7F {
		// printable ASCII easy to deal with -- no encodings
		mod := ModNone
		if d.escaped {
			mod = ModAlt
			d.escaped = false
		}
		ev := NewEvent(Rune, rune(b[0]), mod) // one event for everyone
		for _, cons := range d.receivers {
			cons <- ev
		}
		if _, err := buf.ReadByte(); err != nil {
			return false, false, err
		}
		return true, true, nil
	}

	if b[0] < 0x80 {
		// Low numbered values are control keys, not runes.
		return false, false, nil
	}

	utfBytes := make([]byte, 12)
	for l := 1; l <= len(b); l++ {
		d.decoder.Reset()
		nout, nin, err := d.decoder.Transform(utfBytes, b[:l], true)
		if err == transform.ErrShortSrc {
			continue
		}
		if nout != 0 {
			r, _ := utf8.DecodeRune(utfBytes[:nout]) // not interested in rune size
			if r != utf8.RuneError {
				mod := ModNone
				if d.escaped {
					mod = ModAlt
					d.escaped = false
				}
				ev := NewEvent(Rune, r, mod) // one event for everyone
				for _, cons := range d.receivers {
					cons <- ev
				}
			}
			for nin > 0 {
				if _, err := buf.ReadByte(); err != nil {
					return false, false, err
				}
				nin--
			}
			return true, true, nil
		}
	}
	// Looks like potential escape
	return true, false, nil
}

// readSGR attempts to locate an SGR mouse record at the start of the buffer.
// It returns true, true if it found one, and the associated bytes be removed from the buffer.
// It returns true, false if the buffer might contain such an event, but more bytes are necessary (partial match), and false, false if the content is definitely *not* an SGR mouse record.
func (d *eventDispatcher) readSGR(buf *bytes.Buffer) (bool, error) {
	b := buf.Bytes()
	state := 0
	dig := false
	neg := false
	i := 0
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
			dig = true // stay in state

		case ';':
			switch state {
			case 3:
				neg, dig, state = false, false, 4
			case 4:
				neg, dig, state = false, false, 5
			default:
				return false, nil
			}
		case 'm', 'M':
			if state != 5 {
				return false, nil
			}
			// consume the event bytes
			for i >= 0 {
				if _, err := buf.ReadByte(); err != nil {
					return false, err
				}
				i--
			}
			return true, nil
		}
	}
	// incomplete & inconclusive at this point
	return false, nil
}

// readXTerm is like readSGR, but it parses a legacy X11 mouse record.
func (d *eventDispatcher) readXTerm(buf *bytes.Buffer) (bool, error) {
	b := buf.Bytes()
	state := 0
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
			state++
		case 4:
			state++
		case 5:
			for i >= 0 {
				if _, err := buf.ReadByte(); err != nil {
					return false, err
				}
				i--
			}
			return true, nil
		}
	}
	return false, nil
}

// scanInput reads input from *os.File via input channel
func (d *eventDispatcher) scanInput(buf *bytes.Buffer, expire bool) error {
	d.Lock()
	defer d.Unlock()

	for {
		byts := buf.Bytes()
		if len(byts) == 0 {
			buf.Reset()
			break
		}
		// check if it's a mouse event (this triggers false key events)
		if comp, _ := d.readXTerm(buf); comp {
			continue
		}
		if comp, _ := d.readSGR(buf); comp {
			continue
		}
		// now lookup for normal keys
		partials := 0
		part, comp, err := d.readRuneKey(buf)
		if err != nil {
			return err
		}
		if comp {
			continue
		} else if part {
			partials++
		}
		// last lookup for function keys
		part, comp, err = d.readFuncKey(buf)
		if err != nil {
			return err
		}
		if comp {
			continue
		} else if part {
			partials++
		}

		if partials == 0 || expire {
			if byts[0] == '\x1b' {
				if len(byts) == 1 {
					ev := NewEvent(Esc, 0, ModNone) // one event for everyone
					for _, cons := range d.receivers {
						cons <- ev
					}
					d.escaped = false
				} else {
					d.escaped = true
				}
				if _, err := buf.ReadByte(); err != nil {
					return err
				}
				continue
			}
			// Nothing was going to match, or we timed out waiting for more data -- just deliver the characters to the app & let them sort it out.
			// Possibly we should only do this for control characters like ESC.
			by, _ := buf.ReadByte()
			mod := ModNone
			if d.escaped {
				d.escaped = false
				mod = ModAlt
			}
			ev := NewEvent(Rune, rune(by), mod) // one event for everyone
			for _, cons := range d.receivers {
				cons <- ev
			}
			continue
		}
		// well we have some partial data, wait until we get some more
		break
	}
	return nil
}

// lifeCycle starts a number of goroutines, see comments below
func (d *eventDispatcher) lifeCycle() {
	d.Once.Do(
		func() {
			// input listener routine
			go func(cx context.Context) {
				buf := &bytes.Buffer{}
				for {
					select {
					case <-cx.Done(): // context is done, exiting goroutine
						d.keyTimer.Stop()
						if d.finalizer != nil { // order matters, otherwise the finalizer won't get called
							d.finalizer() // call finalizer function
						}
						close(d.died) // notifying our death to a dispatcher (which listens in register)
						return
					case <-d.keyTimer.C:
						// If the timer fired, and the current time is after the expiration of the escape sequence, then we assume the escape sequence reached it's conclusion, and process the chunk independently.
						// This lets us detect conflicts such as a lone ESC.
						if buf.Len() > 0 {
							if time.Now().After(d.keyExpire) {
								if err := d.scanInput(buf, true); err != nil {
									if Debug {
										log.Printf("error scanning input : %v", err)
									}
								}
							}
						}
						if buf.Len() > 0 {
							if !d.keyTimer.Stop() {
								select {
								case <-d.keyTimer.C:
								default:
								}
							}
							d.keyTimer.Reset(d.keyTimerDuration)
						}
					case chunk := <-d.inputCh:
						buf.Write(chunk)
						d.keyExpire = time.Now().Add(d.keyTimerDuration)
						if err := d.scanInput(buf, false); err != nil {
							if Debug {
								log.Printf("error scanning input : %v", err)
							}
						}
						if !d.keyTimer.Stop() {
							select {
							case <-d.keyTimer.C:
							default:
							}
						}
						if buf.Len() > 0 {
							d.keyTimer.Reset(d.keyTimerDuration)
						}
					}
				}
			}(d.ctx)
		},
	)
}
