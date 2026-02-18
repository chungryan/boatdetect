package gdal

import (
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	os.Setenv("BOATDETECT_GDAL_MODE", "local")
	os.Exit(m.Run())
}
