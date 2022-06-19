package main

import (
	"strings"
	"time"

	"github.com/OutOfBedlam/ots/geom"
	"github.com/OutOfBedlam/ots/logging"
	"github.com/OutOfBedlam/ots/tiles"
)

type DataSource interface {
	Close()

	IntersectsBounds(bound geom.Bound) (*ResultSet, error)

	GetWay(id int64) (*tiles.Way, bool)
	GetNode(id int64) (*tiles.Node, bool)
	GetRelation(id int64) (*tiles.Relation, bool)

	SearchNodes(tag string, keyword string) []*tiles.Node
	SearchWays(tag string, keyword string) []*tiles.Way
	SearchRelations(tag string, keyword string) []*tiles.Relation
}

func NewDataSource(dsaddr string, buffSize int) (DataSource, error) {
	var ds DataSource

	log := logging.GetLog("datasource")

	if strings.HasPrefix(dsaddr, "tcp://") {
		// data source is a remote grpc server
		log.Infof("osm data source: %s", dsaddr)
		rds := &remoteOsmd{
			addr: dsaddr,
		}
		if buffSize > 0 {
			rds.grpcMaxRecvMsgSize = buffSize
		}
		ds = rds
	} else {
		// data source is local file
		log.Infof("reading osm data from %s ...", dsaddr)
		startLoad := time.Now()
		data, err := loadOsmData(dsaddr)
		if err != nil {
			return nil, err
		}
		log.Infof("loaded. %+v", time.Since(startLoad))
		ds = data
	}

	return ds, nil
}

type ResultSet struct {
	Nodes     []*tiles.Node
	Ways      []*tiles.Way
	Relations []*tiles.Relation
}

func (rs *ResultSet) LenObjs() int {
	return len(rs.Nodes) + len(rs.Ways) + len(rs.Relations)
}

func (rs *ResultSet) LenNodes() int {
	return len(rs.Nodes)
}

func (rs *ResultSet) LenWays() int {
	return len(rs.Ways)
}

func (rs *ResultSet) LenRelations() int {
	return len(rs.Relations)
}
