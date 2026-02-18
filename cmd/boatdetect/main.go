package main

import (
	"context"
	"fmt"
	"os"

	"boatdetect/internal/gdal"
)

func main() {
	ctx := context.Background()
	if err := gdal.Initialize(ctx); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer gdal.Shutdown()

	opts, err := parseFlags()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}

	if err := runDetect(ctx, os.Stdout, opts); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
