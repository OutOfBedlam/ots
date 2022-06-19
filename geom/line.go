package geom

import "math"

type Line struct {
	Start LatLon
	End   LatLon
}

func NewLine(start LatLon, end LatLon) Line {
	return Line{
		Start: start,
		End:   end,
	}
}

func (l Line) DistanceToPoint(pt Point) float64 {
	a, b, c := l.Coefficients()
	return math.Abs(a*pt[0]+b*pt[1]+c) / math.Sqrt(a*a+b*b)
}

// Cartesian distance
func (l Line) DistanceTo(coord LatLon) float64 {
	a, b, c := l.Coefficients()
	return math.Abs(a*coord.Lon+b*coord.Lat+c) / math.Sqrt(a*a+b*b)
}

// returns the three coefficients that define a line
// A line can be defined by following equation.
//
// ax + by + c = 0
//
func (l Line) Coefficients() (a, b, c float64) {
	a = l.Start.Lat - l.End.Lat
	b = l.End.Lon - l.Start.Lon
	c = l.Start.Lon*l.End.Lat - l.End.Lon*l.Start.Lat
	return a, b, c
}

func (l Line) SeekMostDistant(points []LatLon) (idx int, maxDist float64) {
	for i, p := range points {
		d := l.DistanceTo(p)
		if d > maxDist {
			maxDist = d
			idx = i
		}
	}
	return
}

func (l Line) SeekMostDistantPoint(points []Point) (idx int, maxDist float64) {
	for i, p := range points {
		d := l.DistanceToPoint(p)
		if d > maxDist {
			maxDist = d
			idx = i
		}
	}
	return
}
