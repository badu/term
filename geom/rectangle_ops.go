package geom

// Empty returns if the Rectangle is zero rows and zero columns
func (r *Rectangle) Empty() bool {
	return r.topCorner.Row >= r.bottomCorner.Row || r.topCorner.Column >= r.bottomCorner.Column
}

// Union returns the smallest rectangle that contains both r and s.
func (r *Rectangle) Union(s *Rectangle) *Rectangle {
	if r.Empty() {
		return s
	}
	if s.Empty() {
		return r
	}
	if r.topCorner.Row > s.topCorner.Row {
		r.topCorner.Row = s.topCorner.Row
	}
	if r.topCorner.Column > s.topCorner.Column {
		r.topCorner.Column = s.topCorner.Column
	}
	if r.bottomCorner.Row < s.bottomCorner.Row {
		r.bottomCorner.Row = s.bottomCorner.Row
	}
	if r.bottomCorner.Column < s.bottomCorner.Column {
		r.bottomCorner.Column = s.bottomCorner.Column
	}
	return r
}

// Intersect returns the largest rectangle contained by both r and s. If the
// two rectangles do not overlap then the zero rectangle will be returned.
func (r *Rectangle) Intersect(s *Rectangle) *Rectangle {
	if r.topCorner.Column < s.topCorner.Column {
		r.topCorner.Column = s.topCorner.Column
	}
	if r.topCorner.Row < s.topCorner.Row {
		r.topCorner.Row = s.topCorner.Row
	}
	if r.bottomCorner.Column > s.bottomCorner.Column {
		r.bottomCorner.Column = s.bottomCorner.Column
	}
	if r.bottomCorner.Row > s.bottomCorner.Row {
		r.bottomCorner.Row = s.bottomCorner.Row
	}
	// Letting r0 and s0 be the values of r and s at the time that the method is called, this next line is equivalent to:
	// if max(r0.topCorner.Column, s0.topCorner.Column) >= min(r0.bottomCorner.Column, s0.bottomCorner.Column) || likewiseForRow { etc }
	if r.Empty() {
		return &Rectangle{}
	}
	return r
}

// Overlaps reports whether r and s have a non-empty intersection.
func (r *Rectangle) Overlaps(s *Rectangle) bool {
	return !r.Empty() && !s.Empty() && r.topCorner.Column < s.bottomCorner.Column && s.topCorner.Column < r.bottomCorner.Column && r.topCorner.Row < s.bottomCorner.Row && s.topCorner.Row < r.bottomCorner.Row
}

// In reports whether every point in r is in s.
func (r *Rectangle) In(s *Rectangle) bool {
	if r.Empty() {
		return true
	}
	// Note that r.bottomCorner is an exclusive bound for r, so that r.In(s)/ does not require that r.bottomCorner.In(s).
	return s.topCorner.Column <= r.topCorner.Column && r.bottomCorner.Column <= s.bottomCorner.Column && s.topCorner.Row <= r.topCorner.Row && r.bottomCorner.Row <= s.bottomCorner.Row
}
