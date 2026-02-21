// xml-reactive demonstrates every bind: attribute in WillowUI's XML template system.
//
// Patterns shown:
//   - bind:text    with a plain Ref[string] and with a ternary expression
//   - bind:value   on TextInput, Toggle, Checkbox, Slider, ProgressBar
//   - bind:checked on Checkbox (alias for bind:value on bool widgets)
//   - bind:selected on Select, OptionRotator, Radio, TabBar
//   - bind:enabled  gates controls reactively from a Ref[bool]
//   - ui:show       toggles visibility based on a Ref[bool]
//   - Controller buttons mutate refs from Go; the template reacts automatically
package main

import (
	_ "embed"
	"fmt"
	"log"

	"github.com/devthicket/willow"
	ui "github.com/devthicket/willowui"
)

//go:embed template.xml
var templateXML []byte

const (
	screenW = 900
	screenH = 680
)

// ---------------------------------------------------------------------------
// Controller
// ---------------------------------------------------------------------------

var (
	classNames = []string{"Warrior", "Mage", "Rogue", "Paladin"}
	diffNames  = []string{"Easy", "Normal", "Hard", "Nightmare"}
	sizeNames  = []string{"Small", "Medium", "Large"}
)

type reactiveController struct {
	name       *ui.Ref[string]
	enabled    *ui.Ref[bool]
	agreed     *ui.Ref[bool]
	volume     *ui.Ref[float64]
	classIndex *ui.Ref[int]
	diffIndex  *ui.Ref[int]
	sizeIndex  *ui.Ref[int]

	volumeLabel *ui.Computed[string]
	summary     *ui.Computed[string]
}

func (c *reactiveController) OnCreate(s *ui.Screen) {
	c.name = ui.NewRef("")
	c.enabled = ui.NewRef(true)
	c.agreed = ui.NewRef(false)
	c.volume = ui.NewRef(0.5)
	c.classIndex = ui.NewRef(0)
	c.diffIndex = ui.NewRef(1) // Normal
	c.sizeIndex = ui.NewRef(0)

	c.volumeLabel = ui.NewComputed(func() string {
		return fmt.Sprintf("Volume: %.0f%%", c.volume.Get()*100)
	})

	c.summary = ui.NewComputed(func() string {
		ci := c.classIndex.Get()
		di := c.diffIndex.Get()
		si := c.sizeIndex.Get()
		cls := safeIdx(classNames, ci)
		dif := safeIdx(diffNames, di)
		siz := safeIdx(sizeNames, si)
		return fmt.Sprintf("class=%s  diff=%s  size=%s", cls, dif, siz)
	})

	font := ui.MustLoadDefaultFont()

	reg := ui.NewTemplateRegistry()
	reg.SetFonts(nil, font)
	reg.SetFontSize(15)

	if err := reg.RegisterXML("reactive", templateXML); err != nil {
		log.Fatalf("register template: %v", err)
	}

	comp, err := reg.Instantiate("reactive", c, s)
	if err != nil {
		log.Fatalf("instantiate template: %v", err)
	}
	comp.X = 30
	comp.Y = 56
	s.Add(comp)

	lh := font.LineHeight(0, false, false)
	scale := 18.0 / lh
	titleNode := willow.NewText("page-title", "WillowUI: XML Reactive Bindings Demo", font)
	titleNode.TextBlock.Color = willow.RGBA(1, 1, 1, 1)
	titleNode.SetScale(scale, scale)
	titleNode.SetPosition(24, 14)
	s.AddNode(titleNode)
}

func (c *reactiveController) OnUpdate(_ float64) {}
func (c *reactiveController) OnDestroy()         {}

// DataProvider — maps names to reactive values for the template.
func (c *reactiveController) LookupRef(path string) any {
	switch path {
	case "name":
		return c.name
	case "enabled":
		return c.enabled
	case "agreed":
		return c.agreed
	case "volume":
		return c.volume
	case "classIndex":
		return c.classIndex
	case "diffIndex":
		return c.diffIndex
	case "sizeIndex":
		return c.sizeIndex
	case "volumeLabel":
		return c.volumeLabel // *Computed[string] — unwrapped automatically
	case "summary":
		return c.summary // *Computed[string]
	}
	return nil
}

func (c *reactiveController) CallMethod(name string) bool {
	switch name {
	case "nextClass":
		c.classIndex.Update(func(i int) int { return (i + 1) % len(classNames) })
	case "nextDiff":
		c.diffIndex.Update(func(i int) int { return (i + 1) % len(diffNames) })
	case "cycleSize":
		c.sizeIndex.Update(func(i int) int { return (i + 1) % len(sizeNames) })
	case "toggleEnable":
		c.enabled.Update(func(v bool) bool { return !v })
	}
	return true
}

// ---------------------------------------------------------------------------
// Main
// ---------------------------------------------------------------------------

func main() {
	ui.Stage.Add(ui.NewScreen(ui.WithController(&reactiveController{})))

	ui.Setup(ui.StageConfig{
		Title:      "WillowUI — XML Reactive Demo",
		Width:      screenW,
		Height:     screenH,
		ClearColor: willow.RGBA(0.08, 0.08, 0.10, 1),
	})
}

func safeIdx(s []string, i int) string {
	if i >= 0 && i < len(s) {
		return s[i]
	}
	return "?"
}
