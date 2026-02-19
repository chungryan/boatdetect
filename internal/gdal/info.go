package gdal

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
)

// RasterInfo describes basic raster metadata from gdalinfo.
type RasterInfo struct {
	Width        int
	Height       int
	GeoTransform [6]float64
	WGS84BBox    *[4]float64
}

// GetInfo runs gdalinfo and extracts raster size and geotransform.
func GetInfo(ctx context.Context, path string) (RasterInfo, error) {
	stdout, _, err := Run(ctx, "gdalinfo", "-json", path)
	if err != nil {
		return RasterInfo{}, fmt.Errorf("gdalinfo: %w", err)
	}

	var payload struct {
		Size         []int     `json:"size"`
		GeoTransform []float64 `json:"geoTransform"`
		WGS84Extent  *struct {
			Type        string        `json:"type"`
			Coordinates [][][]float64 `json:"coordinates"`
		} `json:"wgs84Extent"`
	}
	if err := json.Unmarshal([]byte(stdout), &payload); err != nil {
		return RasterInfo{}, fmt.Errorf("parse gdalinfo json: %w", err)
	}

	if len(payload.Size) != 2 {
		return RasterInfo{}, fmt.Errorf("unexpected gdalinfo size length: %d", len(payload.Size))
	}
	if len(payload.GeoTransform) != 6 {
		return RasterInfo{}, fmt.Errorf("unexpected gdalinfo geotransform length: %d", len(payload.GeoTransform))
	}

	info := RasterInfo{
		Width:  payload.Size[0],
		Height: payload.Size[1],
	}
	for i := 0; i < 6; i++ {
		info.GeoTransform[i] = payload.GeoTransform[i]
	}

	if payload.WGS84Extent != nil {
		bbox := wgs84BBoxFromExtent(payload.WGS84Extent.Coordinates)
		if bbox != nil {
			info.WGS84BBox = bbox
		}
	}

	return info, nil
}

func wgs84BBoxFromExtent(coords [][][]float64) *[4]float64 {
	if len(coords) == 0 || len(coords[0]) == 0 {
		return nil
	}

	minLon, maxLon := coords[0][0][0], coords[0][0][0]
	minLat, maxLat := coords[0][0][1], coords[0][0][1]

	for _, ring := range coords {
		minLon, minLat, maxLon, maxLat = updateBBoxFromRing(ring, minLon, minLat, maxLon, maxLat)
	}

	return &[4]float64{minLon, minLat, maxLon, maxLat}
}

func updateBBoxFromRing(ring [][]float64, minLon, minLat, maxLon, maxLat float64) (float64, float64, float64, float64) {
	for _, pt := range ring {
		if len(pt) < 2 {
			continue
		}

		minLon = math.Min(minLon, pt[0])
		maxLon = math.Max(maxLon, pt[0])
		minLat = math.Min(minLat, pt[1])
		maxLat = math.Max(maxLat, pt[1])
	}

	return minLon, minLat, maxLon, maxLat
}
