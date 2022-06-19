package main

import (
	"context"
	"strings"

	"github.com/OutOfBedlam/ots/geom"
	"github.com/OutOfBedlam/ots/tiles"
	"google.golang.org/grpc"
)

type remoteOsmd struct {
	DataSource
	addr               string
	grpcMaxRecvMsgSize int
	grpcConn           *grpc.ClientConn
}

func (r *remoteOsmd) Close() {
	if r.grpcConn != nil {
		r.grpcConn.Close()
		r.grpcConn = nil
	}
}

func (r *remoteOsmd) _getObjById(id int64, typ tiles.GetRequest_Type) (*tiles.GetResponse, error) {
	// connect to server
	if r.grpcConn == nil {
		if strings.HasPrefix(r.addr, "tcp://") {
			r.addr = r.addr[6:]
		}
		var err error
		r.grpcConn, err = grpc.Dial(r.addr, grpc.WithInsecure(), grpc.WithBlock())
		if err != nil {
			return nil, err
		}
	}

	client := tiles.NewTileClient(r.grpcConn)
	rsp, err := client.Get(context.Background(),
		&tiles.GetRequest{
			Type: typ,
			Id:   id,
		})
	return rsp, err
}

func (r *remoteOsmd) GetWay(id int64) (*tiles.Way, bool) {
	rsp, err := r._getObjById(id, tiles.GetRequest_WAY)
	if err == nil && rsp.Way != nil {
		return rsp.Way, true
	} else {
		return nil, false
	}
}

func (r *remoteOsmd) GetNode(id int64) (*tiles.Node, bool) {
	rsp, err := r._getObjById(id, tiles.GetRequest_NODE)
	if err == nil && rsp.Node != nil {
		return rsp.Node, true
	} else {
		return nil, false
	}
}

func (r *remoteOsmd) GetRelation(id int64) (*tiles.Relation, bool) {
	rsp, err := r._getObjById(id, tiles.GetRequest_RELATION)
	if err == nil && rsp.Relation != nil {
		return rsp.Relation, true
	} else {
		return nil, false
	}
}

func (r *remoteOsmd) _scanObj(tag string, keyword string, scope tiles.ScanRequest_Scope) (*tiles.ScanResponse, error) {
	// connect to server
	if strings.HasPrefix(r.addr, "tcp://") {
		r.addr = r.addr[6:]
	}
	callOpt := grpc.WithDefaultCallOptions(
		grpc.MaxCallRecvMsgSize(1024 * 1024 * 100),
	)
	conn, err := grpc.Dial(r.addr, callOpt, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	client := tiles.NewTileClient(conn)
	rsp, err := client.Scan(context.Background(),
		&tiles.ScanRequest{
			Scope:   scope,
			Tag:     tag,
			Keyword: keyword,
		})
	return rsp, err
}

func (r *remoteOsmd) SearchNodes(tag string, keyword string) []*tiles.Node {
	rsp, _ := r._scanObj(tag, keyword, tiles.ScanRequest_NODE)
	return rsp.Nodes
}
func (r *remoteOsmd) SearchWays(tag string, keyword string) []*tiles.Way {
	rsp, _ := r._scanObj(tag, keyword, tiles.ScanRequest_WAY)
	return rsp.Ways
}
func (r *remoteOsmd) SearchRelations(tag string, keyword string) []*tiles.Relation {
	rsp, _ := r._scanObj(tag, keyword, tiles.ScanRequest_RELATION)
	return rsp.Relations
}

func (r *remoteOsmd) IntersectsBounds(bounds geom.Bound) (*ResultSet, error) {
	// connect to server
	if strings.HasPrefix(r.addr, "tcp://") {
		r.addr = r.addr[6:]
	}
	callOpt := grpc.WithDefaultCallOptions(
		grpc.MaxCallRecvMsgSize(r.grpcMaxRecvMsgSize),
	)
	conn, err := grpc.Dial(r.addr, callOpt, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	client := tiles.NewTileClient(conn)
	rsp, err := client.Find(context.Background(),
		&tiles.FindRequest{
			MinLat: bounds.Min.Lat,
			MinLon: bounds.Min.Lon,
			MaxLat: bounds.Max.Lat,
			MaxLon: bounds.Max.Lon,
		})
	if err != nil {
		return nil, err
	}
	return &ResultSet{rsp.Nodes, rsp.Ways, rsp.Relations}, nil
}
