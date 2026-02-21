// xml-layout demonstrates all WillowUI layout modes via a single XML template.
// A TabBar switches between four tabs: Stacks (VBox/HBox), Grid, Alignment,
// and AnchorLayout. All content is defined in layout.xml.
package main

import (
	_ "embed"
	"log"

	"github.com/devthicket/willow"
	ui "github.com/devthicket/willowui"
)

//go:embed layout.xml
var layoutXML []byte

const (
	screenW = 800
	screenH = 650
)

type layoutController struct {
	font *willow.FontFamily
}

func (c *layoutController) OnCreate(s *ui.Screen) {
	font := c.font

	// Title.
	title := willow.NewText("title", "WillowUI -- XML Layout Demo", font)
	title.TextBlock.FontSize = 22
	title.TextBlock.Color = willow.RGBA(1, 1, 1, 1)
	title.SetPosition(24, 12)
	s.AddNode(title)

	divider := ui.NewDivider("divider", screenW-48)
	divider.SetPosition(24, 42)
	s.AddNode(divider)

	// Template registry.
	reg := ui.NewTemplateRegistry()
	reg.SetFonts(nil, font)
	reg.SetFontSize(14)

	if err := reg.RegisterXML("layout", layoutXML); err != nil {
		log.Fatalf("register layout: %v", err)
	}

	comp, err := reg.InstantiateStatic("layout", s)
	if err != nil {
		log.Fatalf("instantiate layout: %v", err)
	}
	comp.SetPosition(24, 52)
	comp.Width = screenW - 48
	comp.Height = screenH - 60
	comp.MarkLayoutDirty()
	s.Add(comp)
}

func (c *layoutController) OnUpdate(dt float64) {}
func (c *layoutController) OnDestroy()          {}

func main() {
	font := ui.MustLoadDefaultFont()

	ui.Stage.Add(ui.NewScreen(ui.WithController(&layoutController{font: font})))
	ui.Setup(ui.StageConfig{
		Title:      "WillowUI -- XML Layout Demo",
		Width:      screenW,
		Height:     screenH,
		ClearColor: willow.RGBA(0.08, 0.08, 0.10, 1),
	})
}
