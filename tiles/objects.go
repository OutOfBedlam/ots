package tiles

import (
	"image/color"
	"math"

	"github.com/OutOfBedlam/ots/geom"
	"github.com/fogleman/gg"
)

type Object interface {
	Draw(dc *gg.Context, coordTrans CoordTransFunc)
	// returns euclidean distance from 'from' coordinates
	DistanceFrom(from geom.LatLon) float64
	Layer() Layer
	SourceInfo() string
	Visible(zoom int) bool
}

//#region Label

type Label struct {
	text        string
	coord       geom.LatLon
	rotate      float64
	textColor   color.Color
	icon        Icon
	iconSize    float64
	sourceInfo  string
	visibleFunc func(int) bool
}

func (label *Label) Layer() Layer {
	return LayerLabel
}

func (label *Label) Visible(zoom int) bool {
	if label.visibleFunc != nil {
		return label.visibleFunc(zoom)
	}
	return false
}

func (label *Label) SourceInfo() string {
	return label.sourceInfo
}

func (label *Label) DistanceFrom(from geom.LatLon) float64 {
	return geom.DistanceEuclidean(label.coord.Point(), from.Point())
}

func (label *Label) Draw(dc *gg.Context, transCoord CoordTransFunc) {
	if len(label.text) == 0 && label.icon == nil {
		return
	}

	dc.Push()
	if label.textColor != nil {
		dc.SetColor(label.textColor)
	} else {
		dc.SetHexColor("#000000")
	}

	var iconSize = 28.0
	if label.iconSize > 0 {
		iconSize = label.iconSize
	}

	x, y := transCoord(label.coord)
	if label.icon != nil {
		label.icon.Draw(dc, x, y, iconSize)
	}
	if len(label.text) > 0 {
		if label.rotate != 0.0 {
			dc.RotateAbout(label.rotate, x, y)
		}

		ax, ay := 0.5, 0.5
		if label.icon != nil {
			y += iconSize / 2
			ay = 0
		}
		dc.DrawStringWrapped(label.text, x, y, ax, ay, 0.8, 1.06, gg.AlignCenter)
	}

	dc.Pop()
}

//#endregion

//#region PolygonObject
type PolygonObject struct {
	outer       []geom.LatLon
	inner       []geom.LatLon
	layer       Layer
	fillColor   color.Color
	lineColor   color.Color
	lineWidth   float64
	lineDash    []float64
	area        float64
	sourceInfo  string
	visibleFunc func(int) bool
}

func (obj *PolygonObject) SetLayer(l Layer) {
	obj.layer = l
}

func (obj *PolygonObject) Layer() int {
	if obj.layer == 0 {
		return LayerNature
	}
	return obj.layer
}

func (obj *PolygonObject) SourceInfo() string {
	return obj.sourceInfo
}

func (obj *PolygonObject) Visible(zoom int) bool {
	if obj.visibleFunc != nil {
		return obj.visibleFunc(zoom)
	}
	return false
}

func (obj *PolygonObject) Area() float64 {
	if obj.area == 0 {
		obj.area = _calcArea(obj.outer)
	}
	return obj.area
}

func (obj *PolygonObject) DistanceFrom(from geom.LatLon) float64 {
	lenOuter := len(obj.outer)
	if lenOuter == 0 {
		return math.MaxFloat64
	} else if lenOuter == 1 {
		return geom.DistanceEuclidean(from.Point(), obj.outer[0].Point())
	}

	return _minDistanceFrom(obj.outer, from)
}

func (obj *PolygonObject) Draw(dc *gg.Context, transCoord CoordTransFunc) {
	if len(obj.outer) < 2 {
		return
	}
	dc.Push()
	if len(obj.inner) > 2 {
		x, y := transCoord(obj.inner[0])
		dc.MoveTo(x, y)
		for _, in := range obj.inner[1:] {
			x, y := transCoord(in)
			dc.LineTo(x, y)
		}
		dc.Clip()
		dc.InvertMask()
		dc.ClearPath()
	}

	if obj.fillColor != nil {
		x, y := transCoord(obj.outer[0])
		dc.MoveTo(x, y)
		for _, n := range obj.outer[1:] {
			x, y = transCoord(n)
			dc.LineTo(x, y)
		}
		dc.SetColor(obj.fillColor)
		dc.Fill()
		dc.ClearPath()
	}
	if obj.lineColor != nil {
		x, y := transCoord(obj.outer[0])
		dc.MoveTo(x, y)
		for _, n := range obj.outer[1:] {
			x, y = transCoord(n)
			dc.LineTo(x, y)
		}
		if len(obj.lineDash) > 0 {
			dc.SetDash(obj.lineDash...)
		}
		dc.SetColor(obj.lineColor)
		if obj.lineWidth > 0 {
			dc.SetLineWidth(obj.lineWidth)
		}
		dc.Stroke()
	}
	dc.ResetClip()
	dc.Pop()
}

//#endregion

//#region MultiPolygonObject
type MultiPolygonObject struct {
	outers      [][]geom.LatLon
	inners      [][]geom.LatLon
	layer       Layer
	fillColor   color.Color
	lineColor   color.Color
	lineWidth   float64
	lineDash    []float64
	area        float64
	sourceInfo  string
	visibleFunc func(int) bool
}

func (obj *MultiPolygonObject) DistanceFrom(from geom.LatLon) float64 {
	min := math.MaxFloat64
	for _, outer := range obj.outers {
		if len(outer) == 0 {
			continue
		}
		d := _minDistanceFrom(outer, from)
		if d < min {
			min = d
		}
	}
	return min
}

func (mp *MultiPolygonObject) Draw(dc *gg.Context, transCoord CoordTransFunc) {
	if len(mp.outers) == 0 {
		return
	}
	dc.Push()
	for _, in := range mp.inners {
		x, y := transCoord(in[0])
		dc.MoveTo(x, y)
		for _, n := range in[1:] {
			x, y = transCoord(n)
			dc.LineTo(x, y)
		}
	}
	if len(mp.inners) > 0 {
		dc.Clip()
		dc.InvertMask()
		dc.ClearPath()
	}

	if mp.fillColor != nil {
		dc.SetColor(mp.fillColor)
		for _, out := range mp.outers {
			x, y := transCoord(out[0])
			dc.MoveTo(x, y)
			for _, n := range out[1:] {
				x, y = transCoord(n)
				dc.LineTo(x, y)
			}
		}
		dc.Fill()
		dc.ClearPath()
	}

	if mp.lineColor != nil {
		for _, out := range mp.outers {
			x, y := transCoord(out[0])
			dc.MoveTo(x, y)
			for _, n := range out[1:] {
				x, y = transCoord(n)
				dc.LineTo(x, y)
			}
		}
		if len(mp.lineDash) > 0 {
			dc.SetDash(mp.lineDash...)
		}
		dc.SetColor(mp.lineColor)
		if mp.lineWidth > 0 {
			dc.SetLineWidth(mp.lineWidth)
		}
		dc.Stroke()
	}
	dc.ResetClip()
	dc.Pop()
}
func (mp *MultiPolygonObject) Layer() Layer {
	return mp.layer
}

func (obj *MultiPolygonObject) Visible(zoom int) bool {
	if obj.visibleFunc != nil {
		return obj.visibleFunc(zoom)
	}
	return false
}

func (mp *MultiPolygonObject) SourceInfo() string {
	return mp.sourceInfo
}

func (mp *MultiPolygonObject) Area() float64 {
	if mp.area == 0 {
		for _, o := range mp.outers {
			mp.area += _calcArea(o)
		}
	}
	return mp.area
}

//#endregion

func _minDistanceFrom(points []geom.LatLon, from geom.LatLon) float64 {
	min := math.MaxFloat64
	lenOuter := len(points)
	if lenOuter == 0 {
		return min
	} else if lenOuter == 1 {
		return geom.DistanceEuclidean(points[0].Point(), from.Point())
	}
	for i := 1; i < lenOuter; i++ {
		p1 := points[i-1]
		p2 := points[i]
		d := _calcDistanceFromLine(p1, p2, from)

		// only use distance from line
		if d < min {
			min = d
		}
		/*
			bound := geom.MakeBound(p1, p2)
			fromBound := geom.NewBound(from.Lat-d*3, from.Lon-d*3, from.Lat+d*3, from.Lon+d*3)
			if bound.Intersects(fromBound) || fromBound.Intersects(bound) {
				if d < min {
					min = d
				}
			} else {
				d1 := geom.DistanceEuclidean(from, p1)
				d2 := geom.DistanceEuclidean(from, p2)
				if d1 < d2 && d1 < min {
					min = d1
				} else if d2 < d1 && d2 < min {
					min = d2
				}
			}
		*/
	}
	return min
}

func _calcDistanceFromLine(start, end, coord geom.LatLon) float64 {
	a := start.Lat - end.Lat
	b := end.Lon - start.Lon
	c := start.Lon*end.Lat - end.Lon*start.Lat
	return math.Abs(a*coord.Lon+b*coord.Lat+c) / math.Sqrt(a*a+b*b)
}

func _calcArea(linestring []geom.LatLon) float64 {
	var minx, miny float64 = linestring[0].Lon, linestring[0].Lat
	var maxx, maxy float64 = linestring[0].Lon, linestring[0].Lat

	for _, n := range linestring {
		var x, y = n.Lon, n.Lat
		if x < minx {
			minx = x
		}
		if x > maxx {
			maxx = x
		}
		if y < miny {
			miny = y
		}
		if y > maxy {
			maxy = y
		}
	}
	return (maxx - minx) * (maxy - miny)
}
