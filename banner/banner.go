package banner

import (
	"fmt"
	"strings"

	"github.com/mbndr/figlet4go"
)

type Render interface {
	Render(label string) (string, error)
}

///////////////////////////////////
// Ascii Render
type render struct {
	figlet *figlet4go.AsciiRender
}

func NewRender() Render {
	return &render{
		figlet: figlet4go.NewAsciiRender(),
	}
}

func (r *render) Render(label string) (string, error) {
	return r.figlet.Render(label)
}

//////////////////////////////////

func GenBootBanner(pname string, version string) string {
	fig := figlet4go.NewAsciiRender()
	logo, _ := fig.Render(pname)
	return fmt.Sprintf("boot\n%v     %v\n",
		strings.TrimRight(logo, "\r\n"),
		version)
}
