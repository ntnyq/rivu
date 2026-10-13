package main

import "github.com/egoist/mygo/ui"
import "github.com/ntnyq/rivu/internal/session"

type palette struct{ surface, sidebar, alt, text, muted, line, blue, soft, green, orange, track ui.Color }

func designTheme(c *ui.Context) palette {
	p := palette{ui.Hex("#ffffff"), ui.Hex("#f0f1f5"), ui.Hex("#f7f8fa"), ui.Hex("#20232a"), ui.Hex("#626975"), ui.Hex("#e3e5e9"), ui.Hex("#0868db"), ui.Hex("#eaf2ff"), ui.Hex("#25834c"), ui.Hex("#a75a0a"), ui.Hex("#e5e8ed")}
	if c.Theme().Dark {
		p = palette{ui.Hex("#202226"), ui.Hex("#27292e"), ui.Hex("#272a30"), ui.Hex("#f1f2f5"), ui.Hex("#a8aeb9"), ui.Hex("#3b3e45"), ui.Hex("#5aa2ff"), ui.Hex("#283d59"), ui.Hex("#6bc18b"), ui.Hex("#e9af67"), ui.Hex("#40444d")}
	}
	theme := *c.Theme()
	theme.Background = p.surface
	theme.Surface = p.surface
	theme.SurfaceHover = p.track
	theme.SurfacePressed = p.soft
	theme.Text = p.text
	theme.TextMuted = p.muted
	theme.Border = p.line
	theme.Accent = p.blue
	theme.Success = p.green
	theme.Warning = p.orange
	theme.FontSize = 12
	theme.Radius = 6
	c.SetTheme(&theme)
	return p
}
func taskColor(c *ui.Context, phase session.Phase) ui.Color {
	t := c.Theme()
	switch phase {
	case "completed", "seeding":
		return t.Success
	case "error":
		return t.Warning
	case "paused":
		return ui.Hex("#9ca7b9")
	}
	return t.Accent
}
func progressTrack(c *ui.Context, value float64, color, track ui.Color, height float32) {
	ui.Row(c).FillWidth().Height(height).Radius(height).Background(track).Children(func() {
		ui.Box(c).WidthPercent(float32(min(1, max(0, value)) * 100)).Height(height).Radius(height).Background(color)
	})
}
