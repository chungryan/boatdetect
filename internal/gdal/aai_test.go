package gdal

import (
	"reflect"
	"strings"
	"testing"
)

func TestParseAAIGrid(t *testing.T) {
	input := strings.TrimSpace(`
		ncols 3
		nrows 2
		xllcorner 0
		yllcorner 0
		cellsize 1
		NODATA_value -9999
		1 2 3
		4 5 6
	`)

	grid, err := ParseAAIGrid(strings.NewReader(input))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if grid.Width != 3 || grid.Height != 2 {
		t.Fatalf("unexpected size: %+v", grid)
	}
	if grid.NoData != -9999 {
		t.Fatalf("unexpected nodata: %v", grid.NoData)
	}
	want := []float64{1, 2, 3, 4, 5, 6}
	if !reflect.DeepEqual(grid.Data, want) {
		t.Fatalf("unexpected data: %#v", grid.Data)
	}
}
