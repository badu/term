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

func NewPosition(column, row int) *Position {
	return &Position{
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

func Width(p1, p2 *Position) int {
	return Abs(p1.Column-p2.Column) + 1
}

func Height(p1, p2 *Position) int {
	return Abs(p1.Row-p2.Row) + 1
}

func Center(p1, p2 *Position) *Position {
	rows := Height(p1, p2)
	columns := Width(p1, p2)
	// assuming the caller knows what it's doing (both should be odd)
	if columns%2 == 1 && rows%2 == 1 { // both are odd - center will be even (except 1,1 which doesn't have a center)
		return NewPosition(columns>>1, rows>>1)
	}
	if columns%2 == 1 && rows%2 == 0 { // cols are odd, rows are even
		return NewPosition(columns>>1, rows>>1-1)
	}
	if columns%2 == 0 && rows%2 == 1 { // cols are even, rows are odd
		return NewPosition(columns>>1-1, rows>>1)
	}
	// worst case, both are even
	return NewPosition(columns>>1-1, rows>>1-1)
}
