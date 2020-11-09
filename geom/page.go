package geom

import (
	"context"
	"errors"
	"log"
	"sync"

	"github.com/badu/term"
)

// PageOption
type PageOption func(p *Page)

// Owners
type Owners []Rectangle

// Add
func (ow *Owners) Add(pr *Rectangle) {
	*ow = append(*ow, *pr)
}

// MoveToFront
func (ow *Owners) MoveToFront(pr *Rectangle) {
	o := *ow
	r := *pr
	if len(o) == 0 || o[0].id == r.id {
		*ow = o
		return
	}
	var prev Rectangle
	for i, elem := range o {
		switch {
		case i == 0:
			o[0] = r
			prev = elem
		case elem.id == r.id:
			o[i] = prev
			*ow = o
			return
		default:
			o[i] = prev
			prev = elem
		}
	}
	*ow = append(o, prev)
}

// ForgetFirst
func (ow *Owners) ForgetFirst() {
	o := *ow
	switch len(o) {
	case 0:
		return
	case 1:
		o = make([]Rectangle, 0)
	default:
		copy(o[1:], o[2:])
		o = o[:len(o)-1]
	}
	*ow = o
}

// Page
type Page struct {
	sync.RWMutex
	ctx            context.Context
	rectangles     Tree
	engine         term.Engine
	incomingMouse  chan term.MouseEvent
	incomingKey    chan term.KeyEvent
	incomingResize chan term.ResizeEvent
	died           chan struct{}
	owners         map[int]Owners // map[position_hash]Owners
	hidden         bool
}

// WithEngine
func WithEngine(engine term.Engine) PageOption {
	return func(p *Page) {
		p.RLock()
		defer p.RUnlock()
		p.engine = engine
	}
}

// NewPage
func NewPage(ctx context.Context, opts ...PageOption) (*Page, error) {
	res := &Page{
		ctx:            ctx,
		died:           make(chan struct{}),
		incomingMouse:  make(chan term.MouseEvent),
		incomingKey:    make(chan term.KeyEvent),
		incomingResize: make(chan term.ResizeEvent),
		owners:         make(map[int]Owners),
	}
	for _, opt := range opts {
		opt(res)
	}
	if res.engine == nil {
		return nil, errors.New("page requires Engine to register itself to events")
	}
	log.Printf("new page cols : %03d x rows : %03d", res.engine.Size().Width, res.engine.Size().Height)
	res.engine.ResizeDispatcher().Register(res)
	res.engine.MouseDispatcher().Register(res)
	res.engine.KeyDispatcher().Register(res)
	go func() {
		for {
			select {
			case <-ctx.Done():
				res.Shutdown()
				res.died <- struct{}{} // tell that we're done
				return
			case ke := <-res.incomingKey:
				if res.hidden {
					return
				}
				_ = ke
			case me := <-res.incomingMouse:
				if res.hidden {
					return
				}
				// TODO : tell only rectangle in that bounds
				_ = me
			case se := <-res.incomingResize:
				log.Printf("resize : %v", se.Size())
				// TODO : tell only rectangles that have variable size
				//res.Resize(se.Size())
			}
		}
	}()
	return res, nil
}

func (p *Page) pixels() []term.PixelGetter {
	p.RLock()
	defer p.RUnlock()
	return nil
}

func (p *Page) Shutdown() {
	p.Lock()
	defer p.Unlock()
}

func (p *Page) MouseListen() chan term.MouseEvent {
	p.RLock()
	defer p.RUnlock()
	return p.incomingMouse
}

func (p *Page) KeyListen() chan term.KeyEvent {
	p.RLock()
	defer p.RUnlock()
	return p.incomingKey
}

func (p *Page) ResizeListen() chan term.ResizeEvent {
	p.RLock()
	defer p.RUnlock()
	return p.incomingResize
}

func (p *Page) DyingChan() chan struct{} {
	p.RLock()
	defer p.RUnlock()
	return p.died
}

// Activate
func (p *Page) Activate() {
	p.Lock()
	defer p.Unlock()
	p.hidden = false
}

// Deactivate
func (p *Page) Deactivate() {
	p.Lock()
	defer p.Unlock()
	p.hidden = true
}
