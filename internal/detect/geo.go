package detect

// PixelToLonLat converts pixel coordinates to lon/lat using a GDAL-style
// affine geotransform.
func PixelToLonLat(gt [6]float64, px, py float64) (lon, lat float64) {
	lon = gt[0] + px*gt[1] + py*gt[2]
	lat = gt[3] + px*gt[4] + py*gt[5]
	return lon, lat
}
