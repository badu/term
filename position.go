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

func Hash(column, row int) int {
	return ((column & 0xFFFF) << 16) | (row & 0xFFFF)
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
