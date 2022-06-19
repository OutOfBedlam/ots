package main

import (
	"context"
	"fmt"
	"os"
	"sort"

	"github.com/paulmach/osm"
	"github.com/paulmach/osm/osmpbf"
	"github.com/tidwall/btree"
)

type tagCounter struct {
	name  string
	count int
}

type Counter struct {
	OsmDataSource string `arg:"" required:"" name:"osm data source" help:"osm data source, eg) ./data/my.osm.pbf or tcp://host:port"`

	changesetCount   int
	noteCount        int
	userCount        int
	boundsCount      int
	nodeCount        int
	wayCount         int
	relationCount    int
	nodeTagNames     *btree.Map[string, int]
	wayTagNames      *btree.Map[string, int]
	relationTagNames *btree.Map[string, int]
}

func (c *Counter) count(doPrint bool) {
	c.nodeTagNames = &btree.Map[string, int]{}
	c.wayTagNames = &btree.Map[string, int]{}
	c.relationTagNames = &btree.Map[string, int]{}

	f, err := os.Open(c.OsmDataSource)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	scanner := osmpbf.New(context.Background(), f, 3)
	defer scanner.Close()

	for scanner.Scan() {
		switch obj := scanner.Object().(type) {
		case *osm.Node:
			c.nodeCount++
			for _, tag := range obj.Tags {
				if v, b := c.nodeTagNames.Get(tag.Key); b {
					c.nodeTagNames.Set(tag.Key, v+1)
				} else {
					c.nodeTagNames.Set(tag.Key, 1)
				}
			}
		case *osm.Way:
			c.wayCount++
			for _, tag := range obj.Tags {
				if v, b := c.nodeTagNames.Get(tag.Key); b {
					c.wayTagNames.Set(tag.Key, v+1)
				} else {
					c.wayTagNames.Set(tag.Key, 1)
				}
			}
		case *osm.Relation:
			c.relationCount++
			for _, tag := range obj.Tags {
				if v, b := c.relationTagNames.Get(tag.Key); b {
					c.relationTagNames.Set(tag.Key, v+1)
				} else {
					c.relationTagNames.Set(tag.Key, 1)
				}
			}
		case *osm.Changeset:
			c.changesetCount++
		case *osm.Note:
			c.noteCount++
		case *osm.User:
			c.userCount++
		case *osm.Bounds:
			c.boundsCount++
		}
	}
	scanErr := scanner.Err()
	if scanErr != nil {
		panic(scanErr)
	}

	if doPrint {
		c.printResult()
	}
}

func printCounters(label string, count int, tagNames *btree.Map[string, int]) {
	tc := make([]tagCounter, 0)
	tagNames.Scan(func(key string, val int) bool {
		tc = append(tc, tagCounter{name: key, count: val})
		return true
	})

	sort.Slice(tc, func(i, j int) bool {
		return tc[i].count > tc[j].count
	})

	fmt.Printf("%s    %d\n", label, count)
	for _, t := range tc {
		if t.count < 10 {
			break
		}
		fmt.Printf("    %s    %d\n", t.name, t.count)
	}
}

func (c *Counter) printResult() {
	printCounters("Nodes", c.nodeCount, c.nodeTagNames)
	fmt.Printf("---------------------------\n")
	printCounters("Ways", c.wayCount, c.wayTagNames)
	fmt.Printf("---------------------------\n")
	printCounters("Relations", c.relationCount, c.relationTagNames)
	fmt.Printf("---------------------------\n")

	if c.changesetCount > 0 {
		fmt.Printf("Changesets %d\n", c.changesetCount)
	}
	if c.noteCount > 0 {
		fmt.Printf("Notes %d\n", c.noteCount)
	}
	if c.userCount > 0 {
		fmt.Printf("Users %d\n", c.userCount)
	}
	if c.boundsCount > 0 {
		fmt.Printf("Bounds %d\n", c.boundsCount)
	}

}
