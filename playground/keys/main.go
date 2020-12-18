package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"
	"runtime/pprof"
	"time"

	"github.com/badu/term"
	"github.com/badu/term/core"
	initLog "github.com/badu/term/log"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
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

// go run main.go -cpuprof cpu.out -memprof mem.out
var cpuprofile = flag.String("cpuprof", "", "write cpu profile to `file`")
var memprofile = flag.String("memprof", "", "write memory profile to `file`")

func main() {

	initLog.InitLogger()

	flag.Parse()
	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal("could not create CPU profile: ", err)
		}
		defer f.Close() // error handling omitted for example
		if err := pprof.StartCPUProfile(f); err != nil {
			log.Fatal("could not start CPU profile: ", err)
		}
		defer pprof.StopCPUProfile()
	}

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
		percents, _ := cpu.Percent(time.Second, true)
		v, _ := mem.VirtualMemory()
		for seconds > 0 {
			<-time.After(1 * time.Second)
			log.Printf("[key] waiting %d seconds\n", seconds)
			msg := "[Processor]"
			for idx, percent := range percents {
				msg += fmt.Sprintf("[#%d: %.2f]", idx+1, percent)
			}
			log.Print(msg)
			log.Printf("[Memory] Total: %v, Free:%v, UsedPercent:%f%%\n", v.Total, v.Free, v.UsedPercent)
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

	if *memprofile != "" {
		f, err := os.Create(*memprofile)
		if err != nil {
			log.Fatal("could not create memory profile: ", err)
		}
		defer f.Close() // error handling omitted for example
		runtime.GC()    // get up-to-date statistics
		if err := pprof.WriteHeapProfile(f); err != nil {
			log.Fatal("could not write memory profile: ", err)
		}
	}

}
