package detect

import (
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	os.Setenv("DRIFTWATCH_GDAL_MODE", "local")
	os.Exit(m.Run())
}
