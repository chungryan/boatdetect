package detect

import (
	"math"
	"testing"
)

func TestMeanStdIgnoresNoDataAndNaN(t *testing.T) {
	data := []float64{1, 2, 3, -9999, math.NaN()}
	mean, std := MeanStd(data, -9999)

	if math.Abs(mean-2) > 1e-9 {
		t.Fatalf("unexpected mean: %v", mean)
	}

	expectedStd := math.Sqrt(2.0 / 3.0)
	if math.Abs(std-expectedStd) > 1e-9 {
		t.Fatalf("unexpected std: %v", std)
	}
}

func TestMeanStdAllIgnored(t *testing.T) {
	data := []float64{-1, -1, math.NaN()}
	mean, std := MeanStd(data, -1)

	if mean != 0 || std != 0 {
		t.Fatalf("expected zero stats, got mean=%v std=%v", mean, std)
	}
}
