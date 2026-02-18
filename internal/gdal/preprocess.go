package gdal

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// Preprocess runs GDAL commands to warp to EPSG:4326 and produce a Byte-scaled GeoTIFF.
func Preprocess(ctx context.Context, inputPath, outputDir string, bbox [4]float64) (byteTifPath string, err error) {
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return "", fmt.Errorf("create output dir: %w", err)
	}

	base := strings.TrimSuffix(filepath.Base(inputPath), filepath.Ext(inputPath))
	hash := preprocessHash(inputPath, bbox)
	tmpPath := filepath.Join(outputDir, fmt.Sprintf("%s_%s_tmp.tif", base, hash))
	bytePath := filepath.Join(outputDir, fmt.Sprintf("%s_%s_byte.tif", base, hash))

	_, _, err = Run(ctx, "gdalwarp",
		"-t_srs", "EPSG:4326",
		"-te", strconv.FormatFloat(bbox[0], 'f', -1, 64),
		strconv.FormatFloat(bbox[1], 'f', -1, 64),
		strconv.FormatFloat(bbox[2], 'f', -1, 64),
		strconv.FormatFloat(bbox[3], 'f', -1, 64),
		"-r", "bilinear",
		"-overwrite",
		inputPath,
		tmpPath,
	)
	if err != nil {
		return "", fmt.Errorf("gdalwarp: %w", err)
	}

	_, _, err = Run(ctx, "gdal_translate",
		"-ot", "Byte",
		"-scale",
		tmpPath,
		bytePath,
	)
	if err != nil {
		return "", fmt.Errorf("gdal_translate: %w", err)
	}

	return bytePath, nil
}

func preprocessHash(inputPath string, bbox [4]float64) string {
	key := fmt.Sprintf("%s|%.8f|%.8f|%.8f|%.8f", inputPath, bbox[0], bbox[1], bbox[2], bbox[3])
	sum := sha256.Sum256([]byte(key))
	return hex.EncodeToString(sum[:8])
}
