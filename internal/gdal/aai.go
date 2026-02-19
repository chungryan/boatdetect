package gdal

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// Grid represents an ESRI ASCII grid (AAIGrid).
type Grid struct {
	Width  int
	Height int
	NoData float64
	Data   []float64
}

// ParseAAIGrid reads an ESRI ASCII grid from r.
func ParseAAIGrid(r io.Reader) (Grid, error) {
	reader := bufio.NewReader(r)
	fields, err := parseHeaderFields(reader)
	if err != nil {
		return Grid{}, err
	}

	width, height, err := parseGridDimensions(fields)
	if err != nil {
		return Grid{}, err
	}

	err = validateRequiredFloatHeaders(fields)
	if err != nil {
		return Grid{}, err
	}

	// nodata_value is optional, default to -9999 if not present
	nodata, err := parseNoDataValue(fields)
	if err != nil {
		return Grid{}, err
	}

	expected := width * height
	data, err := parseGridData(reader, expected, nodata)
	if err != nil {
		return Grid{}, err
	}

	err = validateNoTrailingData(reader)
	if err != nil {
		return Grid{}, err
	}

	return Grid{
		Width:  width,
		Height: height,
		NoData: nodata,
		Data:   data,
	}, nil
}

func parseHeaderFields(reader *bufio.Reader) (map[string]string, error) {
	fields := make(map[string]string, 6)
	for len(fields) < 6 {
		key, value, err := scanHeaderPair(reader)
		if err != nil {
			return nil, err
		}

		fields[strings.ToLower(key)] = value
	}

	return fields, nil
}

func scanHeaderPair(reader *bufio.Reader) (string, string, error) {
	key, err := scanHeaderToken(reader, "key")
	if err != nil {
		return "", "", err
	}

	value, err := scanHeaderToken(reader, "value")
	if err != nil {
		return "", "", err
	}

	return key, value, nil
}

func scanHeaderToken(reader *bufio.Reader, tokenName string) (string, error) {
	var token string
	_, err := fmt.Fscan(reader, &token)
	if err == nil {
		return token, nil
	}
	if err == io.EOF {
		return "", fmt.Errorf("parse header: unexpected EOF")
	}

	return "", fmt.Errorf("parse header %s: %w", tokenName, err)
}

func parseNoDataValue(fields map[string]string) (float64, error) {
	nodata := -9999.0
	value, ok := fields["nodata_value"]
	if !ok {
		return nodata, nil
	}

	parsed, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0, fmt.Errorf("parse header: nodata_value=%q: %w", value, err)
	}

	return parsed, nil
}

func parseGridData(reader *bufio.Reader, expected int, nodata float64) ([]float64, error) {
	data := make([]float64, 0, expected)
	for len(data) < expected {
		value, err := scanDataValue(reader)
		if err == nil {
			data = append(data, value)
			continue
		}

		if err != io.EOF {
			return nil, fmt.Errorf("parse data value: %w", err)
		}

		if len(data) < (expected*99)/100 {
			return nil, fmt.Errorf("parse data: expected %d values, got %d", expected, len(data))
		}

		break
	}

	if len(data) == expected {
		return data, nil
	}

	missing := expected - len(data)
	for i := 0; i < missing; i++ {
		data = append(data, nodata)
	}

	return data, nil
}

func scanDataValue(reader *bufio.Reader) (float64, error) {
	var value float64
	_, err := fmt.Fscan(reader, &value)
	if err != nil {
		return 0, err
	}

	return value, nil
}

func validateNoTrailingData(reader *bufio.Reader) error {
	var extra string
	_, err := fmt.Fscan(reader, &extra)
	if err == io.EOF {
		return nil
	}
	if err != nil {
		return fmt.Errorf("parse data: %w", err)
	}

	return fmt.Errorf("parse data: unexpected trailing value %q", extra)
}

func parseGridDimensions(fields map[string]string) (int, int, error) {
	width, err := parseHeaderInt(fields, "ncols")
	if err != nil {
		return 0, 0, err
	}

	height, err := parseHeaderInt(fields, "nrows")
	if err != nil {
		return 0, 0, err
	}

	return width, height, nil
}

func validateRequiredFloatHeaders(fields map[string]string) error {
	requiredKeys := []string{"xllcorner", "yllcorner", "cellsize"}
	for _, key := range requiredKeys {
		_, err := parseHeaderFloat(fields, key)
		if err != nil {
			return err
		}
	}

	return nil
}

func parseHeaderInt(fields map[string]string, key string) (int, error) {
	value, ok := fields[key]
	if !ok {
		return 0, fmt.Errorf("parse header: missing %s", key)
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("parse header: %s=%q: %w", key, value, err)
	}
	return parsed, nil
}

func parseHeaderFloat(fields map[string]string, key string) (float64, error) {
	value, ok := fields[key]
	if !ok {
		return 0, fmt.Errorf("parse header: missing %s", key)
	}
	parsed, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0, fmt.Errorf("parse header: %s=%q: %w", key, value, err)
	}
	return parsed, nil
}
