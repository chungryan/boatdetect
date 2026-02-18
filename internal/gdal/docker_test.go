package gdal

import (
	"context"
	"os"
	"strings"
	"testing"
)

func TestNewClient(t *testing.T) {
	ctx := context.Background()
	client, err := NewClient(ctx)
	if err != nil {
		t.Fatalf("Failed to create Docker client: %v", err)
	}
	if client == nil {
		t.Fatal("Docker client is nil")
	}
	defer client.Close()
}

func TestConvertPath(t *testing.T) {
	cwd, _ := os.Getwd()
	client := &Client{workDir: cwd}

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "relative path",
			input:    "test.txt",
			expected: "/work/test.txt",
		},
		{
			name:     "relative path with subdirs",
			input:    "subdir/test.txt",
			expected: "/work/subdir/test.txt",
		},
		{
			name:     "absolute path within workdir",
			input:    cwd + "/test.txt",
			expected: "/work/test.txt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := client.convertPath(tt.input)
			if result != tt.expected {
				t.Errorf("convertPath(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestFormatCommand(t *testing.T) {
	tests := []struct {
		name     string
		cmdName  string
		args     []string
		contains string
	}{
		{
			name:     "simple command",
			cmdName:  "gdalinfo",
			args:     []string{"input.tif"},
			contains: "gdalinfo",
		},
		{
			name:     "command with multiple args",
			cmdName:  "gdalwarp",
			args:     []string{"-t_srs", "EPSG:4326", "input.tif", "output.tif"},
			contains: "gdalwarp",
		},
		{
			name:     "command with special characters",
			cmdName:  "gdal_translate",
			args:     []string{"-of", "AAIGrid", "input.tif", "output.asc"},
			contains: "gdal_translate",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatCommand(tt.cmdName, tt.args)
			if !strings.Contains(result, tt.contains) {
				t.Errorf("formatCommand(%q, %v) = %q, want to contain %q", tt.cmdName, tt.args, result, tt.contains)
			}
		})
	}
}

func TestQuoteArg(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple arg",
			input:    "test",
			expected: "test",
		},
		{
			name:     "arg with space",
			input:    "hello world",
			expected: `"hello world"`,
		},
		{
			name:     "empty string",
			input:    "",
			expected: `""`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := quoteArg(tt.input)
			if result != tt.expected {
				t.Errorf("quoteArg(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}
