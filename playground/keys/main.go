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
	incoming chan term.KeyEvent // We need a channel on which we will listen for incoming events
	died     chan struct{}      // this is a buffered channel of size one
}

func (r *listener) KeyListen() chan term.KeyEvent {
	return r.incoming
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
				log.Println("[key] context is done")
				r.died <- struct{}{}
				return
			case ev := <-r.incoming:
				log.Printf("[key] key : name=%s key=%c modname=%s", ev.Name(), ev.Rune(), ev.ModName())
			}
		}
	}()
}

func NewReceiver(ctx context.Context) term.KeyListener {
	receiver := &listener{
		died:     make(chan struct{}),      // init of died channel, a buffered channel of exactly one
		incoming: make(chan term.KeyEvent), // init of incoming channel}
	}
	receiver.lifeCycle(ctx)
	return receiver
}

func main() {
	initLog.InitLogger()

	engine, err := core.NewCore(os.Getenv("TERM"), core.WithFinalizer(func() {
		log.Println("[key] core finalizer called")
	}))
	if err != nil {
		log.Printf("[key] error : %v", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithCancel(context.Background())
	if err := engine.Start(ctx); err != nil {
		log.Printf("[key] error : %v", err)
		os.Exit(1)
	}

	rCtx, cancel2 := context.WithCancel(ctx)
	receiver := NewReceiver(rCtx)
	engine.KeyDispatcher().Register(receiver)

	seconds := 10
	wait := 11 * time.Second
	go func() {
		for seconds > 0 {
			<-time.After(1 * time.Second)
			log.Printf("[key] waiting %d seconds", seconds)
			seconds--
			if seconds == 0 {
				log.Println("[key] 0 seconds existing counting goroutine")
				return
			}
		}
	}()
	<-time.After(wait)
	cancel2()
	cancel()
	log.Println("[key] waiting for engine to finalize correctly")
	<-engine.DyingChan()
	log.Println("[key] done.")
}
