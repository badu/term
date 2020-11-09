package geom

import (
	"log"
)

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

// ExtractRows extracts from (width, height) until the end
func (r *Rectangle) ExtractRows(width, height int) []px {
	if r.HasColumns() {
		log.Println("bad call to Rectangle.ExtractRows : orientation is not vertical (rows)")
		return nil
	}
	result := make([]px, 0)
	cHeight := height
	for cHeight < len(r.rows) {
		if r.rows[cHeight] != nil {
			minWidth := 0
			if cHeight == height {
				minWidth = width
			}
			result = append(result, r.rows[cHeight][minWidth:]...)
		}
		cHeight++
	}
	return result
}

// ExtractOnlyRows from (width, height) until (newWidth, newHeight)
func (r *Rectangle) ExtractOnlyRows(width, height, newWidth, newHeight int) []px {
	if r.HasColumns() {
		log.Println("bad call to Rectangle.ExtractOnlyRows : orientation is not vertical (rows)")
		return nil
	}
	result := make([]px, 0)
	cHeight := height
	for cHeight <= newHeight && cHeight < len(r.rows) {
		minWidth := 0
		if cHeight == height {
			minWidth = width
		}
		if r.rows[cHeight] != nil {
			maxWidth := newWidth
			if cHeight != newHeight {
				maxWidth = len(r.rows[cHeight]) - 1
			}
			result = append(result, r.rows[cHeight][minWidth:maxWidth]...)
		}
		cHeight++
	}
	return result
}

func (r *Rectangle) setNewLine(value px, newWidth int, line []px) []px {
	if newWidth >= len(line) {
		line = append(line, make([]px, newWidth-len(line)+1)...)
	}
	line[newWidth] = value
	return line
}

// GrowRows sets a value at (newWidth, newHeight) growing the matrix as necessary
func (r *Rectangle) GrowRows(value px, newWidth int, newHeight int) {
	if r.HasColumns() {
		log.Println("bad call to Rectangle.GrowRows : orientation is not vertical (rows)")
		return
	}
	if newHeight >= len(r.rows) {
		r.rows = append(r.rows, make([][]px, newHeight-len(r.rows)+1)...)
	}
	r.rows[newHeight] = r.setNewLine(value, newWidth, r.rows[newHeight])
}

// ExtractColumns extracts from (height, width) until the end
func (r *Rectangle) ExtractColumns(height, width int) []px {
	if r.HasRows() {
		log.Println("bad call to Rectangle.ExtractColumns : orientation is not horizontal (columns)")
		return nil
	}
	result := make([]px, 0)
	cWidth := width
	for cWidth < len(r.cols) {
		if r.cols[cWidth] != nil {
			minHeight := 0
			if cWidth == width {
				minHeight = height
			}
			result = append(result, r.cols[cWidth][minHeight:]...)
		}
		cWidth++
	}
	return result
}

// ExtractOnlyColumns from (height, width) until (newHeight, newWidth)
func (r *Rectangle) ExtractOnlyColumns(height, width, newHeight, newWidth int) []px {
	if r.HasRows() {
		log.Println("bad call to Rectangle.ExtractOnlyColumns : orientation is not horizontal (columns)")
		return nil
	}
	result := make([]px, 0)
	cWidth := width
	for cWidth <= newWidth && cWidth < len(r.cols) {
		minHeight := 0
		if cWidth == width {
			minHeight = height
		}
		if r.cols[cWidth] != nil {
			maxHeight := newHeight
			if cWidth != newWidth {
				maxHeight = len(r.cols[cWidth]) - 1
			}
			result = append(result, r.cols[cWidth][minHeight:maxHeight]...)
		}
		cWidth++
	}
	return result
}

func (r *Rectangle) setNewColumn(value px, newHeight int, column []px) []px {
	if newHeight >= len(column) {
		column = append(column, make([]px, newHeight-len(column)+1)...)
	}
	column[newHeight] = value
	return column
}

// GrowColumns sets a value at (newHeight, newWidth) growing the matrix as necessary
func (r *Rectangle) GrowColumns(value px, newHeight, newWidth int) {
	if r.HasRows() {
		log.Println("bad call to Rectangle.GrowColumns : orientation is not horizontal (columns)")
		return
	}
	if newWidth >= len(r.cols) {
		r.cols = append(r.cols, make([][]px, newWidth-len(r.cols)+1)...)
	}
	r.cols[newWidth] = r.setNewColumn(value, newHeight, r.cols[newWidth])
}
