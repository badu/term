package geom

import (
	"log"

	"github.com/badu/term"
	"github.com/badu/term/style"
)

// RootRectangle interface
type RootRectangle interface {
	Orientation() style.Orientation
	HasRows() bool
	NumRows() int
	Rows() PixelsMatrix
	Row(index int) Pixels
	HasColumns() bool
	NumColumns() int
	Columns() PixelsMatrix
	Column(index int) Pixels
}

type root struct {
	orientation  style.Orientation // orientation dictates pixel slices above (rows or cols). Default orientation is style.Vertical
	topCorner    *term.Position    // The current top corner of the Rectangle
	bottomCorner *term.Position    // The current top corner of the Rectangle
	pxs          map[int]px        // map[position_hash]pixel, for fast access to pixels
	rows         PixelsMatrix      // when organized by rows, for fast access to rows
	cols         PixelsMatrix      // when organized by cols, for fast access to columns
}

// Orientation
func (r *root) Orientation() style.Orientation {
	return r.orientation
}

// HasRows
func (r *root) HasRows() bool {
	return r.orientation == style.Vertical
}

// HasColumns
func (r *root) HasColumns() bool {
	return r.orientation == style.Horizontal
}

// Row returns the row of pixels at index (absolute, starting with zero)
func (r *root) Row(index int) Pixels {
	if r.HasColumns() {
		if index <= 0 {
			if Debug {
				log.Println("bad call to Rectangle.Row : bad index")
			}
			return nil
		}
		if len(r.cols) <= 0 {
			if Debug {
				log.Println("bad call to Rectangle.Row : horizontal orientation, but columns are empty")
			}
		}
		var result Pixels
		for _, column := range r.cols {
			for idx, row := range column {
				if idx == index {
					result = append(result, row)
				}
			}
		}
		return result
	}
	// vertical orientation
	if index <= 0 {
		if Debug {
			log.Println("bad call to Rectangle.Row : bad index")
		}
		return nil
	}
	if index-1 >= len(r.rows) {
		if Debug {
			log.Println("bad call to Rectangle.Row : index outside number of rows")
		}
		return nil
	}
	return r.rows[index-1]
}

// NumRows - depends on orientation
func (r *root) NumRows() int {
	if r.HasColumns() {
		if len(r.cols) <= 0 {
			if Debug {
				log.Println("bad call to Rectangle.NumRows : cannot calculate number of rows (columns are empty)")
			}
			return 0
		}
		return len(r.cols[0])
	}
	return len(r.rows)
}

// Rows
func (r *root) Rows() PixelsMatrix {
	if r.HasColumns() {
		if len(r.cols) <= 0 {
			if Debug {
				log.Println("bad call to Rectangle.Rows : cannot return rotated (columns are empty)")
			}
			return nil
		}
		return rotate(r.cols)
	}
	return r.rows
}

// Column returns the column of pixels at index (absolute, starting with zero)
func (r *root) Column(index int) Pixels {
	if r.HasRows() {
		// vertical orientation column
		if index <= 0 {
			if Debug {
				log.Println("bad call to Rectangle.Column : bad index")
			}
			return nil
		}
		if len(r.rows) <= 0 {
			if Debug {
				log.Println("bad call to Rectangle.Column : vertical orientation, but rows are empty")
			}
		}
		var result Pixels
		for _, row := range r.rows {
			for idx, column := range row {
				if idx == index {
					result = append(result, column)
				}
			}
		}
		return result
	}
	// horizontal direction
	if index <= 0 {
		if Debug {
			log.Println("bad call to Rectangle.Column : bad index")
		}
		return nil
	}
	if index-1 > len(r.cols) {
		if Debug {
			log.Println("bad call to Rectangle.Column : index outside number of columns")
		}
		return nil
	}
	return r.cols[index-1]
}

// NumColumns - depends on orientation
func (r *root) NumColumns() int {
	if r.HasRows() {
		if len(r.rows) <= 0 {
			if Debug {
				log.Println("bad call to Rectangle.NumColumns : cannot calculate number of columns (rows are empty)")
			}
			return 0
		}
		return len(r.rows[0])
	}
	return len(r.cols)
}

// Columns
func (r *root) Columns() PixelsMatrix {
	if r.HasRows() {
		if len(r.rows) <= 0 {
			if Debug {
				log.Println("bad call to Rectangle.Columns : cannot return rotated (rows are empty)")
			}
			return nil
		}
		return rotate(r.rows)
	}
	return r.cols
}

// resize on a rectangle updates the new size of this object.
// If it has a stroke width this will cause it to Refresh.
func (r *root) resize(size *term.Size) {
	switch r.orientation {
	case style.Vertical:
		if len(r.rows) == 0 {
			if Debug {
				log.Println("bad root rectangle (no height)")
			}
			return
		}
		if len(r.rows[0]) == 0 {
			if Debug {
				log.Println("bad root rectangle (no width)")
			}
			return
		}
		numRows, numColumns := len(r.rows), len(r.rows[0])
		if numColumns == size.Columns && numRows == size.Rows {
			return
		}
		r.verticalResize(size.Columns, size.Rows)
	case style.Horizontal:
		if len(r.cols) == 0 {
			if Debug {
				log.Println("bad root rectangle (no width)")
			}
			return
		}
		if len(r.cols[0]) == 0 {
			if Debug {
				log.Println("bad root rectangle (no height)")
			}
			return
		}
		numColumns, numRows := len(r.cols), len(r.cols[0])
		if numColumns == size.Columns && numRows == size.Rows {
			return
		}
		r.horizontalResize(size.Columns, size.Rows)
	}
}

// startup
func (r *root) startup(engine term.Engine) {
	r.pxs = make(map[int]px)
	pixels := make([]term.PixelGetter, 0)
	columns := r.bottomCorner.Column - r.topCorner.Column
	rows := r.bottomCorner.Row - r.topCorner.Row
	switch r.orientation {
	case style.Vertical:
		r.rows = make(PixelsMatrix, rows)
		for row := 0; row < rows; row++ {
			r.rows[row] = make(Pixels, columns)
			for column := 0; column < columns; column++ {
				pixel := newPixel(column, row)
				r.pxs[pixel.PositionHash()] = pixel
				pixels = append(pixels, &pixel)
				r.rows[row][column] = pixel
			}
		}
	case style.Horizontal:
		r.cols = make(PixelsMatrix, columns)
		for column := 0; column < columns; column++ {
			r.cols[column] = make(Pixels, rows)
			for row := 0; row < rows; row++ {
				pixel := newPixel(column, row)
				r.pxs[pixel.PositionHash()] = pixel
				pixels = append(pixels, &pixel)
				r.cols[column][row] = pixel
			}
		}
	}
	engine.ActivePixels(pixels)
}

func (r *root) horizontalResize(newRows, newColumns int) { // Note : the params are inverted (rows, columns)
	currRows := len(r.cols)
	if currRows < newRows { // grow or shrink our rows
		r.cols = append(r.cols, make(PixelsMatrix, newRows-currRows)...)
	} else if currRows > newRows {
		r.cols = r.cols[:newRows]
		// TODO : announce dead pixels
	}
	for column := range r.cols { // iterate through our columns to grow or shrink their rows
		currRows := len(r.cols[column])
		if currRows < newColumns { // grow or shrink our rows
			r.cols[column] = append(r.cols[column], make([]Cell, newColumns-currRows)...)
			for row := currRows; row < newColumns; row++ {
				pixel := newPixel(column, row)
				r.pxs[pixel.PositionHash()] = pixel
				r.cols[column][row] = pixel
			}
		} else if currRows > newColumns {
			r.cols[column] = r.cols[column][:newColumns]
			// TODO : announce dead pixels
		}
	}
}

func (r *root) verticalResize(newColumns, newRows int) { // Note : the params are inverted (columns, rows)
	currRows := len(r.rows)
	if currRows < newRows { // grow or shrink our rows
		r.rows = append(r.rows, make(PixelsMatrix, newRows-currRows)...)
	} else if currRows > newRows {
		r.rows = r.rows[:newRows]
		// TODO : announce dead pixels
	}
	for row := range r.rows { // iterate through our rows to grow or shrink their columns
		currColumn := len(r.rows[row])
		if currColumn < newColumns { // grow or shrink our columns
			r.rows[row] = append(r.rows[row], make([]Cell, newColumns-currColumn)...)
			for column := currColumn; column < newColumns; column++ {
				pixel := newPixel(column, row)
				r.pxs[pixel.PositionHash()] = pixel
				r.rows[row][column] = pixel
			}
		} else if currColumn > newColumns {
			r.rows[row] = r.rows[row][:newColumns]
			// TODO : announce dead pixels
		}
	}
}

func rotate(matrix PixelsMatrix) PixelsMatrix {
	m, n := len(matrix), len(matrix[0])
	result := make(PixelsMatrix, n)
	for i := 0; i < n; i++ {
		result[i] = make(Pixels, m)
	}
	for i := 0; i < m; i++ {
		for j := 0; j < n; j++ {
			result[j][m-i-1] = matrix[i][j]
		}
	}
	return result
}
