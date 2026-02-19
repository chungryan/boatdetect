package gdal

import (
	"context"
	"fmt"
	"os"
)

// ToAAIGrid converts a GeoTIFF to an Arc/Info ASCII Grid.
func ToAAIGrid(ctx context.Context, inputTif, outputAsc string) error {
	if err := removeIfExists(outputAsc); err != nil {
		return err
	}

	_, _, err := Run(ctx, "gdal_translate", "-of", "AAIGrid", inputTif, outputAsc)
	if err != nil {
		return fmt.Errorf("gdal_translate: %w", err)
	}

	return nil
}

func removeIfExists(path string) error {
	_, err := os.Stat(path)
	if err == nil {
		return removeFile(path)
	}
	if os.IsNotExist(err) {
		return nil
	}
	return fmt.Errorf("stat output: %w", err)
}

func removeFile(path string) error {
	err := os.Remove(path)
	if err != nil {
		return fmt.Errorf("remove output: %w", err)
	}

	return nil
}
