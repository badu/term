package style

// Alignment represents the alignment of an object, and consists of  either or both of horizontal and vertical alignment.
type Alignment int

const (
	Left    Alignment           = 1 << iota // Left indicates alignment on the left edge.
	HCenter                                 // HCenter indicates horizontally centered.
	Right                                   // Right indicates alignment on the right edge.
	Top                                     // Top indicates alignment on the top edge.
	VCenter                                 // VCenter indicates vertically centered.
	Bottom                                  // Bottom indicates alignment on the bottom edge.
	Begin   = Left | Top                    // Begin indicates alignment at the top left corner.
	End     = Right | Bottom                // End indicates alignment at the bottom right corner.
	Middle  = HCenter | VCenter             // Middle indicates full centering.
)

// Orientation represents the direction of a widget or layout.
type Orientation int

const (
	Unset      Orientation = 1 << iota
	Horizontal             // Horizontal indicates left to right orientation.
	Vertical               // Vertical indicates top to bottom orientation.
)
