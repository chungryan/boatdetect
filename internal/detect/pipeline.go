package detect

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"boatdetect/internal/gdal"
)

// Candidate represents a detected signal candidate.
type Candidate struct {
	Lon    float64
	Lat    float64
	Score  float64
	AreaPx int
}

// DetectCandidates runs the candidate detection pipeline for a GeoTIFF.
func DetectCandidates(ctx context.Context, byteTifPath string, k float64, percentile float64, invert bool, minAreaPx int) ([]Candidate, error) {
	info, err := gdal.GetInfo(ctx, byteTifPath)
	if err != nil {
		return nil, fmt.Errorf("get raster info: %w", err)
	}

	// Create temp file in current working directory (accessible by Docker container)
	tempDir := filepath.Join(".tmp", "temp")
	if err := os.MkdirAll(tempDir, 0o755); err != nil {
		return nil, fmt.Errorf("create temp dir: %w", err)
	}
	defer os.RemoveAll(".tmp")

	ascFile, err := os.CreateTemp(tempDir, "*.asc")
	if err != nil {
		return nil, fmt.Errorf("create temp grid: %w", err)
	}
	ascPath := ascFile.Name()
	if err := ascFile.Close(); err != nil {
		return nil, fmt.Errorf("close temp grid: %w", err)
	}
	defer os.Remove(ascPath)

	if err := gdal.ToAAIGrid(ctx, byteTifPath, ascPath); err != nil {
		return nil, fmt.Errorf("convert to ascii grid: %w", err)
	}

	gridFile, err := os.Open(ascPath)
	if err != nil {
		return nil, fmt.Errorf("open ascii grid: %w", err)
	}
	defer gridFile.Close()

	grid, err := gdal.ParseAAIGrid(gridFile)
	if err != nil {
		return nil, fmt.Errorf("parse ascii grid: %w", err)
	}

	threshold, err := calculateThreshold(grid, k, percentile, invert)
	if err != nil {
		return nil, err
	}

	components := Components(grid, threshold, invert, minAreaPx)
	candidates := make([]Candidate, 0, len(components))
	for _, component := range components {
		lon, lat := PixelToLonLat(info.GeoTransform, component.Cx, component.Cy)
		candidates = append(candidates, Candidate{
			Lon:    lon,
			Lat:    lat,
			Score:  component.Sum / float64(component.Area),
			AreaPx: component.Area,
		})
	}

	return candidates, nil
}

func calculateThreshold(grid gdal.Grid, k, percentile float64, invert bool) (float64, error) {
	if percentile > 0 {
		return calculatePercentileThreshold(grid, percentile, invert)
	}
	return calculateStdDevThreshold(grid, k, invert), nil
}

func calculatePercentileThreshold(grid gdal.Grid, percentile float64, invert bool) (float64, error) {
	effectivePercentile := percentile
	if invert {
		effectivePercentile = 100 - percentile
	}
	pct, err := Percentile(grid.Data, grid.NoData, effectivePercentile)
	if err != nil {
		return 0, fmt.Errorf("percentile threshold: %w", err)
	}
	return pct, nil
}

func calculateStdDevThreshold(grid gdal.Grid, k float64, invert bool) float64 {
	mean, std := MeanStd(grid.Data, grid.NoData)
	if invert {
		return mean - k*std
	}
	return mean + k*std
}
