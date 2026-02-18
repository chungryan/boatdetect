package geojson

import "boatdetect/internal/detect"

// BuildBoatsFC builds a GeoJSON feature collection from detected candidates.
func BuildBoatsFC(sceneID string, candidates []detect.Candidate) FeatureCollection {
	features := make([]Feature, 0, len(candidates))
	for _, candidate := range candidates {
		features = append(features, Feature{
			Type: featureType,
			Geometry: Geometry{
				Type:        geometryPointType,
				Coordinates: []float64{candidate.Lon, candidate.Lat},
			},
			Properties: map[string]interface{}{
				"scene_id": sceneID,
				"score":    candidate.Score,
				"area_px":  candidate.AreaPx,
			},
		})
	}

	return FeatureCollection{
		Type:     featureCollectionType,
		Features: features,
	}
}
