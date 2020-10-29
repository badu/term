package app

import (
	"context"

	"github.com/badu/term/geom"
)

// PageOption
type PageOption func(p *Page)

// Page
type Page struct {
	geom.Rectangle
	ctx context.Context
}

// WithContext
func WithContext(ctx context.Context) PageOption {
	return func(p *Page) {
		p.ctx = ctx
	}
}

// NewPage
func NewPage(ctx context.Context, opts ...PageOption) *Page {
	res := &Page{}
	for _, opt := range opts {
		opt(res)
	}
	return res
}

// Context
func (p *Page) Context() context.Context {
	return p.ctx
}

// Activate
func (p *Page) Activate() {

}

// Deactivate
func (p *Page) Deactivate() {

}
