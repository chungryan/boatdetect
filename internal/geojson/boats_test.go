package geojson

import (
	"reflect"
	"testing"

	"boatdetect/internal/detect"
)

func TestBuildBoatsFC(t *testing.T) {
	candidates := []detect.Candidate{
		{
			Lon:    -122.5,
			Lat:    37.9,
			Score:  1.25,
			AreaPx: 42,
		},
		{
			Lon:    -122.6,
			Lat:    38.0,
			Score:  0.75,
			AreaPx: 7,
		},
	}

	fc := BuildBoatsFC("scene-123", candidates)

	if fc.Type != featureCollectionType {
		t.Fatalf("expected type %q, got %q", featureCollectionType, fc.Type)
	}
	if len(fc.Features) != len(candidates) {
		t.Fatalf("expected %d features, got %d", len(candidates), len(fc.Features))
	}

	for i, feature := range fc.Features {
		if feature.Type != featureType {
			t.Fatalf("feature %d type: expected %q, got %q", i, featureType, feature.Type)
		}
		if feature.Geometry.Type != geometryPointType {
			t.Fatalf("feature %d geometry type: expected %q, got %q", i, geometryPointType, feature.Geometry.Type)
		}
		coords, ok := feature.Geometry.Coordinates.([]float64)
		if !ok {
			t.Fatalf("feature %d coordinates: expected []float64, got %T", i, feature.Geometry.Coordinates)
		}

		wantCoords := []float64{candidates[i].Lon, candidates[i].Lat}
		if !reflect.DeepEqual(coords, wantCoords) {
			t.Fatalf("feature %d coordinates: expected %v, got %v", i, wantCoords, coords)
		}

		if feature.Properties["scene_id"] != "scene-123" {
			t.Fatalf("feature %d scene_id: expected %q, got %v", i, "scene-123", feature.Properties["scene_id"])
		}
		if score, ok := feature.Properties["score"].(float64); !ok || score != candidates[i].Score {
			t.Fatalf("feature %d score: expected %v, got %v", i, candidates[i].Score, feature.Properties["score"])
		}
		if area, ok := feature.Properties["area_px"].(int); !ok || area != candidates[i].AreaPx {
			t.Fatalf("feature %d area_px: expected %v, got %v", i, candidates[i].AreaPx, feature.Properties["area_px"])
		}
	}
}

func TestBuildBoatsFCEmpty(t *testing.T) {
	fc := BuildBoatsFC("scene-123", nil)
	if fc.Type != featureCollectionType {
		t.Fatalf("expected type %q, got %q", featureCollectionType, fc.Type)
	}
	if len(fc.Features) != 0 {
		t.Fatalf("expected no features, got %d", len(fc.Features))
	}
}
