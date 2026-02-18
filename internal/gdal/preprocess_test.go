package gdal

import "testing"

func TestPreprocessHashStable(t *testing.T) {
	bbox := [4]float64{-122.5, 37.7, -122.3, 37.9}
	first := preprocessHash("/tmp/input.tif", bbox)
	second := preprocessHash("/tmp/input.tif", bbox)
	if first != second {
		t.Fatalf("expected stable hash, got %q and %q", first, second)
	}
}

func TestPreprocessHashDiffersByBBox(t *testing.T) {
	first := preprocessHash("/tmp/input.tif", [4]float64{-122.5, 37.7, -122.3, 37.9})
	second := preprocessHash("/tmp/input.tif", [4]float64{-122.5, 37.7, -122.1, 37.9})
	if first == second {
		t.Fatalf("expected different hashes for different bbox, got %q", first)
	}
}
