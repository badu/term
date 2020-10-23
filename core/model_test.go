package core_test

import (
	"context"
	"testing"
	"time"

	"github.com/badu/term"
	"github.com/badu/term/core"
)

type fakeCharsHandler struct {
	t *testing.T
}

func (f fakeCharsHandler) OnChange(value []rune) {
	f.t.Logf("Chars changed : %s", string(value))
}

func TestCharsModel(t *testing.T) {
	c, err := core.NewChars(nil)
	if err != nil {
		t.Fatalf("error : %v", err)
	}
	h := &fakeCharsHandler{t: t}
	initial := []rune{'ﾐ', 'ﾑ', 'ﾒ', 'ﾓ', 'ﾔ', 'ﾕ', 'ﾖ', 'ﾗ', 'ﾘ', 'ﾙ', 'ﾚ', 'ﾛ', 'ﾜ', 'ﾝ', 'ﾞ', 'ﾟ'}
	shorter := []rune{'ﾐ', 'ﾑ', 'ﾒ', 'ﾓ', 'ﾔ'}
	longer := []rune{'ﾐ', 'ﾑ', 'ﾒ', 'ﾓ', 'ﾔ', 'ﾕ', 'ﾖ', 'ﾗ', 'ﾘ', 'ﾙ', 'ﾚ', 'ﾛ', 'ﾜ', 'ﾝ', 'ﾞ', 'ﾟ', 'A'}
	idx := 0
	ctx, cancel := context.WithCancel(context.Background())
	producer := func(t *testing.T, c core.Chars) {
		for {
			select {
			case <-ctx.Done():
				t.Log("context done. exiting")
				return
			case <-time.After(time.Millisecond * 300):
				switch idx {
				case 0, 4:
					if err := c.Set(shorter); err != nil {
						t.Logf("error : %v", err)
					}
				case 1, 5:
					if err := c.Set(longer); err != nil {
						t.Logf("error : %v", err)
					}
				case 2, 6:
					if err := c.Set(initial); err != nil {
						t.Logf("error : %v", err)
					}
				case 7, 8:
					if err := c.Set(initial); err != nil {
						t.Logf("error : %v", err)
					}
				case 3:
					if err := c.Set(nil); err != nil {
						t.Logf("error : %v", err)
					}
				}
				idx++
			}
		}
	}

	fn, err := core.PropagateCharsChange(ctx, c, h)
	if err != nil {
		t.Fatalf("error : %v", err)
	}

	go fn()
	go producer(t, c)
	// wait three seconds here
	<-time.After(time.Second * 3)
	cancel()
}

func TestParentGoRoutineKilled(t *testing.T) {
	cells := make([]term.PixelGetter, 0)
	aggregateChan := make(chan term.PixelGetter)
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		t.Logf("[core] mounting pixel aggregate routine")
		for _, pix := range cells {
			go func(pixCh chan term.PixelGetter) {
				for msg := range pixCh {
					aggregateChan <- msg
				}
			}(pix.DrawCh())
		}
		for {
			select {
			case <-ctx.Done():
				t.Log("[core] previous pixel context was cancelled -> returning")
				return
			case pixel := <-aggregateChan:
				t.Logf("draw request %#v", pixel)
				// check if this is the cancellation pixel, we're exiting the goroutine
				if pixel.Position().X == -1 && pixel.Position().Y == -1 {
					t.Log("[core] received shutdown pixel -> returning")
					return
				}
				t.Log("ok, fine")
			}
		}
	}()

	<-time.After(10 * time.Second)
	cancel()
}
