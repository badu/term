package term

// Size describes something with width and height.
type Size struct {
	Columns int // The number of units along the horizontal axis.
	Rows    int // The number of units along the vertical axis.
}

// NewSize returns a newly allocated Size of the specified dimensions.
func NewSize(columns int, rows int) *Size {
	return &Size{columns, rows}
}

// Add returns a new Size that is the result of increasing the current size by s2 Width and Height.
func (s *Size) Add(s2 *Size) Size {
	return Size{s.Columns + s2.Columns, s.Rows + s2.Rows}
}

// IsZero returns whether the Size has zero width and zero height.
func (s *Size) IsZero() bool {
	return s.Columns == 0 && s.Rows == 0
}

// Max returns a new Size that is the maximum of the current Size and s2.
func (s *Size) Max(s2 Size) *Size {
	maxW := Max(s.Columns, s2.Columns)
	maxH := Max(s.Rows, s2.Rows)

	return NewSize(maxW, maxH)
}

// Min returns a new Size that is the minimum of the current Size and s2.
func (s *Size) Min(s2 Size) *Size {
	minW := Min(s.Columns, s2.Columns)
	minH := Min(s.Rows, s2.Rows)

	return NewSize(minW, minH)
}

// Subtract returns a new Size that is the result of decreasing the current size by s2 Width and Height.
func (s *Size) Subtract(s2 Size) Size {
	return Size{s.Columns - s2.Columns, s.Rows - s2.Rows}
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
