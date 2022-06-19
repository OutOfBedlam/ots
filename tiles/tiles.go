package tiles

import (
	"context"
	"fmt"
	"image/color"
	"io"
	"math"

	"github.com/OutOfBedlam/ots/geom"
	"github.com/OutOfBedlam/ots/logging"
	"github.com/OutOfBedlam/ots/projection"
	"github.com/fogleman/gg"
	"github.com/golang/freetype/truetype"
	"github.com/paulmach/osm"
)

type TileBuilder interface {
	AddWays(ways ...*Way)
	AddNodes(nodes ...*Node)
	AddRelations(rels ...*Relation)
	Build(ctx context.Context) (*Tile, error)
	SetBuildLayerRange(start, end int)
	SetHideLabels(bool)
	SetVerbose(bool)
	SetWatermark(string)
	SetTint(bool)
}

type Tile struct {
	width, height   int
	defaultFont     *truetype.Font
	objs            []Object
	coordTranslator CoordTransFunc
	watermark       string
	tint            bool
}

func TilesToBounds(x, y, z int) geom.Bound {
	maxLat, minLon := projection.Tile2LatLon(x, y, z)
	minLat, maxLon := projection.Tile2LatLon(x+1, y+1, z)
	return geom.Bound{
		Min: geom.LatLon{Lat: minLat, Lon: minLon},
		Max: geom.LatLon{Lat: maxLat, Lon: maxLon},
	}
}

func NewBuilder(x, y, z int) TileBuilder {
	builder := &DefaultBuilder{
		log:             logging.GetLog(fmt.Sprintf("tiles-%d-%d-%d", z, x, y)),
		canvasWidth:     512,
		canvasHeight:    512,
		zoom:            z,
		buildLayerStart: 0,
		buildLayerEnd:   math.MaxInt,
	}

	// requested bounds
	maxLat, minLon := projection.Tile2LatLon(x, y, z)
	minLat, maxLon := projection.Tile2LatLon(x+1, y+1, z)

	pixelPerLat := projection.TileSize / (maxLat - minLat)
	pixelPerLon := projection.TileSize / (maxLon - minLon)

	dpiXScale := builder.canvasWidth / projection.TileSize
	dpiYScale := builder.canvasHeight / projection.TileSize

	builder.bounds = geom.MakeBound(minLat, minLon, maxLat, maxLon)

	// converter: lat/lon to local (gg.Context) x,y coord
	builder.transCoordToXY = func(p geom.LatLon) (float64, float64) {
		x := math.Ceil((p.Lon-minLon)*pixelPerLon) * dpiXScale
		y := math.Ceil((maxLat-p.Lat)*pixelPerLat) * dpiYScale
		return x, y
	}

	return builder
}

func NewBuilderBounds(bounds geom.Bound, outputWidth, outputHeight float64) TileBuilder {
	builder := &DefaultBuilder{
		log:          logging.GetLog("bounds"),
		canvasWidth:  outputWidth,
		canvasHeight: outputHeight,
		zoom:         projection.TileZoom(5), // 5 meters/pixel
	}

	tileSize := math.Min(outputWidth, outputHeight)

	minLon := bounds.Min.Lon
	maxLat := bounds.Max.Lat
	pixelPerLat := tileSize / (maxLat - bounds.Min.Lat)
	pixelPerLon := tileSize / (bounds.Max.Lon - minLon)

	pixelPerCoord := math.Min(pixelPerLat, pixelPerLon)
	pixelPerLat = pixelPerCoord
	pixelPerLon = pixelPerCoord

	dpiXScale := 1.0
	dpiYScale := 1.0

	builder.bounds = bounds
	builder.transCoordToXY = func(p geom.LatLon) (float64, float64) {
		x := math.Ceil((p.Lon-minLon)*pixelPerLon) * dpiXScale
		y := math.Ceil((maxLat-p.Lat)*pixelPerLat) * dpiYScale
		return x, y
	}

	return builder
}

func (r *Relation) FindTag(key string) string {
	if v, b := r.Tags[key]; b {
		return v
	} else {
		return ""
	}
}

func (w *Way) FindTag(key string) string {
	if v, b := w.Tags[key]; b {
		return v
	} else {
		return ""
	}
}

func (n *Node) FindTag(key string) string {
	if v, b := n.Tags[key]; b {
		return v
	} else {
		return ""
	}
}

func (t *Tile) addLast(obj ...Object) {
	if obj == nil || len(obj) == 0 {
		return
	}
	if t.objs == nil {
		t.objs = make([]Object, 0)
	}
	t.objs = append(t.objs, obj...)
}

func (t *Tile) addFirst(obj ...Object) {
	if obj == nil || len(obj) == 0 {
		return
	}
	if t.objs == nil {
		t.objs = make([]Object, 0)
	}
	t.objs = append(obj, t.objs...)
}

func (t *Tile) CountObjects() int {
	return len(t.objs)
}

func (t *Tile) EncodePNG(writer io.Writer) error {
	// canvas
	canvas := gg.NewContext(t.width, t.height)

	if t.defaultFont != nil {
		face := truetype.NewFace(t.defaultFont, &truetype.Options{Size: 20})
		canvas.SetFontFace(face)
		defer face.Close()
	}
	for _, obj := range t.objs {
		obj.Draw(canvas, t.coordTranslator)
	}
	err := canvas.EncodePNG(writer)
	return err
}

func (t *Tile) AddWatermark(text string, tint bool) {
	face := truetype.NewFace(t.defaultFont, &truetype.Options{Size: 60})
	wm := &Watermark{
		text:      text,
		size:      math.Min(float64(t.width), float64(t.height)),
		fontFace:  face,
		textColor: color.RGBA{R: 0x00, G: 0x00, B: 0x00, A: 0x30},
	}

	if tint {
		wm.tintColor = color.RGBA{R: 0x00, G: 0x00, B: 0x00, A: 0x0F}
	}

	t.addLast(wm)
}

func RelationMemberType(t osm.Type) Relation_MemberType {
	switch t {
	default: // will not support other types
		// TypeChangeset, TypeNote, TypeUser
		return Relation_UNKNOWN
	case osm.TypeNode:
		return Relation_NODE
	case osm.TypeWay:
		return Relation_WAY
	case osm.TypeRelation:
		return Relation_RELATION
	case osm.TypeBounds:
		return Relation_BOUNDS
	}
}
