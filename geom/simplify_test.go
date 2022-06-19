package geom_test

import (
	"testing"

	. "github.com/OutOfBedlam/ots/geom"
)

func TestSimplifyPath(t *testing.T) {
	points := []Point{
		{0, 0},
		{1, 2},
		{2, 7},
		{3, 1},
		{4, 8},
		{5, 2},
		{6, 8},
		{7, 3},
		{8, 3},
		{9, 0},
	}

	t.Run("Threshold=0", func(t *testing.T) {
		if len(SimplifyPath(points, 0)) != 10 {
			t.Error("simplified path should have all points")
		}
	})

	t.Run("Threshold=2", func(t *testing.T) {
		if len(SimplifyPath(points, 2)) != 7 {
			t.Error("simplified path should only have 7 points")
		}
	})

	t.Run("Threshold=5", func(t *testing.T) {
		if len(SimplifyPath(points, 100)) != 2 {
			t.Error("simplified path should only have two points")
		}
	})
}

func TestSimplifyTrajectory(t *testing.T) {
	points := []LatLon{
		{0, 0},
		{2, 1},
		{7, 2},
		{1, 3},
		{8, 4},
		{2, 5},
		{8, 6},
		{3, 7},
		{3, 8},
		{0, 9},
	}

	t.Run("Threshold=0", func(t *testing.T) {
		if len(SimplifyTrajectory(points, 0)) != 10 {
			t.Error("simplified path should have all points")
		}
	})

	t.Run("Threshold=2", func(t *testing.T) {
		if len(SimplifyTrajectory(points, 2)) != 7 {
			t.Error("simplified path should only have 7 points")
		}
	})

	t.Run("Threshold=5", func(t *testing.T) {
		if len(SimplifyTrajectory(points, 100)) != 2 {
			t.Error("simplified path should only have two points")
		}
	})
}
func TestSeekMostDistantPoint(t *testing.T) {
	l := Line{Start: LatLon{Lat: 0, Lon: 0}, End: LatLon{Lat: 0, Lon: 10}}
	points := []Point{
		{13, 13},
		{1, 15},
		{1, 1},
		{3, 6},
	}

	idx, maxDist := l.SeekMostDistantPoint(points)

	if idx != 1 {
		t.Error("failed to find most distant point away from a line")
	}

	if maxDist != 15 {
		t.Error("maximum distance is incorrect")
	}
}
