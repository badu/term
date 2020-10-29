package app

import (
	"context"
	"sync"
)

// ApplicationOption
type ApplicationOption func(a *Application)

// Application
type Application struct {
	sync.RWMutex
	ctx     context.Context
	cancels map[*Page]func()
	pages   []Page
}

// WithPage
func WithPage(p *Page) ApplicationOption {
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
func (a *Application) AddPage(p *Page) {
	pageCtx, cancel := context.WithCancel(a.ctx)
	a.cancels[p] = cancel
	WithContext(pageCtx)
	a.pages = append(a.pages, *p)
}

// DeactivatePage
func (a *Application) DeactivatePage(p *Page) {

}

// RemovePage
func (a *Application) RemovePage(p *Page) {
	a.Lock()
	defer a.Unlock()
	a.cancels[p]()
}
