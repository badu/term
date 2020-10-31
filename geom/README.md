#### Line, Column, Row, Rectangle, Grid

All primitives are just a way of indicating a group of `term.Pixel`. 
A `Line` can be a collection of `term.Pixel` which can be horizontal, making it a `Row` or vertical, making it a `Column`.
Using the `LineFromPositions` function from `geom` package, you can have non typical lines, like a diagonal or other arbitrary positioned line.
Now, thinking about `Rect` : it's just a thicker `Line`, isn't it? 
The only particularity of a `Rect` is the `orientation` : a `Rect` can be organized `vertically`, making it an array of `Row`s - or horizontally, making it an array of `Column`s. 
A `Grid` is a collection of `Rect', right ? It is only special as it is "flexible", in terms that it should resize it's children to the desired sizes. 

#### Center of a rectangle

|  Cases               | Column Even (e.g:2)   | Column Odd (e.g:3)    |
|----------------------|-----------------------|-----------------------|
|  Row Even (e.g:2)    | Center Odd  (0,0)     | Center SemiEven (1,0) |
|  Row Odd  (e.g:3)    | Center SemiEven (0,1) | Center Even (1,1)     |

#### Page and it's context.Context

When a `Page` gets created, a cancellable `context.Context` should be passed to it, so it can free the collection of `term.Pixel` it orchestrates. 
Stepping back a little, a `Page` it's just a `Rect` which just fills the entire screen.
Another particularity of a `Page` it's that will listen for `term.ResizeEvent` in order to fill the entire screen. 
Another important one is inheritance of BackgroundColor from the `Page` beneath itself.
Last but least, a `Page` should be able to keep a `stack` of owners of a `term.Pixel` (think `Dropdown` or `Autocomplete` needs to "borrow" some pixels from a neighbour for a while).  
`Page` listens for mouse events and dispatches them only to `Rectangles` that are in that position.


#### 
