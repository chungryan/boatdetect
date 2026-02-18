package gdal

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

func TestRunCapturesStdoutStderr(t *testing.T) {
	useLocalGDAL(t)

	ctx := context.Background()
	stdout, stderr, err := Run(ctx, "sh", "-c", "echo out; echo err 1>&2")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if strings.TrimSpace(stdout) != "out" {
		t.Fatalf("unexpected stdout: %q", stdout)
	}
	if strings.TrimSpace(stderr) != "err" {
		t.Fatalf("unexpected stderr: %q", stderr)
	}
}

func TestRunErrorIncludesCommandAndStderr(t *testing.T) {
	useLocalGDAL(t)

	ctx := context.Background()
	stdout, stderr, err := Run(ctx, "sh", "-c", "echo out; echo err 1>&2; exit 2")
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if strings.TrimSpace(stdout) != "out" {
		t.Fatalf("unexpected stdout: %q", stdout)
	}
	if strings.TrimSpace(stderr) != "err" {
		t.Fatalf("unexpected stderr: %q", stderr)
	}
	msg := err.Error()
	if !strings.Contains(msg, "sh -c") {
		t.Fatalf("error missing command: %q", msg)
	}
	if !strings.Contains(msg, "err") {
		t.Fatalf("error missing stderr: %q", msg)
	}
}

func TestRunContextTimeout(t *testing.T) {
	useLocalGDAL(t)

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	_, _, err := Run(ctx, "sh", "-c", "sleep 1")
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return
	}
	msg := err.Error()
	if strings.Contains(msg, "context deadline exceeded") {
		return
	}
	if strings.Contains(msg, "signal: killed") {
		return
	}
	t.Fatalf("expected context timeout-related error, got %v", err)
}
