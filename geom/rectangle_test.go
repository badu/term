package geom_test

import (
	"context"
	"testing"

	"github.com/badu/term"
	"github.com/badu/term/color"
	"github.com/badu/term/geom"
)

func TestRectangleValid(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	_, err := geom.NewRectangle(ctx)
	if err == nil {
		t.Fatal("error : rectangle should return error")
	}
	_, err = geom.NewRectangle(ctx, geom.WithAcquisitionChan(make(chan term.Position)))
	if err == nil {
		t.Fatal("error : rectangle should return error")
	}
	_, err = geom.NewRectangle(ctx, geom.WithAcquisitionChan(make(chan term.Position)), geom.WithTopCorner(term.NewPosition(1, 1)))
	if err == nil {
		t.Fatal("error : rectangle should return error")
	}
	r, err := geom.NewRectangle(ctx, geom.WithAcquisitionChan(make(chan term.Position)), geom.WithTopCorner(term.NewPosition(1, 1)), geom.WithBottomCorner(term.NewPosition(1, 1)))
	if err != nil {
		t.Fatalf("error : there should be no error - %v", err)
	}
	if r.Bg() != color.Default {
		t.Fatal("error : background color should be default")
	}
	if r.Fg() != color.Default {
		t.Fatal("error : foreground color should be default")
	}
	if r.Size().Width != 1 || r.Size().Height != 1 {
		t.Fatal("error : width and height should be 1")
	}
	if px := r.Column(1); px != nil {
		t.Fatal("error : column should be empty (we have default horizontal orientation)")
	}
	if px := r.Row(1); px == nil {
		t.Fatal("error : we should have a row")
	}
	if px := r.Row(1); len(px) != 1 {
		t.Fatalf("error : row should have one pixel but has %d pixels", len(px))
	}
	cancel()
}
