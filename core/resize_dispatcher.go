package core

import (
	"context"
	"io"
	"log"
	"os"

	"github.com/badu/term"
)

// for readability
type channels []chan term.ResizeEvent

// delete removes the element at index from channels.
// Note that this is the fastest version, which changes order of elements inside the slice
// Yes, this is repeated code, because avoiding use of interface{}
func (c *channels) delete(idx int) {
	(*c)[idx] = (*c)[len(*c)-1] // Copy last element to index i.
	(*c)[len(*c)-1] = nil       // Erase last element (write zero value).
	*c = (*c)[:len(*c)-1]       // Truncate slice.
}

// EventResize is sent when the window size changes.
type EventResize struct {
	size *term.Size
}

// NewResizeEvent
func NewResizeEvent(cols, rows int) *EventResize {
	return &EventResize{size: &term.Size{Columns: cols, Rows: rows}}
}

// Size
func (e *EventResize) Size() *term.Size {
	return e.size
}

// readerCtx
type readerCtx struct {
	ctx      context.Context
	r        io.Reader
	mouseCh  chan []byte
	keyCh    chan []byte
	hasMouse bool
}

// ioret
type ioret struct {
	n   int
	err error
}

// Read
func (r *readerCtx) Read(_ []byte) (int, error) {
	inBuf := make([]byte, 128)
	ret := make(chan ioret, 1)

	go func() {
		n, err := r.r.Read(inBuf)
		ret <- ioret{n, err}
		close(ret)
	}()
	// blocking wait for one of the channels (either we have reads or context cancellation)
	select {
	case ret := <-ret:
		if r.hasMouse {
			r.mouseCh <- inBuf[:ret.n]
		}
		r.keyCh <- inBuf[:ret.n]
		return ret.n, ret.err
	case <-r.ctx.Done():
		return 0, r.ctx.Err()
	}
}

// newContextReader gets a context-aware io.Reader.
func newContextReader(ctx context.Context, r io.Reader, keyChan, mouseChan chan []byte, hasMouse bool) io.Reader {
	return &readerCtx{
		ctx:      ctx,
		r:        r,
		mouseCh:  mouseChan,
		keyCh:    keyChan,
		hasMouse: hasMouse,
	}
}

// lifeCycle all lifecycle goroutines
func (c *core) lifeCycle(ctx context.Context) {
	// goroutine for listening inputs and distribute them to listeners
	go func(cx context.Context) {
		reader := newContextReader(cx, c.in, c.keyDispatcher.InChan(), c.mouseDispatcher.InChan(), c.comm.HasMouse)
		for {
			// by default we just listen whatever comes
			_, err := reader.Read(nil)
			switch err {
			case io.EOF, nil: // ok
			case context.Canceled:
				if Debug {
					log.Println("context cancelled : reader no longer reads.")
				}
				return // probably killed by internalShutdown, so we exit
			default:
				if Debug {
					log.Printf("[core] read error has occurred : %v", err)
				}
				return
			}
		}
	}(ctx)
	// goroutine for gracefully shutting down
	go func(cx context.Context) {
		<-cx.Done() // block here until we're done
		if Debug {
			log.Println("[core] init'ing shutdown sequence.")
		}
		c.Lock()
		defer c.Unlock()
		// performing shutdown
		c.resize(0, 0, true) // important : it will cancel pixels listener context
		c.comm.PutShowCursor(c.out)
		c.comm.PutAttrOff(c.out)
		c.comm.PutClear(c.out)
		c.comm.PutExitCA(c.out)
		c.comm.PutExitKeypad(c.out)
		c.comm.PutDisableMouse(c.out)
		if err := c.internalShutdown(); err != nil {
			if Debug {
				log.Printf("[core] internal shutdown error : %v", err)
			}
		}
		c.comm.PutClear(os.Stdout) // clears the terminal screen after shutdown
		if Debug {
			log.Println("[core] shutdown complete")
		}
		// order matters, otherwise the finalizer won't get called
		if c.finalizer != nil {
			c.finalizer()
		}
		close(c.died) // notifying our death to a dispatcher (which listens in register)
	}(ctx)
	// goroutine for watching size changes
	go func(cx context.Context) {
		for {
			select {
			case <-cx.Done():
				if Debug {
					log.Println("[core] context done - exiting resize listener")
				}
				return
			case enable := <-c.mouseSwitch:
				if enable {
					c.comm.PutEnableMouse(c.out)
				} else {
					c.comm.PutDisableMouse(c.out)
				}
			case <-c.winSizeCh:
				c.Lock()
				w, h, err := c.readWinSize() // read new width and height information
				if err != nil {
					if Debug {
						log.Printf("error in win size reader : %v", err)
					}
				}
				c.resize(w, h, false)              // store resize comm
				ev := &EventResize{size: c.size}   // create one event for everyone
				for _, cons := range c.receivers { // multiplexing
					cons <- ev // Important note : yes, there is the risk of writing to close channels
				}
				c.Unlock()
			}
		}
	}(ctx)
}

// Register is registering receivers
func (c *core) Register(r term.ResizeListener) {
	c.Lock()
	defer c.Unlock()
	if c.ctx == nil {
		if Debug {
			log.Fatal("context not set : cannot listen context.Done()")
		}
		return
	}
	// check against double registration
	alreadyRegistered := false
	for _, ch := range c.receivers {
		// Two channel values are considered equal if they originated from the same make call (meaning they refer to the same channel value in memory).
		if ch == r.ResizeListen() {
			alreadyRegistered = true
			break
		}
	}
	if alreadyRegistered {
		if Debug {
			log.Fatal("warning : ResizeListen chan is nil")
		}
		return
	}
	if r.ResizeListen() == nil {
		if Debug {
			log.Fatal("error : ResizeListen chan is nil")
		}
		return
	}
	// we're fine, lets register it
	c.receivers = append(c.receivers, r.ResizeListen())
	// mounting a go routine to listen bye-bye when the listener's context get cancelled
	go func() {
		select {
		case <-c.ctx.Done():
			if Debug {
				log.Println("[core] context is done. Existing death listening routine in Register")
			}
			return
		case <-r.DyingChan():
			// now lookup for that very channel and forget it
			for idx, ch := range c.receivers {
				// Two channel values are considered equal if they originated from the same make call (meaning they refer to the same channel value in memory).
				if ch == r.ResizeListen() {
					c.receivers.delete(idx)
					break
				}
			}
		}
	}()
}

// DyingChan implements the term.Engine interface, listened in composer for waiting graceful shutdown
func (c *core) DyingChan() chan struct{} {
	c.Lock()
	defer c.Unlock()

	return c.died
}
