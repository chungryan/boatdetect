package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/tabwriter"

	"boatdetect/internal/detect"
	"boatdetect/internal/gdal"
	"boatdetect/internal/geojson"
)

const (
	defaultK             = 2.0
	defaultPercentile    = 99.5
	defaultInvert        = true
	defaultMinAreaPx     = 2
	defaultMaxCandidates = 200
)

type detectOptions struct {
	input string
	out   string
}

type candidateRecord struct {
	sceneID   string
	candidate detect.Candidate
}

func parseFlags() (detectOptions, error) {
	var opts detectOptions
	flag.StringVar(&opts.input, "input", "", "Input folder containing .tif or .SAFE")
	flag.StringVar(&opts.out, "out", "", "Output GeoJSON path")

	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage: boatdetect --input <dir> --out <file>\n\n")
		flag.PrintDefaults()
	}

	flag.Parse()

	if opts.input == "" {
		return detectOptions{}, fmt.Errorf("input is required")
	}
	if opts.out == "" {
		return detectOptions{}, fmt.Errorf("out is required")
	}
	return opts, nil
}

func runDetect(ctx context.Context, out io.Writer, opts detectOptions) error {
	inputFiles, err := findTifFiles(opts.input)
	if err != nil {
		return err
	}
	if len(inputFiles) == 0 {
		return fmt.Errorf("no .tif files found in %s", opts.input)
	}

	if err := ensureOutputDir(opts.out); err != nil {
		return err
	}

	minLon, minLat, maxLon, maxLat, err := bboxFromFiles(ctx, inputFiles)
	if err != nil {
		return err
	}

	preprocessDir := filepath.Join(filepath.Dir(opts.out), ".tmp", "preprocess")
	bbox := [4]float64{minLon, minLat, maxLon, maxLat}

	records, sceneOrder, err := processCandidates(ctx, inputFiles, preprocessDir, bbox)
	if err != nil {
		return err
	}

	records = limitCandidates(records, defaultMaxCandidates)
	byScene := groupCandidates(sceneOrder, records)

	if err := writeSummaryTable(out, sceneOrder, byScene); err != nil {
		return err
	}

	if err := writeGeojson(opts.out, sceneOrder, byScene); err != nil {
		return err
	}

	return cleanupTemp(opts.out)
}

func processCandidates(ctx context.Context, inputFiles []string, preprocessDir string, bbox [4]float64) ([]candidateRecord, []string, error) {
	records := make([]candidateRecord, 0)
	seenScenes := make(map[string]struct{})
	sceneOrder := make([]string, 0)

	for _, inputPath := range inputFiles {
		byteTif, err := gdal.Preprocess(ctx, inputPath, preprocessDir, bbox)
		if err != nil {
			return nil, nil, fmt.Errorf("preprocess %s: %w", inputPath, err)
		}

		candidates, err := detect.DetectCandidates(ctx, byteTif, defaultK, defaultPercentile, defaultInvert, defaultMinAreaPx)
		if err != nil {
			return nil, nil, fmt.Errorf("detect %s: %w", inputPath, err)
		}

		sceneID := sceneIDFromPath(inputPath)
		if _, ok := seenScenes[sceneID]; !ok {
			seenScenes[sceneID] = struct{}{}
			sceneOrder = append(sceneOrder, sceneID)
		}

		for _, candidate := range candidates {
			records = append(records, candidateRecord{
				sceneID:   sceneID,
				candidate: candidate,
			})
		}
	}

	return records, sceneOrder, nil
}

func writeGeojson(outPath string, sceneOrder []string, byScene map[string][]detect.Candidate) error {
	features := make([]geojson.Feature, 0)
	for _, sceneID := range sceneOrder {
		fc := geojson.BuildBoatsFC(sceneID, byScene[sceneID])
		features = append(features, fc.Features...)
	}

	fc := geojson.FeatureCollection{
		Type:     "FeatureCollection",
		Features: features,
	}
	if err := geojson.WriteFeatureCollection(outPath, fc); err != nil {
		return fmt.Errorf("write geojson: %w", err)
	}
	return nil
}

func cleanupTemp(outPath string) error {
	tmpDir := filepath.Join(filepath.Dir(outPath), ".tmp")
	return os.RemoveAll(tmpDir)
}

func findTifFiles(inputDir string) ([]string, error) {
	info, err := os.Stat(inputDir)
	if err != nil {
		return nil, fmt.Errorf("stat input: %w", err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("input must be a directory")
	}

	files := make([]string, 0)
	err = filepath.WalkDir(inputDir, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		if isTifFile(path) {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("walk input: %w", err)
	}

	sort.Strings(files)
	return files, nil
}

func isTifFile(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	return ext == ".tif" || ext == ".tiff"
}

func sceneIDFromPath(path string) string {
	parts := strings.Split(filepath.Clean(path), string(filepath.Separator))
	for _, part := range parts {
		if strings.HasSuffix(part, ".SAFE") {
			return strings.TrimSuffix(part, ".SAFE")
		}
	}

	base := filepath.Base(path)
	return strings.TrimSuffix(base, filepath.Ext(base))
}

func limitCandidates(records []candidateRecord, maxCandidates int) []candidateRecord {
	if maxCandidates <= 0 || len(records) <= maxCandidates {
		return records
	}

	sort.Slice(records, func(i, j int) bool {
		if records[i].candidate.Score != records[j].candidate.Score {
			return records[i].candidate.Score > records[j].candidate.Score
		}
		if records[i].candidate.AreaPx != records[j].candidate.AreaPx {
			return records[i].candidate.AreaPx > records[j].candidate.AreaPx
		}
		if records[i].sceneID != records[j].sceneID {
			return records[i].sceneID < records[j].sceneID
		}
		if records[i].candidate.Lat != records[j].candidate.Lat {
			return records[i].candidate.Lat < records[j].candidate.Lat
		}
		return records[i].candidate.Lon < records[j].candidate.Lon
	})

	return records[:maxCandidates]
}

func groupCandidates(sceneOrder []string, records []candidateRecord) map[string][]detect.Candidate {
	byScene := make(map[string][]detect.Candidate, len(sceneOrder))
	for _, sceneID := range sceneOrder {
		byScene[sceneID] = nil
	}
	for _, record := range records {
		byScene[record.sceneID] = append(byScene[record.sceneID], record.candidate)
	}
	return byScene
}

func writeSummaryTable(w io.Writer, sceneOrder []string, byScene map[string][]detect.Candidate) error {
	tw := tabwriter.NewWriter(w, 0, 4, 2, ' ', 0)
	if _, err := fmt.Fprintln(tw, "scene_id\tcandidates\tscore_mean\tscore_max\tarea_min\tarea_max"); err != nil {
		return err
	}

	for _, sceneID := range sceneOrder {
		candidates := byScene[sceneID]
		if err := writeSummaryRow(tw, sceneID, candidates); err != nil {
			return err
		}
	}

	return tw.Flush()
}

func writeSummaryRow(tw *tabwriter.Writer, sceneID string, candidates []detect.Candidate) error {
	if len(candidates) == 0 {
		_, err := fmt.Fprintf(tw, "%s\t0\t0.00\t0.00\t0\t0\n", sceneID)
		return err
	}

	stats := calculateCandidateStats(candidates)
	_, err := fmt.Fprintf(tw, "%s\t%d\t%.2f\t%.2f\t%d\t%d\n",
		sceneID, len(candidates), stats.meanScore, stats.maxScore, stats.minArea, stats.maxArea)
	return err
}

type candidateStats struct {
	meanScore float64
	maxScore  float64
	minArea   int
	maxArea   int
}

func calculateCandidateStats(candidates []detect.Candidate) candidateStats {
	scoreSum := 0.0
	scoreMax := candidates[0].Score
	areaMin := candidates[0].AreaPx
	areaMax := candidates[0].AreaPx

	for _, c := range candidates {
		scoreSum += c.Score
		if c.Score > scoreMax {
			scoreMax = c.Score
		}
		if c.AreaPx < areaMin {
			areaMin = c.AreaPx
		}
		if c.AreaPx > areaMax {
			areaMax = c.AreaPx
		}
	}

	return candidateStats{
		meanScore: scoreSum / float64(len(candidates)),
		maxScore:  scoreMax,
		minArea:   areaMin,
		maxArea:   areaMax,
	}
}

func ensureOutputDir(outPath string) error {
	dir := filepath.Dir(outPath)
	if dir == "." {
		return nil
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create output dir: %w", err)
	}
	return nil
}

func bboxFromFiles(ctx context.Context, inputFiles []string) (minLon, minLat, maxLon, maxLat float64, err error) {
	minLon = math.Inf(1)
	minLat = math.Inf(1)
	maxLon = math.Inf(-1)
	maxLat = math.Inf(-1)

	for _, path := range inputFiles {
		info, err := gdal.GetInfo(ctx, path)
		if err != nil {
			return 0, 0, 0, 0, fmt.Errorf("get raster info: %w", err)
		}

		var bbox [4]float64
		if info.WGS84BBox != nil {
			bbox = *info.WGS84BBox
		} else {
			bbox, err = calculateBboxFromCorners(info)
			if err != nil {
				return 0, 0, 0, 0, err
			}
		}

		minLon = math.Min(minLon, bbox[0])
		minLat = math.Min(minLat, bbox[1])
		maxLon = math.Max(maxLon, bbox[2])
		maxLat = math.Max(maxLat, bbox[3])
	}

	if math.IsInf(minLon, 1) || math.IsInf(minLat, 1) || math.IsInf(maxLon, -1) || math.IsInf(maxLat, -1) {
		return 0, 0, 0, 0, fmt.Errorf("no valid raster extents found")
	}

	return minLon, minLat, maxLon, maxLat, nil
}

func calculateBboxFromCorners(info gdal.RasterInfo) ([4]float64, error) {
	minLon, minLat := math.Inf(1), math.Inf(1)
	maxLon, maxLat := math.Inf(-1), math.Inf(-1)

	corners := [][2]float64{
		{0, 0},
		{float64(info.Width), 0},
		{0, float64(info.Height)},
		{float64(info.Width), float64(info.Height)},
	}
	for _, corner := range corners {
		lon, lat := detect.PixelToLonLat(info.GeoTransform, corner[0], corner[1])
		minLon = math.Min(minLon, lon)
		maxLon = math.Max(maxLon, lon)
		minLat = math.Min(minLat, lat)
		maxLat = math.Max(maxLat, lat)
	}

	return [4]float64{minLon, minLat, maxLon, maxLat}, nil
}
