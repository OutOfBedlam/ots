package main

import (
	"context"
	_ "embed"
	"fmt"
	"os"
	"time"

	"github.com/OutOfBedlam/ots/geom"
	"github.com/OutOfBedlam/ots/logging"
	"github.com/OutOfBedlam/ots/tiles"
)

type RenderCmd struct {
	OsmDataSource string   `arg:"" required:"" name:"osm data source" help:"osm data source, eg) ./data/my.osm.pbf or tcp://host:port"`
	Output        string   `arg:"" required:"" name:"output file name" help:"output file name, eg) ./out.png"`
	TypeAndIds    []string `arg:"" required:"" name:"TYPE_IDs" help:"comma seperated multiple Ids that combine type(WAY | REL | TILE) with colon"`
	LayerRange    string   `arg:"" optional:"" name:"LAYER_RANGE" default:"0:-" help:"rendering layer 'start:end'"`
	Width         int      `short:"W" default:"1024" help:"width of output image"`
	Height        int      `short:"H" default:"1024" help:"height of output image"`
	Verbose       bool     `short:"v" default:"false" help:"verbose"`
	Time          bool     `negatable:"" default:"false" help:"show elapse time"`
	ShowWatermark bool     `negatable:"" default:"false" help:"show watermark"`
	ShowLabels    bool     `negatable:"" default:"true" help:"show labels"`

	targetTiles []renderTileTarget
	targetIds   []renderIdTarget
	layerStart  int
	layerEnd    int
}

type renderIdTarget struct {
	typ tiles.GetRequest_Type
	id  int64
}

type renderTileTarget struct {
	z, x, y int
}

// RELATION 6114540   // 잠실주경기장
// RELATION 8824257   // 롯데월드타워
// WAY      129600971 // 영동일 고등학교
// WAY      636527742 // 롯데월드
func (opt *RenderCmd) render() {

	logging.SetDefaultLogging(&renderLogger{name: "render"})
	if opt.Verbose {
		logging.SetDefaultLevel(logging.LevelTrace)
	} else {
		logging.SetDefaultLevel(logging.LevelInfo)
	}
	logging.SetDefaultPrefixWidth(10)

	ds, err := NewDataSource(opt.OsmDataSource, 0)
	if err != nil {
		panic(err)
	}

	if len(opt.targetIds) > 0 {
		err = opt.renderIds(ds)
	} else if len(opt.targetTiles) > 0 {
		err = opt.renderTiles(ds)
	}

	if err != nil {
		panic(err)
	}
}

func (opt *RenderCmd) renderTiles(ds DataSource) error {
	z, x, y := opt.targetTiles[0].z, opt.targetTiles[0].x, opt.targetTiles[0].y

	log := logging.GetLog("render")
	log.Tracef("Target Coord: %d/%d/%d", z, x, y)

	//// search ways in the bounds
	t0 := time.Now()
	tileBounds := tiles.TilesToBounds(x, y, z).Pad(0.001)
	rset, err := ds.IntersectsBounds(tileBounds)
	if err != nil {
		return err
	}

	//// make builder
	t1 := time.Now()
	builder := tiles.NewBuilder(x, y, z)
	builder.SetVerbose(opt.Verbose)
	builder.SetBuildLayerRange(opt.layerStart, opt.layerEnd)
	builder.AddWays(rset.Ways...)
	builder.AddNodes(rset.Nodes...)
	builder.AddRelations(rset.Relations...)
	builder.SetHideLabels(!opt.ShowLabels)
	if opt.ShowWatermark {
		builder.SetWatermark(fmt.Sprintf("%d/%d/%d", z, x, y))
	}

	log.Infof("objects %d/%d/%d objs:%d ways:%d nodes:%d relations:%d",
		z, x, y, rset.LenObjs(), rset.LenWays(), rset.LenNodes(), rset.LenRelations())

	//// build tile
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	tile, err := builder.Build(ctx)
	if err != nil {
		return err
	}

	log.Infof("builder %d/%d/%d objs:%d bounds: %f,%f %f,%f",
		z, x, y, tile.CountObjects(), tileBounds.Min.Lat, tileBounds.Min.Lon, tileBounds.Max.Lat, tileBounds.Max.Lon)

	t2 := time.Now()
	file, err := os.Create(fmt.Sprintf("%s", opt.Output))
	defer file.Close()
	tile.EncodePNG(file)

	log.Infof("timing  %d/%d/%d objs:%d query:%s compile:%s render:%s\n",
		z, x, y, tile.CountObjects(), t1.Sub(t0), t2.Sub(t1), time.Since(t2))

	return nil
}

func (opt *RenderCmd) renderIds(ds DataSource) error {
	var builderBounds geom.Bound
	var builderObjs = make([]any, 0)

	log := logging.GetLog("render")
	log.Tracef("Target Id: %+v", opt.targetIds)

	tm0 := time.Now()

	for i, t := range opt.targetIds {
		switch t.typ {
		default:
			return fmt.Errorf("Unsupported type: '%s'", t.typ)
		case tiles.GetRequest_WAY:
			w, ok := ds.GetWay(t.id)
			if !ok {
				return fmt.Errorf("WAY[%d] not found", t.id)
			}
			if i == 0 {
				builderBounds = geom.Bound{
					Min: geom.LatLon{Lat: w.MinLat, Lon: w.MinLon},
					Max: geom.LatLon{Lat: w.MaxLat, Lon: w.MaxLon},
				}
			} else {
				builderBounds = builderBounds.Extend(geom.LatLon{Lat: w.MinLat, Lon: w.MinLon})
				builderBounds = builderBounds.Extend(geom.LatLon{Lat: w.MaxLat, Lon: w.MaxLon})
			}
			builderObjs = append(builderObjs, w)
		case tiles.GetRequest_RELATION:
			r, b := ds.GetRelation(t.id)
			if !b {
				return fmt.Errorf("RELATION[%d] not found", t.id)
			}
			if i == 0 {
				builderBounds = geom.Bound{
					Min: geom.LatLon{Lat: r.MinLat, Lon: r.MinLon},
					Max: geom.LatLon{Lat: r.MaxLat, Lon: r.MaxLon},
				}
			} else {
				builderBounds = builderBounds.Extend(geom.LatLon{Lat: r.MinLat, Lon: r.MinLon})
				builderBounds = builderBounds.Extend(geom.LatLon{Lat: r.MaxLat, Lon: r.MaxLon})
			}
			builderObjs = append(builderObjs, r)

			for _, m := range r.Members {
				if m.Type == tiles.Relation_WAY {
					if w, b := ds.GetWay(m.Id); b {
						builderObjs = append(builderObjs, w)
					}
				} else if m.Type == tiles.Relation_NODE {
					if n, b := ds.GetNode(m.Id); b {
						builderObjs = append(builderObjs, n)
					}
				} else if m.Type == tiles.Relation_RELATION {
					if r, b := ds.GetRelation(m.Id); b {
						builderObjs = append(builderObjs, r)
					}
				}
			}
		}
	}

	log.Tracef("Create Builder: %+v", builderBounds)
	tm1 := time.Now()
	builder := tiles.NewBuilderBounds(builderBounds, float64(opt.Width), float64(opt.Height))
	builder.SetVerbose(opt.Verbose)
	builder.SetBuildLayerRange(opt.layerStart, opt.layerEnd)
	builder.SetHideLabels(!opt.ShowLabels)
	for _, obj := range builderObjs {
		switch o := obj.(type) {
		case *tiles.Node:
			builder.AddNodes(o)
		case *tiles.Way:
			builder.AddWays(o)
		case *tiles.Relation:
			builder.AddRelations(o)
		}
	}

	tm2 := time.Now()
	tm3 := time.Now()
	if builder != nil {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()
		log.Tracef("Building....")
		tile, err := builder.Build(ctx)
		if err != nil {
			panic(err)
		}
		tm3 = time.Now()
		file, err := os.Create(fmt.Sprintf("%s", opt.Output))
		defer file.Close()

		log.Tracef("Save PNG....")
		tile.EncodePNG(file)
	}

	log.Infof("builder bounds: %+v", builderBounds)
	if opt.Time {
		log.Infof("timing query:%s builder:%s render:%s encode:%s\n",
			tm1.Sub(tm0), tm2.Sub(tm1), tm3.Sub(tm2), time.Since(tm3))

	}
	return nil
}

type renderLogger struct {
	name string
}

func (rl *renderLogger) Name() string {
	return rl.name
}

func (rl *renderLogger) Printf(format string, v ...interface{}) {
	fmt.Fprintf(os.Stdout, format, v...)
}
func (rl *renderLogger) Write(p []byte) (n int, err error) {
	return os.Stdout.Write(p)
}
