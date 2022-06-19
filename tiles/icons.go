package tiles

import (
	_ "embed"

	"github.com/fogleman/gg"
	"github.com/golang/freetype/truetype"
)

//go:embed fonts/fa-solid-900.ttf
var faSold900Data []byte
var FontFaSolid900 *truetype.Font

func init() {
	var err error
	if FontFaSolid900, err = truetype.Parse(faSold900Data); err != nil {
		panic("invalid font")
	}
}

type Icon interface {
	Draw(dc *gg.Context, x, y, size float64)
	DrawAnchored(dc *gg.Context, x, y, size, ax, ay float64)
}

var (
	// find more font-awesome at https://fontawesome.com/v5/cheatsheet
	// https://www.w3schools.com/icons/fontawesome5_icons_maps.asp
	fa_swimmer       = newFaIcon('\uf5c4')
	fa_anchor        = newFaIcon('\uf13d')
	fa_parking       = newFaIcon('\uf540')
	fa_school        = newFaIcon('\uf549')
	fa_gasPump       = newFaIcon('\uf52f')
	fa_helicopter    = newFaIcon('\uf533')
	fa_hostpital_alt = newFaIcon('\uf47d')
	fa_child         = newFaIcon('\uf1ae')
	fa_book          = newFaIcon('\uf02d')
	fa_university    = newFaIcon('\uf19c')
	fa_fire          = newFaIcon('\uf06d')
	fa_user_shield   = newFaIcon('\uf505')
)

func newFaIcon(code rune) Icon {
	return &faIcon{
		code: code,
	}
}

type faIcon struct {
	code rune
}

func (icn *faIcon) Size() (float64, float64) {
	return 0, 0
}

func (icn *faIcon) Draw(dc *gg.Context, x, y, size float64) {
	icn.DrawAnchored(dc, x, y, size, 0.5, 0.5)
}

func (icn *faIcon) DrawAnchored(dc *gg.Context, x, y, size, ax, ay float64) {
	face := truetype.NewFace(FontFaSolid900, &truetype.Options{Size: size})
	defer face.Close()

	dc.Push()
	dc.SetFontFace(face)
	x = x - size*ax
	y = y + size*ay
	dc.DrawString(string(icn.code), x, y)
	dc.Pop()
}
