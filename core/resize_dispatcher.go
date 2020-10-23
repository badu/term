package core

import (
	"context"
	"fmt"
	"io"
	"log"

	"github.com/badu/term"
)

// for readability
type channels []chan term.ResizeEvent

// delete removes the element at index from channels. This is fast version, but changes order, therefore we cannot rely on index for register / unregister
func (c *channels) delete(idx int) {
	(*c)[idx] = (*c)[len(*c)-1] // Copy last element to index i.
	(*c)[len(*c)-1] = nil       // Erase last element (write zero value).
	*c = (*c)[:len(*c)-1]       // Truncate slice.
}

// EventResize is sent when the window size changes.
type EventResize struct {
	size term.Size
}

func (e EventResize) Size() term.Size {
	return e.size
}

type readerCtx struct {
	ctx      context.Context
	r        io.Reader
	mouseCh  chan []byte
	keyCh    chan []byte
	hasMouse bool
}

type ioret struct {
	n   int
	err error
}

func (r *readerCtx) Read(_ []byte) (int, error) {
	inBuf := make([]byte, 128)

	c := make(chan ioret, 1)

	go func() {
		n, err := r.r.Read(inBuf)
		c <- ioret{n, err}
		close(c)
	}()

	select {
	case ret := <-c:
		chunkCopy := make([]byte, ret.n)
		copy(chunkCopy, inBuf[:ret.n])
		if r.hasMouse {
			r.mouseCh <- chunkCopy
		}
		r.keyCh <- chunkCopy
		return ret.n, ret.err
	case <-r.ctx.Done():
		return 0, r.ctx.Err()
	}
}

// NewReader gets a context-aware io.Reader.
func NewReader(ctx context.Context, r io.Reader, keyChan, mouseChan chan []byte, hasMouse bool) io.Reader {
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
	go func() {
		reader := NewReader(ctx, c.in, c.keyDispatcher.InChan(), c.mouseDispatcher.InChan(), len(c.info.Mouse) != 0)
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
	}()
	// goroutine for gracefully shutting down
	go func(done <-chan struct{}) {
		<-done // block here until we're done
		if Debug {
			log.Println("[core] init'ing shutdown sequence.")
		}
		c.Lock()
		defer c.Unlock()
		// performing shutdown
		c.resize(0, 0, true) // important : it will cancel pixels listener context
		c.info.PutShowCursor(c.out)
		c.info.PutAttrOff(c.out)
		c.info.PutClear(c.out)
		c.info.PutExitCA(c.out)
		c.info.PutExitKeypad(c.out)
		c.info.PutDisableMouse(c.out)
		if err := c.internalShutdown(); err != nil {
			if Debug {
				log.Printf("[core] internal shutdown error : %v", err)
			}
		}
		fmt.Println(c.info.Clear) // clears the screen after shutdown
		if Debug {
			log.Println("[core] shutdown complete")
		}
		// order matters, otherwise the finalizer won't get called
		if c.finalizer != nil {
			c.finalizer()
		}

		c.died <- struct{}{} // notifying our death to a dispatcher (which listens in register)
	}(ctx.Done())
	// goroutine for watching size changes
	go func(done <-chan struct{}) {
		for {
			select {
			case <-ctx.Done():
				return
			case <-c.winSizeCh:
				c.Lock()
				w, h, err := c.readWinSize() // read new width and height information
				if err != nil {
					if Debug {
						log.Printf("error in win size reader : %v", err)
					}
				}
				c.resize(w, h, false)              // store resize info
				ev := EventResize{size: *c.size}   // create one event for everyone
				for _, cons := range c.receivers { // multiplexing
					cons <- ev // Important note : yes, there is the risk of writing to close channels
				}
				c.Unlock()
			}
		}
	}(ctx.Done())
}

// Register is registering receivers
func (c *core) Register(r term.ResizeListener) {
	c.Lock()
	defer c.Unlock()

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
		// wait for death announcement
		<-r.DyingChan()
		// now lookup for that very channel and forget it
		for idx, ch := range c.receivers {
			// Two channel values are considered equal if they originated from the same make call (meaning they refer to the same channel value in memory).
			if ch == r.ResizeListen() {
				c.receivers.delete(idx)
				break
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