package main

import (
	"bytes"
	"context"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"io/ioutil"
	"log"
	"os"
	"sync"

	"github.com/badu/term"
	"github.com/badu/term/color"
	"github.com/badu/term/core"
	enc "github.com/badu/term/encoding"
	"github.com/badu/term/geom"
	"github.com/badu/term/key"
	initLog "github.com/badu/term/log"
)

type page struct {
	sync.Once
	incomingMouse     chan term.MouseEvent  // We need a channel on which we will listen for incoming events
	incomingKey       chan term.KeyEvent    // We need a channel on which we will listen for incoming events
	incomingResize    chan term.ResizeEvent // We need a channel on which we will listen for incoming events
	died              chan struct{}         // this is a buffered channel of size one
	engine            term.Engine           //
	image             [][][2]color.Color
	pixels            [][]term.Pixel
	imageOffsetRow    int
	imageOffsetColumn int
	imageWidth        int
	imageHeight       int
	size              *term.Size
}

func (p *page) MouseListen() chan term.MouseEvent {
	return p.incomingMouse
}

func (p *page) KeyListen() chan term.KeyEvent {
	return p.incomingKey
}

func (p *page) ResizeListen() chan term.ResizeEvent {
	return p.incomingResize
}

func (p *page) DyingChan() chan struct{} {
	return p.died
}

func (p *page) adjustAndRegisterPixels(ev term.ResizeEvent) {
	p.size = ev.Size()
	p.init()
}

func (p *page) init() {
	getters := make([]term.PixelGetter, 0)
	p.pixels = make([][]term.Pixel, p.size.Width)
	imageColumn := p.imageOffsetColumn
	for column := 0; column < p.size.Width; column++ {
		imageRow := p.imageOffsetRow
		p.pixels[column] = make([]term.Pixel, p.size.Height)
		for row := 0; row < p.size.Height; row++ {
			fg := geom.WithForeground(p.image[imageColumn][imageRow][0])
			bg := geom.WithBackground(p.image[imageColumn][imageRow][1])
			px, _ := geom.NewPixel(geom.WithRune('▀'), fg, bg, geom.WithPoint(image.Point{X: row, Y: column})) // row is X, column is Y
			getters = append(getters, px)
			p.pixels[column][row] = px
			imageRow++
		}
		imageColumn++
	}
	p.engine.ActivePixels(getters)
}

func (p *page) redraw() {
	imageRow := p.imageOffsetRow
	for row := 0; row < p.size.Height; row++ {
		imageColumn := p.imageOffsetColumn
		for column := 0; column < p.size.Width; column++ {
			p.pixels[column][row].SetFgBg(p.image[imageColumn][imageRow][0], p.image[imageColumn][imageRow][1])
			imageColumn++
		}
		imageRow++
	}
	p.engine.HideCursor()
}

// listen listens for incoming events
func (p *page) lifeCycle(ctx context.Context, cancel func()) {
	p.Once.Do(
		func() {
			p.size = p.engine.Size()
			p.engine.HideCursor() // hide cursor
			p.engine.Clear()      // clear
			log.Printf("initial screen size w = %03d h = %03d", p.size.Width, p.size.Height)
			imageWidth := len(p.image)
			imageHeight := len(p.image[0])
			p.imageOffsetRow = (imageHeight - p.size.Height) / 2
			p.imageOffsetColumn = (imageWidth - p.size.Width) / 2
			log.Printf("loaded image size %04d,%04d", imageWidth, imageHeight)
			log.Printf("displaying %04d,%04d -> %04d,%04d", p.imageOffsetRow, p.imageOffsetColumn, p.imageOffsetRow+p.size.Width, p.imageOffsetColumn+p.size.Height)
			p.init()
			go func() {
				escapeCount := 0
				for {
					select {
					case <-ctx.Done():
						log.Println("[app] context is done.")
						p.died <- struct{}{}
						return
					case ev := <-p.incomingKey:
						switch ev.Key() {
						case key.Escape:
							escapeCount++
							if escapeCount > 1 {
								log.Println("[app] waiting for engine to finalize correctly")
								cancel()
								return
							}
						case key.Up:
							log.Printf("going up 10 px ? %04d,%04d", p.imageOffsetRow-10, imageHeight)
							if p.imageOffsetRow-10 > 0 {
								p.imageOffsetRow -= 10
								p.redraw()
							}
						case key.Down:
							log.Printf("going down 10 px ? %04d,%04d", p.imageOffsetRow+p.size.Height+10, imageHeight)
							if p.imageOffsetRow+p.size.Height+10 < imageHeight {
								p.imageOffsetRow += 10
								p.redraw()
							}
						case key.Left:
							log.Printf("going left 10 px ? %04d,%04d", p.imageOffsetColumn-10, imageWidth)
							if p.imageOffsetColumn-10 > 0 {
								p.imageOffsetColumn -= 10
								p.redraw()
							}
						case key.Right:
							log.Printf("going right 10 px ? %04d,%04d", p.imageOffsetColumn+p.size.Width+10, imageWidth)
							if p.imageOffsetColumn+p.size.Width+10 < imageWidth {
								p.imageOffsetColumn += 10
								p.redraw()
							}
						}
					case ev := <-p.incomingResize:
						p.adjustAndRegisterPixels(ev)
					}

				}
			}()
		},
	)
}

// creates the "Application"
func NewPage(ctx context.Context, engine term.Engine, cancel func(), image [][][2]color.Color, imageWidth, imageHeight int) *page {
	result := &page{
		died:           make(chan struct{}),        // init of died channel, a buffered channel of exactly one
		incomingMouse:  make(chan term.MouseEvent), // init of incoming channel
		incomingKey:    make(chan term.KeyEvent),   // init of incoming channel
		incomingResize: make(chan term.ResizeEvent),
		engine:         engine,
		image:          image,
		imageWidth:     imageWidth,
		imageHeight:    imageHeight,
	}
	result.lifeCycle(ctx, cancel)
	return result
}

// gather each pixel color but each image row is condensed into one, which holds the two colors
func makeColors(r io.Reader) ([][][2]color.Color, int, int, error) {
	img, format, err := image.Decode(r)
	if err != nil {
		log.Printf("decode error : %v %s\n", err, format)
		return nil, 0, 0, err
	}
	bounds := img.Bounds()
	width := bounds.Max.X
	height := bounds.Max.Y
	result := make([][][2]color.Color, width)
	for col := 0; col < width; col++ {
		result[col] = make([][2]color.Color, height/2+1)
		rrow := 0
		for row := 0; row < height; row += 2 {
			result[col][rrow] = [2]color.Color{}
			c1 := color.NewRGBAColor(img.At(col, row).RGBA())
			result[col][rrow][0] = c1
			if !c1.IsRGB() {
				return nil, 0, 0, fmt.Errorf("error : color 0x%06X is not RGB", c1.Hex())
			}
			if row+1 < height {
				c2 := color.NewRGBAColor(img.At(col, row+1).RGBA())
				result[col][rrow][1] = c2
				if !c2.IsRGB() {
					return nil, 0, 0, fmt.Errorf("error : color 0x%06X is not RGB", c1.Hex())
				}
			}
			rrow++
		}
	}
	return result, width, height/2 + 1, nil
}

const usage = `piximage [pattern|url]
Examples:
    piximage path/to/image.jpg
    piximage https://example.com/image.jpg`

func main() {
	if len(os.Args) == 1 {
		fmt.Println(usage)
		os.Exit(1)
	}
	if os.Args[1] == "-h" || os.Args[1] == "--help" {
		fmt.Println(usage)
		os.Exit(0)
	}
	var err error
	var colors [][][2]color.Color
	url := os.Args[1]
	w, h := 0, 0
	log.Printf("loading %q\n", url)

	if file, err := ioutil.ReadFile(url); err != nil {
		fmt.Printf("error loading image %q : %v", url, err)
	} else {
		colors, w, h, err = makeColors(bytes.NewReader(file))
		if err != nil {
			fmt.Printf("error reading image %q to colors : %v", url, err)
		}
		fmt.Printf("%d bytes were read. [w=%d,h=%d]\n", len(file), w, h)
	}

	enc.Register()
	initLog.InitLogger()
	engine, err := core.NewCore(
		os.Getenv("TERM"),
		core.WithFinalizer(func() {
			log.Println("[app] core finalizer called")
		}),
		core.WithIsIntensiveDraw(true),
	)
	if err != nil {
		log.Printf("error : %v", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithCancel(context.Background())
	if err := engine.Start(ctx); err != nil {
		log.Printf("error : %v", err)
		os.Exit(1)
	}
	log.Printf("image %s has size %04d,%04d", url, w, h)
	pCtx, _ := context.WithCancel(ctx)
	page := NewPage(pCtx, engine, cancel, colors, w, h)
	engine.KeyDispatcher().Register(page)
	engine.ResizeDispatcher().Register(page)
	log.Printf("[app] registered listeners.")
	<-engine.DyingChan()
	log.Println("[app] done.")
}
