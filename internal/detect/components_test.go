package detect

import (
	"math"
	"sort"
	"testing"

	"boatdetect/internal/gdal"
)

const componentEps = 1e-9

func TestComponentsFourNeighborhood(t *testing.T) {
	grid := gdal.Grid{
		Width:  4,
		Height: 3,
		NoData: -9999,
		Data: []float64{
			1, 1, 0, 0,
			1, 0, 0, 2,
			0, 0, 2, 2,
		},
	}

	got := Components(grid, 1, false, 1)
	if len(got) != 2 {
		t.Fatalf("expected 2 components, got %d", len(got))
	}

	sort.Slice(got, func(i, j int) bool {
		if got[i].Cx == got[j].Cx {
			return got[i].Cy < got[j].Cy
		}
		return got[i].Cx < got[j].Cx
	})

	want := []Component{
		{
			Area: 3,
			Sum:  3,
			Cx:   1.0 / 3.0,
			Cy:   1.0 / 3.0,
		},
		{
			Area: 3,
			Sum:  6,
			Cx:   8.0 / 3.0,
			Cy:   5.0 / 3.0,
		},
	}

	for i := range want {
		assertComponentClose(t, got[i], want[i])
	}
}

func TestComponentsDiagonalSeparate(t *testing.T) {
	grid := gdal.Grid{
		Width:  2,
		Height: 2,
		NoData: -9999,
		Data: []float64{
			1, 0,
			0, 1,
		},
	}

	got := Components(grid, 1, false, 1)
	if len(got) != 2 {
		t.Fatalf("expected 2 components, got %d", len(got))
	}

	sort.Slice(got, func(i, j int) bool {
		if got[i].Cx == got[j].Cx {
			return got[i].Cy < got[j].Cy
		}
		return got[i].Cx < got[j].Cx
	})

	want := []Component{
		{Area: 1, Sum: 1, Cx: 0, Cy: 0},
		{Area: 1, Sum: 1, Cx: 1, Cy: 1},
	}

	for i := range want {
		assertComponentClose(t, got[i], want[i])
	}
}

func TestComponentsIgnoreNoDataAndMinArea(t *testing.T) {
	grid := gdal.Grid{
		Width:  3,
		Height: 2,
		NoData: -9999,
		Data: []float64{
			1, -9999, 1,
			math.NaN(), 1, 0,
		},
	}

	got := Components(grid, 1, false, 1)
	if len(got) != 3 {
		t.Fatalf("expected 3 components, got %d", len(got))
	}

	sort.Slice(got, func(i, j int) bool {
		if got[i].Cx == got[j].Cx {
			return got[i].Cy < got[j].Cy
		}
		return got[i].Cx < got[j].Cx
	})

	want := []Component{
		{Area: 1, Sum: 1, Cx: 0, Cy: 0},
		{Area: 1, Sum: 1, Cx: 1, Cy: 1},
		{Area: 1, Sum: 1, Cx: 2, Cy: 0},
	}

	for i := range want {
		assertComponentClose(t, got[i], want[i])
	}

	got = Components(grid, 1, false, 2)
	if len(got) != 0 {
		t.Fatalf("expected 0 components with min area, got %d", len(got))
	}
}

func assertComponentClose(t *testing.T, got, want Component) {
	t.Helper()
	if got.Area != want.Area {
		t.Fatalf("expected area %d, got %d", want.Area, got.Area)
	}
	if math.Abs(got.Sum-want.Sum) > componentEps {
		t.Fatalf("expected sum %v, got %v", want.Sum, got.Sum)
	}
	if math.Abs(got.Cx-want.Cx) > componentEps {
		t.Fatalf("expected cx %v, got %v", want.Cx, got.Cx)
	}
	if math.Abs(got.Cy-want.Cy) > componentEps {
		t.Fatalf("expected cy %v, got %v", want.Cy, got.Cy)
	}
}
