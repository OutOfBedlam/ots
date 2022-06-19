package geom

// Ram-Douglas-Peucker simplify
func SimplifyPath(points []Point, ep float64) []Point {
	if len(points) <= 2 {
		return points
	}

	l := Line{Start: LatLngFromPoint(points[0]), End: LatLngFromPoint(points[len(points)-1])}

	idx, maxDist := l.SeekMostDistantPoint(points)
	if maxDist >= ep {
		left := SimplifyPath(points[:idx+1], ep)
		right := SimplifyPath(points[idx:], ep)
		return append(left[:len(left)-1], right...)
	}

	return []Point{points[0], points[len(points)-1]}
}

// Ram-Douglas-Peucker simplify
func SimplifyTrajectory(points []LatLon, ep float64) []LatLon {
	if len(points) <= 2 {
		return points
	}

	l := NewLine(points[0], points[len(points)-1])

	idx, maxDist := l.SeekMostDistant(points)
	if maxDist >= ep {
		left := SimplifyTrajectory(points[:idx+1], ep)
		right := SimplifyTrajectory(points[idx:], ep)
		return append(left[:len(left)-1], right...)
	}

	return []LatLon{points[0], points[len(points)-1]}
}
