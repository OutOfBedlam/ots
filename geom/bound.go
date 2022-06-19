package geom

import (
	"fmt"
	"math"
	"os"
)

type Bound struct {
	Min LatLon
	Max LatLon
}

func NewBound(p1, p2 LatLon) Bound {
	minLt := math.Min(p1.Lat, p2.Lat)
	minLn := math.Min(p1.Lon, p2.Lon)
	maxLt := math.Max(p1.Lat, p2.Lat)
	maxLn := math.Max(p1.Lon, p2.Lon)

	return Bound{
		Min: LatLon{Lat: minLt, Lon: minLn},
		Max: LatLon{Lat: maxLt, Lon: maxLn},
	}
}

func MakeBound(minLat, minLon, maxLat, maxLon float64) Bound {
	return Bound{
		Min: LatLon{Lat: minLat, Lon: minLon},
		Max: LatLon{Lat: maxLat, Lon: maxLon},
	}
}

func (b Bound) Pad(d float64) Bound {
	b.Min.Lat -= d
	b.Min.Lon -= d

	b.Max.Lat += d
	b.Max.Lon += d
	return b
}

// Extend grows the bound to include the new point.
func (b Bound) Extend(point LatLon) Bound {
	// already included
	if b.Contains(point) {
		return b
	}

	return Bound{
		Min: LatLon{
			Lat: math.Min(b.Min.Lat, point.Lat),
			Lon: math.Min(b.Min.Lon, point.Lon),
		},
		Max: LatLon{
			Lat: math.Max(b.Max.Lat, point.Lat),
			Lon: math.Max(b.Max.Lon, point.Lon),
		},
	}
}

// Union extends this bound to contain the union of this and the given bound.
func (b Bound) Union(other Bound) Bound {
	if other.IsEmpty() {
		return b
	}

	b = b.Extend(other.Min)
	b = b.Extend(other.Max)
	b = b.Extend(other.LeftTop())
	b = b.Extend(other.RightBottom())

	return b
}

func (bound Bound) Intersects(b2 Bound) bool {
	return bound.IntersectsCoord(b2.Min.Lat, b2.Min.Lon, b2.Max.Lat, b2.Max.Lon)
}

func (bound Bound) IntersectsCoord(minLat, minLon, maxLat, maxLon float64) bool {
	if (maxLat < bound.Min.Lat) ||
		(minLat > bound.Max.Lat) ||
		(maxLon < bound.Min.Lon) ||
		(minLon > bound.Max.Lon) {
		return false
	}

	return true
}

func (bound Bound) Contains(p LatLon) bool {
	return bound.ContainsCoord(p.Lat, p.Lon)
}

func (bound Bound) ContainsCoord(lat, lng float64) bool {
	if lat < bound.Min.Lat || bound.Max.Lat < lat {
		return false
	}

	if lng < bound.Min.Lon || bound.Max.Lon < lng {
		return false
	}

	return true
}

func (bound Bound) OnProjects(bearing float64, p LatLon) bool {
	return bound.OnProjectsCoord(bearing, p.Lat, p.Lon, "")
}

func (bound Bound) OnProjectsCoord(bearing float64, lat, lng float64, debug string) bool {
	d01 := BearingCoords(bound.Min.Lat, bound.Min.Lon, bound.Max.Lat, bound.Max.Lon)
	d02 := BearingCoords(bound.Max.Lat, bound.Max.Lon, bound.Min.Lat, bound.Min.Lon)
	d1 := BearingCoords(bound.Min.Lat, bound.Min.Lon, lat, lng)
	d2 := BearingCoords(bound.Max.Lat, bound.Max.Lon, lat, lng)
	t1 := d1 - d01
	if t1 > 180 {
		t1 = t1 - 360
	}
	// else if t1 < 180 {
	// 	t1 = t1 + 360
	// }
	t2 := d2 - d02
	if t2 > 180 {
		t2 = t2 - 360
	}
	// else if t2 < 180 {
	// 	t2 = t2 + 360
	// }
	if len(debug) > 0 {
		fmt.Fprintf(os.Stderr, "%s min:%5f,%5f max:%.5f,%.5f d01:%.f d1:%.f d02:%.f d2:%f t1:%.f t2:%.f\n",
			debug,
			bound.Min.Lat, bound.Min.Lon, bound.Max.Lat, bound.Max.Lon,
			d01, d1, d02, d2, t1, t2)
	}
	return math.Abs(t1) <= 90 && math.Abs(t2) <= 90
}

// Center returns the center of the bounds by "averaging" the x and y coords.
func (b Bound) Center() LatLon {
	return LatLon{
		Lat: (b.Min.Lat + b.Max.Lat) / 2.0,
		Lon: (b.Min.Lon + b.Max.Lon) / 2.0,
	}
}

// Top returns the top of the bound.
func (b Bound) Top() float64 {
	return b.Max.Lat
}

// Bottom returns the bottom of the bound.
func (b Bound) Bottom() float64 {
	return b.Min.Lat
}

// Right returns the right of the bound.
func (b Bound) Right() float64 {
	return b.Max.Lon
}

// Left returns the left of the bound.
func (b Bound) Left() float64 {
	return b.Min.Lon
}

// LeftTop returns the upper left point of the bound.
func (b Bound) LeftTop() LatLon {
	return LatLon{Lon: b.Left(), Lat: b.Top()}
}

// RightBottom return the lower right point of the bound.
func (b Bound) RightBottom() LatLon {
	return LatLon{Lon: b.Right(), Lat: b.Bottom()}
}

// IsEmpty returns true if it contains zero area or if
// it's in some malformed negative state where the left point is larger than the right.
// This can be caused by padding too much negative.
func (b Bound) IsEmpty() bool {
	return b.Min.Lon > b.Max.Lon || b.Min.Lat > b.Max.Lat
}
