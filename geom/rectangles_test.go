package geom_test

import (
	"context"
	"testing"
	"time"

	"github.com/badu/term"
	. "github.com/badu/term/geom"
	"github.com/badu/term/style"
)

const (
	defaultWidth  = 146
	defaultHeight = 36
)

func TestTree(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	aq := make(chan term.Position)
	fakeEngine := NewFakeEngine(t, ctx, 132, 43)
	mainTree := NewTree(ctx, defaultWidth, defaultHeight, aq, style.NoOrientation, fakeEngine.RD) // oriented as default, style.Vertical
	r1, _ := NewRectangle(ctx, WithAcquisitionChan(aq), WithHeight(20))                           // 20% height, 100% width (optional) of the parent (root)
	r2, _ := NewRectangle(ctx, WithAcquisitionChan(aq), WithHeight(60))                           // 60% height, 100% width (optional) of the parent (root)
	r3, _ := NewRectangle(ctx, WithAcquisitionChan(aq), WithHeight(20))                           // 20% height, 100% width (optional) of the parent (root)
	mainTree.Register(r1, r2, r3)
	r4, _ := NewRectangle(ctx, WithAcquisitionChan(aq), WithWidth(30), WithHeight(30)) // 30% width and height (level+1)
	mainTree.Register(r4)
	cancel()
}

func TestResize(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	aq := make(chan term.Position)
	fakeEngine := NewFakeEngine(t, ctx, 132, 43)
	r1, _ := NewRectangle(ctx, WithAcquisitionChan(aq), WithMinSize(term.NewSize(20, 20)), WithCore(fakeEngine.RD))
	_ = r1
	fakeEngine.RD.Dispatch(10, 10)
	t.Log("cancelling")
	cancel()
	<-time.After(500 * time.Millisecond)
}
