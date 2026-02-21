// hotreload demonstrates WillowUI's hot reload feature: edit template.xml
// while the app is running and see the UI update live.
//
// Run with:
//
//	go run -tags hotreload ./examples/templating/hotreload/
//
// Then edit examples/templating/hotreload/template.xml and save — the UI reloads
// automatically.
package main

import (
	"fmt"
	"log"
	"os"

	"github.com/devthicket/willow"
	ui "github.com/devthicket/willowui"
)

const (
	screenW = 800
	screenH = 600
	xmlPath = "examples/templating/hotreload/template.xml"
)

type hotReloadController struct {
	counter   *ui.Ref[int]
	title     *ui.Ref[string]
	showExtra *ui.Ref[bool]
	reloader  *ui.HotReloader
}

func (c *hotReloadController) OnCreate(s *ui.Screen) {
	c.counter = ui.NewRef(0)
	c.title = ui.NewRef("Hot Reload Demo")
	c.showExtra = ui.NewRef(false)

	reg := ui.NewTemplateRegistry()

	font := ui.MustLoadDefaultFont()
	reg.SetFonts(nil, font)
	reg.SetFontSize(16)

	// Register initial template from disk.
	xmlData, err := os.ReadFile(xmlPath)
	if err != nil {
		log.Fatalf("read template: %v", err)
	}
	if err := reg.RegisterXML("demo", xmlData); err != nil {
		log.Fatalf("register template: %v", err)
	}

	comp, err := reg.Instantiate("demo", c, s)
	if err != nil {
		log.Fatalf("instantiate template: %v", err)
	}

	s.Add(comp)

	// Start hot reloader.
	c.reloader, err = ui.NewHotReloader(reg, s, c, "demo", xmlPath)
	if err != nil {
		log.Printf("warning: hot reload unavailable: %v", err)
	} else {
		log.Println("hot reload: watching", xmlPath)
	}

	titleNode := willow.NewText("page-title", "Watching: "+xmlPath, font)
	titleNode.TextBlock.FontSize = 16
	titleNode.TextBlock.Color = willow.RGBA(1, 1, 1, 1)
	titleNode.SetPosition(24, 16)
	s.AddNode(titleNode)

	divider := willow.NewSprite("divider", willow.TextureRegion{})
	divider.SetPosition(24, 48)
	divider.SetScale(screenW-48, 1)
	divider.SetColor(willow.RGBA(0.25, 0.3, 0.35, 1))
	s.AddNode(divider)
}

func (c *hotReloadController) OnUpdate(dt float64) {}

func (c *hotReloadController) OnDestroy() {
	if c.reloader != nil {
		c.reloader.Stop()
	}
}

func (c *hotReloadController) LookupRef(path string) any {
	switch path {
	case "title":
		return c.title
	case "statusText":
		return fmt.Sprintf("Count: %d", c.counter.Get())
	case "showExtra":
		return c.showExtra
	}
	return nil
}

func (c *hotReloadController) CallMethod(name string) bool {
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
	}
	return false
}

func main() {
	ui.Stage.Add(ui.NewScreen(ui.WithController(&hotReloadController{})))

	ui.Setup(ui.StageConfig{
		Title:      "WillowUI — Hot Reload Demo",
		Width:      screenW,
		Height:     screenH,
		ClearColor: willow.RGBA(0.08, 0.08, 0.10, 1),
	})
}
