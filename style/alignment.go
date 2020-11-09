package style

import (
	"strings"
)

// Alignment represents the alignment of an object, and consists of  either or both of horizontal and vertical alignment.
type Alignment int

const (
	NoAlignment Alignment           = 1 << iota // NoAlignment indicates missing alignment info.
	Left                                        // Left indicates alignment on the left edge.
	HCenter                                     // HCenter indicates horizontally centered.
	Right                                       // Right indicates alignment on the right edge.
	Top                                         // Top indicates alignment on the top edge.
	VCenter                                     // VCenter indicates vertically centered.
	Bottom                                      // Bottom indicates alignment on the bottom edge.
	Begin       = Left | Top                    // Begin indicates alignment at the top left corner (default)
	End         = Right | Bottom                // End indicates alignment at the bottom right corner.
	Middle      = HCenter | VCenter             // Middle indicates full centering.
)

// Stringer implementation
func (a Alignment) String() string {
	var sb strings.Builder
	switch a {
	// orientation must be Horizontal
	case Left:
		sb.WriteString("left-edge")
	case HCenter:
		sb.WriteString("horizontal-center")
	case Right:
		sb.WriteString("right-edge")
		// orientation must be Vertical
	case Top:
		sb.WriteString("top-edge")
	case VCenter:
		sb.WriteString("vertical-center")
	case Bottom:
		sb.WriteString("bottom-edge")
		// orientation must be both Horizontal and Vertical
	case Begin:
		sb.WriteString("top-left-corner")
	case End:
		sb.WriteString("bottom-right-corner")
	case Middle:
		sb.WriteString("full-center")
	default:
		sb.WriteString("not-set")
	}
	return sb.String()
}

// Orientation represents the direction of a widget or layout.
type Orientation int

const (
	NoOrientation Orientation             = 1 << iota // NoOrientation indicates orientation has not been indicated.
	Horizontal                                        // Horizontal indicates left to right orientation.
	Vertical                                          // Vertical indicates top to bottom orientation.
	Absolute      = Vertical | Horizontal             // Absolute ???
)

// Stringer implementation
func (o Orientation) String() string {
	var sb strings.Builder
	switch o {
	case Horizontal:
		sb.WriteString("columns")
	case Vertical:
		sb.WriteString("rows")
	case Absolute:
		sb.WriteString("absolute")
	default:
		sb.WriteString("not-set")
	}
	return sb.String()
}
