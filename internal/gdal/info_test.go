package gdal

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestGetInfoParsesFields(t *testing.T) {
	useLocalGDAL(t)

	ctx := context.Background()
	tempDir := t.TempDir()
	fake := filepath.Join(tempDir, "gdalinfo")
	writeScript(t, fake, `#!/bin/sh
cat <<'EOF'
{"size":[123,456],"geoTransform":[1,2,3,4,5,6],"wgs84Extent":{"type":"Polygon","coordinates":[[[10,20],[30,20],[30,40],[10,40],[10,20]]]}}
EOF
`)

	prependPath(t, tempDir)

	info, err := GetInfo(ctx, "/tmp/example.tif")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if info.Width != 123 || info.Height != 456 {
		t.Fatalf("unexpected size: %+v", info)
	}
	if info.GeoTransform != ([6]float64{1, 2, 3, 4, 5, 6}) {
		t.Fatalf("unexpected geotransform: %+v", info.GeoTransform)
	}
	if info.WGS84BBox == nil {
		t.Fatalf("expected wgs84 bbox to be set")
	}
	if *info.WGS84BBox != ([4]float64{10, 20, 30, 40}) {
		t.Fatalf("unexpected wgs84 bbox: %+v", *info.WGS84BBox)
	}
}

func TestGetInfoValidatesLengths(t *testing.T) {
	useLocalGDAL(t)

	ctx := context.Background()
	tempDir := t.TempDir()
	fake := filepath.Join(tempDir, "gdalinfo")
	writeScript(t, fake, `#!/bin/sh
cat <<'EOF'
{"size":[123],"geoTransform":[1,2,3,4,5,6]}
EOF
`)

	prependPath(t, tempDir)

	_, err := GetInfo(ctx, "/tmp/example.tif")
	if err == nil {
		t.Fatalf("expected error, got nil")
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
