package geom

import (
	"context"
	"errors"
	"log"
	"sync"

	"github.com/badu/term"
	"github.com/badu/term/style"
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
	sync.RWMutex                         //
	sync.Once                            // required for registering lifecycle goroutines exactly once
	root                                 //
	rectangles     Tree                  //
	engine         term.Engine           //
	incomingMouse  chan term.MouseEvent  //
	incomingKey    chan term.KeyEvent    //
	incomingResize chan term.ResizeEvent //
	died           chan struct{}         //
	owners         map[int]Owners        // map[position_hash]Owners
	hidden         bool                  //
}

// WithEngine
func WithEngine(engine term.Engine) PageOption {
	return func(p *Page) {
		p.RLock()
		defer p.RUnlock()
		p.engine = engine
	}
}

// WithPageOrientation
func WithPageOrientation(o style.Orientation) PageOption {
	return func(p *Page) {
		p.RLock()
		defer p.RUnlock()
		p.orientation = o
	}
}

// NewPage
func NewPage(ctx context.Context, opts ...PageOption) (*Page, error) {
	res := &Page{
		died:           make(chan struct{}),
		incomingMouse:  make(chan term.MouseEvent),
		incomingKey:    make(chan term.KeyEvent),
		incomingResize: make(chan term.ResizeEvent),
		owners:         make(map[int]Owners),
		root: root{
			orientation:  style.Vertical,           // default orientation
			topCorner:    term.NewPosition(0, 0),   // by default, rectangle is on top corner
			bottomCorner: term.NewPosition(-1, -1), // by default, rectangle is nowhere
		},
	}
	for _, opt := range opts {
		opt(res)
	}
	if res.engine == nil {
		return nil, errors.New("page requires Engine to register itself to events")
	}
	initialSize := res.engine.Size()
	res.bottomCorner.Column = initialSize.Columns
	res.bottomCorner.Row = initialSize.Rows
	log.Printf("new page cols : %03d x rows : %03d", res.engine.Size().Columns, res.engine.Size().Rows)
	res.engine.ResizeDispatcher().Register(res)
	if res.engine.HasMouse() {
		res.engine.MouseDispatcher().Register(res)
	}
	res.engine.KeyDispatcher().Register(res)
	res.lifeCycle(ctx)
	return res, nil
}

func (p *Page) lifeCycle(ctx context.Context) {
	p.Once.Do(func() {
		p.startup(p.engine)
		go func(cx context.Context) {
			for {
				select {
				case <-cx.Done():
					p.Shutdown()
					p.died <- struct{}{} // tell that we're done
					return
				case ke := <-p.incomingKey:
					if p.hidden {
						return
					}
					_ = ke
				case me := <-p.incomingMouse:
					if p.hidden {
						return
					}
					// TODO : tell only rectangle in that bounds
					_ = me
				case se := <-p.incomingResize:
					p.resize(se.Size())
				}
			}
		}(ctx)
	})
}

func (p *Page) pixels() []term.PixelGetter {
	p.RLock()
	defer p.RUnlock()
	return nil
}

// Shutdown
func (p *Page) Shutdown() {
	p.Lock()
	defer p.Unlock()
	// TODO : implement me on context.Context
}

// MouseListen
func (p *Page) MouseListen() chan term.MouseEvent {
	p.RLock()
	defer p.RUnlock()
	return p.incomingMouse
}

// KeyListen
func (p *Page) KeyListen() chan term.KeyEvent {
	p.RLock()
	defer p.RUnlock()
	return p.incomingKey
}

// ResizeListen
func (p *Page) ResizeListen() chan term.ResizeEvent {
	p.RLock()
	defer p.RUnlock()
	return p.incomingResize
}

// DyingChan
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
	// TODO : activate underlying pixels
}

// Deactivate
func (p *Page) Deactivate() {
	p.Lock()
	defer p.Unlock()
	p.hidden = true
	// TODO : deactivate underlying pixels
}
