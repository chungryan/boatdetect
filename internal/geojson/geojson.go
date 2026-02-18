package geojson

import (
	"encoding/json"
	"os"
)

const (
	featureCollectionType = "FeatureCollection"
	featureType           = "Feature"
	geometryPointType     = "Point"
)

type FeatureCollection struct {
	Type     string    `json:"type"`
	Features []Feature `json:"features"`
}

type Feature struct {
	Type       string                 `json:"type"`
	Geometry   Geometry               `json:"geometry"`
	Properties map[string]interface{} `json:"properties,omitempty"`
}

type Geometry struct {
	Type        string      `json:"type"`
	Coordinates interface{} `json:"coordinates"`
}

func WriteFeatureCollection(path string, fc FeatureCollection) error {
	if fc.Type == "" {
		fc.Type = featureCollectionType
	}
	for i := range fc.Features {
		if fc.Features[i].Type == "" {
			fc.Features[i].Type = featureType
		}
	}

	data, err := json.Marshal(fc)
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0o644)
}
