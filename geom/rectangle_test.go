package geom_test

import (
	"context"
	"testing"

	"github.com/badu/term/color"
	"github.com/badu/term/geom"
	"github.com/badu/term/style"
)

func TestRectangleValid(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	_, err := geom.NewRectangle(ctx)
	if err == nil {
		t.Fatal("error : rectangle should return error")
	}
	_, err = geom.NewRectangle(ctx, testAcquisitionChan())
	if err == nil {
		t.Fatal("error : rectangle should return error")
	}
	_, err = geom.NewRectangle(ctx, testAcquisitionChan(), geom.WithTopCorner(1, 1))
	if err == nil {
		t.Fatal("error : rectangle should return error")
	}
	// a one pixel rectangle
	r, err := geom.NewRectangle(ctx, testAcquisitionChan(), geom.WithTopCorner(1, 1), geom.WithBottomCorner(1, 1))
	if err != nil {
		t.Fatalf("error : there should be no error - %v", err)
	}
	if r.Bg() != color.Default {
		t.Fatal("error : background color should be default")
	}
	if r.Fg() != color.Default {
		t.Fatal("error : foreground color should be default")
	}
	// since it's a pixel, width and height should be one
	if r.Size().Width != 1 || r.Size().Height != 1 {
		t.Fatal("error : width and height should be 1")
	}
	// same functionality as in Size()
	if r.Width() != 1 {
		t.Fatal("error : width should be 1")
	}
	if r.Height() != 1 {
		t.Fatal("error : height should be 1")
	}
	// default alignment is style.Vertical
	if px := r.Column(1); px != nil {
		t.Fatal("error : column should be empty (we have default horizontal orientation)")
	}
	// same style.Vertical check
	if px := r.Row(1); px == nil {
		t.Fatal("error : we should have a row")
	}
	// same style.Vertical check
	if px := r.Row(1); len(px) != 1 {
		t.Fatalf("error : row should have one pixel but has %d pixels", len(px))
	}
	// check for center of a single pixel (doesn't have a center => col = 0 ,row = 0)
	center := r.Center()
	if center.Row != 0 {
		t.Fatalf("center (row) should be 0, but it's %d", center.Row)
	}
	if center.Column != 0 {
		t.Fatalf("center (column) should be 0, but it's %d", center.Column)
	}
	// checking for the center of 4 cols x 4 rows Rectangle
	r5, err := geom.NewRectangle(ctx, testAcquisitionChan(), geom.WithTopCorner(1, 1), geom.WithBottomCorner(4, 4))
	if err != nil {
		t.Fatalf("error : there should be no error - %v", err)
	}
	if r5.Size().Width != 4 || r5.Size().Height != 4 {
		t.Fatal("error : width and height should be 4")
	}
	center = r5.Center()
	if center.Row != 1 {
		t.Fatalf("center (row) should be 1, but it's %d", center.Row)
	}
	if center.Column != 1 {
		t.Fatalf("center (column) should be 1, but it's %d", center.Column)
	}
	// now checking for the center of 3 cols x 3 rows Rectangle
	r4, err := geom.NewRectangle(ctx, testAcquisitionChan(), geom.WithTopCorner(1, 1), geom.WithBottomCorner(3, 3))
	if err != nil {
		t.Fatalf("error : there should be no error - %v", err)
	}
	if r4.Size().Width != 3 || r4.Size().Height != 3 {
		t.Fatal("error : width and height should be 3")
	}
	center = r4.Center()
	if center.Row != 1 {
		t.Fatalf("center (row) should be 1, but it's %d", center.Row)
	}
	if center.Column != 1 {
		t.Fatalf("center (column) should be 1, but it's %d", center.Column)
	}
	// now checking for the center of Line (3 cols x 1 row)
	r2, err := geom.NewRectangle(ctx, testAcquisitionChan(), geom.WithTopCorner(1, 1), geom.WithBottomCorner(3, 1))
	if err != nil {
		t.Fatalf("error : there should be no error - %v", err)
	}
	if r2.Size().Width != 3 || r2.Size().Height != 1 {
		t.Fatal("error : width should be 3 and height should be 1")
	}
	center = r2.Center()
	if center.Row != 0 {
		t.Fatalf("center (row) should be 0, but it's %d", center.Row)
	}
	if center.Column != 1 {
		t.Fatalf("center (column) should be 1, but it's %d", center.Column)
	}
	// now checking for the center of Column (1 cols x 3 rows)
	r3, err := geom.NewRectangle(ctx, testAcquisitionChan(), geom.WithOrientation(style.Horizontal), geom.WithTopCorner(1, 1), geom.WithBottomCorner(1, 3))
	if err != nil {
		t.Fatalf("error : there should be no error - %v", err)
	}
	if r3.Size().Width != 1 || r3.Size().Height != 3 {
		t.Fatal("error : width should be 1 height should be 3")
	}
	center = r3.Center()
	if center.Row != 1 {
		t.Fatalf("center (row) should be 1, but it's %d", center.Row)
	}
	if center.Column != 0 {
		t.Fatalf("center (column) should be 0, but it's %d", center.Column)
	}
	// now checking for the center of 2 cols x 2 rows Rectangle
	r4, err = geom.NewRectangle(ctx, testAcquisitionChan(), geom.WithTopCorner(1, 1), geom.WithBottomCorner(2, 2))
	if err != nil {
		t.Fatalf("error : there should be no error - %v", err)
	}
	if r4.Size().Width != 2 || r4.Size().Height != 2 {
		t.Fatal("error : width and height should be 2")
	}
	center = r4.Center()
	if center.Row != 0 {
		t.Fatalf("center (row) should be 0, but it's %d", center.Row)
	}
	if center.Column != 0 {
		t.Fatalf("center (column) should be 0, but it's %d", center.Column)
	}
	cancel()
}

func TestRectangleUnion(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	r1, err := geom.NewRectangle(ctx, testAcquisitionChan(), geom.WithTopCorner(1, 1), geom.WithBottomCorner(2, 2))
	if err != nil {
		t.Fatalf("error : there should be no error - %v", err)
	}
	r2, err := geom.NewRectangle(ctx, testAcquisitionChan(), geom.WithTopCorner(2, 2), geom.WithBottomCorner(3, 3))
	if err != nil {
		t.Fatalf("error : there should be no error - %v", err)
	}
	r3 := r1.Union(r2)
	if w := r3.Width(); w != 3 {
		t.Fatal("width should be 3", w)
	}
	if h := r3.Height(); h != 3 {
		t.Fatal("height should be 3", h)
	}
	r2, err = geom.NewRectangle(ctx, testAcquisitionChan(), geom.WithTopCorner(12, 12), geom.WithBottomCorner(13, 13))
	if err != nil {
		t.Fatalf("error : there should be no error - %v", err)
	}
	r3 = r1.Union(r2)
	if w := r3.Width(); w != 13 {
		t.Fatal("width should be 13", w)
	}
	if h := r3.Height(); h != 13 {
		t.Fatal("height should be 13", h)
	}
	cancel()
}

func TestIntersect(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	r1, err := geom.NewRectangle(ctx, testAcquisitionChan(), geom.WithTopCorner(1, 1), geom.WithBottomCorner(12, 12))
	if err != nil {
		t.Fatalf("error : there should be no error - %v", err)
	}
	r2, err := geom.NewRectangle(ctx, testAcquisitionChan(), geom.WithTopCorner(5, 5), geom.WithBottomCorner(20, 8))
	if err != nil {
		t.Fatalf("error : there should be no error - %v", err)
	}
	r3 := r1.Intersect(r2)
	if w := r3.Width(); w != 8 {
		t.Fatal("width should be 8", w)
	}
	if h := r3.Height(); h != 4 {
		t.Fatal("height should be 4", h)
	}
	cancel()
}

func TestIntersectInside(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	r1, err := geom.NewRectangle(ctx, testAcquisitionChan(), geom.WithTopCorner(1, 1), geom.WithBottomCorner(12, 12))
	if err != nil {
		t.Fatalf("error : there should be no error - %v", err)
	}
	r2, err := geom.NewRectangle(ctx, testAcquisitionChan(), geom.WithTopCorner(3, 3), geom.WithBottomCorner(9, 9))
	if err != nil {
		t.Fatalf("error : there should be no error - %v", err)
	}
	r3 := r1.Intersect(r2)
	if w := r3.Width(); w != 7 {
		t.Fatal("width should be 7", w)
	}
	if h := r3.Height(); h != 7 {
		t.Fatal("height should be 7", h)
	}
	cancel()
}

func TestInside(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	r1, err := geom.NewRectangle(ctx, testAcquisitionChan(), geom.WithTopCorner(1, 1), geom.WithBottomCorner(12, 12))
	if err != nil {
		t.Fatalf("error : there should be no error - %v", err)
	}
	r2, err := geom.NewRectangle(ctx, testAcquisitionChan(), geom.WithTopCorner(3, 3), geom.WithBottomCorner(9, 9))
	if err != nil {
		t.Fatalf("error : there should be no error - %v", err)
	}
	r3, err := geom.NewRectangle(ctx, testAcquisitionChan(), geom.WithTopCorner(13, 13), geom.WithBottomCorner(19, 19))
	if err != nil {
		t.Fatalf("error : there should be no error - %v", err)
	}
	if r1.In(r2) {
		t.Fatal("error rectangle should NOT be inside")
	}
	if !r2.In(r1) {
		t.Fatal("error rectangle SHOULD be inside")
	}
	if r3.In(r1) || r1.In(r3) {
		t.Fatal("error these rectangle do NOT have something in common")
	}
	cancel()
}

func TestOverlap(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	r1, err := geom.NewRectangle(ctx, testAcquisitionChan(), geom.WithTopCorner(1, 1), geom.WithBottomCorner(12, 12))
	if err != nil {
		t.Fatalf("error : there should be no error - %v", err)
	}
	r2, err := geom.NewRectangle(ctx, testAcquisitionChan(), geom.WithTopCorner(3, 3), geom.WithBottomCorner(9, 9))
	if err != nil {
		t.Fatalf("error : there should be no error - %v", err)
	}
	r3, err := geom.NewRectangle(ctx, testAcquisitionChan(), geom.WithTopCorner(13, 13), geom.WithBottomCorner(19, 19))
	if err != nil {
		t.Fatalf("error : there should be no error - %v", err)
	}
	if r1.Overlaps(r3) {
		t.Fatal("rectangles should NOT overlap")
	}
	if !r2.Overlaps(r1) || !r1.Overlaps(r2) {
		t.Fatal("rectangles SHOULD overlap")
	}
	cancel()
}
