package gdal

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// Run executes a GDAL command in a Docker container.
// It captures stdout/stderr separately and returns a detailed error if the command fails.
// The Docker client must be initialized via Initialize() before calling this.
func Run(ctx context.Context, name string, args ...string) (stdout string, stderr string, err error) {
	mode := strings.ToLower(strings.TrimSpace(os.Getenv("BOATDETECT_GDAL_MODE")))

	if mode == "local" {
		return runLocal(ctx, name, args...)
	}

	client := GetClient()
	if client == nil {
		if mode == "docker" {
			return "", "", fmt.Errorf("docker client not initialized - call Initialize() first")
		}
		return runLocal(ctx, name, args...)
	}

	return client.RunDocker(ctx, name, args...)
}

func runLocal(ctx context.Context, name string, args ...string) (stdout string, stderr string, err error) {
	cmd := exec.CommandContext(ctx, name, args...)

	var stdoutBuf bytes.Buffer
	var stderrBuf bytes.Buffer
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf

	err = cmd.Run()
	stdout = stdoutBuf.String()
	stderr = stderrBuf.String()
	if err != nil {
		cmdStr := formatCommand(name, args)
		detail := strings.TrimSpace(stderr)
		if detail != "" {
			return stdout, stderr, fmt.Errorf("command %s failed: %w: %s", cmdStr, err, detail)
		}
		return stdout, stderr, fmt.Errorf("command %s failed: %w", cmdStr, err)
	}

	return stdout, stderr, nil
}
