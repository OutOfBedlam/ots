package tiles

import (
	"image/color"
	"strconv"
	"strings"
)

func rgb(r, g, b uint8) color.Color {
	return color.RGBA{R: r, G: g, B: b, A: 0xFF}
}

func RgbHex(str string) color.Color {
	if !strings.HasPrefix(str, "#") {
		return nil
	}

	v, err := strconv.ParseInt(str[1:3], 16, 32)
	if err != nil {
		return nil
	}
	r := uint8(v)

	v, err = strconv.ParseInt(str[3:5], 16, 32)
	if err != nil {
		return nil
	}
	g := uint8(v)

	v, err = strconv.ParseInt(str[5:7], 16, 32)
	if err != nil {
		return nil
	}
	b := uint8(v)

	return color.RGBA{R: r, G: g, B: b, A: 0xFF}
}

/*
func RgbHex(str string) color.Color {
	if !strings.HasPrefix(str, "#") {
		return nil
	}

	v, err := strconv.ParseInt(str[1:3], 16, 16)
	if err != nil {
		return nil
	}
	r := uint8(v)

	v, err = strconv.ParseInt(str[3:5], 16, 16)
	if err != nil {
		return nil
	}
	g := uint8(v)

	v, err = strconv.ParseInt(str[5:7], 16, 16)
	if err != nil {
		return nil
	}
	b := uint8(v)

	return &HexColor{R: r, G: g, B: b, A: 0xFF}
}

type HexColor struct {
	R, G, B, A uint8
}

func (c *HexColor) RGBA() (r, g, b, a uint32) {
	return uint32(c.R), uint32(c.G), uint32(c.B), uint32(c.A)
}
*/

// color hex codes from https://github.com/gaithoben/material-ui-colors/tree/master/src
// color table: https://materialui.co/colors/
var (
	Clear = color.RGBA{R: 0xFF, G: 0xFF, B: 0xFF, A: 0xFF}

	// Red
	Red     = RgbHex("#f44336")
	Red50   = RgbHex("#ffebee")
	Red100  = RgbHex("#ffcdd2")
	Red200  = RgbHex("#ef9a9a")
	Red300  = RgbHex("#e57373")
	Red400  = RgbHex("#ef5350")
	Red500  = RgbHex("#f44336")
	Red600  = RgbHex("#e53935")
	Red700  = RgbHex("#d32f2f")
	Red800  = RgbHex("#c62828")
	Red900  = RgbHex("#b71c1c")
	RedA100 = RgbHex("#ff8a80")
	RedA200 = RgbHex("#ff5252")
	RedA400 = RgbHex("#ff1744")
	RedA700 = RgbHex("#d50000")

	// Pink
	Pink     = RgbHex("#e91e63")
	Pink50   = RgbHex("#fce4ec")
	Pink100  = RgbHex("#f8bbd0")
	Pink200  = RgbHex("#f48fb1")
	Pink300  = RgbHex("#f06292")
	Pink400  = RgbHex("#ec407a")
	Pink500  = RgbHex("#e91e63")
	Pink600  = RgbHex("#d81b60")
	Pink700  = RgbHex("#c2185b")
	Pink800  = RgbHex("#ad1457")
	Pink900  = RgbHex("#880e4f")
	PinkA100 = RgbHex("#ff80ab")
	PinkA200 = RgbHex("#ff4081")
	PinkA400 = RgbHex("#f50057")
	PinkA700 = RgbHex("#c51162")

	// Purple
	Purple     = RgbHex("#9c27b0")
	Purple50   = RgbHex("#f3e5f5")
	Purple100  = RgbHex("#e1bee7")
	Purple200  = RgbHex("#ce93d8")
	Purple300  = RgbHex("#ba68c8")
	Purple400  = RgbHex("#ab47bc")
	Purple500  = RgbHex("#9c27b0")
	Purple600  = RgbHex("#8e24aa")
	Purple700  = RgbHex("#7b1fa2")
	Purple800  = RgbHex("#6a1b9a")
	Purple900  = RgbHex("#4a148c")
	PurpleA100 = RgbHex("#ea80fc")
	PurpleA200 = RgbHex("#e040fb")
	PurpleA400 = RgbHex("#d500f9")
	PurpleA700 = RgbHex("#aa00ff")

	// DeepPurple
	DeepPurple     = RgbHex("#673ab7")
	DeepPurple50   = RgbHex("#ede7f6")
	DeepPurple100  = RgbHex("#d1c4e9")
	DeepPurple200  = RgbHex("#b39ddb")
	DeepPurple300  = RgbHex("#9575cd")
	DeepPurple400  = RgbHex("#7e57c2")
	DeepPurple500  = RgbHex("#673ab7")
	DeepPurple600  = RgbHex("#5e35b1")
	DeepPurple700  = RgbHex("#512da8")
	DeepPurple800  = RgbHex("#4527a0")
	DeepPurple900  = RgbHex("#311b92")
	DeepPurpleA100 = RgbHex("#b388ff")
	DeepPurpleA200 = RgbHex("#7c4dff")
	DeepPurpleA400 = RgbHex("#651fff")
	DeepPurpleA700 = RgbHex("#6200ea")

	// Indigo
	Indigo     = RgbHex("#3f51b5")
	Indigo50   = RgbHex("#e8eaf6")
	Indigo100  = RgbHex("#c5cae9")
	Indigo200  = RgbHex("#9fa8da")
	Indigo300  = RgbHex("#7986cb")
	Indigo400  = RgbHex("#5c6bc0")
	Indigo500  = RgbHex("#3f51b5")
	Indigo600  = RgbHex("#3949ab")
	Indigo700  = RgbHex("#303f9f")
	Indigo800  = RgbHex("#283593")
	Indigo900  = RgbHex("#1a237e")
	IndigoA100 = RgbHex("#8c9eff")
	IndigoA200 = RgbHex("#536dfe")
	IndigoA400 = RgbHex("#3d5afe")
	IndigoA700 = RgbHex("#304ffe")

	// Blue
	Blue     = RgbHex("#2196f3")
	Blue50   = RgbHex("#e3f2fd")
	Blue100  = RgbHex("#bbdefb")
	Blue200  = RgbHex("#90caf9")
	Blue300  = RgbHex("#64b5f6")
	Blue400  = RgbHex("#42a5f5")
	Blue500  = RgbHex("#2196f3")
	Blue600  = RgbHex("#1e88e5")
	Blue700  = RgbHex("#1976d2")
	Blue800  = RgbHex("#1565c0")
	Blue900  = RgbHex("#0d47a1")
	BlueA100 = RgbHex("#82b1ff")
	BlueA200 = RgbHex("#448aff")
	BlueA400 = RgbHex("#2979ff")
	BlueA700 = RgbHex("#2962ff")

	// LightBlue
	LightBlue     = RgbHex("#03a9f4")
	LightBlue50   = RgbHex("#e1f5fe")
	LightBlue100  = RgbHex("#b3e5fc")
	LightBlue200  = RgbHex("#81d4fa")
	LightBlue300  = RgbHex("#4fc3f7")
	LightBlue400  = RgbHex("#29b6f6")
	LightBlue500  = RgbHex("#03a9f4")
	LightBlue600  = RgbHex("#039be5")
	LightBlue700  = RgbHex("#0288d1")
	LightBlue800  = RgbHex("#0277bd")
	LightBlue900  = RgbHex("#01579b")
	LightBlueA100 = RgbHex("#80d8ff")
	LightBlueA200 = RgbHex("#40c4ff")
	LightBlueA400 = RgbHex("#00b0ff")
	LightBlueA700 = RgbHex("#0091ea")

	// Cyan
	Cyan     = RgbHex("#00bcd4")
	Cyan50   = RgbHex("#e0f7fa")
	Cyan100  = RgbHex("#b2ebf2")
	Cyan200  = RgbHex("#80deea")
	Cyan300  = RgbHex("#4dd0e1")
	Cyan400  = RgbHex("#26c6da")
	Cyan500  = RgbHex("#00bcd4")
	Cyan600  = RgbHex("#00acc1")
	Cyan700  = RgbHex("#0097a7")
	Cyan800  = RgbHex("#00838f")
	Cyan900  = RgbHex("#006064")
	CyanA100 = RgbHex("#84ffff")
	CyanA200 = RgbHex("#18ffff")
	CyanA400 = RgbHex("#00e5ff")
	CyanA700 = RgbHex("#00b8d4")

	// Teal
	Teal     = RgbHex("#009688")
	Teal50   = RgbHex("#e0f2f1")
	Teal100  = RgbHex("#b2dfdb")
	Teal200  = RgbHex("#80cbc4")
	Teal300  = RgbHex("#4db6ac")
	Teal400  = RgbHex("#26a69a")
	Teal500  = RgbHex("#009688")
	Teal600  = RgbHex("#00897b")
	Teal700  = RgbHex("#00796b")
	Teal800  = RgbHex("#00695c")
	Teal900  = RgbHex("#004d40")
	TealA100 = RgbHex("#a7ffeb")
	TealA200 = RgbHex("#64ffda")
	TealA400 = RgbHex("#1de9b6")
	TealA700 = RgbHex("#00bfa5")

	// Green
	Green     = RgbHex("#4caf50")
	Green50   = RgbHex("#e8f5e9")
	Green100  = RgbHex("#c8e6c9")
	Green200  = RgbHex("#a5d6a7")
	Green300  = RgbHex("#81c784")
	Green400  = RgbHex("#66bb6a")
	Green500  = RgbHex("#4caf50")
	Green600  = RgbHex("#43a047")
	Green700  = RgbHex("#388e3c")
	Green800  = RgbHex("#2e7d32")
	Green900  = RgbHex("#1b5e20")
	GreenA100 = RgbHex("#b9f6ca")
	GreenA200 = RgbHex("#69f0ae")
	GreenA400 = RgbHex("#00e676")
	GreenA700 = RgbHex("#00c853")

	// LightGreen
	LightGreen     = RgbHex("#8bc34a")
	LightGreen50   = RgbHex("#f1f8e9")
	LightGreen100  = RgbHex("#dcedc8")
	LightGreen200  = RgbHex("#c5e1a5")
	LightGreen300  = RgbHex("#aed581")
	LightGreen400  = RgbHex("#9ccc65")
	LightGreen500  = RgbHex("#8bc34a")
	LightGreen600  = RgbHex("#7cb342")
	LightGreen700  = RgbHex("#689f38")
	LightGreen800  = RgbHex("#558b2f")
	LightGreen900  = RgbHex("#33691e")
	LightGreenA100 = RgbHex("#ccff90")
	LightGreenA200 = RgbHex("#b2ff59")
	LightGreenA400 = RgbHex("#76ff03")
	LightGreenA700 = RgbHex("#64dd17")

	// Lime
	Lime     = RgbHex("#cddc39")
	Lime50   = RgbHex("#f9fbe7")
	Lime100  = RgbHex("#f0f4c3")
	Lime200  = RgbHex("#e6ee9c")
	Lime300  = RgbHex("#dce775")
	Lime400  = RgbHex("#d4e157")
	Lime500  = RgbHex("#cddc39")
	Lime600  = RgbHex("#c0ca33")
	Lime700  = RgbHex("#afb42b")
	Lime800  = RgbHex("#9e9d24")
	Lime900  = RgbHex("#827717")
	LimeA100 = RgbHex("#f4ff81")
	LimeA200 = RgbHex("#eeff41")
	LimeA400 = RgbHex("#c6ff00")
	LimeA700 = RgbHex("#aeea00")

	// Yellow
	Yellow     = RgbHex("#ffeb3b")
	Yellow50   = RgbHex("#fffde7")
	Yellow100  = RgbHex("#fff9c4")
	Yellow200  = RgbHex("#fff59d")
	Yellow300  = RgbHex("#fff176")
	Yellow400  = RgbHex("#ffee58")
	Yellow500  = RgbHex("#ffeb3b")
	Yellow600  = RgbHex("#fdd835")
	Yellow700  = RgbHex("#fbc02d")
	Yellow800  = RgbHex("#f9a825")
	Yellow900  = RgbHex("#f57f17")
	YellowA100 = RgbHex("#ffff8d")
	YellowA200 = RgbHex("#ffff00")
	YellowA400 = RgbHex("#ffea00")
	YellowA700 = RgbHex("#ffd600")

	// Amber
	Amber     = RgbHex("#ffc107")
	Amber50   = RgbHex("#fff8e1")
	Amber100  = RgbHex("#ffecb3")
	Amber200  = RgbHex("#ffe082")
	Amber300  = RgbHex("#ffd54f")
	Amber400  = RgbHex("#ffca28")
	Amber500  = RgbHex("#ffc107")
	Amber600  = RgbHex("#ffb300")
	Amber700  = RgbHex("#ffa000")
	Amber800  = RgbHex("#ff8f00")
	Amber900  = RgbHex("#ff6f00")
	AmberA100 = RgbHex("#ffe57f")
	AmberA200 = RgbHex("#ffd740")
	AmberA400 = RgbHex("#ffc400")
	AmberA700 = RgbHex("#ffab00")

	// Orange
	Orange     = RgbHex("#ff9800")
	Orange50   = RgbHex("#fff3e0")
	Orange100  = RgbHex("#ffe0b2")
	Orange200  = RgbHex("#ffcc80")
	Orange300  = RgbHex("#ffb74d")
	Orange400  = RgbHex("#ffa726")
	Orange500  = RgbHex("#ff9800")
	Orange600  = RgbHex("#fb8c00")
	Orange700  = RgbHex("#f57c00")
	Orange800  = RgbHex("#ef6c00")
	Orange900  = RgbHex("#e65100")
	OrangeA100 = RgbHex("#ffd180")
	OrangeA200 = RgbHex("#ffab40")
	OrangeA400 = RgbHex("#ff9100")
	OrangeA700 = RgbHex("#ff6d00")

	// DeepOrange
	DeppOrange     = RgbHex("#ff5722")
	DeppOrange50   = RgbHex("#fbe9e7")
	DeppOrange100  = RgbHex("#ffccbc")
	DeppOrange200  = RgbHex("#ffab91")
	DeppOrange300  = RgbHex("#ff8a65")
	DeppOrange400  = RgbHex("#ff7043")
	DeppOrange500  = RgbHex("#ff5722")
	DeppOrange600  = RgbHex("#f4511e")
	DeppOrange700  = RgbHex("#e64a19")
	DeppOrange800  = RgbHex("#d84315")
	DeppOrange900  = RgbHex("#bf360c")
	DeppOrangeA100 = RgbHex("#ff9e80")
	DeppOrangeA200 = RgbHex("#ff6e40")
	DeppOrangeA400 = RgbHex("#ff3d00")
	DeppOrangeA700 = RgbHex("#dd2c00")

	// Brown
	Brown     = RgbHex("#795548")
	Brown50   = RgbHex("#efebe9")
	Brown100  = RgbHex("#d7ccc8")
	Brown200  = RgbHex("#bcaaa4")
	Brown300  = RgbHex("#a1887f")
	Brown400  = RgbHex("#8d6e63")
	Brown500  = RgbHex("#795548")
	Brown600  = RgbHex("#6d4c41")
	Brown700  = RgbHex("#5d4037")
	Brown800  = RgbHex("#4e342e")
	Brown900  = RgbHex("#3e2723")
	BrownA100 = RgbHex("#d7ccc8")
	BrownA200 = RgbHex("#bcaaa4")
	BrownA400 = RgbHex("#8d6e63")
	BrownA700 = RgbHex("#5d4037")

	// Gray
	Gray     = RgbHex("#9e9e9e")
	Gray50   = RgbHex("#fafafa")
	Gray100  = RgbHex("#f5f5f5")
	Gray200  = RgbHex("#eeeeee")
	Gray300  = RgbHex("#e0e0e0")
	Gray400  = RgbHex("#bdbdbd")
	Gray500  = RgbHex("#9e9e9e")
	Gray600  = RgbHex("#757575")
	Gray700  = RgbHex("#616161")
	Gray800  = RgbHex("#424242")
	Gray900  = RgbHex("#212121")
	GrayA100 = RgbHex("#d5d5d5")
	GrayA200 = RgbHex("#aaaaaa")
	GrayA400 = RgbHex("#303030")
	GrayA700 = RgbHex("#616161")

	// BlueGray
	BlueGray     = RgbHex("#607d8b")
	BlueGray50   = RgbHex("#eceff1")
	BlueGray100  = RgbHex("#cfd8dc")
	BlueGray200  = RgbHex("#b0bec5")
	BlueGray300  = RgbHex("#90a4ae")
	BlueGray400  = RgbHex("#78909c")
	BlueGray500  = RgbHex("#607d8b")
	BlueGray600  = RgbHex("#546e7a")
	BlueGray700  = RgbHex("#455a64")
	BlueGray800  = RgbHex("#37474f")
	BlueGray900  = RgbHex("#263238")
	BlueGrayA100 = RgbHex("#cfd8dc")
	BlueGrayA200 = RgbHex("#b0bec5")
	BlueGrayA400 = RgbHex("#78909c")
	BlueGrayA700 = RgbHex("#455a64")
)
