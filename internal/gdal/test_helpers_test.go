package gdal

import "testing"

func useLocalGDAL(t *testing.T) {
	t.Helper()
	t.Setenv("BOATDETECT_GDAL_MODE", "local")
}
