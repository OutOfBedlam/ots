package main

import (
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"

	"github.com/OutOfBedlam/ots/banner"
	"github.com/OutOfBedlam/ots/tiles"
	"github.com/alecthomas/kong"
	konghcl "github.com/alecthomas/kong-hcl/v2"
)

func main() {
	var cli struct {
		Version    struct{}         `cmd:"" help:"show version"`
		TileServer TileServerConfig `cmd:"" name:"server" help:"osm tile server"`
		Render     RenderCmd        `cmd:"" help:"render specified object"`
		Search     Search           `cmd:"" help:"search osm elements"`
		Count      Counter          `cmd:"" help:"count osm data features"`
	}

	var cmd *kong.Context

	if os.Args[1] == "server" {
		tsc := TileServerConfig{}
		parser, err := kong.New(&tsc, kong.Configuration(konghcl.Loader))
		if err != nil {
			panic(err)
		}

		cmd, err = parser.Parse(os.Args[2:])
		if err != nil {
			panic(err)
		}
		tile_server(&tsc)
	} else {
		cmd = kong.Parse(&cli)
		switch cmd.Command() {
		case "version":
			fmt.Println(banner.Version())
		case "search <osm data source> <KEYWORD>":
			cli.Search.search()
		case "count <osm data source>":
			cli.Count.count(true)
		case "render <osm data source> <output file name> <TYPE_IDs>":
			_render(&cli.Render)
		default:
			panic(cmd.Command())
		}
	}
}

func _render(cmd *RenderCmd) {
	if len(cmd.LayerRange) > 0 {
		tok := strings.Split(cmd.LayerRange, ":")
		if len(tok) != 2 {
			fmt.Fprintf(os.Stderr, "invalid LAYER_RANGE\n")
			os.Exit(1)
		}
		start := 0
		end := math.MaxInt
		var err error
		if tok[0] != "-" {
			start, err = strconv.Atoi(tok[0])
			if err != nil {
				fmt.Fprintf(os.Stderr, "invalid LAYER_RANGE: %s\n", tok[0])
				os.Exit(1)
			}
		}

		if tok[1] != "-" {
			end, err = strconv.Atoi(tok[1])
			if err != nil {
				fmt.Fprintf(os.Stderr, "invalid LAYER_RANGE: %s\n", tok[1])
				os.Exit(1)
			}
		}
		cmd.layerStart = start
		cmd.layerEnd = end
	}

	for _, rawid := range cmd.TypeAndIds {
		id := strings.ToUpper(rawid)
		if strings.HasPrefix(id, "REL:") || strings.HasPrefix(id, "WAY:") {
			target := tiles.GetRequest_RELATION
			strid := id[4:]
			if strings.HasPrefix(id, "WAY:") {
				target = tiles.GetRequest_WAY
			}
			targetId, err := strconv.ParseInt(strid, 10, 64)
			if err != nil {
				fmt.Fprintf(os.Stderr, "invalid id: %s\n", id)
				os.Exit(1)
			}
			cmd.targetIds = append(cmd.targetIds,
				renderIdTarget{typ: target, id: targetId})
		} else if strings.HasPrefix(id, "TILE:") {
			strid := id[5:]
			toks := strings.Split(strid, "/")
			if len(toks) != 3 {
				fmt.Fprintf(os.Stderr, "tile id should be 'z/x/y'\n")
				os.Exit(1)
			}
			coords := make([]int, 3)
			for i, t := range toks {
				d, err := strconv.ParseInt(t, 10, 32)
				if err != nil {
					fmt.Fprintf(os.Stderr, "invalid ID: %s\n", t)
					os.Exit(1)
				}
				coords[i] = int(d)
			}
			cmd.targetTiles = append(cmd.targetTiles,
				renderTileTarget{z: coords[0], x: coords[1], y: coords[2]})
		} else {
			fmt.Fprintf(os.Stderr, "invalid TYPE: %s\n", id)
			os.Exit(1)
		}
	}

	cmd.render()
}
