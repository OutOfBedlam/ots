package geom

import (
	"math"

	"github.com/paulmach/orb"
	"github.com/paulmach/orb/geo"
)

// Distance returns the distance between two points on the earth.
func Distance(p1, p2 Point) float64 {
	return geo.Distance(orb.Point{p1[0], p1[1]}, orb.Point{p2[0], p2[1]})
}

// DistanceHaversine computes the distance on the earth using the
// more accurate haversine formula.
func DistanceHaversine(p1, p2 Point) float64 {
	return geo.DistanceHaversine(orb.Point{p1[0], p1[1]}, orb.Point{p2[0], p2[1]})
}

func DistanceEuclidean(p1, p2 Point) float64 {
	return math.Sqrt(math.Pow(p2[0]-p1[0], 2) + math.Pow(p2[1]-p1[1], 2))
}

// Bearing computes the direction one must start traveling on earth
// to be heading from, to the given points.
func Bearing(from, to Point) float64 {
	return geo.Bearing(orb.Point{from[0], from[1]}, orb.Point{to[0], to[1]})
}

// Midpoint returns the half-way point along a great circle path between the two points.
func Midpoint(p1, p2 Point) Point {
	p := geo.Midpoint(orb.Point{p1[0], p1[1]}, orb.Point{p2[0], p2[1]})
	return Point{p[0], p[1]}
}

// PointAtBearingAndDistance returns the point at the given bearing and distance in meters from the point
func PointAtBearingAndDistance(p Point, bearing, distance float64) Point {
	r := geo.PointAtBearingAndDistance(orb.Point{p[0], p[1]}, bearing, distance)
	return Point{r[0], r[1]}
}
