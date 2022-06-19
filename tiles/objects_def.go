package tiles

import (
	"image/color"

	"github.com/OutOfBedlam/ots/geom"
	"github.com/fogleman/gg"
	"golang.org/x/image/font"
)

type TileBackground struct {
	width  float64
	height float64
	color  color.Color
}

func (t *TileBackground) Layer() Layer {
	return LayerBackground
}

func (t *TileBackground) SourceInfo() string {
	return "background"
}

func (label *TileBackground) DistanceFrom(from geom.LatLon) float64 {
	return 0
}

func (t *TileBackground) Draw(dc *gg.Context, transCoord CoordTransFunc) {
	dc.Push()
	dc.SetColor(t.color)
	dc.Clear()
	dc.Pop()
}

func (t *TileBackground) Visible(zoom int) bool {
	return true
}

type Watermark struct {
	text      string
	size      float64
	fontFace  font.Face
	textColor color.Color
	tintColor color.Color
}

func (wm *Watermark) Layer() Layer {
	return LayerWatermark
}

func (wm *Watermark) Visible(zoom int) bool {
	return true
}

func (wm *Watermark) SourceInfo() string {
	return "watermark"
}

func (label *Watermark) DistanceFrom(from geom.LatLon) float64 {
	return 0
}

func (wm *Watermark) Draw(dc *gg.Context, transCoord CoordTransFunc) {
	S := wm.size
	dc.Push()
	if wm.tintColor != nil {
		dc.SetColor(wm.tintColor)
		dc.DrawRectangle(0, 0, S, S)
		dc.Fill()
	}

	if len(wm.text) > 0 {
		dc.SetFontFace(wm.fontFace)
		dc.SetColor(wm.textColor)
		dc.DrawStringAnchored(wm.text, S/2, S/4, 0.5, 0.5)
	}
	dc.Pop()
}
