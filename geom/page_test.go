package geom_test

import (
	"context"
	"log"
	"testing"
	"time"

	"github.com/badu/term"
	"github.com/badu/term/geom"
	initLog "github.com/badu/term/log"
)

func testAcquisitionChan() geom.RectangleOption {
	return geom.WithAcquisitionChan(make(chan term.Position))
}

func TestMoveToFront(t *testing.T) {
	o := make(geom.Owners, 0)
	ctx, cancel := context.WithCancel(context.Background())
	r, err := geom.NewRectangle(ctx, geom.WithTopCorner(10, 10), geom.WithBottomCorner(20, 20), testAcquisitionChan())
	if err != nil {
		t.Fatalf("error : %v", err)
	}
	o.Add(r)
	r2, err := geom.NewRectangle(ctx, geom.WithTopCorner(30, 30), geom.WithBottomCorner(32, 32), testAcquisitionChan())
	if err != nil {
		t.Fatalf("error : %v", err)
	}
	o.Add(r2)
	o.MoveToFront(r)
	t.Logf("size : %d", len(o))
	for _, r := range o {
		//t.Logf("rect %d : %#v", r.Id(), r)
		_ = r
	}
	o.ForgetFirst()
	t.Logf("size : %d", len(o))
	for _, r := range o {
		//t.Logf("rect %d : %#v", r.Id(), r)
		_ = r
	}
	r3, err := geom.NewRectangle(ctx, geom.WithTopCorner(20, 20), geom.WithBottomCorner(22, 22), testAcquisitionChan())
	if err != nil {
		t.Fatalf("error : %v", err)
	}
	o.Add(r3)
	o.MoveToFront(r3)
	t.Logf("size : %d", len(o))
	for _, r := range o {
		//t.Logf("rect %d : %#v", r.Id(), r)
		_ = r
	}
	cancel()
}

func TestRectangleWithPage(t *testing.T) {
	initLog.InitLogger()
	log.Print("starting test")
	ctx, cancel := context.WithCancel(context.Background())
	fakeEngine := NewFakeEngine(t, ctx, 132, 43)
	pageCtx, pageCancel := context.WithCancel(ctx)
	p, err := geom.NewPage(pageCtx, geom.WithEngine(fakeEngine))
	if err != nil {
		t.Fatalf("error : %v", err)
	}
	log.Print("resizing to 140 x 50")
	fakeEngine.SetSize(140, 50)
	log.Print("resizing to 130 x 40")
	fakeEngine.SetSize(130, 40)
	_ = p
	<-time.After(4 * time.Second)
	pageCancel()
	<-time.After(1 * time.Second)
	cancel()
}

func Test(t *testing.T) {

}
