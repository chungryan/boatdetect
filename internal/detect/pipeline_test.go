package detect

import (
	"context"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestDetectCandidates(t *testing.T) {
	t.Setenv("BOATDETECT_GDAL_MODE", "local")

	ctx := context.Background()

	tempDir := t.TempDir()

	infoPath := filepath.Join(tempDir, "gdalinfo")
	writeScript(t, infoPath, `#!/bin/sh
cat <<'EOF'
{"size":[3,2],"geoTransform":[10,2,0,20,0,-2]}
EOF
`)

	translatePath := filepath.Join(tempDir, "gdal_translate")
	writeScript(t, translatePath, "#!/bin/sh\n"+
		"cat > \"$4\" <<'EOF'\n"+
		"ncols 3\n"+
		"nrows 2\n"+
		"xllcorner 0\n"+
		"yllcorner 0\n"+
		"cellsize 1\n"+
		"NODATA_value -9999\n"+
		"1 2 3\n"+
		"4 5 6\n"+
		"EOF\n")

	prependPath(t, tempDir)

	candidates, err := DetectCandidates(ctx, "/tmp/input.tif", 0.5, 0, false, 1)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(candidates) != 1 {
		t.Fatalf("expected 1 candidate, got %d", len(candidates))
	}

	got := candidates[0]
	assertFloatClose(t, got.Lon, 13)
	assertFloatClose(t, got.Lat, 18)
	assertFloatClose(t, got.Score, 5.5)
	if got.AreaPx != 2 {
		t.Fatalf("expected area 2, got %d", got.AreaPx)
	}
}

func TestDetectCandidatesPropagatesInfoError(t *testing.T) {
	t.Setenv("BOATDETECT_GDAL_MODE", "local")

	ctx := context.Background()
	tempDir := t.TempDir()

	infoPath := filepath.Join(tempDir, "gdalinfo")
	writeScript(t, infoPath, "#!/bin/sh\n"+
		"echo nope 1>&2\n"+
		"exit 2\n")

	prependPath(t, tempDir)

	_, err := DetectCandidates(ctx, "/tmp/input.tif", 0.5, 0, false, 1)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "get raster info") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func assertFloatClose(t *testing.T, got, want float64) {
	t.Helper()
	if math.Abs(got-want) > 1e-9 {
		t.Fatalf("expected %v, got %v", want, got)
	}
}

func writeScript(t *testing.T, path, contents string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(contents), 0o755); err != nil {
		t.Fatalf("write script: %v", err)
	}
	if runtime.GOOS == "windows" {
		t.Fatalf("test does not support windows")
	}
}

func prependPath(t *testing.T, dir string) {
	t.Helper()
	old := os.Getenv("PATH")
	t.Setenv("PATH", dir+string(os.PathListSeparator)+old)
}
