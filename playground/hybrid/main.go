package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/badu/term"
	"github.com/badu/term/core"
	initLog "github.com/badu/term/log"
)

type listener struct {
	incomingMouse  chan term.MouseEvent  // We need a channel on which we will listen for incoming events
	incomingKey    chan term.KeyEvent    // We need a channel on which we will listen for incoming events
	incomingResize chan term.ResizeEvent // We need a channel on which we will listen for incoming events
	died           chan struct{}         // this is a buffered channel of size one
}

func (r *listener) MouseListen() chan term.MouseEvent {
	return r.incomingMouse
}

func (r *listener) KeyListen() chan term.KeyEvent {
	return r.incomingKey
}

func (r *listener) ResizeListen() chan term.ResizeEvent {
	return r.incomingResize
}

func (r *listener) DyingChan() chan struct{} {
	return r.died
}

// listen listens for incoming events
func (r *listener) lifeCycle(ctx context.Context) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				log.Println("[hybrid] context is done.")
				r.died <- struct{}{}
				return
			case ev := <-r.incomingMouse:
				x, y := ev.Position()
				log.Printf("[hybrid] mouse : %s x=%d y=%d %s", ev.ButtonNames(), x, y, ev.ModName())
			case ev := <-r.incomingKey:
				log.Printf("[hybrid] key : name=%s key=%c modname=%s", ev.Name(), ev.Rune(), ev.ModName())
			case ev := <-r.incomingResize:
				log.Printf("[hybrid] resize : width = %d  height = %d", ev.Size().Columns, ev.Size().Rows)
			}
		}
	}()
}

func NewReceiver(ctx context.Context) *listener {
	receiver := &listener{
		died:           make(chan struct{}),        // init of died channel, a buffered channel of exactly one
		incomingMouse:  make(chan term.MouseEvent), // init of incoming channel
		incomingKey:    make(chan term.KeyEvent),   // init of incoming channel
		incomingResize: make(chan term.ResizeEvent),
	}
	receiver.lifeCycle(ctx)
	return receiver
}

func main() {
	initLog.InitLogger()
	engine, err := core.NewCore(os.Getenv("TERM"), core.WithFinalizer(func() {
		log.Println("[hybrid] core finalizer called")
	}))
	if err != nil {
		log.Printf("[hybrid] error : %v", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithCancel(context.Background())
	if err := engine.Start(ctx); err != nil {
		log.Printf("[hybrid] error : %v", err)
		os.Exit(1)
	}

	rCtx, _ := context.WithCancel(ctx)
	receiver := NewReceiver(rCtx)
	engine.MouseDispatcher().Register(receiver)
	engine.KeyDispatcher().Register(receiver)
	engine.ResizeDispatcher().Register(receiver)

	seconds := 10
	wait := 10 * time.Second
	go func() {
		for seconds > 0 {
			<-time.After(1 * time.Second)
			log.Printf("[hybrid] waiting %d seconds", seconds)
			seconds--
			// last three seconds disable mouse, no events should be triggered
			if seconds == 3 {
				engine.MouseDispatcher().Disable()
			}
			if seconds == 0 {
				return
			}
		}
	}()
	<-time.After(wait)
	cancel()
	log.Println("[hybrid] waiting for engine to finalize correctly")
	<-engine.DyingChan()
	log.Println("[hybrid] done.")
	os.Exit(0)
}
