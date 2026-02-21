// custom-component demonstrates how to build a reusable custom widget and
// register it for use in XML templates via RegisterWidget.
//
// The example defines a StarRating widget that composes Panel + Labels,
// handles focus and keyboard input, and supports reactive binding.
package main

import (
	_ "embed"
	"fmt"
	"log"
	"strconv"

	"github.com/devthicket/willow"
	ui "github.com/devthicket/willowui"
	"github.com/hajimehoshi/ebiten/v2"
)

//go:embed template.xml
var templateXML []byte

const (
	screenW = 800
	screenH = 600
)

// ---------------------------------------------------------------------------
// StarRating — custom widget
// ---------------------------------------------------------------------------

// StarRating is a clickable 1-N star rating widget. It demonstrates all four
// custom component integration points: composition, focus+keyboard, ticking,
// and reactive binding.
type StarRating struct {
	root     *ui.Panel
	stars    []*ui.Label
	maxStars int
	value    int
	onChange func(int)

	font     *willow.FontFamily
	fontSize float64
}

// newStarRating creates a star rating widget.
func newStarRating(name string, font *willow.FontFamily, fontSize float64) *StarRating {
	sr := &StarRating{
		maxStars: 5,
		font:     font,
		fontSize: fontSize,
	}

	sr.root = ui.NewPanel(name)
	sr.root.Layout = ui.LayoutHBox
	sr.root.Spacing = 4

	// Make the root panel focusable for keyboard input.
	sr.root.Focusable = true
	sr.root.AllowTab = true

	// Store typed reference in UserData so XML setters can cast it back.
	sr.root.SetUserData(sr)

	// Register with the focus manager so Tab cycling reaches this widget.
	ui.FM.Register(&sr.root.Component)

	sr.buildStars()

	// Set a default size and hit shape so the widget is clickable and focusable.
	sr.SetWidth(float64(sr.maxStars) * (fontSize + 4))

	// OnUpdate polls keyboard each frame for arrow key input.
	sr.root.Node().OnUpdate = func(_ float64) {
		if !sr.root.IsFocused() || !sr.root.IsEnabled() {
			return
		}
		if ui.Input.IsKeyJustAvailable(ebiten.KeyRight) {
			sr.SetValue(sr.value + 1)
			ui.Input.Consume(ebiten.KeyRight)
		}
		if ui.Input.IsKeyJustAvailable(ebiten.KeyLeft) {
			sr.SetValue(sr.value - 1)
			ui.Input.Consume(ebiten.KeyLeft)
		}
	}

	return sr
}

// buildStars creates the star labels based on maxStars.
func (sr *StarRating) buildStars() {
	// Remove existing stars.
	for _, s := range sr.stars {
		sr.root.RemoveChild(s)
	}
	sr.stars = nil

	for i := 0; i < sr.maxStars; i++ {
		idx := i + 1 // 1-based rating value
		label := ui.NewLabel(
			fmt.Sprintf("star-%d", idx),
			"-",
			sr.font,
			sr.fontSize,
		)
		// Click on a star sets the rating to that star's index.
		label.Node().OnClick(func(_ willow.ClickContext) {
			if sr.root.IsEnabled() {
				sr.SetValue(idx)
			}
		})
		sr.root.AddChild(label)
		sr.stars = append(sr.stars, label)
	}
	sr.updateDisplay()
}

// Node returns the root node for scene integration.
func (sr *StarRating) Node() *willow.Node {
	return sr.root.Node()
}

// Component returns the root Component for XML template integration.
func (sr *StarRating) Component() *ui.Component {
	return &sr.root.Component
}

// SetValue sets the current rating, clamped to [0, maxStars].
func (sr *StarRating) SetValue(v int) {
	if v < 0 {
		v = 0
	}
	if v > sr.maxStars {
		v = sr.maxStars
	}
	if v == sr.value {
		return
	}
	sr.value = v
	sr.updateDisplay()
	if sr.onChange != nil {
		sr.onChange(v)
	}
	// Fire the "change" event so on:change in XML templates works.
	sr.root.FireEvent("change")
}

// Value returns the current rating.
func (sr *StarRating) Value() int {
	return sr.value
}

// SetMaxStars changes the number of stars and rebuilds the display.
func (sr *StarRating) SetMaxStars(n int) {
	if n < 1 {
		n = 1
	}
	if n == sr.maxStars {
		return
	}
	sr.maxStars = n
	if sr.value > n {
		sr.value = n
	}
	sr.buildStars()
	sr.SetWidth(float64(n) * (sr.fontSize + 4))
}

// SetOnChange registers a callback invoked when the rating changes.
func (sr *StarRating) SetOnChange(fn func(int)) {
	sr.onChange = fn
}

// updateDisplay refreshes the star labels to show filled (*) or empty (-).
func (sr *StarRating) updateDisplay() {
	for i, label := range sr.stars {
		if i < sr.value {
			label.SetText("*")
		} else {
			label.SetText("-")
		}
	}
}

// SetWidth sets the width of the root panel and hit shape.
func (sr *StarRating) SetWidth(w float64) {
	h := sr.fontSize + 8
	sr.root.SetSize(w, h)
	sr.root.Node().HitShape = willow.HitRect{Width: w, Height: h}
}

// ---------------------------------------------------------------------------
// XML Controller
// ---------------------------------------------------------------------------

type controller struct {
	rating *ui.Ref[int]
}

func (c *controller) OnCreate(s *ui.Screen) {
	c.rating = ui.NewRef(0)

	reg := ui.NewTemplateRegistry()

	font := ui.MustLoadDefaultFont()
	reg.SetFonts(nil, font)
	reg.SetFontSize(16)

	// Register the custom StarRating widget for XML templates.
	reg.RegisterWidget("StarRating", func(name string) (*ui.Component, error) {
		sr := newStarRating(name, font, 16)
		// Wire onChange to update the controller's ref so that bind:value
		// flows back from the widget to the data model (two-way binding).
		sr.SetOnChange(func(v int) {
			c.rating.Set(v)
		})
		return sr.Component(), nil
	}, map[string]ui.AttrSetter{
		"value": func(comp *ui.Component, val any) {
			if sr, ok := comp.UserData().(*StarRating); ok {
				switch v := val.(type) {
				case int:
					sr.SetValue(v)
				case float64:
					sr.SetValue(int(v))
				case string:
					n, _ := strconv.Atoi(v)
					sr.SetValue(n)
				}
			}
		},
		"maxStars": func(comp *ui.Component, val any) {
			if sr, ok := comp.UserData().(*StarRating); ok {
				switch v := val.(type) {
				case int:
					sr.SetMaxStars(v)
				case float64:
					sr.SetMaxStars(int(v))
				case string:
					n, _ := strconv.Atoi(v)
					sr.SetMaxStars(n)
				}
			}
		},
	})

	if err := reg.RegisterXML("demo", templateXML); err != nil {
		log.Fatalf("register template: %v", err)
	}

	comp, err := reg.Instantiate("demo", c, s)
	if err != nil {
		log.Fatalf("instantiate template: %v", err)
	}

	comp.X = 40
	comp.Y = 40
	s.Add(comp)
}

func (c *controller) OnUpdate(dt float64) {}
func (c *controller) OnDestroy()          {}

func (c *controller) LookupRef(path string) any {
	switch path {
	case "title":
		return "Custom Component: StarRating"
	case "rating":
		return c.rating
	case "ratingText":
		return fmt.Sprintf("Current rating: %d", c.rating.Get())
	}
	return nil
}

func (c *controller) CallMethod(name string) bool {
	switch name {
	case "reset":
		c.rating.Set(0)
		return true
	case "setThree":
		c.rating.Set(3)
		return true
	case "onRatingChange":
		// Fired by on:change on the StarRating via FireEvent("change").
		// The ref is already updated by onChange, so we can read it here.
		fmt.Printf("Rating changed to: %d\n", c.rating.Peek())
		return true
	}
	return false
}

// ---------------------------------------------------------------------------
// Main
// ---------------------------------------------------------------------------

func main() {
	ui.Stage.Add(ui.NewScreen(ui.WithController(&controller{})))

	ui.Setup(ui.StageConfig{
		Title:      "WillowUI -- Custom Component Demo",
		Width:      screenW,
		Height:     screenH,
		ClearColor: willow.RGBA(0.08, 0.08, 0.10, 1),
	})
}
