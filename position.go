package term

import (
	"strconv"
)

const (
	MinusOneMinusOne = 4294967295
)

type Position struct {
	Row    int
	Column int
	hash   int
}

// Hash - combines a row and a column into a single integer. Note that it doesn't work with very large numbers.
func Hash(column, row int) int {
	return ((column & 0xFFFF) << 16) | (row & 0xFFFF)
}

// UnHash - given a hash built with above function, return the original column and row. Note that negative values are not returned correctly. Use UnHashNeg below.
func UnHash(hash int) (int, int) {
	return hash >> 16, hash & 0xFFFF
}

// UnHashNeg - given a hash built with above function, return the original column and row. Note that negative values are "special"
func UnHashNeg(hash int) (int, int) {
	column := hash >> 16
	if column > 0x8000 { // negative column
		column = -(column ^ 0xFFFF) - 1
	}
	row := hash & 0xFFFF
	if row > 0x8000 { // negative row
		row = -(row ^ 0xFFFF) - 1
	}
	return column, row
}

func NewPosition(column, row int) Position {
	return Position{
		Row:    row,
		Column: column,
		hash:   Hash(column, row),
	}
}

func (p *Position) Hash() int {
	return p.hash
}

func (p *Position) UpdateHash() {
	p.hash = Hash(p.Column, p.Row)
}

func (p Position) String() string {
	return "col:" + strconv.Itoa(p.Column) + ", row:" + strconv.Itoa(p.Row)
}
