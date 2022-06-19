package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/OutOfBedlam/ots/geom"
	"github.com/OutOfBedlam/ots/logging"
	"github.com/OutOfBedlam/ots/tiles"
	"github.com/paulmach/osm"
	"github.com/paulmach/osm/osmpbf"
	"github.com/tidwall/btree"
	"github.com/tidwall/rtree"
)

type osmdata struct {
	DataSource
	log           logging.Log
	relations     *btree.Map[osm.RelationID, *osm.Relation]
	ways          *btree.Map[osm.WayID, *osm.Way]
	nodes         *btree.Map[osm.NodeID, *osm.Node]
	relationIndex *rtree.Generic[*osm.Relation]
	wayIndex      *rtree.Generic[*osm.Way]
	nodeIndex     *rtree.Generic[*osm.Node]
}

func (data *osmdata) Close() {
}

func (data *osmdata) insertWay(way *osm.Way) {
	b := way.Bounds
	data.wayIndex.Insert([2]float64{b.MinLon, b.MinLat}, [2]float64{b.MaxLon, b.MaxLat}, way)
}

func (data *osmdata) insertRelation(rel *osm.Relation) {
	b := rel.Bounds
	data.relationIndex.Insert([2]float64{b.MinLon, b.MinLat}, [2]float64{b.MaxLon, b.MaxLat}, rel)
}

func (data *osmdata) insertNode(node *osm.Node) {
	data.nodeIndex.Insert([2]float64{node.Lon, node.Lat}, [2]float64{node.Lon, node.Lat}, node)
}

func (data *osmdata) searchRelation(bound *osm.Bounds, cb func(b *osm.Bounds, value *osm.Relation) bool) {
	data.relationIndex.Search(
		[2]float64{bound.MinLon, bound.MinLat},
		[2]float64{bound.MaxLon, bound.MaxLat},
		func(min, max [2]float64, value *osm.Relation) bool {
			return cb(&osm.Bounds{MinLon: min[0], MinLat: min[1], MaxLon: max[0], MaxLat: max[1]}, value)
		})
}

func (data *osmdata) searchWay(bound *osm.Bounds, cb func(b *osm.Bounds, value *osm.Way) bool) {
	data.wayIndex.Search(
		[2]float64{bound.MinLon, bound.MinLat},
		[2]float64{bound.MaxLon, bound.MaxLat},
		func(min, max [2]float64, value *osm.Way) bool {
			return cb(&osm.Bounds{MinLon: min[0], MinLat: min[1], MaxLon: max[0], MaxLat: max[1]}, value)
		})
}

func (data *osmdata) searchNode(bound *osm.Bounds, cb func(lat, lon float64, value *osm.Node) bool) {
	data.nodeIndex.Search(
		[2]float64{bound.MinLon, bound.MinLat},
		[2]float64{bound.MaxLon, bound.MaxLat},
		func(min, max [2]float64, value *osm.Node) bool {
			return cb(min[1], min[0], value)
		})
}

func loadOsmData(osmPbfPath string) (*osmdata, error) {
	f, err := os.Open(osmPbfPath)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	// scanner is not thread-safe
	scanner := osmpbf.New(context.Background(), f, 3)
	defer scanner.Close()

	data := &osmdata{
		log:           logging.GetLog("osm-data"),
		relationIndex: &rtree.Generic[*osm.Relation]{},
		wayIndex:      &rtree.Generic[*osm.Way]{},
		nodeIndex:     &rtree.Generic[*osm.Node]{},
		relations:     &btree.Map[osm.RelationID, *osm.Relation]{},
		ways:          &btree.Map[osm.WayID, *osm.Way]{},
		nodes:         &btree.Map[osm.NodeID, *osm.Node]{},
	}

	tick := time.Now()
	for scanner.Scan() {
		switch obj := scanner.Object().(type) {
		case *osm.Node:
			data.nodes.Set(obj.ID, obj)
		case *osm.Way:
			data.ways.Set(obj.ID, obj)
		case *osm.Relation:
			data.relations.Set(obj.ID, obj)
		}
	}
	scanErr := scanner.Err()
	if scanErr != nil {
		return nil, scanErr
	}
	data.log.Debugf("loading osm data time elapse: %s", time.Since(tick))

	var extend = func(b *osm.Bounds, node *osm.Node) {
		if b.ContainsNode(node) {
			return
		}
		if node.Lat < b.MinLat {
			b.MinLat = node.Lat
		}
		if node.Lat > b.MaxLat {
			b.MaxLat = node.Lat
		}
		if node.Lon < b.MinLon {
			b.MinLon = node.Lon
		}
		if node.Lon > b.MaxLon {
			b.MaxLon = node.Lon
		}
	}

	for _, node := range data.nodes.Values() {
		data.insertNode(node)
	}

	tick = time.Now()
	closeWay := 0
	openWay := 0
	for _, way := range data.ways.Values() {
		for i, n := range way.Nodes {
			node, b := data.nodes.Get(n.ID)
			if !b {
				continue
			}
			way.Nodes[i].Version = node.Version
			way.Nodes[i].Lat = node.Lat
			way.Nodes[i].Lon = node.Lon
			if way.Bounds == nil {
				way.Bounds = &osm.Bounds{
					MinLat: node.Lat, MinLon: node.Lon,
					MaxLat: node.Lat, MaxLon: node.Lon,
				}
			} else {
				extend(way.Bounds, node)
			}
		}
		if len(way.Nodes) > 0 {
			if way.Nodes[0].Lat == way.Nodes[len(way.Nodes)-1].Lat &&
				way.Nodes[0].Lon == way.Nodes[len(way.Nodes)-1].Lon {
				closeWay++
			} else {
				openWay++
			}
		} else {
			openWay++
		}
		if way.Bounds != nil {
			data.insertWay(way)
		}
	}
	data.log.Debugf("loading ways time elapse: %s (close:%d open:%d)", time.Since(tick), closeWay, openWay)

	tick = time.Now()
	for _, relation := range data.relations.Values() {
		for m := range relation.Members {
			if relation.Members[m].Type == osm.TypeNode {
				node, b := data.nodes.Get(osm.NodeID(relation.Members[m].Ref))
				if !b {
					continue
				}
				relation.Members[m].Lat = node.Lat
				relation.Members[m].Lon = node.Lon
				relation.Members[m].Version = node.Version
				if relation.Bounds == nil {
					relation.Bounds = &osm.Bounds{
						MinLat: node.Lat, MinLon: node.Lon,
						MaxLat: node.Lat, MaxLon: node.Lon,
					}
				} else {
					extend(relation.Bounds, node)
				}
			} else if relation.Members[m].Type == osm.TypeWay {
				way, b := data.ways.Get(osm.WayID(relation.Members[m].Ref))
				if !b {
					continue
				}
				relation.Members[m].Version = way.Version
				relation.Members[m].Nodes = make([]osm.WayNode, len(way.Nodes))
				for i, node := range way.Nodes {
					relation.Members[m].Nodes[i] = node
				}

				strPoints := make([]string, 0)
				for i, n := range relation.Members[m].Nodes {
					node, b := data.nodes.Get(n.ID)
					if !b {
						continue
					}

					strPoints = append(strPoints, fmt.Sprintf("%f %f", node.Lon, node.Lat))

					relation.Members[m].Nodes[i].Lat = node.Lat
					relation.Members[m].Nodes[i].Lon = node.Lon

					if relation.Bounds == nil {
						relation.Bounds = &osm.Bounds{
							MinLat: node.Lat, MinLon: node.Lon,
							MaxLat: node.Lat, MaxLon: node.Lon,
						}
					} else {
						extend(relation.Bounds, node)
					}
				}
			} else if relation.Members[m].Type == osm.TypeRelation {
				continue
			} else {
				continue
			}
		}
		if relation.Bounds != nil {
			data.insertRelation(relation)
		}
	}
	data.log.Debugf("loading relations time elapse: %s", time.Since(tick))

	return data, nil
}

func (data *osmdata) GetWay(id int64) (*tiles.Way, bool) {
	way, b := data.ways.Get(osm.WayID(id))
	if !b {
		return nil, false
	}
	wayMin := geom.LatLon{Lat: way.Bounds.MinLat, Lon: way.Bounds.MinLon}
	wayMax := geom.LatLon{Lat: way.Bounds.MaxLat, Lon: way.Bounds.MaxLon}
	wayBounds := geom.Bound{Min: wayMin, Max: wayMax}

	w := &tiles.Way{
		Id:    int64(way.ID),
		Tags:  way.TagMap(),
		Nodes: make([]*tiles.Way_NodeRef, 0),
	}
	w.MinLat = wayBounds.Min.Lat
	w.MinLon = wayBounds.Min.Lon
	w.MaxLat = wayBounds.Max.Lat
	w.MaxLon = wayBounds.Max.Lon
	if len(way.Nodes) > 0 && w.Nodes == nil {
		w.Nodes = make([]*tiles.Way_NodeRef, 0)
	}

	for _, n := range way.Nodes {
		w.Nodes = append(w.Nodes, &tiles.Way_NodeRef{
			Id:  int64(n.ID),
			Lat: n.Lat,
			Lon: n.Lon,
		})
	}
	return w, true
}

func (data *osmdata) GetNode(id int64) (*tiles.Node, bool) {
	n, b := data.nodes.Get(osm.NodeID(id))
	if !b {
		return nil, false
	}
	return &tiles.Node{
		Id:   int64(n.ID),
		Tags: n.TagMap(),
		Lat:  n.Lat,
		Lon:  n.Lon,
	}, true
}

func (data *osmdata) GetRelation(id int64) (*tiles.Relation, bool) {
	rel, b := data.relations.Get(osm.RelationID(id))
	if !b {
		return nil, false
	}

	r := &tiles.Relation{
		Id:      int64(rel.ID),
		Tags:    rel.TagMap(),
		MinLat:  rel.Bounds.MinLat,
		MinLon:  rel.Bounds.MinLon,
		MaxLat:  rel.Bounds.MaxLat,
		MaxLon:  rel.Bounds.MaxLon,
		Members: make([]*tiles.Relation_Member, len(rel.Members)),
	}

	for i, m := range rel.Members {
		r.Members[i] = &tiles.Relation_Member{
			Id:   int64(m.Ref),
			Type: tiles.RelationMemberType(m.Type),
			Role: m.Role,
		}
	}
	return r, true
}

func (data *osmdata) SearchNodes(tag string, keyword string) []*tiles.Node {
	rt := make([]*tiles.Node, 0)
	for _, node := range data.nodes.Values() {
		tagValue := node.Tags.Find(tag)
		if strings.Contains(tagValue, keyword) {
			if n, b := data.GetNode(int64(node.ID)); b {
				rt = append(rt, n)
			}
		}
	}
	return rt
}

func (data *osmdata) SearchWays(tag string, keyword string) []*tiles.Way {
	rt := make([]*tiles.Way, 0)
	for _, way := range data.ways.Values() {
		tagValue := way.Tags.Find(tag)
		if strings.Contains(tagValue, keyword) {
			if n, b := data.GetWay(int64(way.ID)); b {
				rt = append(rt, n)
			}
		}
	}
	return rt
}

func (data *osmdata) SearchRelations(tag string, keyword string) []*tiles.Relation {
	rt := make([]*tiles.Relation, 0)
	for _, rel := range data.relations.Values() {
		tagValue := rel.Tags.Find(tag)
		if strings.Contains(tagValue, keyword) {
			if n, b := data.GetRelation(int64(rel.ID)); b {
				rt = append(rt, n)
			}
		}
	}
	return rt
}

func _intersects(bounds geom.Bound, objBounds geom.Bound) bool {
	//return bounds.Intersects(objBounds) || objBounds.Intersects(bounds)
	return bounds.Intersects(objBounds)
}

func _osmWayToTileWay(way *osm.Way) *tiles.Way {
	wayMin := geom.LatLon{Lat: way.Bounds.MinLat, Lon: way.Bounds.MinLon}
	wayMax := geom.LatLon{Lat: way.Bounds.MaxLat, Lon: way.Bounds.MaxLon}
	wayBounds := geom.Bound{Min: wayMin, Max: wayMax}

	w := &tiles.Way{
		Id:    int64(way.ID),
		Tags:  way.TagMap(),
		Nodes: make([]*tiles.Way_NodeRef, 0),
	}
	w.MinLat = wayBounds.Min.Lat
	w.MinLon = wayBounds.Min.Lon
	w.MaxLat = wayBounds.Max.Lat
	w.MaxLon = wayBounds.Max.Lon

	if len(way.Nodes) > 0 && w.Nodes == nil {
		w.Nodes = make([]*tiles.Way_NodeRef, 0)
	}

	for _, n := range way.Nodes {
		w.Nodes = append(w.Nodes, &tiles.Way_NodeRef{
			Id:  int64(n.ID),
			Lat: n.Lat,
			Lon: n.Lon,
		})
	}
	return w
}

func (data *osmdata) _searchNode(b *osm.Bounds) *btree.Map[int64, *osm.Node] {
	rawNodes := btree.Map[int64, *osm.Node]{}
	data.searchNode(b, func(lat, lon float64, node *osm.Node) bool {
		rawNodes.Set(int64(node.ID), node)
		return true
	})
	return &rawNodes
}

func (data *osmdata) _searchWay(b *osm.Bounds, rset *ResultSet, rawNodes *btree.Map[int64, *osm.Node]) {
	data.searchWay(b, func(b *osm.Bounds, way *osm.Way) bool {
		w := _osmWayToTileWay(way)
		for _, n := range w.Nodes {
			// 반환할 node list에서 해당 node를 (way에 포함되었으므로) 제외시킨다.
			rawNodes.Delete(n.Id)
		}
		rset.Ways = append(rset.Ways, w)
		return true
	})
}

func (data *osmdata) _searchRelation(b *osm.Bounds, rset *ResultSet, rawNodes *btree.Map[int64, *osm.Node]) {
	data.searchRelation(b, func(b *osm.Bounds, obj *osm.Relation) bool {
		r := &tiles.Relation{
			Id:      int64(obj.ID),
			Tags:    obj.TagMap(),
			MinLat:  obj.Bounds.MinLat,
			MinLon:  obj.Bounds.MinLon,
			MaxLat:  obj.Bounds.MaxLat,
			MaxLon:  obj.Bounds.MaxLon,
			Members: make([]*tiles.Relation_Member, len(obj.Members)),
		}

		for i, m := range obj.Members {
			r.Members[i] = &tiles.Relation_Member{
				Id:   int64(m.Ref),
				Type: tiles.RelationMemberType(m.Type),
				Role: m.Role,
			}
			switch r.Members[i].Type {
			case tiles.Relation_NODE:
				// 반환할 node list에서 해당 node를 제외시킨다.
				rawNodes.Delete(int64(m.Ref))
			case tiles.Relation_WAY:
				contains := false
				for _, w := range rset.Ways {
					if w.Id == m.Ref {
						contains = true
						break
					}
				}
				if !contains {
					if way, b := data.ways.Get(osm.WayID(m.Ref)); b {
						w := _osmWayToTileWay(way)
						rset.Ways = append(rset.Ways, w)
					} else {
						//data.log.Tracef("REL:%d missing [%d] WAY: %d\n", obj.ID, i, m.Ref)
					}
				}
			}
		}
		rset.Relations = append(rset.Relations, r)
		return true
	})
}

func (data *osmdata) IntersectsBounds(bounds geom.Bound) (rset *ResultSet, err error) {
	rset = &ResultSet{
		Nodes:     make([]*tiles.Node, 0),
		Ways:      make([]*tiles.Way, 0),
		Relations: make([]*tiles.Relation, 0),
	}

	if data.log != nil && data.log.DebugEnabled() {
		t1 := time.Now()
		defer func() {
			data.log.Debugf("bound:%v rels:%d ways:%d nodes:%d %s",
				bounds, rset.LenRelations(), rset.LenWays(), rset.LenNodes(), time.Since(t1))
		}()
	}

	searchBound := &osm.Bounds{
		MinLat: bounds.Min.Lat, MinLon: bounds.Min.Lon,
		MaxLat: bounds.Max.Lat, MaxLon: bounds.Max.Lon}

	rawNodes := data._searchNode(searchBound)
	data._searchWay(searchBound, rset, rawNodes)
	//// TODO: fix the performance issue in search relations, it takes 99% of time
	data._searchRelation(searchBound, rset, rawNodes)

	for _, v := range rawNodes.Values() {
		n := &tiles.Node{
			Id:   int64(v.ID),
			Tags: v.TagMap(),
			Lat:  v.Lat,
			Lon:  v.Lon,
		}
		rset.Nodes = append(rset.Nodes, n)
	}

	return
}
