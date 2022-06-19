package tiles_test

import (
	"fmt"
	"testing"

	"github.com/fogleman/gg"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font/gofont/goregular"
)

const outputDir = "../../tmp/"

func TestInvertMask(t *testing.T) {
	dc := gg.NewContext(1024, 1024)

	dc.Push()
	dc.DrawRectangle(0, 0, 1024, 1024)
	dc.SetRGB(1.0, 1.0, 1.0)
	dc.Fill()
	dc.Pop()

	dc.DrawCircle(512, 512, 384)
	dc.Clip()
	dc.InvertMask()

	dc.Push()
	dc.MoveTo(0, 0)
	dc.LineTo(1000, 1000)
	dc.SetRGB(0, 0, 0)
	dc.Stroke()
	dc.Pop()

	dc.SavePNG(outputDir + "out.png")
}

func TestRoatedText(t *testing.T) {
	const S = 400
	dc := gg.NewContext(S, S)
	dc.SetRGB(1, 1, 1)
	dc.Clear()
	dc.SetRGB(0, 0, 0)
	font, err := truetype.Parse(goregular.TTF)
	if err != nil {
		panic("")
	}
	face := truetype.NewFace(font, &truetype.Options{
		Size: 40,
	})
	dc.SetFontFace(face)
	text := "Hello, world!"
	w, h := dc.MeasureString(text)
	dc.Rotate(gg.Radians(10))
	dc.DrawRectangle(100, 180, w, h)
	dc.Stroke()
	dc.DrawStringAnchored(text, 100, 180, 0.0, 0.0)
	dc.SavePNG(outputDir + "out.png")
}

func TestCustomFontText(t *testing.T) {
	const S = 400
	dc := gg.NewContext(S, S)
	dc.SetRGB(1, 1, 1)
	dc.Clear()
	dc.SetRGB(0, 0, 0)
	face, err := gg.LoadFontFace("./fonts/fa-solid-900.ttf", 30)
	if err != nil {
		panic("")
	}
	dc.SetFontFace(face)
	text := "\uf0f3Alarm"
	w, h := dc.MeasureString(text)
	fmt.Printf("measurestring: %.f x %.f\n", w, h)
	dc.DrawRectangle(100, 180, w, h)
	dc.Stroke()
	dc.DrawStringAnchored(text, 100, 180, 0, 0)
	dc.SavePNG(outputDir + "out.png")
}

func TestFillArea(t *testing.T) {
	const S = 512
	dc := gg.NewContext(S, S)
	dc.SetRGB(1, 1, 1)
	dc.Clear()

	coords := [][]float64{
		{-190.000000, 114.000000},
		{-98.000000, 34.000000},
		{-36.000000, 138.000000},
		{90.000000, 448.000000},
		{-22.000000, 508.000000},
		{-36.000000, 464.000000},
		{-58.000000, 436.000000},
		{-88.000000, 420.000000},
		{-158.000000, 170.000000},
		{-166.000000, 146.000000},
	}
	path := func() {
		dc.ClearPath()
		for i, c := range coords {
			if i == 0 {
				dc.MoveTo(c[0], c[1])
			} else {
				dc.LineTo(c[0], c[1])
			}
		}
		dc.ClosePath()
	}

	path()
	dc.SetRGB(0, 0, 0)
	dc.StrokePreserve()

	path()
	dc.SetRGB(0.7, 0, 0)
	dc.FillPreserve()

	dc.SavePNG(outputDir + "out.png")
}
