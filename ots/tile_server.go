package main

import (
	"bufio"
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/OutOfBedlam/ots/banner"
	"github.com/OutOfBedlam/ots/geom"
	"github.com/OutOfBedlam/ots/httpsvr"
	"github.com/OutOfBedlam/ots/logging"
	"github.com/OutOfBedlam/ots/tiles"
	"github.com/alecthomas/kong"
	"github.com/gin-gonic/gin"
	lru "github.com/hashicorp/golang-lru"
	"github.com/pkg/errors"
	"github.com/soheilhy/cmux"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type tileServer struct {
	tiles.TileServer

	log       logging.Log
	ds        DataSource
	quit      chan os.Signal
	options   *TileServerOptions
	tileCache *lru.Cache
}

type TileServerConfig struct {
	Config        kong.ConfigFlag   `short:"c" type:"existingfile" placeholder:"<path>" help:"path to config file"`
	OsmDataSource string            `short:"i" placeholder:"<datasource>" help:"osm data source, eg) ./data/my.osm.pbf or tcp://host:port"`
	Bind          string            `short:"b" default:"127.0.0.1" help:"bind address"`
	Port          int               `short:"p" default:"1919" help:"bind port"`
	CacheSize     int               `default:"2000" name:"cache-size" help:"lru cache size for generated images"`
	Options       TileServerOptions `embed:"" prefix:""`
	//// Caution!! by inconsistency (bug?) b/w kong and kong-hcl, do not use "group" tag, it will not work
	HttpLogConfig   logging.Config `embed:"" name:"httplog" prefix:"httplog-"`
	ServerLogConfig logging.Config `embed:"" name:"log" prefix:"log-"`
}

type TileServerOptions struct {
	Pname              string `default:"tilesvr" name:"pname" help:"server instance name"`
	GrpcMaxRecvMsgSize int    `default:"10" help:"grpc max recv message size in MB"`
	GrpcMaxSendMsgSize int    `default:"10" help:"grpc max send message size in MB"`
	ShowWatermark      bool   `default:"false" negatable:"" help:"show watermark"`
	ShowLabels         bool   `default:"true" negatable:"" help:"show labels"`
	Debug              bool   `default:"false" help:"debug mode"`
	HttpConsoleColor   bool   `default:"false" help:"http colored console log"`
	HttpDebugMode      bool   `default:"false" help:"http debug mode"`
}

func tile_server(conf *TileServerConfig) {
	conf.ServerLogConfig.Name = "tile-server"
	conf.HttpLogConfig.Name = "http-log"

	logging.Configure(&conf.ServerLogConfig)
	log := logging.GetLog("tilesvr")

	// Banner
	log.Info(banner.GenBootBanner(conf.Options.Pname, banner.Version()))

	var err error

	var tileCache *lru.Cache
	if conf.CacheSize > 0 {
		if tileCache, err = lru.New(conf.CacheSize); err != nil {
			log.Errorf("fail to create cache")
			os.Exit(1)
		}
	}

	lsnrAddr := fmt.Sprintf("%s:%d", conf.Bind, conf.Port)

	lsnr, err := net.Listen("tcp", lsnrAddr)
	if err != nil {
		log.Errorf("fail to listen port %s", err)
		os.Exit(1)
	}

	ds, err := NewDataSource(conf.OsmDataSource, conf.Options.GrpcMaxRecvMsgSize*1024*1024)
	if err != nil {
		log.Errorf("datasource %s loading failed, %s", conf.OsmDataSource, err.Error())
		os.Exit(1)
	}
	defer ds.Close()

	svr := tileServer{
		log:       log,
		ds:        ds,
		quit:      make(chan os.Signal, 1),
		options:   &conf.Options,
		tileCache: tileCache,
	}

	// New Mux Server
	mux := cmux.New(lsnr)
	grpcL := mux.MatchWithWriters(
		cmux.HTTP2MatchHeaderFieldSendSettings("content-type", "application/grpc"),
	)
	httpL := mux.Match(cmux.HTTP2(), cmux.HTTP1())

	// Start GRPC Server
	grpcOpt := []grpc.ServerOption{
		grpc.MaxRecvMsgSize(conf.Options.GrpcMaxRecvMsgSize * 1024 * 1024),
		grpc.MaxSendMsgSize(conf.Options.GrpcMaxSendMsgSize * 1024 * 1024),
	}

	grpcS := grpc.NewServer(grpcOpt...)
	tiles.RegisterTileServer(grpcS, &svr)
	reflection.Register(grpcS)

	httpSvr := httpsvr.NewServer(&httpsvr.HttpServerConfig{
		DisableConsoleColor: !conf.Options.HttpConsoleColor,
		DebugMode:           conf.Options.HttpDebugMode,
		BindHostPort:        lsnrAddr,
		LoggingConfig:       &conf.HttpLogConfig,
	})

	httpSvr.GET("tiles/:Z/:X/:Y", svr.handleGetTile)
	httpSvr.GET("", svr.handleDemoPage)
	log.Infof("grpc on tcp://%s", lsnrAddr)

	httpSvr.Start(httpL)
	go grpcS.Serve(grpcL)
	go mux.Serve()

	signal.Notify(svr.quit, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	<-svr.quit

	mux.Close()
	grpcS.Stop()
	httpSvr.Stop()
}

//go:embed tile_server.html
var htmlData []byte

func (svr *tileServer) handleDemoPage(c *gin.Context) {
	c.Data(http.StatusOK, "text/html", htmlData)
}

func (svr *tileServer) handleGetTile(c *gin.Context) {
	z, x, y, err := _parseZXY(c)
	if err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}

	cacheKey := fmt.Sprintf("%d/%d/%d", z, x, y)
	if svr.tileCache != nil {
		if a, ok := svr.tileCache.Get(cacheKey); ok {
			pngBytes := a.([]byte)
			c.Data(http.StatusOK, "image/png", pngBytes)
			c.Writer.Flush()
			return
		}
	}

	//// search objects that intersect the bounds
	t1 := time.Now()
	tileBounds := tiles.TilesToBounds(x, y, z).Pad(0.001)
	rset, err := svr.ds.IntersectsBounds(tileBounds)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	resultSetCount := rset.LenObjs()

	//// make builder
	t2 := time.Now()
	builder := tiles.NewBuilder(x, y, z)
	builder.SetVerbose(svr.options.Debug)
	builder.SetHideLabels(!svr.options.ShowLabels)
	builder.AddWays(rset.Ways...)
	builder.AddNodes(rset.Nodes...)
	builder.AddRelations(rset.Relations...)

	if svr.options.ShowWatermark {
		builder.SetWatermark(fmt.Sprintf("%d/%d/%d", z, x, y))
		builder.SetTint(x%2 == y%2)
	}

	//// build tile
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	tile, err := builder.Build(ctx)
	cancel()
	if err != nil {
		svr.log.Errorf("Builder timeout error %d/%d/%d", z, x, y)
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	objsCount := tile.CountObjects()

	t3 := time.Now()
	var b bytes.Buffer
	var bw = bufio.NewWriter(&b)
	if err := tile.EncodePNG(bw); err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	bw.Flush()
	pngBytes := b.Bytes()

	if svr.tileCache != nil {
		svr.tileCache.Add(cacheKey, pngBytes)
	}

	c.Data(http.StatusOK, "image/png", pngBytes)
	c.Writer.Flush()
	svr.log.Infof("%s query:%s %d compile:%s %d render:%s",
		cacheKey, t2.Sub(t1), resultSetCount, t3.Sub(t2), objsCount, time.Since(t3))
}

func _parseZXY(c *gin.Context) (z, x, y int, err error) {
	z, err = strconv.Atoi(c.Param("Z"))
	if err != nil {
		err = errors.New("invalid Z")
		return
	}
	if z < 11 || z > 19 {
		err = errors.New("unsupported Z level")
		return
	}

	x, err = strconv.Atoi(c.Param("X"))
	if err != nil {
		err = errors.New("invalid X")
		return
	}
	stry := c.Param("Y")
	if !strings.HasSuffix(stry, ".png") {
		err = errors.New("unsupported file extension")
		return
	}
	y, err = strconv.Atoi(stry[:len(stry)-4]) // remove '.png' suffix
	if err != nil {
		err = errors.New("invalid Y")
		return
	}
	return
}

func (svr *tileServer) Find(ctx context.Context, req *tiles.FindRequest) (*tiles.FindResponse, error) {
	tick := time.Now()

	findBounds := geom.Bound{
		Min: geom.LatLon{Lat: req.MinLat, Lon: req.MinLon},
		Max: geom.LatLon{Lat: req.MaxLat, Lon: req.MaxLon},
	}.Pad(0.001)

	rset, err := svr.ds.IntersectsBounds(findBounds)
	rsp := &tiles.FindResponse{
		Nodes:     rset.Nodes,
		Ways:      rset.Ways,
		Relations: rset.Relations,
	}

	if err == nil {
		rsp.Code = 0
		rsp.Reason = "ok"
	} else {
		rsp.Code = 1
		rsp.Reason = err.Error()
	}
	rsp.Elapsed = time.Since(tick).String()

	return rsp, nil
}

func (svr *tileServer) Get(ctx context.Context, req *tiles.GetRequest) (*tiles.GetResponse, error) {
	tick := time.Now()
	rsp := &tiles.GetResponse{}
	switch req.Type {
	case tiles.GetRequest_UNKNOWN:
	case tiles.GetRequest_NODE:
		if r, b := svr.ds.GetNode(req.Id); b {
			rsp.Node = r
		}
	case tiles.GetRequest_WAY:
		if w, b := svr.ds.GetWay(req.Id); b {
			rsp.Way = w
		}
	case tiles.GetRequest_RELATION:
		if r, b := svr.ds.GetRelation(req.Id); b {
			rsp.Relation = r
		}
	}
	rsp.Elapsed = time.Since(tick).String()
	return rsp, nil
}

func (svr *tileServer) Scan(ctx context.Context, req *tiles.ScanRequest) (*tiles.ScanResponse, error) {
	tick := time.Now()
	rsp := &tiles.ScanResponse{}

	switch req.Scope {
	case tiles.ScanRequest_UNKNOWN:
	case tiles.ScanRequest_NODE:
		rsp.Nodes = svr.ds.SearchNodes(req.Tag, req.Keyword)
	case tiles.ScanRequest_WAY:
		rsp.Ways = svr.ds.SearchWays(req.Tag, req.Keyword)
	case tiles.ScanRequest_RELATION:
		rsp.Relations = svr.ds.SearchRelations(req.Tag, req.Keyword)
	}

	rsp.Elapsed = time.Since(tick).String()
	return rsp, nil
}
