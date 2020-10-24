package style

import (
	"github.com/badu/term/color"
)

// Theme defines the colors used across Applications.
type Theme struct {
	BorderColor             color.Color // Color for borders.
	MainBackgroundColor     color.Color // Main background color for primitives.
	HoverBackgroundColor    color.Color // Background color for hovered elements.
	SelectedBackgroundColor color.Color // Background color for selected elements.
	InverseBackgroundColor  color.Color //
	TextColor               color.Color // Main text color.
	HoverTextColor          color.Color // Hovered text color.
	SelectedTextColor       color.Color // Selected text color.
	InverseTextColor        color.Color //
}
