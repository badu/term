# A terminal voice, 05:06 am on 8th of October 2020

Pixels. I don't like pixels. Some people call them cells. Like prison cells... 
They are all just abstraction, while me, a terminal, I'm real.

My author here, attempted to write me better. He likes channels. He is more suitable to work on Danube-Black Sea channel.
Or just get a life, instead of using Go channels to experiment on me. 

He reaped the work of a lifetime from other people for that.
Just a bunch of constants and concepts and patterns, if you ask me.

He thinks that something goes faster if you have less allocation. We will see in the end.
He also thinks in separation of concerns, whatever that means. 
It's probably abstraction too. Just like pixels, and I don't like pixels.

# What is this?

Well, everything started while playing with [tcell](https://github.com/gdamore/tcell). 
I've asked myself why there is no separation of concerns inside of it. I mean, the problem it's simple: we have a file reader - for reading input from keyboard and mouse - and a file writer to display stuff.
Everything else are just rules and functionality which terminals provide, a bunch of predefined `[]byte` which are used to send commands.

While thinking about separation of concerns, I've asked myself about `context.Context` : [is it](https://dave.cheney.net/2017/01/26/context-is-for-cancelation) [or it is not](https://dave.cheney.net/2017/08/20/context-isnt-for-cancellation) for cancellation?
It seems that in this context, the `context.Context` is suitable for cancellation usage. The `Application` will provide a cancellable `context.Context` and upon shutdown (think `CTRL+Q` or double `ESC`), it will cancel it.

## Package `info`

This holds various terminal specific commands, which are `registered` and looked up by the `core`.
For using as little allocation as possible, I've created `[]byte` slices for each and every used command, so instead of passing `string` around, we're just using those slices to write to output.
Also, there are caches for `goto` and `colors`, so `[]byte` required to be written in output is cached.
Despite the fact that is has public methods and properties, it's not intended for direct usage, being `core`'s responsibility to orchestrate the writes to output. 

## Package `core`

Creates key, event and resize dispatchers. All events are passed via channels, to avoid allocations.

The `core` constructor supports functional options : `NewCore(termEnv string, options ...Option)`.

Possible options are : 

* `WithFinalizer` - for the case when `Application` want to execute a function prior shutdown.
* `WithWinSizeBufferedChannelSize` - `Application` can set the size of the buffered channel. Defaults to `runtime.NumCPU()`.
* `WithRunesFallback` - `Application` can set the runes fallback upon constructing.
* `WithTrueColor` - a functional option so `Application` can send "disable" to disable true color

Constructor returns an interface. The `Engine` interface:

* `DyingChan() chan struct{}` - it's a channel that needs to be listened inside `Application`, to allow gracefully shutdowns.
* `Start(ctx context.Context) error` - the `Application` call this method with a cancellable context, so `core` will shutdown upon cancellation.              
* `ResizeDispatcher() ResizeDispatcher` - exposes the resize dispatcher, so `Components` can Register themselves to listening events.          
* `KeyDispatcher() KeyDispatcher` - exposes the key dispatcher, so `Components` can Register themselves to listening events.               
* `MouseDispatcher() MouseDispatcher` - exposes the mouse dispatcher, so `Components` can Register themselves to listening events.                         
* `CanDisplay(r rune, checkFallbacks bool) bool` - checks if the rune can be displayed in terminal.
* `CharacterSet() string` - returns the current char set.
* `SetRuneFallback(orig rune, fallback string)` - sets the fallback for a rune.
* `UnsetRuneFallback(orig rune)` - forgets the fallback set above.
* `NumColors() int` - returns the number of colors that terminal supports.
* `Size() *Size` - returns the current size of the window.
* `HasTrueColor() bool` - returns the terminal support for true colors.
* `Palette() []color.Color` - returns the terminal palette
* `Colors() map[color.Color]color.Color` - returns the terminal color map.
* `ActivePixels(pixels []PixelGetter)` - used by `Application` to orchestrate pixels. Pages will be able to have their own set of pixels.
* `Redraw(pixels []PixelGetter)` - temporary exposed for `Application` to force draw.
* `ShowCursor(where *image.Point)` - displays input cursor at coordinates.
* `HideCursor()` - hides input cursor.
* `Cursor() *image.Point` - returns current input cursor position.
* `Clear()` - clears the screen.
 
`ResizeEvent` is an interface has only one method `Size() Size` and Size has - of course - Width and Height properties. 

`Application` must call `Start(ctx context.Context) error` with a cancellable context, in order to use `ActivePixels(pixels []PixelGetter)` registration.

## Package `geom` 

A `Pixel` is an interface which is known by both `Application` and `Engine`. The setters will write to a channel, so `core` can receive the draw request, when a property of the pixel has changed.

A `Pixel` constructor accepts the following functional options:
* `WithBackground` - presets the background color.
* `WithForeground` - presets the foreground (text) color.
* `WithPoint` - is required, and sets the position of the pixel : row is X, column is Y.
* `WithRune` - presets the rune.
* `WithUnicode` - presets the unicode (optional, for that reason it is a pointer).
* `WithAttrs` - presets the `style.Mask` of the `Pixel`.

The `Pixel` interface (includes `PixelGetter` and `PixelSetter` interfaces):

* `DrawCh() chan Pixel` - the channel used by `core` to listen redraw request.
* `Position() *image.Point` - the position of the `Pixel`.
* `BgCol() color.Color` - background color of the `Pixel`.
* `FgCol() color.Color` - foreground color of the `Pixel`.
* `Attrs() style.Mask` - `Pixel` `style.Mask`.
* `Rune() rune` - rune.
* `Width() int` - rune and `Unicode` width.
* `HasUnicode() bool ` - exposes if `Unicode` is a nil pointer or not.
* `Unicode() Unicode` - `Unicode` if declared.
* `Set(r rune, fg, bg color.Color)` - setter for rune and colors. If any of them changed, redraw request gets triggered.
* `SetFgBg(fg, bg color.Color)` - setter for colors. If any of them changed, redraw request gets triggered.
* `SetForeground(c color.Color)` - setter just for foreground.
* `SetBackground(c color.Color)` - setter just for background.
* `SetAttrs(m style.Mask)` - setter for `style.Mask`. 
* `SetUnicode(u Unicode)` - setter for `Unicode`.
* `SetRune(r rune)` - setter for rune.
* `SetAll(bg, fg color.Color, m style.Mask, r rune, u Unicode)` - setter for everything. If any of them changed, redraw request gets triggered.

## Package `key`

* `Register(r KeyListener)` - used by `Components` to register to events listening. Events come via a channel (listener must implement `KeyListener` interface).
* `LifeCycle(ctx context.Context)` - used by `core` upon `Start(ctx context.Context) error` call.
* `HasKey(k Key) bool` - checks if terminal supports the key provided as parameter.   
* `DyingChan() chan struct{}` - `core` listens to this channel to check if dispatcher has finished shutdown, upon context cancellation.
* `InChan() chan []byte` - `core` uses this channel to send input from terminal.
	
## Package `mouse`

* `Register(r MouseListener)`- used by `Components` to register to events listening. Events come via a channel (listener must implement `MouseListener` interface).
* `LifeCycle(ctx context.Context, out *os.File)`- used by `core` upon `Start(ctx context.Context) error` call.
* `Enable()` - enables mouse support
* `Disable()` - disables mouse support
* `ResizeListen() chan ResizeEvent` - `core` uses this channel to send resize events.
* `DyingChan() chan struct{}` - `core` listens to this channel to check if dispatcher has finished shutdown, upon context cancellation.
* `InChan() chan []byte` - `core` uses this channel to send input from terminal.

## Package `style`

* `Palette() []color.Color` - returns the known palette
* `Colors() map[color.Color]color.Color` - returns all colors map
