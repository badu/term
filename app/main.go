package app

import (
	"context"
	"sync"

	"github.com/badu/term/geom"
)

// ApplicationOption
type ApplicationOption func(a *Application)

// Application
type Application struct {
	sync.RWMutex
	ctx     context.Context
	cancels map[*geom.Page]func()
	pages   []geom.Page
}

// WithPage
func WithPage(p *geom.Page) ApplicationOption {
	return func(a *Application) {
		a.AddPage(p)
	}
}

// NewApplication
func NewApplication(ctx context.Context, opts ...ApplicationOption) *Application {
	res := &Application{}
	for _, opt := range opts {
		opt(res)
	}
	res.Start(ctx)
	return res
}

// Start
func (a *Application) Start(ctx context.Context) {

}

// AddPage
func (a *Application) AddPage(p *geom.Page) {
	pageCtx, cancel := context.WithCancel(a.ctx)
	a.cancels[p] = cancel
	geom.WithEngine(pageCtx)
	a.pages = append(a.pages, *p)
}

// DeactivatePage
func (a *Application) DeactivatePage(p *geom.Page) {

}

// RemovePage
func (a *Application) RemovePage(p *geom.Page) {
	a.Lock()
	defer a.Unlock()
	a.cancels[p]()
}
