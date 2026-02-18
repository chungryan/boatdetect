package gdal

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestToAAIGridRunsTranslate(t *testing.T) {
	useLocalGDAL(t)

	ctx := context.Background()
	tempDir := t.TempDir()
	fake := filepath.Join(tempDir, "gdal_translate")
	writeScript(t, fake, "#!/bin/sh\n"+
		"if [ \"$1\" != \"-of\" ] || [ \"$2\" != \"AAIGrid\" ]; then\n"+
		"  echo \"bad args\" 1>&2\n"+
		"  exit 3\n"+
		"fi\n"+
		"echo input=$3 output=$4 > \"$4\"\n")

	prependPath(t, tempDir)

	inputPath := filepath.Join(tempDir, "input.tif")
	outputPath := filepath.Join(tempDir, "output.asc")
	if err := os.WriteFile(inputPath, []byte("data"), 0o644); err != nil {
		t.Fatalf("write input: %v", err)
	}
	if err := os.WriteFile(outputPath, []byte("old"), 0o644); err != nil {
		t.Fatalf("write output: %v", err)
	}

	if err := ToAAIGrid(ctx, inputPath, outputPath); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("read output: %v", err)
	}
	if !strings.Contains(string(content), "input="+inputPath) {
		t.Fatalf("unexpected output content: %q", string(content))
	}
}

func TestToAAIGridReturnsError(t *testing.T) {
	useLocalGDAL(t)

	ctx := context.Background()
	tempDir := t.TempDir()
	fake := filepath.Join(tempDir, "gdal_translate")
	writeScript(t, fake, "#!/bin/sh\n"+
		"echo nope 1>&2\n"+
		"exit 2\n")

	prependPath(t, tempDir)

	err := ToAAIGrid(ctx, "/tmp/input.tif", filepath.Join(tempDir, "output.asc"))
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "gdal_translate") {
		t.Fatalf("expected gdal_translate error, got %v", err)
	}
}
