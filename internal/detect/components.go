package detect

import (
	"math"

	"boatdetect/internal/gdal"
)

// Component describes a connected component on a thresholded grid.
type Component struct {
	Area int
	Sum  float64
	Cx   float64
	Cy   float64
}

// Components extracts 4-neighborhood connected components from a thresholded grid.
func Components(grid gdal.Grid, threshold float64, invert bool, minAreaPx int) []Component {
	if grid.Width <= 0 || grid.Height <= 0 {
		return nil
	}

	expected := grid.Width * grid.Height
	if len(grid.Data) < expected {
		return nil
	}

	visited := make([]bool, expected)
	checkNoData := !math.IsNaN(grid.NoData)
	components := make([]Component, 0)

	qualifies := func(v float64) bool {
		if math.IsNaN(v) {
			return false
		}
		if checkNoData && v == grid.NoData {
			return false
		}
		if invert {
			return v <= threshold
		}
		return v >= threshold
	}

	for idx := 0; idx < expected; idx++ {
		if visited[idx] {
			continue
		}
		v := grid.Data[idx]
		if !qualifies(v) {
			visited[idx] = true
			continue
		}

		component := floodFillComponent(grid, idx, visited, qualifies)
		area := component.Area

		if area == 0 || area < minAreaPx {
			continue
		}

		components = append(components, Component{
			Area: area,
			Sum:  component.Sum,
			Cx:   component.Cx,
			Cy:   component.Cy,
		})
	}

	return components
}

func floodFillComponent(grid gdal.Grid, startIdx int, visited []bool, qualifies func(float64) bool) Component {
	area := 0
	sum := 0.0
	sumX := 0.0
	sumY := 0.0

	stack := []int{startIdx}
	visited[startIdx] = true

	for len(stack) > 0 {
		n := len(stack) - 1
		cur := stack[n]
		stack = stack[:n]

		cv := grid.Data[cur]
		if !qualifies(cv) {
			continue
		}

		x := cur % grid.Width
		y := cur / grid.Width

		area++
		sum += cv
		sumX += float64(x)
		sumY += float64(y)

		stack = addNeighbors(cur, x, y, grid.Width, grid.Height, visited, stack)
	}

	if area == 0 {
		return Component{}
	}

	return Component{
		Area: area,
		Sum:  sum,
		Cx:   sumX / float64(area),
		Cy:   sumY / float64(area),
	}
}

func addNeighbors(cur, x, y, width, height int, visited []bool, stack []int) []int {
	if x > 0 {
		stack = addIfUnvisited(cur-1, visited, stack)
	}
	if x+1 < width {
		stack = addIfUnvisited(cur+1, visited, stack)
	}
	if y > 0 {
		stack = addIfUnvisited(cur-width, visited, stack)
	}
	if y+1 < height {
		stack = addIfUnvisited(cur+width, visited, stack)
	}
	return stack
}

func addIfUnvisited(idx int, visited []bool, stack []int) []int {
	if visited[idx] {
		return stack
	}
	visited[idx] = true
	return append(stack, idx)
}
