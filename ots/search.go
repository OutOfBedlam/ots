package main

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/OutOfBedlam/ots/tiles"
)

type Search struct {
	OsmDataSource string `arg:"" required:"" name:"osm data source" help:"osm data source, eg) ./data/my.osm.pbf or tcp://host:port"`
	Keyword       string `arg:"" required:"" name:"KEYWORD" help:"search keyword"`
	Scope         string `short:"s" default:"" help:"search scope w(ays), n(odes), r(elatations)"`
	Tag           string `short:"t" help:"search tag fields"`
	ShowCoords    bool   `default:"false" negatable:"" help:"show coordinates"`
	ShowTags      bool   `default:"true" negatable:"" help:"show tags"`
}

func (s *Search) search() {
	var ds DataSource

	if strings.HasPrefix(s.OsmDataSource, "tcp://") {
		ds = &remoteOsmd{
			addr: s.OsmDataSource,
		}
	} else {
		var err error
		ds, err = loadOsmData(s.OsmDataSource)
		if err != nil {
			panic(err)
		}
	}

	if s.Scope == "" && s.Tag == "" {
		K := strings.ToUpper(s.Keyword)
		if strings.HasPrefix(K, "REL:") {
			s.Scope = "r"
			s.Tag = "id"
			s.Keyword = s.Keyword[4:]
		} else if strings.HasPrefix(K, "WAY:") {
			s.Scope = "w"
			s.Tag = "id"
			s.Keyword = s.Keyword[4:]
		} else if strings.HasPrefix(K, "NODE:") {
			s.Scope = "n"
			s.Tag = "id"
			s.Keyword = s.Keyword[5:]
		} else {
			return
		}
	}

	var targetID int64 = 0
	if strings.ToUpper(s.Tag) == "ID" {
		var err error
		if targetID, err = strconv.ParseInt(s.Keyword, 10, 64); err != nil {
			panic(err)
		}
	}

	var printTags = func(tags map[string]string) {
		if tags == nil || len(tags) == 0 {
			return
		}
		keys := make([]string, 0, len(tags))
		for k := range tags {
			keys = append(keys, k)
		}
		sort.Sort(sort.StringSlice(keys))
		for _, k := range keys {
			v := tags[k]
			fmt.Printf("  %s=%s\n", k, v)
		}
	}

	if strings.Contains(s.Scope, "n") { //// node
		var rset []*tiles.Node
		if targetID > 0 {
			if n, b := ds.GetNode(targetID); b {
				rset = []*tiles.Node{n}
			}
		} else {
			rset = ds.SearchNodes(s.Tag, s.Keyword)
		}
		for _, node := range rset {
			tagValue := node.Tags[s.Tag]
			fmt.Printf("Node[%d] %s\n", node.Id, tagValue)

			if s.ShowCoords {
				fmt.Printf("     point: %f,%f\n", node.Lat, node.Lon)
			}

			if s.ShowTags {
				printTags(node.Tags)
			}
		}
	}

	if strings.Contains(s.Scope, "w") { //// way
		var rset []*tiles.Way
		if targetID > 0 {
			if w, b := ds.GetWay(targetID); b {
				rset = []*tiles.Way{w}
			}
		} else {
			rset = ds.SearchWays(s.Tag, s.Keyword)
		}
		for _, way := range rset {
			fmt.Printf("Way[%d] nodes:%d\n", way.Id, len(way.Nodes))

			if s.ShowCoords {
				fmt.Printf("     bounds: %f,%f %f,%f\n", way.MinLat, way.MinLon, way.MaxLat, way.MaxLon)
			}
			if s.ShowTags {
				printTags(way.Tags)
			}
		}
	}

	if strings.Contains(s.Scope, "r") { //// relation
		var rset []*tiles.Relation
		if targetID > 0 {
			if r, b := ds.GetRelation(targetID); b {
				rset = []*tiles.Relation{r}
			}
		} else {
			rset = ds.SearchRelations(s.Tag, s.Keyword)
		}
		for _, rel := range rset {
			fmt.Printf("Relation[%d] members:%d\n", rel.Id, len(rel.Members))
			if s.ShowCoords {
				fmt.Printf("     bounds: %f,%f %f,%f\n", rel.MinLat, rel.MinLon, rel.MaxLat, rel.MaxLon)
			}
			if s.ShowTags {
				printTags(rel.Tags)
			}
			for _, m := range rel.Members {
				fmt.Printf("     [%v] %d %s\n", m.Type, m.Id, m.Role)
			}
		}
	}
}
