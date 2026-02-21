// xml-binary demonstrates WillowUI's binary template format (.xmlui): the same
// demo as xml-basic but using a pre-compiled binary template loaded via
// go:embed and RegisterBinary instead of RegisterXML.
//
// To regenerate the .xmlui file:
//
//	go run ./cmd/xmluic examples/templating/xml-binary/template.xml examples/templating/xml-binary/template.xmlui
package main

import (
	_ "embed"
	"fmt"
	"log"

	"github.com/devthicket/willow"
	ui "github.com/devthicket/willowui"
)

//go:embed template.xmlui
var templateXMLUI []byte

const (
	screenW = 800
	screenH = 600
)

// ---------------------------------------------------------------------------
// Controller with DataProvider
// ---------------------------------------------------------------------------

type binaryDemoController struct {
	counter   *ui.Ref[int]
	title     *ui.Ref[string]
	showExtra *ui.Ref[bool]
}

func (c *binaryDemoController) OnCreate(s *ui.Screen) {
	c.counter = ui.NewRef(0)
	c.title = ui.NewRef("Binary Template Demo")
	c.showExtra = ui.NewRef(false)

	reg := ui.NewTemplateRegistry()

	font := ui.MustLoadDefaultFont()
	reg.SetFonts(nil, font)
	reg.SetFontSize(16)

	if err := reg.RegisterBinary("demo", templateXMLUI); err != nil {
		log.Fatalf("register template: %v", err)
	}

	comp, err := reg.Instantiate("demo", c, s)
	if err != nil {
		log.Fatalf("instantiate template: %v", err)
	}

	comp.X = 40
	comp.Y = 60
	s.Add(comp)

	lh := font.LineHeight(0, false, false)
	scaleTitle := 16.0 / lh
	titleNode := willow.NewText("page-title", "WillowUI: Binary Template Demo", font)
	titleNode.TextBlock.Color = willow.RGBA(1, 1, 1, 1)
	titleNode.SetScale(scaleTitle, scaleTitle)
	titleNode.SetPosition(24, 16)
	s.AddNode(titleNode)

	divider := willow.NewSprite("divider", willow.TextureRegion{})
	divider.SetPosition(24, 48)
	divider.SetScale(screenW-48, 1)
	divider.SetColor(willow.RGBA(0.25, 0.3, 0.35, 1))
	s.AddNode(divider)
}

func (c *binaryDemoController) OnUpdate(dt float64) {}
func (c *binaryDemoController) OnDestroy()          {}

func (c *binaryDemoController) LookupRef(path string) any {
	switch path {
	case "title":
		return c.title
	case "statusText":
		return fmt.Sprintf("Count: %d", c.counter.Get())
	case "showExtra":
		return c.showExtra
	case "toggleLabel":
		if c.showExtra.Get() {
			return "Hide Extra"
		}
		return "Show Extra"
	}
	return nil
}

func (c *binaryDemoController) CallMethod(name string) bool {
	switch name {
	case "increment":
		v := c.counter.Peek() + 1
		c.counter.Set(v)
		c.showExtra.Set(v > 3)
		return true
	case "reset":
		c.counter.Set(0)
		c.showExtra.Set(false)
		return true
	case "toggleExtra":
		c.showExtra.Set(!c.showExtra.Peek())
		return true
	}
	return false
}

// ---------------------------------------------------------------------------
// Main
// ---------------------------------------------------------------------------

func main() {
	ui.Stage.Add(ui.NewScreen(ui.WithController(&binaryDemoController{})))

	ui.Setup(ui.StageConfig{
		Title:      "WillowUI — Binary Template Demo",
		Width:      screenW,
		Height:     screenH,
		ClearColor: willow.RGBA(0.08, 0.08, 0.10, 1),
	})
}
