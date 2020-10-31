package term_test

import (
	"testing"

	"github.com/badu/term"
)

func TestHashCollision(t *testing.T) {
	const (
		columns = 1000
		rows    = 500
	)
	hashMap := make(map[int]struct{})
	for col := -columns; col <= columns; col++ {
		for row := -rows; row <= rows; row++ {
			currHash := term.Hash(col, row)
			if _, has := hashMap[currHash]; has {
				t.Fatalf("error! collision detected at col : %04d row : %04d", col, row)
			}
			hashMap[currHash] = struct{}{}
		}
	}
}

func TestUnhash(t *testing.T) {
	const (
		columns = 1000
		rows    = 500
	)
	// make hashes
	hashMap := make(map[int]*term.Position)
	for col := 0; col <= columns; col++ {
		for row := 0; row <= rows; row++ {
			currHash := term.Hash(col, row)
			hashMap[currHash] = term.NewPosition(col, row)
		}
	}
	// check un-hashes
	for k, v := range hashMap {
		col, row := term.UnHash(k)
		if col != v.Column {
			t.Fatalf("error! col : %04d should be col : %04d ", col, v.Column)
		}
		if row != v.Row {
			t.Fatalf("error! row : %04d should be row : %04d", row, v.Row)
		}
	}
}

func TestUnhashNeg(t *testing.T) {
	const (
		columns = 1000
		rows    = 500
	)
	// make hashes
	hashMap := make(map[int]*term.Position)
	for col := -columns; col <= columns; col++ {
		for row := -rows; row <= rows; row++ {
			currHash := term.Hash(col, row)
			hashMap[currHash] = term.NewPosition(col, row)
		}
	}
	// check un-hashes
	for k, v := range hashMap {
		col, row := term.UnHashNeg(k)
		if col != v.Column {
			t.Fatalf("error! col : %04d should be col : %04d ", col, v.Column)
		}
		if row != v.Row {
			t.Fatalf("error! row : %04d should be row : %04d", row, v.Row)
		}
	}
}

func TestHashUnHashMinusOne(t *testing.T) {
	minusOne := term.Hash(-1, -1)
	if minusOne != term.MinusOneMinusOne {
		t.Fatalf("%d != %d and they should be. did you changed the hash function ?", term.MinusOneMinusOne, minusOne)
	}
	colMinusOne, rowMinusOne := term.UnHashNeg(minusOne)
	if colMinusOne != -1 {
		t.Fatal("column should be -1")
	}
	if rowMinusOne != -1 {
		t.Fatal("row should be -1")
	}
}

func TestCenter(t *testing.T) {
	p1 := term.NewPosition(1, 1)
	p2 := term.NewPosition(4, 4)
	p3 := term.NewPosition(3, 3)
	p4 := term.NewPosition(6, 5)
	if s := term.Center(p1, p1); s.Column != 0 || s.Row != 0 {
		t.Fatal("center should be col=0 row=0", s.Column, s.Row)
	}
	if s := term.Center(p1, p2); s.Column != 1 || s.Row != 1 {
		t.Fatal("center should be col=1 row=1", s.Column, s.Row)
	}
	if s := term.Center(p1, p3); s.Column != 1 || s.Row != 1 {
		t.Fatal("center should be col=1 row=1", s.Column, s.Row)
	}
	if s := term.Center(p2, p3); s.Column != 0 || s.Row != 0 {
		t.Fatal("center should be col=0 row=0", s.Column, s.Row)
	}
	if s := term.Center(p1, p4); s.Column != 2 || s.Row != 2 {
		t.Fatal("center should be col=0 row=0", s.Column, s.Row)
	}
	if s := term.Center(p2, p4); s.Column != 1 || s.Row != 0 {
		t.Fatal("center should be col=0 row=0", s.Column, s.Row)
	}
}
