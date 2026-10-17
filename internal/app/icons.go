package app

import (
	"github.com/egoist/mygo/ui"
	"github.com/ntnyq/rivu/resources"
)

var trayPNG, _ = resources.ReadFile("tray.png")
var symbols = map[string]*ui.SVG{}
var appMark = func() *ui.SVG { data, _ := resources.ReadFile("icon.svg"); return ui.MustParseSVG(data) }()

func trayIcon() []byte { return trayPNG }
func symbol(name string) *ui.SVG {
	if svg := symbols[name]; svg != nil {
		return svg
	}
	data, err := resources.ReadFile("ui/" + name + ".svg")
	if err != nil {
		data, _ = resources.ReadFile("ui/file.svg")
	}
	svg := ui.MustParseSVG(data)
	symbols[name] = svg
	return svg
}
