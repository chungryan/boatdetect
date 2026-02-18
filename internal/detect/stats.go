package detect

import (
	"fmt"
	"math"
	"sort"
)

// MeanStd returns the mean and population standard deviation, ignoring nodata and NaN values.
func MeanStd(data []float64, nodata float64) (mean, std float64) {
	count := 0
	mean = 0
	m2 := 0.0

	checkNoData := !math.IsNaN(nodata)
	for _, v := range data {
		if math.IsNaN(v) {
			continue
		}
		if checkNoData && v == nodata {
			continue
		}

		count++
		delta := v - mean
		mean += delta / float64(count)
		delta2 := v - mean
		m2 += delta * delta2
	}

	if count == 0 {
		return 0, 0
	}

	variance := m2 / float64(count)
	return mean, math.Sqrt(variance)
}

// Percentile returns the pth percentile value, ignoring nodata and NaN values.
// p must be in the range (0, 100).
func Percentile(data []float64, nodata float64, p float64) (float64, error) {
	if p <= 0 || p >= 100 {
		return 0, fmt.Errorf("invalid percentile %v", p)
	}

	values := make([]float64, 0, len(data))
	checkNoData := !math.IsNaN(nodata)
	for _, v := range data {
		if math.IsNaN(v) {
			continue
		}
		if checkNoData && v == nodata {
			continue
		}
		values = append(values, v)
	}

	if len(values) == 0 {
		return 0, fmt.Errorf("no valid values")
	}

	sort.Float64s(values)
	idx := int(math.Ceil((p/100.0)*float64(len(values)))) - 1
	if idx < 0 {
		return values[0], nil
	}
	if idx >= len(values) {
		return values[len(values)-1], nil
	}
	return values[idx], nil
}
