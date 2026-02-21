// xml-variants demonstrates custom Panel and Label variants defined in a theme
// JSON file and referenced by name in XML templates. No inline color or
// cornerRadius attributes are needed in the XML — all visual style comes from
// the theme.
package main

import (
	_ "embed"
	"log"

	"github.com/devthicket/willow"
	ui "github.com/devthicket/willowui"
)

//go:embed template.xml
var templateXML []byte

//go:embed theme.json
var themeJSON []byte

const (
	screenW = 560
	screenH = 480
)

type ctrl struct{}

func (c *ctrl) OnCreate(s *ui.Screen) {
	// Compile the theme — this resolves the "variants" array and all named groups.
	th, err := ui.LoadTheme(themeJSON)
	if err != nil {
		log.Fatalf("load theme: %v", err)
	}

	font := ui.MustLoadDefaultFont()

	reg := ui.NewTemplateRegistry()
	reg.SetFonts(nil, font)
	reg.SetFontSize(14)
	// SetTheme registers custom variant names (card, surface, badge-ok, …)
	// so that variant="card" in XML resolves to the correct theme slot.
	reg.SetTheme(th)

	if err := reg.RegisterXML("main", templateXML); err != nil {
		log.Fatalf("register template: %v", err)
	}

	comp, err := reg.Instantiate("main", c, s)
	if err != nil {
		log.Fatalf("instantiate: %v", err)
	}

	// Apply the compiled theme to the root component so all widgets inherit it.
	comp.SetTheme(th)
	comp.X = 40
	comp.Y = 40
	s.Add(comp)
}

func (c *ctrl) OnUpdate(dt float64) {}
func (c *ctrl) OnDestroy()          {}

func main() {
	ui.Stage.Add(ui.NewScreen(ui.WithController(&ctrl{})))

	ui.Setup(ui.StageConfig{
		Title:      "WillowUI — XML Variant Demo",
		Width:      screenW,
		Height:     screenH,
		ClearColor: willow.RGBA(0.07, 0.08, 0.10, 1),
	})
}
