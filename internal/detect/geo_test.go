package detect

import (
	"math"
	"testing"
)

const pixelToGeoEps = 1e-9

func TestPixelToLonLat(t *testing.T) {
	tests := []struct {
		name string
		gt   [6]float64
		px   float64
		py   float64
		lon  float64
		lat  float64
	}{
		{
			name: "identity",
			gt:   [6]float64{0, 1, 0, 0, 0, 1},
			px:   10,
			py:   20,
			lon:  10,
			lat:  20,
		},
		{
			name: "translated",
			gt:   [6]float64{100, 1, 0, -50, 0, 1},
			px:   3,
			py:   4,
			lon:  103,
			lat:  -46,
		},
		{
			name: "rotated",
			gt:   [6]float64{10, 2, 0.5, 20, -1, 3},
			px:   5,
			py:   7,
			lon:  10 + 5*2 + 7*0.5,
			lat:  20 + 5*-1 + 7*3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lon, lat := PixelToLonLat(tt.gt, tt.px, tt.py)
			if math.Abs(lon-tt.lon) > pixelToGeoEps {
				t.Fatalf("expected lon %v, got %v", tt.lon, lon)
			}
			if math.Abs(lat-tt.lat) > pixelToGeoEps {
				t.Fatalf("expected lat %v, got %v", tt.lat, lat)
			}
		})
	}
}
