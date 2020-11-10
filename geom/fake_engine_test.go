package geom_test

import (
	"context"
	"testing"

	"github.com/badu/term"
	"github.com/badu/term/core"
	"github.com/badu/term/key"
	"github.com/badu/term/mouse"
)

// for readability
type mouseChannels []chan term.MouseEvent

// delete removes the element at index from mouseChannels.
// Note that this is the fastest version, which changes order of elements inside the slice
// Yes, this is repeated code, because avoiding use of interface{}
func (c *mouseChannels) delete(idx int) {
	(*c)[idx] = (*c)[len(*c)-1] // Copy last element to index i.
	(*c)[len(*c)-1] = nil       // Erase last element (write zero value).
	*c = (*c)[:len(*c)-1]       // Truncate slice.
}

type FakeMouseDispatcher struct {
	t         *testing.T
	ctx       context.Context
	receivers mouseChannels
}

func (e *FakeMouseDispatcher) Register(l term.MouseListener) {
	if e.ctx == nil {
		e.t.Error("context not set : cannot listen context.Done()")
		return
	}
	// check against double registration
	alreadyRegistered := false
	for _, ch := range e.receivers {
		// Two channel values are considered equal if they originated from the same make call (meaning they refer to the same channel value in memory).
		if ch == l.MouseListen() {
			alreadyRegistered = true
			break
		}
	}
	if alreadyRegistered {
		return
	}
	// we're fine, lets register it
	e.receivers = append(e.receivers, l.MouseListen())
	// mounting a go routine to listen bye-bye life
	go func() {
		select {
		case <-e.ctx.Done():
			if e.t != nil {
				e.t.Log("[fake mouse dispatcher] context died : existing death listener")
			}
			return
		case <-l.DyingChan(): // wait for death announcement
			e.t.Log("component died")
			// now lookup for that very channel and forget it
			for idx, ch := range e.receivers {
				// Two channel values are considered equal if they originated from the same make call (meaning they refer to the same channel value in memory).
				if ch == l.MouseListen() {
					e.receivers.delete(idx)
					break
				}
			}
		}
	}()
}

func (e *FakeMouseDispatcher) DyingChan() chan struct{}            { return nil }
func (e *FakeMouseDispatcher) ResizeListen() chan term.ResizeEvent { return nil }
func (e *FakeMouseDispatcher) InChan() chan []byte                 { return nil }
func (e *FakeMouseDispatcher) LifeCycle(ctx context.Context)       { e.ctx = ctx }
func (e *FakeMouseDispatcher) Enable()                             {}
func (e *FakeMouseDispatcher) Disable()                            {}
func (e *FakeMouseDispatcher) Dispatch(col, row int, button term.ButtonMask, mod term.ModMask) {
	ev := mouse.NewEvent(row, col, button, mod)
	for _, cons := range e.receivers {
		cons <- ev
	}
}

type FakeKeyDispatcher struct {
	t         *testing.T
	ctx       context.Context
	receivers keyChannels
}

// for readability
type keyChannels []chan term.KeyEvent

// delete removes the element at index from keyChannels.
// Note that this is the fastest version, which changes order of elements inside the slice
// Yes, this is repeated code, because avoiding use of interface{}
func (c *keyChannels) delete(idx int) {
	(*c)[idx] = (*c)[len(*c)-1] // Copy last element to index i.
	(*c)[len(*c)-1] = nil       // Erase last element (write zero value).
	*c = (*c)[:len(*c)-1]       // Truncate slice.
}

func (e *FakeKeyDispatcher) Register(l term.KeyListener) {
	if e.ctx == nil {
		e.t.Error("context not set : cannot listen context.Done()")
		return
	}
	// check against double registration
	alreadyRegistered := false
	for _, ch := range e.receivers {
		// Two channel values are considered equal if they originated from the same make call (meaning they refer to the same channel value in memory).
		if ch == l.KeyListen() {
			alreadyRegistered = true
			break
		}
	}
	if alreadyRegistered {
		return
	}
	// we're fine, lets register it
	e.receivers = append(e.receivers, l.KeyListen())
	// mounting a go routine to listen bye-bye life
	go func() {
		select {
		case <-e.ctx.Done():
			e.t.Log("[fake key dispatcher] context died : existing death listener")
			return
		case <-l.DyingChan(): // wait for death announcement
			e.t.Log("component died")
			// now lookup for that very channel and forget it
			for idx, ch := range e.receivers {
				// Two channel values are considered equal if they originated from the same make call (meaning they refer to the same channel value in memory).
				if ch == l.KeyListen() {
					e.receivers.delete(idx)
					break
				}
			}
		}
	}()
}
func (e *FakeKeyDispatcher) DyingChan() chan struct{}      { return nil }
func (e *FakeKeyDispatcher) HasKey(k term.Key) bool        { return false }
func (e *FakeKeyDispatcher) InChan() chan []byte           { return nil }
func (e *FakeKeyDispatcher) LifeCycle(ctx context.Context) { e.ctx = ctx }
func (e *FakeKeyDispatcher) Dispatch(k term.Key, ch rune, mod term.ModMask) {
	ev := key.NewEvent(k, ch, mod) // one event for everyone
	for _, cons := range e.receivers {
		cons <- ev
	}
}

// for readability
type resizeChannels []chan term.ResizeEvent

// delete removes the element at index from resizeChannels.
// Note that this is the fastest version, which changes order of elements inside the slice
// Yes, this is repeated code, because avoiding use of interface{}
func (c *resizeChannels) delete(idx int) {
	(*c)[idx] = (*c)[len(*c)-1] // Copy last element to index i.
	(*c)[len(*c)-1] = nil       // Erase last element (write zero value).
	*c = (*c)[:len(*c)-1]       // Truncate slice.
}

type FakeResizeDispatcher struct {
	t         *testing.T
	ctx       context.Context
	receivers resizeChannels
}

func (e *FakeResizeDispatcher) LifeCycle(ctx context.Context) { e.ctx = ctx }
func (e *FakeResizeDispatcher) DyingChan() chan struct{}      { return nil }
func (e *FakeResizeDispatcher) Register(l term.ResizeListener) {
	if e.ctx == nil {
		e.t.Error("context not set : cannot listen context.Done()")
		return
	}
	// check against double registration
	alreadyRegistered := false
	for _, ch := range e.receivers {
		// Two channel values are considered equal if they originated from the same make call (meaning they refer to the same channel value in memory).
		if ch == l.ResizeListen() {
			alreadyRegistered = true
			break
		}
	}
	if alreadyRegistered {
		return
	}
	// we're fine, lets register it
	e.receivers = append(e.receivers, l.ResizeListen())
	// mounting a go routine to listen bye-bye life
	go func() {
		select {
		case <-e.ctx.Done():
			e.t.Log("[fake resize dispatcher] context died : existing death listener")
			return
		case <-l.DyingChan(): // wait for death announcement
			e.t.Log("component died")
			// now lookup for that very channel and forget it
			for idx, ch := range e.receivers {
				// Two channel values are considered equal if they originated from the same make call (meaning they refer to the same channel value in memory).
				if ch == l.ResizeListen() {
					e.receivers.delete(idx)
					break
				}
			}
		}
	}()
}
func (e *FakeResizeDispatcher) Dispatch(newCols, newRows int) {
	ev := core.NewResizeEvent(newCols, newRows)
	e.t.Logf("dispatching new size %04d cols x %04d rows to %d receivers", newCols, newRows, len(e.receivers))
	for _, cons := range e.receivers {
		cons <- ev
	}
}

// FakeEngine is a core used for tests
type FakeEngine struct {
	t             *testing.T
	Columns, Rows int
	ctx           context.Context
	MD            *FakeMouseDispatcher
	KD            *FakeKeyDispatcher
	RD            *FakeResizeDispatcher
}

func NewFakeEngine(t *testing.T, cols, rows int) *FakeEngine {
	return &FakeEngine{
		t:       t,
		Columns: cols,
		Rows:    rows,
		MD:      &FakeMouseDispatcher{t: t},
		KD:      &FakeKeyDispatcher{t: t},
		RD:      &FakeResizeDispatcher{t: t},
	}
}

func (e *FakeEngine) DyingChan() chan struct{} {
	return nil
}

func (e *FakeEngine) Start(ctx context.Context) error {
	e.ctx = ctx
	e.MD.LifeCycle(ctx)
	e.KD.LifeCycle(ctx)
	e.RD.LifeCycle(ctx)
	return nil
}

func (e *FakeEngine) ResizeDispatcher() term.ResizeDispatcher {
	return e.RD
}

func (e *FakeEngine) KeyDispatcher() term.KeyDispatcher {
	return e.KD
}

func (e *FakeEngine) MouseDispatcher() term.MouseDispatcher {
	return e.MD
}

func (e *FakeEngine) CanDisplay(r rune, checkFallbacks bool) bool {
	return false
}

func (e *FakeEngine) CharacterSet() string {
	return ""
}

func (e *FakeEngine) SetRuneFallback(orig rune, fallback string) {}

func (e *FakeEngine) UnsetRuneFallback(orig rune) {}

func (e *FakeEngine) NumColors() int {
	return 0
}

func (e *FakeEngine) Size() *term.Size {
	return &term.Size{
		Columns: e.Columns,
		Rows:    e.Rows,
	}
}

func (e *FakeEngine) SetSize(columns, rows int) {
	e.Columns = columns
	e.Rows = rows
	e.RD.Dispatch(columns, rows)
}

func (e *FakeEngine) HasTrueColor() bool {
	return false
}

func (e *FakeEngine) Style() term.Style {
	return nil
}

func (e *FakeEngine) ActivePixels(pixels []term.PixelGetter) {
	e.t.Logf("consider that %d pixels were activated", len(pixels))
}

func (e *FakeEngine) Redraw(pixels []term.PixelGetter) {
	e.t.Logf("consider that %d pixels were redrawn", len(pixels))
}

func (e *FakeEngine) ShowCursor(where *term.Position) {
	e.t.Logf("consider cursor at %03d,%03d", where.Row, where.Column)
}

func (e *FakeEngine) HideCursor() {}

func (e *FakeEngine) Cursor() *term.Position {
	return nil
}

func (e *FakeEngine) Clear() {}

func (e *FakeEngine) HasMouse() bool { return true }
