package tiles

import (
	"context"
	_ "embed"
	"fmt"
	"math"
	reflect "reflect"
	"sort"

	"github.com/OutOfBedlam/ots/geom"
	"github.com/OutOfBedlam/ots/logging"
	"github.com/golang/freetype/truetype"
	lru "github.com/hashicorp/golang-lru"
	"github.com/tidwall/btree"
)

//go:embed fonts/D2Coding-Ver1.3.2-20180524.ttf
var ttfD2CodingData []byte
var FontD2Coding *truetype.Font

var objectCache *lru.Cache

func init() {
	var err error
	if FontD2Coding, err = truetype.Parse(ttfD2CodingData); err != nil {
		panic("invalid font")
	}

	if objectCache, err = lru.New(2000); err != nil {
		panic("fail to create cache")
	}
}

type CoordTransFunc func(coord geom.LatLon) (float64, float64)

type DefaultBuilder struct {
	ctx             context.Context
	log             logging.Log
	verbose         bool
	tint            bool
	hideLabels      bool
	watermark       string
	canvasWidth     float64
	canvasHeight    float64
	buildLayerStart int
	buildLayerEnd   int
	bounds          geom.Bound
	transCoordToXY  func(geom.LatLon) (float64, float64)
	ways            btree.Map[int64, *Way]
	relations       btree.Map[int64, *Relation]
	nodes           btree.Map[int64, *Node]
	zoom            int
	customStyler    StyleFunc
}

func (br *DefaultBuilder) SetVerbose(v bool) {
	br.verbose = v
}

func (br *DefaultBuilder) SetWatermark(wm string) {
	br.watermark = wm
}

func (br *DefaultBuilder) SetTint(b bool) {
	br.tint = b
}

func (br *DefaultBuilder) SetHideLabels(b bool) {
	br.hideLabels = b
}

func (br *DefaultBuilder) SetBuildLayerRange(start, end int) {
	br.buildLayerStart = start
	br.buildLayerEnd = end
}

func (br *DefaultBuilder) AddWays(ways ...*Way) {
	for _, way := range ways {
		if way == nil {
			continue
		}
		br.ways.Set(way.Id, way)
	}
}

func (br *DefaultBuilder) AddNodes(nodes ...*Node) {
	for _, node := range nodes {
		if node == nil {
			continue
		}
		br.nodes.Set(node.Id, node)
	}
}

func (br *DefaultBuilder) AddRelations(rels ...*Relation) {
	for _, rel := range rels {
		if rel == nil {
			continue
		}
		br.relations.Set(rel.Id, rel)
	}
}

func (br *DefaultBuilder) Build(ctx context.Context) (*Tile, error) {
	center := br.bounds.Center()
	radius := geom.DistanceEuclidean(center.Point(), br.bounds.Max.Point()) * 1.1 // 10% larger for padding

	objects := make([]Object, 0)
	for _, rel := range br.relations.Values() {
		cacheKey := fmt.Sprintf("REL:%d", rel.Id)
		var rset []Object
		if objs, ok := objectCache.Get(cacheKey); ok {
			rset = objs.([]Object)
		} else {
			rset = br.compileRelation(rel)
			objectCache.Add(cacheKey, rset)
		}
		for _, o := range rset {
			if o.Visible(br.zoom) && o.DistanceFrom(center) <= radius {
				objects = append(objects, o)
			}
		}
	}
	for _, way := range br.ways.Values() {
		cacheKey := fmt.Sprintf("WAY:%d", way.Id)
		var rset []Object
		if objs, ok := objectCache.Get(cacheKey); ok {
			rset = objs.([]Object)
		} else {
			rset = br.compileWay(way)
			objectCache.Add(cacheKey, rset)
		}
		for _, o := range rset {
			if o.Visible(br.zoom) && o.DistanceFrom(center) <= radius {
				objects = append(objects, o)
			}
		}
	}
	for _, node := range br.nodes.Values() {
		cacheKey := fmt.Sprintf("NODE:%d", node.Id)
		var rset []Object
		if objs, ok := objectCache.Get(cacheKey); ok {
			rset = objs.([]Object)
		} else {
			rset = br.compileNode(node)
			objectCache.Add(cacheKey, rset)
		}
		for _, o := range rset {
			if o.Visible(br.zoom) && o.DistanceFrom(center) <= radius {
				objects = append(objects, o)
			}
		}
	}

	br.ctx = ctx
	tile := &Tile{
		width:       int(br.canvasWidth),
		height:      int(br.canvasHeight),
		defaultFont: FontD2Coding,
		objs:        objects,
	}

	// z-order layers
	sort.Slice(tile.objs, func(i, j int) bool {
		lo := tile.objs[i]
		ro := tile.objs[j]
		return LayerCompareOrder(lo, ro)
	})

	// draw subset or fullset of layer?
	var start = br.buildLayerStart
	var end = br.buildLayerEnd
	if len(tile.objs) < end {
		end = len(tile.objs)
	}
	if start > end {
		start = end
	}
	tile.objs = tile.objs[start:end]
	tile.coordTranslator = br.transCoordToXY

	if br.verbose {
		for i, o := range tile.objs {
			text := ""
			if reflect.TypeOf(o).String() == "*tiles.Label" {
				if len(o.(*Label).text) > 0 {
					text = o.(*Label).text
				} else if o.(*Label).icon != nil {
					text = "<icon>"
				} else {
					text = "-"
				}
			}
			br.log.Tracef("z-order %2d %s %s %s",
				start+i, reflect.TypeOf(o), o.SourceInfo(), text)
		}
	}

	// background
	tile.addFirst(&TileBackground{
		color:  Gray50,
		width:  float64(br.canvasWidth),
		height: float64(br.canvasHeight),
	})

	// watermark
	if len(br.watermark) > 0 || br.tint {
		tile.AddWatermark(br.watermark, br.tint)
	}

	return tile, nil
}

type roleItem struct {
	role       string
	points     []geom.LatLon
	sourceInfo string
}

func (r *roleItem) firstLatLon() (float64, float64) {
	if len(r.points) == 0 {
		panic(fmt.Sprintf("no points in a role '%s' %s", r.role, r.sourceInfo))
	}
	return r.points[0].Lat, r.points[0].Lon
}

func (r *roleItem) lastLatLon() (float64, float64) {
	l := len(r.points)
	if l >= 2 {
		return r.points[l-1].Lat, r.points[l-1].Lon
	}
	return 0, 0
}

func (r *roleItem) isClosed() bool {
	if len(r.points) > 2 {
		x1, y1 := r.firstLatLon()
		x2, y2 := r.lastLatLon()
		return x1 == x2 && y1 == y2
	}
	return false

}
func (r *roleItem) canConnectTo(other *roleItem) bool {
	lastLat, lastLon := r.lastLatLon()
	firstLat, fistLon := other.firstLatLon()
	return lastLat == firstLat && lastLon == fistLon
}

func (r *roleItem) hasOriginLatLon(lat, lon float64) bool {
	aLat, aLon := r.firstLatLon()
	return aLat == lat && aLon == lon
}

func (r *roleItem) hasOriginPoint(origin [2]float64) bool {
	aLat, aLon := r.firstLatLon()
	return aLat == origin[0] && aLon == origin[1]
}

func (r *roleItem) hasSameOrigin(other *roleItem) bool {
	aLat, aLon := r.firstLatLon()
	bLat, bLon := other.firstLatLon()
	return aLat == bLat && aLon == bLon
}

func (r *roleItem) hasSameTermination(other *roleItem) bool {
	aLat, aLon := r.lastLatLon()
	bLat, bLon := other.lastLatLon()
	return aLat == bLat && aLon == bLon
}

type roleItemGroup []*roleItem

// find item that has same origin
func (rg roleItemGroup) findOrigin() ([2]float64, bool) {
	var origin [2]float64
	for i, r1 := range rg {
		for n, r2 := range rg {
			if i == n {
				continue
			}
			if r1.hasSameOrigin(r2) {
				lat, lon := r1.firstLatLon()
				origin = [2]float64{lat, lon}
				return origin, true
			}
		}
	}
	return origin, false
}

func (rg roleItemGroup) linearizeCoords() [][]geom.LatLon {
	var rt = make([][]geom.LatLon, 0)

	noneClosed := make([]*roleItem, 0)
	for _, r1 := range rg {
		if r1.isClosed() {
			rt = append(rt, r1.points)
		} else {
			noneClosed = append(noneClosed, r1)
		}
	}
	rg = noneClosed
	if len(rg) == 0 {
		return rt
	}

	leaders := make([]*roleItem, 0)
	others := make([]*roleItem, 0)
	if origin, b := rg.findOrigin(); b {
		for _, r2 := range rg {
			if r2.hasOriginPoint(origin) {
				leaders = append(leaders, r2)
			} else {
				others = append(others, r2)
			}
		}
	} else {
		leaders = append(leaders, rg[0])
		others = append(others, rg[1:]...)
	}

	for _, leader := range leaders {
	repeat:
		remains := make([]*roleItem, 0)
		for _, other := range others {
			if leader.canConnectTo(other) {
				leader.points = append(leader.points, other.points...)
			} else {
				remains = append(remains, other)
			}
		}
		if len(others) != len(remains) {
			others = remains
			//// leader has been appended
			goto repeat
		}
	}

	for _, leader := range leaders {
		//fmt.Fprintf(os.Stderr, "leader: %f %f ~ %f %f\n", leader.points[0][0], leader.points[0][1], leader.points[len(leader.points)-1][0], leader.points[len(leader.points)-1][1])
		rt = append(rt, leader.points)
	}
	// for _, other := range others {
	// 	fmt.Fprintf(os.Stderr, "other : %f %f ~ %f %f\n", other.points[0][0], other.points[0][1], other.points[len(other.points)-1][0], other.points[len(other.points)-1][1])
	// }

	return rt
}

func (br *DefaultBuilder) compileRelation(rel *Relation) []Object {
	var objects []Object
	var sourceInfo = fmt.Sprintf("REL:%d", rel.Id)
	// if br.verbose {
	// 	br.log.Tracef("building... REL:%d", rel.Id)
	// }
	var style *Style = styleFromTags(&StyleParam{Tags: rel.Tags}, br.customStyler)

	var label *Label
	var name = rel.FindTag("name")
	if len(name) > 0 {
		sourceInfo += " " + name
		clat, clon := rel.MinLat+(rel.MaxLat-rel.MinLat)/2, rel.MinLon+(rel.MaxLon-rel.MinLon)/2
		label = &Label{
			text:       name,
			textColor:  style.MarkerColor,
			coord:      geom.LatLon{Lat: clat, Lon: clon},
			rotate:     0,
			sourceInfo: sourceInfo,
			visibleFunc: func(z int) bool {
				return !br.hideLabels && style.MarkerVisible(z)
			},
		}
		objects = append(objects, label)
	}

	// 한강: ./tmp/osmd render -i tcp://127.0.0.1:1918 -o ./tmp/render_out.png -v REL 152336
	var roleItems = make([]*roleItem, 0)
	for _, m := range rel.Members {
		if m.Type != Relation_WAY {
			// TODO: it can be NODE, RELATION
			continue
		}
		way, ok := br.ways.Get(m.Id)
		if way == nil || !ok {
			//br.log.Errorf("REL[%d] not found member WAY:%d", rel.Id, m.Id)
			continue
		}

		var points = make([]geom.LatLon, 0)
		for _, n := range way.Nodes {
			points = append(points, geom.LatLon{Lat: n.Lat, Lon: n.Lon})
		}

		ritem := &roleItem{
			role:       m.Role,
			points:     points,
			sourceInfo: fmt.Sprintf("WAY:%d", way.Id),
		}
		roleItems = append(roleItems, ritem)
	}

	var outerItems = make([]*roleItem, 0)
	var innerItems = make([]*roleItem, 0)
	for _, itm := range roleItems {
		if itm.role == "outer" {
			outerItems = append(outerItems, itm)
		} else if itm.role == "inner" {
			innerItems = append(innerItems, itm)
		} else {
			obj := br.buildPolygonLineString(itm.points, style, itm.sourceInfo)
			objects = append(objects, obj)
		}
	}

	inners := roleItemGroup(innerItems).linearizeCoords()
	outers := roleItemGroup(outerItems).linearizeCoords()

	if len(outers) > 0 {
		maskedObj := &MultiPolygonObject{
			inners:      inners,
			outers:      outers,
			lineWidth:   style.LineWidth,
			lineColor:   style.LineColor,
			lineDash:    style.LineDash,
			fillColor:   style.FillColor,
			layer:       style.BaseLayer,
			sourceInfo:  sourceInfo,
			visibleFunc: func(z int) bool { return true },
		}
		if maskedObj != nil {
			objects = append(objects, maskedObj)
		}
	}

	return objects
}

func (br *DefaultBuilder) compileNode(node *Node) []Object {
	return []Object{}
}

func (br *DefaultBuilder) compileWay(way *Way) []Object {
	var objects []Object
	sourceInfo := fmt.Sprintf("WAY:%d", way.Id)

	closed := false
	if len(way.Nodes) >= 3 {
		firstNode := way.Nodes[0]
		lastNode := way.Nodes[len(way.Nodes)-1]
		closed = firstNode.Id == lastNode.Id || (firstNode.Lat == lastNode.Lat && firstNode.Lon == lastNode.Lon)
	}
	var style *Style = styleFromTags(&StyleParam{Tags: way.Tags, Closed: closed}, br.customStyler)

	polygon := br.buildPolygon(way, style, sourceInfo)
	objects = append(objects, polygon)

	labelText := way.FindTag("name")
	if len(labelText) > 0 {
		sourceInfo += labelText
	}

	if len(labelText) > 0 {
		var labelRotate = 0.0
		var latLon geom.LatLon
		if style.FillColor == nil {
			count := len(polygon.outer)
			latLon = polygon.outer[0]
			if count > 1 {
				lon1, lat1 := polygon.outer[0].Lon, polygon.outer[0].Lat
				lon2, lat2 := polygon.outer[1].Lon, polygon.outer[1].Lat
				labelRotate = math.Atan2(lat2-lat1, lon2-lon1)
				latLon.Lon, latLon.Lat = lon1+(lon2-lon1)/2, lat1+(lat2-lat1)/2
			}
		} else {
			latLon.Lon = way.MinLon + (way.MaxLon-way.MinLon)/2
			latLon.Lat = way.MinLat + (way.MaxLat-way.MinLat)/2
		}
		label := &Label{
			text:       labelText,
			textColor:  style.MarkerColor,
			coord:      latLon,
			rotate:     labelRotate,
			icon:       style.Marker,
			sourceInfo: sourceInfo,
			visibleFunc: func(z int) bool {
				return style.MarkerVisible(br.zoom) && !br.hideLabels
			},
		}
		objects = append(objects, label)
	}

	for _, n := range way.Nodes {
		br.nodes.Delete(n.Id)
	}
	return objects
}

func (br *DefaultBuilder) buildPolygon(way *Way, style *Style, sourceInfo string) *PolygonObject {
	var coords = make([]geom.LatLon, 0)
	for _, n := range way.Nodes {
		latLon := geom.LatLon{Lat: n.Lat, Lon: n.Lon}
		coords = append(coords, latLon)
	}

	return br.buildPolygonLineString(coords, style, sourceInfo)
}

func (br *DefaultBuilder) buildPolygonLineString(coords []geom.LatLon, style *Style, sourceInfo string) *PolygonObject {
	obj := &PolygonObject{
		outer:      coords,
		lineWidth:  style.LineWidth,
		lineColor:  style.LineColor,
		lineDash:   style.LineDash,
		fillColor:  style.FillColor,
		layer:      style.BaseLayer,
		sourceInfo: sourceInfo,
		visibleFunc: func(z int) bool {
			return true
		},
	}
	return obj
}

func (br *DefaultBuilder) dumpWay(way *Way) {
	tagMap := way.Tags
	name := tagMap["name"]
	br.log.Debugf("Way[%d] %s (nodes:%d)", way.Id, name, len(way.Nodes))
	for k, v := range tagMap {
		br.log.Debugf("    %s = %s", k, v)
	}
}

func (br *DefaultBuilder) dumpRelation(rel *Relation, tags bool) {
	name := rel.FindTag("name")
	typ := rel.FindTag("type")
	subType := rel.FindTag(typ)
	landuse := rel.FindTag("landuse")
	br.log.Tracef("Relation[%d] name:%s members:%d type:%s %s:%s landuse:%s",
		rel.Id, name, len(rel.Members), typ, typ, subType, landuse)
	if tags {
		for k, v := range rel.Tags {
			br.log.Tracef("    %s = %s", k, v)
		}
	}
}
