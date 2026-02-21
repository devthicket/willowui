// xml-forms demonstrates WillowUI's XML template system with form input widgets:
// TextInput, TextArea, Toggle, Checkbox, Slider, NumberStepper, Select,
// OptionRotator, and Radio with RadioButton children.
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
	screenW = 800
	screenH = 700
)

// ---------------------------------------------------------------------------
// Controller
// ---------------------------------------------------------------------------

type formsController struct {
	name       *ui.Ref[string]
	email      *ui.Ref[string]
	notes      *ui.Ref[string]
	enabled    *ui.Ref[bool]
	agreed     *ui.Ref[bool]
	volume     *ui.Ref[float64]
	quantity   *ui.Ref[float64]
	classIndex *ui.Ref[int]
	diffIndex  *ui.Ref[int]
	status     *ui.Ref[string]

	volumeLabel   *ui.Computed[string]
	quantityLabel *ui.Computed[string]
}

func (c *formsController) OnCreate(s *ui.Screen) {
	c.name = ui.NewRef("")
	c.email = ui.NewRef("")
	c.notes = ui.NewRef("")
	c.enabled = ui.NewRef(true)
	c.agreed = ui.NewRef(false)
	c.volume = ui.NewRef(75.0)
	c.quantity = ui.NewRef(1.0)
	c.classIndex = ui.NewRef(0)
	c.diffIndex = ui.NewRef(1)
	c.status = ui.NewRef("Fill out the form above.")

	c.volumeLabel = ui.NewComputed(func() string {
		return fmt.Sprintf("Volume: %.0f", c.volume.Get())
	})
	c.quantityLabel = ui.NewComputed(func() string {
		return fmt.Sprintf("Quantity: %.0f", c.quantity.Get())
	})

	reg := ui.NewTemplateRegistry()
	font := ui.MustLoadDefaultFont()
	reg.SetFonts(nil, font)
	reg.SetFontSize(16)

	if err := reg.RegisterXML("forms", templateXML); err != nil {
		log.Fatalf("register template: %v", err)
	}

	comp, err := reg.Instantiate("forms", c, s)
	if err != nil {
		log.Fatalf("instantiate template: %v", err)
	}
	comp.X = 40
	comp.Y = 60
	s.Add(comp)

	lh := font.LineHeight(0, false, false)
	scale := 20.0 / lh
	titleNode := willow.NewText("page-title", "WillowUI: XML Forms Demo", font)
	titleNode.TextBlock.Color = willow.RGBA(1, 1, 1, 1)
	titleNode.SetScale(scale, scale)
	titleNode.SetPosition(24, 16)
	s.AddNode(titleNode)
}

func (c *formsController) OnUpdate(_ float64) {}
func (c *formsController) OnDestroy()         {}

// DataProvider — maps names to reactive values.
func (c *formsController) LookupRef(path string) any {
	switch path {
	case "name":
		return c.name
	case "email":
		return c.email
	case "notes":
		return c.notes
	case "enabled":
		return c.enabled
	case "agreed":
		return c.agreed
	case "volume":
		return c.volume
	case "quantity":
		return c.quantity
	case "classIndex":
		return c.classIndex
	case "diffIndex":
		return c.diffIndex
	case "status":
		return c.status
	case "volumeLabel":
		return c.volumeLabel
	case "quantityLabel":
		return c.quantityLabel
	}
	return nil
}

var classNames = []string{"Warrior", "Mage", "Rogue", "Paladin"}
var diffNames = []string{"Easy", "Normal", "Hard", "Nightmare"}

func (c *formsController) CallMethod(name string) bool {
	switch name {
	case "onNameChange":
		c.status.Set("Name: " + c.name.Peek())
	case "onToggleChange":
		if c.enabled.Peek() {
			c.status.Set("Toggle: on")
		} else {
			c.status.Set("Toggle: off")
		}
	case "onAgreedChange":
		if c.agreed.Peek() {
			c.status.Set("Terms accepted")
		} else {
			c.status.Set("Terms not accepted")
		}
	case "onVolumeChange":
		c.status.Set(fmt.Sprintf("Volume changed to %.0f", c.volume.Peek()))
	case "onQuantityChange":
		c.status.Set(fmt.Sprintf("Quantity changed to %.0f", c.quantity.Peek()))
	case "onClassChange":
		idx := c.classIndex.Peek()
		if idx >= 0 && idx < len(classNames) {
			c.status.Set("Class: " + classNames[idx])
		}
	case "onDiffChange":
		idx := c.diffIndex.Peek()
		if idx >= 0 && idx < len(diffNames) {
			c.status.Set("Difficulty: " + diffNames[idx])
		}
	case "onThemeChange":
		c.status.Set("Theme changed")
	case "onSubmit":
		cls := ""
		if idx := c.classIndex.Peek(); idx >= 0 && idx < len(classNames) {
			cls = classNames[idx]
		}
		c.status.Set(fmt.Sprintf(
			"Submitted -- name=%q class=%s vol=%.0f qty=%.0f agreed=%v",
			c.name.Peek(), cls, c.volume.Peek(), c.quantity.Peek(), c.agreed.Peek(),
		))
	}
	return true
}

// ---------------------------------------------------------------------------
// Main
// ---------------------------------------------------------------------------

func main() {
	ui.Stage.Add(ui.NewScreen(ui.WithController(&formsController{})))

	ui.Setup(ui.StageConfig{
		Title:      "WillowUI — XML Forms Demo",
		Width:      screenW,
		Height:     screenH,
		ClearColor: willow.RGBA(0.08, 0.08, 0.10, 1),
	})
}
