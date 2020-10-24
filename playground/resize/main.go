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
	incoming chan term.ResizeEvent // We need a channel on which we will listen for incoming events
	died     chan struct{}         // this is a buffered channel of size one
}

func (r *listener) ResizeListen() chan term.ResizeEvent {
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
				r.died <- struct{}{}
				return
			case ev := <-r.incoming:
				log.Printf("event : %v", ev)
			}
		}
	}()
}

func NewReceiver(ctx context.Context) term.ResizeListener {
	receiver := &listener{
		died:     make(chan struct{}),         // init of died channel, a buffered channel of exactly one
		incoming: make(chan term.ResizeEvent), // init of incoming channel}
	}
	receiver.lifeCycle(ctx)
	return receiver
}

func main() {
	initLog.InitLogger()

	engine, err := core.NewCore(os.Getenv("TERM"), core.WithFinalizer(func() {
		log.Println("Core finalizer called")
	}))
	if err != nil {
		log.Printf("error : %v", err)
		os.Exit(1)
	}
	ctx, cancel := context.WithCancel(context.Background())
	if err := engine.Start(ctx); err != nil {
		log.Printf("error : %v", err)
		os.Exit(1)
	}

	rCtx, cancel2 := context.WithCancel(ctx)
	receiver := NewReceiver(rCtx)
	engine.ResizeDispatcher().Register(receiver)

	seconds := 30
	wait := 30 * time.Second
	go func() {
		for seconds > 0 {
			<-time.After(1 * time.Second)
			log.Printf("waiting %d seconds", seconds)
			seconds--
		}
	}()
	<-time.After(wait)
	cancel2()
	cancel()
	log.Println("waiting for engine to finalize correctly")
	<-engine.DyingChan()
}
