package h3

import "testing"

func TestLatLngToCellAndBack(t *testing.T) {
	cell := LatLngToCell(6.5244, 3.3792, 8)
	lat, lng := CellToLatLng(cell)
	if cell == "" {
		t.Fatal("expected non-empty cell")
	}
	if lat == 0 || lng == 0 {
		t.Fatalf("expected non-zero lat/lng from cell, got %f %f", lat, lng)
	}
}

func TestKRingCells(t *testing.T) {
	cells := KRingCells("r8:100:200", 1)
	if len(cells) != 9 {
		t.Fatalf("expected 9 cells in simplified k-ring, got %d", len(cells))
	}
}

func TestDistanceBetweenCells(t *testing.T) {
	distance := DistanceBetweenCells("r8:10:10", "r8:12:13")
	if distance != 5 {
		t.Fatalf("expected grid distance 5, got %d", distance)
	}
}
