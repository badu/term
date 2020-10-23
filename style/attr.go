package style

// Mask represents a mask of text attributes, apart from color.
// Note that support for attributes may vary widely across terminals.
type Mask int

// Attributes are not colors, but affect the display of text.
// They can be combined.
const (
	Bold Mask = 1 << iota
	Blink
	Reverse
	Underline
	Dim
	Italic
	StrikeThrough
	Invalid          // Mark the style or attributes invalid
	None    Mask = 0 // Just normal text.
)
