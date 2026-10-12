package main

import (
	"embed"
	"github.com/egoist/mygo/ui"
)

//go:embed resources/tray.png
var trayPNG []byte

//go:embed resources/ui/*.svg resources/icon.svg
var artwork embed.FS
var symbols = map[string]*ui.SVG{}
var appMark = func() *ui.SVG { data, _ := artwork.ReadFile("resources/icon.svg"); return ui.MustParseSVG(data) }()

func trayIcon() []byte { return trayPNG }
func symbol(name string) *ui.SVG {
	if svg := symbols[name]; svg != nil {
		return svg
	}
	data, err := artwork.ReadFile("resources/ui/" + name + ".svg")
	if err != nil {
		data, _ = artwork.ReadFile("resources/ui/file.svg")
	}
	svg := ui.MustParseSVG(data)
	symbols[name] = svg
	return svg
}
func glyph(c *ui.Context, name string, size float32, color ui.Color) {
	ui.Icon(c, symbol(name)).Size(size, size).TextColor(color)
}
func iconButton(c *ui.Context, name, label string, enabled bool) bool {
	e := ui.Row(c).Key(label).Size(29, 29).Radius(6).AlignItems(ui.Center).Justify(ui.Center).Label(label).Tooltip(label).Focusable().Disabled(!enabled)
	if e.Hovered() {
		e.Background(c.Theme().SurfaceHover)
	}
	e.Children(func() { glyph(c, name, 17, c.Theme().TextMuted) })
	return e.Clicked() || e.Shortcut(0, ui.KeyEnter) || e.Shortcut(0, ui.KeySpace)
}
