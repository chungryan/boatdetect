package docker

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	// GDALImage is the Docker image to use for GDAL operations.
	// Uses GitHub Container Registry (ghcr.io) for reliability.
	GDALImage = "ghcr.io/osgeo/gdal:latest"
	// ContainerWorkDir is the working directory inside the container.
	ContainerWorkDir = "/work"
)

// Client handles Docker operations for GDAL using the docker CLI.
type Client struct {
	workDir string
}

// New creates a new Docker client for GDAL operations.
func New(ctx context.Context) (*Client, error) {
	// Verify Docker is installed and running
	cmd := exec.CommandContext(ctx, "docker", "ps")
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("docker not available: %w", err)
	}

	cwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("get working directory: %w", err)
	}

	return &Client{workDir: cwd}, nil
}

// Close is a no-op for CLI-based Docker client.
func (c *Client) Close() error {
	return nil
}

// Run executes a GDAL command in a Docker container using docker run.
// It mounts the current working directory and captures stdout/stderr.
func (c *Client) Run(ctx context.Context, name string, args ...string) (stdout string, stderr string, err error) {
	// Ensure the image is available
	if err := c.ensureImage(ctx); err != nil {
		return "", "", fmt.Errorf("ensure image: %w", err)
	}

	// Convert absolute paths to container paths, but skip flags and non-path arguments
	convertedArgs := make([]string, len(args))
	for i, arg := range args {
		convertedArgs[i] = c.convertArgForDocker(arg)
	}

	// Build docker run command
	dockerArgs := []string{
		"run",
		"--rm",
		"-v", fmt.Sprintf("%s:%s", c.workDir, ContainerWorkDir),
		"-w", ContainerWorkDir,
		GDALImage,
		name,
	}
	dockerArgs = append(dockerArgs, convertedArgs...)

	// Execute the command
	cmd := exec.CommandContext(ctx, "docker", dockerArgs...)

	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf

	err = cmd.Run()
	stdout = stdoutBuf.String()
	stderr = stderrBuf.String()

	if err == nil {
		return stdout, stderr, nil
	}

	cmdStr := formatCommand(name, args)
	return stdout, stderr, formatDockerCommandError(cmdStr, err, stderr)
}

func (c *Client) convertArgForDocker(arg string) string {
	if strings.HasPrefix(arg, "-") {
		return arg
	}
	if !shouldConvertPath(arg) {
		return arg
	}

	return c.convertPath(arg)
}

// convertPath converts absolute file paths to container paths.
// If a path is within the working directory, it's converted to a relative path.
func (c *Client) convertPath(filePath string) string {
	// If it's already a relative path, prepend the container work directory
	if !filepath.IsAbs(filePath) {
		return filepath.Join(ContainerWorkDir, filePath)
	}

	// If the absolute path is within the current working directory, convert it
	relPath, err := filepath.Rel(c.workDir, filePath)
	if err == nil && !strings.HasPrefix(relPath, "..") {
		// It's within the workdir
		return filepath.Join(ContainerWorkDir, relPath)
	}

	// For files outside workdir, use just the basename
	return filepath.Join(ContainerWorkDir, filepath.Base(filePath))
}

// ensureImage ensures the GDAL image is available locally.
// If not, it pulls the image from Docker Hub.
func (c *Client) ensureImage(ctx context.Context) error {
	// Check if image exists
	cmd := exec.CommandContext(ctx, "docker", "inspect", GDALImage)
	if err := cmd.Run(); err == nil {
		return nil // Image already exists
	}

	// Pull the image
	cmd = exec.CommandContext(ctx, "docker", "pull", GDALImage)
	var stderrBuf bytes.Buffer
	cmd.Stderr = &stderrBuf

	err := cmd.Run()
	if err == nil {
		return nil
	}

	return formatDockerPullError(err, stderrBuf.String())
}

func formatDockerCommandError(command string, commandErr error, stderr string) error {
	detail := strings.TrimSpace(stderr)
	if detail == "" {
		return fmt.Errorf("command %s failed: %w", command, commandErr)
	}

	return fmt.Errorf("command %s failed: %s", command, detail)
}

func formatDockerPullError(commandErr error, stderr string) error {
	detail := strings.TrimSpace(stderr)
	if detail == "" {
		return fmt.Errorf("pull image: %w", commandErr)
	}

	return fmt.Errorf("pull image: %s", detail)
}

func formatCommand(name string, args []string) string {
	parts := []string{quoteArg(name)}
	for _, arg := range args {
		parts = append(parts, quoteArg(arg))
	}
	return strings.Join(parts, " ")
}

func quoteArg(arg string) string {
	if arg == "" || strings.ContainsAny(arg, " \t\n\r\"\\") {
		return strconv.Quote(arg)
	}
	return arg
}

// shouldConvertPath checks if an argument is likely a file path that needs conversion.
func shouldConvertPath(arg string) bool {
	if arg == "" {
		return false
	}

	// Contains path separators - definitely a path
	if strings.ContainsAny(arg, "/\\") {
		return true
	}

	// Starts with . or ~ (relative paths)
	if strings.HasPrefix(arg, ".") || strings.HasPrefix(arg, "~") {
		return true
	}

	// Check for known file extensions
	lowerArg := strings.ToLower(arg)
	commonGISExts := ".tif.tiff.tif.zip.asc.grd.hdr.jp2.j2k.img.hdf.h5.nc.netcdf.vrt.xml.geojson.json.shp.shx.dbf.gml.gpkg.las.laz"
	for _, ext := range strings.Split(commonGISExts, ".")[1:] {
		if strings.HasSuffix(lowerArg, "."+ext) {
			return true
		}
	}

	return false
}
