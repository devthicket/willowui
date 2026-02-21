// Counter — minimal reactive example.
package main

import (
	"github.com/devthicket/willow"
	ui "github.com/devthicket/willowui"
)

func main() {
	font := ui.MustLoadDefaultFont()

	clicks := ui.NewRef(0)

	btn := ui.NewButton("inc", "+1", font, 16)
	btn.SetOnClick(ui.Increment(clicks, 1))

	label := ui.NewLabel("count", "0", font, 36)
	formatted, _ := ui.BindFormatterf(clicks, "Count: %d") // app-lifetime ref; handle intentionally not stopped
	label.BindText(formatted)

	row := ui.NewHBox("row")
	row.Spacing = 16
	row.SetPosition(20, 15)
	row.AddChild(btn)
	row.AddChild(label)

	ui.Setup(ui.StageConfig{Title: "Counter", Width: 400, Height: 70,
		ClearColor: willow.RGBA(0.08, 0.08, 0.10, 1)}, row)
}
