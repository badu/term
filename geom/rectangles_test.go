package geom_test

import (
	"context"
	"testing"
	"time"

	"github.com/badu/term"
	. "github.com/badu/term/geom"
	"github.com/badu/term/style"
)

const (
	defaultCols = 146
	defaultRows = 36
)

func TestTree(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	aq := make(chan term.Position)
	fakeEngine := NewFakeEngine(t, 132, 43)
	fakeEngine.Start(ctx)
	mainTree := NewTree(ctx, defaultCols, defaultRows, aq, style.NoOrientation, fakeEngine) // oriented as default, style.Vertical
	r1, _ := NewRectangle(ctx, WithAcquisitionChan(aq), WithHeight(20))                     // 20% height, 100% width (optional) of the parent (root)
	r2, _ := NewRectangle(ctx, WithAcquisitionChan(aq), WithHeight(60))                     // 60% height, 100% width (optional) of the parent (root)
	r3, _ := NewRectangle(ctx, WithAcquisitionChan(aq), WithHeight(20))                     // 20% height, 100% width (optional) of the parent (root)
	mainTree.Register(r1, r2, r3)
	r4, _ := NewRectangle(ctx, WithAcquisitionChan(aq), WithWidth(30), WithHeight(30)) // 30% width and height (level+1)
	mainTree.Register(r4)
	fakeEngine.SetSize(100, 40)
	cancel()
}

func print(t *testing.T, r RootRectangle) {

	columns, rows := r.Columns(), r.Rows()
	t.Logf("oriented %s %d columns and %d rows", r.Orientation(), len(columns), len(rows))
	for _, rows := range columns {
		for _, row := range rows {
			if row.HasUnicode() {
				cv := row.Unicode()
				t.Logf("column %s", string(*cv))
			} else {
				t.Log("unicode not set (forgot setting Debug const?)")
			}
		}
	}
	for _, row := range rows {
		for _, column := range row {
			if column.HasUnicode() {
				cv := column.Unicode()
				t.Logf("row %s", string(*cv))
			} else {
				t.Log("unicode not set (forgot setting Debug const?)")
			}
		}
	}
}

func TestVResize(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	fakeEngine := NewFakeEngine(t, 132, 43)
	fakeEngine.Start(ctx)
	page, _ := NewPage(ctx, WithEngine(fakeEngine))
	fakeEngine.SetSize(2, 2)
	<-time.After(500 * time.Millisecond)
	if page.NumRows() != 2 || page.NumColumns() != 2 {
		t.Fatalf("expecting 2 rows and 2 columns : got %d rows and %d columns", page.NumRows(), page.NumColumns())
	}
	fakeEngine.SetSize(3, 3)
	<-time.After(500 * time.Millisecond)
	if page.NumRows() != 3 || page.NumColumns() != 3 {
		t.Fatalf("expecting 3 rows and 3 columns : got %d rows and %d columns", page.NumRows(), page.NumColumns())
	}
	fakeEngine.SetSize(3, 5)
	<-time.After(500 * time.Millisecond)
	if page.NumRows() != 5 || page.NumColumns() != 3 {
		t.Fatalf("expecting 5 rows and 3 columns : got %d rows and %d columns", page.NumRows(), page.NumColumns())
	}
	fakeEngine.SetSize(5, 5)
	<-time.After(500 * time.Millisecond)
	if page.NumRows() != 5 || page.NumColumns() != 5 {
		t.Fatalf("expecting 5 rows and 5 columns : got %d rows and %d columns", page.NumRows(), page.NumColumns())
	}
	fakeEngine.SetSize(5, 3)
	<-time.After(500 * time.Millisecond)
	if page.NumRows() != 3 || page.NumColumns() != 5 {
		t.Fatalf("expecting 3 rows and 5 columns : got %d rows and %d columns", page.NumRows(), page.NumColumns())
	}
	print(t, page)
	cancel()
}

func TestHResize(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	fakeEngine := NewFakeEngine(t, 132, 43)
	fakeEngine.Start(ctx)
	page, _ := NewPage(ctx, WithPageOrientation(style.Horizontal), WithEngine(fakeEngine))
	fakeEngine.SetSize(5, 4)
	if page.NumRows() != 4 || page.NumColumns() != 5 {
		t.Fatalf("expecting 4 rows and 5 columns : got %d rows and %d columns", page.NumRows(), page.NumColumns())
	}
	fakeEngine.SetSize(5, 7)
	<-time.After(500 * time.Millisecond) // wait for goroutines to act
	if page.NumRows() != 7 || page.NumColumns() != 5 {
		t.Fatalf("expecting 7 rows and 5 columns : got %d rows and %d columns", page.NumRows(), page.NumColumns())
	}
	fakeEngine.SetSize(7, 7)
	<-time.After(500 * time.Millisecond) // wait for goroutines to act
	if page.NumRows() != 7 || page.NumColumns() != 7 {
		t.Fatalf("expecting 7 rows and 7 columns : got %d rows and %d columns", page.NumRows(), page.NumColumns())
	}
	fakeEngine.SetSize(5, 7)
	<-time.After(500 * time.Millisecond) // wait for goroutines to act
	if page.NumRows() != 7 || page.NumColumns() != 5 {
		t.Fatalf("expecting 7 rows and 5 columns : got %d rows and %d columns", page.NumRows(), page.NumColumns())
	}
	fakeEngine.SetSize(5, 5)
	<-time.After(500 * time.Millisecond) // wait for goroutines to act
	if page.NumRows() != 5 || page.NumColumns() != 5 {
		t.Fatalf("expecting 5 rows and 5 columns : got %d rows and %d columns", page.NumRows(), page.NumColumns())
	}
	fakeEngine.SetSize(1, 3)
	<-time.After(500 * time.Millisecond) // wait for goroutines to act
	if page.NumRows() != 3 || page.NumColumns() != 1 {
		t.Fatalf("expecting 2 rows and 1 column : got %d rows and %d columns", page.NumRows(), page.NumColumns())
	}
	print(t, page)
	cancel()
}
