package app

import "github.com/egoist/mygo/ui"

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

// actionButton keeps the native button behavior and theme for icon/text actions.
func actionButton(c *ui.Context, name, label string, primary bool) ui.Element {
	var button ui.Element
	if primary {
		button = ui.PrimaryButton(c, "")
	} else {
		button = ui.Button(c, "")
	}
	return button.Label(label).Children(func() {
		ui.Icon(c, symbol(name)).Size(14, 14).Shrink(0)
		ui.Text(c, label).SingleLine()
	})
}

// menuControl centers the disclosure icon independently of the text baseline.
func menuControl(c *ui.Context, label, value string) ui.Element {
	return ui.Button(c, "").Key(label).Label(label).Height(28).MinWidth(112).Padding(0, 10).Gap(10).Children(func() {
		ui.Text(c, value).FontSize(11).SingleLine().Grow(1)
		ui.Icon(c, symbol("chevron-down")).Size(12, 12).Shrink(0)
	})
}

func navigationButton(c *ui.Context, label string, selected bool, p palette) ui.Element {
	button := ui.ButtonBase(c).Key(label).Label(label).Height(36).FillWidth().Padding(0, 10).Gap(9).Radius(6).Justify(ui.Start)
	switch {
	case selected:
		button.Background(p.blue).TextColor(ui.Hex("#ffffff"))
	case button.Pressed():
		button.Background(p.soft)
	case button.Hovered():
		button.Background(p.track)
	}
	return button
}

// Compact controls use the prototype's row and footer sizes, without inheriting
// the larger toolbar hit box that shifts task text and progress tracks.
func compactIconButton(c *ui.Context, name, label string, size, iconSize float32, color ui.Color) bool {
	e := ui.Row(c).Key(label).Size(size, size).Shrink(0).Radius(5).Justify(ui.Center).Label(label).Tooltip(label).Focusable()
	if e.Hovered() {
		e.Background(c.Theme().SurfaceHover)
	}
	e.Children(func() { glyph(c, name, iconSize, color) })
	return e.Clicked() || e.Shortcut(0, ui.KeyEnter) || e.Shortcut(0, ui.KeySpace)
}
