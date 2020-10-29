package main

import (
	"context"
	"fmt"
	"log"
	"math"
	"os"

	"github.com/badu/term"
	"github.com/badu/term/color"
	"github.com/badu/term/core"
	"github.com/badu/term/encoding"
	"github.com/badu/term/geom"
	"github.com/badu/term/key"
	initLog "github.com/badu/term/log"
	"github.com/badu/term/mouse"
	"github.com/badu/term/runewidth"
	"github.com/badu/term/style"
)

func (r *listener) emitStr(column, row int, style *style.Style, str string) {
	for _, c := range str {
		w := runewidth.RuneWidth(c)
		if w == 0 {
			c = encoding.Space
			w = 1
		}
		r.refs[column][row].Set(c, style.Fg, style.Bg)
		column += w
	}
}

func (r *listener) drawBox(col1, row1, col2, row2 int, style *style.Style, c rune) {
	if row2 < row1 {
		row1, row2 = row2, row1
	}
	if col2 < col1 {
		col1, col2 = col2, col1
	}

	for col := col1; col <= col2; col++ {
		r.refs[col][row1].Set(encoding.HLine, style.Fg, style.Bg)
		r.refs[col][row2].Set(encoding.HLine, style.Fg, style.Bg)
	}
	for row := row1 + 1; row < row2; row++ {
		r.refs[col1][row].Set(encoding.VLine, style.Fg, style.Bg)
		r.refs[col2][row].Set(encoding.VLine, style.Fg, style.Bg)
	}
	if row1 != row2 && col1 != col2 {
		// Only add corners if we need to
		r.refs[col1][row1].Set(encoding.ULCorner, style.Fg, style.Bg)
		r.refs[col2][row1].Set(encoding.URCorner, style.Fg, style.Bg)
		r.refs[col1][row2].Set(encoding.LLCorner, style.Fg, style.Bg)
		r.refs[col2][row2].Set(encoding.LRCorner, style.Fg, style.Bg)
	}
	for row := row1 + 1; row < row2; row++ {
		for col := col1 + 1; col < col2; col++ {
			r.refs[col][row].Set(c, style.Fg, style.Bg)
		}
	}

}

type listener struct {
	incomingMouse  chan term.MouseEvent  // We need a channel on which we will listen for incoming events
	incomingKey    chan term.KeyEvent    // We need a channel on which we will listen for incoming events
	incomingResize chan term.ResizeEvent // We need a channel on which we will listen for incoming events
	died           chan struct{}         // this is a buffered channel of size one
	engine         term.Engine           //
	refs           [][]term.Pixel
}

func (r *listener) init(size *term.Size) {
	r.refs = make([][]term.Pixel, size.Width)
	var getters = make([]term.PixelGetter, 0)
	for column := 0; column < size.Width; column++ {
		r.refs[column] = make([]term.Pixel, size.Height)
		for row := 0; row < size.Height; row++ {
			var px term.Pixel
			if column%2 == 0 {
				px, _ = geom.NewPixel(geom.WithPosition(term.NewPosition(column, row)))
			} else {
				px, _ = geom.NewPixel(geom.WithPosition(term.NewPosition(column, row)))
			}
			r.refs[column][row] = px
			getters = append(getters, px)
		}
	}
	r.engine.ActivePixels(getters)
}

func (r *listener) MouseListen() chan term.MouseEvent {
	return r.incomingMouse
}

func (r *listener) KeyListen() chan term.KeyEvent {
	return r.incomingKey
}

func (r *listener) ResizeListen() chan term.ResizeEvent {
	return r.incomingResize
}

func (r *listener) DyingChan() chan struct{} {
	return r.died
}

func (r *listener) drawSelect(col1, row1, col2, row2 int, st *style.Style) []term.PixelSetter {
	if col2 < col1 {
		col1, col2 = col2, col1
	}
	if row2 < row1 {
		row1, row2 = row2, row1
	}
	var affectedPixels []term.PixelSetter
	for col := col1; col <= col2; col++ {
		for row := row1; row <= row2; row++ {
			r.refs[col][row].Set('█', st.Fg, st.Bg)
			affectedPixels = append(affectedPixels, r.refs[col][row])
		}
	}
	return affectedPixels
}

func (r *listener) cleanup(pixelsToCleanup []term.PixelSetter) {
	for _, pixel := range pixelsToCleanup {
		pixel.Set(' ', color.Reset, color.Reset)
	}
}

// listen listens for incoming events
func (r *listener) lifeCycle(ctx context.Context, cancel func()) {
	go func() {
		white := style.NewStyle(style.WithFg(color.White), style.WithBg(color.Blue))
		blue := style.NewStyle(style.WithFg(color.Blue))
		rgb := style.NewStyle(style.WithBg(color.Red), style.WithFg(color.NewRGBAColor(0, 0, math.MaxUint32, math.MaxUint32)))
		bstr := ""
		mousePosition := "%03d,%03d"
		buttonsInfo := "%10s"
		keysInfo := "%10s"

		keyEventName := ""
		escapeCount := 0

		mouseCol, mouseRow := -1, -1
		mouseDownCol, mouseDownRow := -1, -1

		size := r.engine.Size()
		hasChange := true

		r.init(size)
		r.drawBox(1, 1, 42, 7, white, encoding.Space)
		r.emitStr(2, 2, rgb, "Press ESC twice to exit, C to clear.")
		r.emitStr(2, 3, white, "Click and drag to draw a rectangle.")
		const (
			mouseStr   = "Mouse: "
			buttonsStr = "Buttons: "
			keysStr    = "Keys: "
		)
		r.emitStr(2, 4, white, mouseStr)
		r.emitStr(2, 5, white, buttonsStr)
		r.emitStr(2, 6, white, keysStr)
		log.Printf("[app] initial w = %03d h = %03d", size.Width, size.Height)
		pixelsToCleanup := make([]term.PixelSetter, 0)
		for {
			if hasChange {
				r.emitStr(2+len(mouseStr), 4, white, fmt.Sprintf(mousePosition, mouseCol, mouseRow))
				r.emitStr(2+len(buttonsStr), 5, white, fmt.Sprintf(buttonsInfo, bstr))
				r.emitStr(2+len(keysStr), 6, white, fmt.Sprintf(keysInfo, keyEventName))
				r.engine.HideCursor()
				hasChange = false
			}
			select {
			case <-ctx.Done():
				log.Println("[app] context is done.")
				r.died <- struct{}{}
				return
			case ev := <-r.incomingMouse:
				mouseCol, mouseRow = ev.Position()
				bstr = ev.ButtonNames()
				hasChange = true
				switch ev.Buttons() {
				case mouse.ButtonPrimary:
					if mouseDownCol == -1 && mouseDownRow == -1 {
						log.Printf("[app] mouse up at %03d, %03d", mouseCol, mouseRow)
						mouseDownCol, mouseDownRow = mouseCol, mouseRow
					} else {
						pixelsToCleanup = append(pixelsToCleanup, r.drawSelect(mouseDownCol, mouseDownRow, mouseCol, mouseRow, blue)...)
					}
				case mouse.ButtonNone:
					if mouseDownCol > 0 && mouseDownRow > 0 {
						log.Printf("[app] mouse up at %03d, %03d. draw rect with %03d, %03d", mouseCol, mouseRow, mouseDownCol, mouseDownRow)
						r.cleanup(pixelsToCleanup)
						mouseDownCol, mouseDownRow = -1, -1
						pixelsToCleanup = make([]term.PixelSetter, 0)
					}
				}
			case ev := <-r.incomingKey:
				switch ev.Key() {
				case key.Escape:
					escapeCount++
					if escapeCount > 1 {
						log.Println("[app] waiting for engine to finalize correctly")
						cancel()
						return
					}
				case key.CtrlL:
					escapeCount = 0
				default:
					escapeCount = 0
					if ev.Rune() == 'C' || ev.Rune() == 'c' {
						log.Println("[app] clear screen")
						r.engine.Clear()
					}
				}
				keyEventName = ev.Name()
				hasChange = true
			case ev := <-r.incomingResize:
				newSize := ev.Size()
				size = newSize
				hasChange = true
			}

		}
	}()
}

func NewReceiver(ctx context.Context, engine term.Engine, cancel func()) *listener {
	receiver := &listener{
		died:           make(chan struct{}),        // init of died channel, a buffered channel of exactly one
		incomingMouse:  make(chan term.MouseEvent), // init of incoming channel
		incomingKey:    make(chan term.KeyEvent),   // init of incoming channel
		incomingResize: make(chan term.ResizeEvent),
		engine:         engine,
	}
	receiver.lifeCycle(ctx, cancel)
	return receiver
}

func main() {
	encoding.Register()
	initLog.InitLogger()
	engine, err := core.NewCore(
		os.Getenv("TERM"),
		core.WithFinalizer(func() {
			log.Println("[app] core finalizer called")
		}),
		//core.WithIsIntensiveDraw(true),
	)
	if err != nil {
		log.Printf("[app] error : %v", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithCancel(context.Background())
	if err := engine.Start(ctx); err != nil {
		log.Printf("[app] error : %v", err)
		os.Exit(1)
	}

	rCtx, _ := context.WithCancel(ctx)
	receiver := NewReceiver(rCtx, engine, cancel)
	engine.MouseDispatcher().Register(receiver)
	engine.KeyDispatcher().Register(receiver)
	engine.ResizeDispatcher().Register(receiver)

	<-engine.DyingChan()
	log.Println("[app] done.")
}
