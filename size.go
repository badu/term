package term

// Size describes something with width and height.
type Size struct {
	Width  int // The number of units along the X axis.
	Height int // The number of units along the Y axis.
}

// NewSize returns a newly allocated Size of the specified dimensions.
func NewSize(w int, h int) Size {
	return Size{w, h}
}

// Add returns a new Size that is the result of increasing the current size by s2 Width and Height.
func (s Size) Add(s2 Size) Size {
	return Size{s.Width + s2.Width, s.Height + s2.Height}
}

// IsZero returns whether the Size has zero width and zero height.
func (s Size) IsZero() bool {
	return s.Width == 0 && s.Height == 0
}

// Max returns a new Size that is the maximum of the current Size and s2.
func (s Size) Max(s2 Size) Size {
	maxW := Max(s.Width, s2.Width)
	maxH := Max(s.Height, s2.Height)

	return NewSize(maxW, maxH)
}

// Min returns a new Size that is the minimum of the current Size and s2.
func (s Size) Min(s2 Size) Size {
	minW := Min(s.Width, s2.Width)
	minH := Min(s.Height, s2.Height)

	return NewSize(minW, minH)
}

// Subtract returns a new Size that is the result of decreasing the current size by s2 Width and Height.
func (s Size) Subtract(s2 Size) Size {
	return Size{s.Width - s2.Width, s.Height - s2.Height}
}

// Min returns the smaller of the passed values.
func Min(x, y int) int {
	if x < y {
		return x
	}
	return y
}

// Max returns the larger of the passed values.
func Max(x, y int) int {
	if x > y {
		return x
	}
	return y
}

// Abs returns the absolute value
func Abs(a int) int {
	if a <= 0 {
		return -a
	}
	return a
}
