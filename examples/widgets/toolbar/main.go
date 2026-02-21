// ToolBar demonstrates WillowUI's ToolBar with icon buttons in three modes:
//   - Action buttons (click-once): Save, Open, Undo, Redo
//   - Toggle buttons (independent on/off): Bold, Italic, Underline
//   - Radio group (mutually exclusive): Cursor, Pencil, Brush, Wand
//
// Icons are loaded from the famfamfam Silk spritesheet.
package main

import (
	"image"
	"image/png"
	"log"
	"os"

	"github.com/devthicket/willow"
	"github.com/hajimehoshi/ebiten/v2"
	ui "github.com/devthicket/willowui"
)

const (
	screenW = 800
	screenH = 400
)

// iconDef maps an icon name to its spritesheet coordinates.
type iconDef struct {
	x, y int
}

// Icons used in this demo (coordinates from famfamfam-silk.json).
var icons = map[string]iconDef{
	"disk":           {288, 192},
	"folder":         {32, 320},
	"arrow_undo":     {32, 112},
	"arrow_redo":     {112, 0},
	"help":           {64, 336},
	"text_bold":      {480, 16},
	"text_italic":    {480, 208},
	"text_underline": {16, 480},
	"cursor":         {16, 272},
	"pencil":         {416, 240},
	"paintbrush":     {416, 160},
	"wand":           {496, 192},
}

func main() {
	font := ui.MustLoadDefaultFont()

	// Load spritesheet.
	sheet := loadSheet("examples/_assets/famfamfam-silk.png")

	const (
		sizeLarge  = 24.0
		sizeMedium = 14.0
		sizeSmall  = 13.0
		btnSize    = 32.0
	)

	screen := ui.NewScreen()

	// Title.
	title := willow.NewText("title", "WillowUI -- ToolBar Demo", font)
	title.TextBlock.FontSize = sizeLarge
	title.TextBlock.Color = willow.RGBA(1, 1, 1, 1)
	title.SetPosition(24, 16)
	screen.AddNode(title)

	div := ui.NewDivider("divider", screenW-48)
	div.SetPosition(24, 48)
	screen.AddNode(div)

	statusLabel := ui.NewLabel("status", "Click toolbar buttons to interact", font, sizeSmall)
	statusLabel.SetPosition(24, 340)
	screen.Add(statusLabel)

	// =====================================================================
	// Section 1: Action buttons (click-once)
	// =====================================================================
	addSectionLabel(screen, font, sizeSmall, "Action buttons (click-once)", 24, 58)

	actionBar := ui.NewToolBar("action-toolbar")
	actionBar.SetSize(screenW-48, 40)
	actionBar.SetPosition(24, 76)

	saveBtn := makeIconBtn("save", sheet, "disk", btnSize)
	saveBtn.SetOnClick(func() { statusLabel.SetText("Save clicked") })

	openBtn := makeIconBtn("open", sheet, "folder", btnSize)
	openBtn.SetOnClick(func() { statusLabel.SetText("Open clicked") })

	undoBtn := makeIconBtn("undo", sheet, "arrow_undo", btnSize)
	undoBtn.SetOnClick(func() { statusLabel.SetText("Undo clicked") })

	redoBtn := makeIconBtn("redo", sheet, "arrow_redo", btnSize)
	redoBtn.SetOnClick(func() { statusLabel.SetText("Redo clicked") })

	helpBtn := makeIconBtn("help", sheet, "help", btnSize)
	helpBtn.SetOnClick(func() { statusLabel.SetText("Help clicked") })

	actionBar.AddItem(saveBtn)
	actionBar.AddItem(openBtn)
	actionBar.AddSeparator()
	actionBar.AddItem(undoBtn)
	actionBar.AddItem(redoBtn)
	actionBar.AddSpacer()
	actionBar.AddItem(helpBtn)

	screen.Add(actionBar)

	// =====================================================================
	// Section 2: Toggle buttons (independent on/off)
	// =====================================================================
	addSectionLabel(screen, font, sizeSmall, "Toggle buttons (independent on/off)", 24, 126)

	toggleBar := ui.NewToolBar("toggle-toolbar")
	toggleBar.SetSize(200, 40)
	toggleBar.SetPosition(24, 144)

	boldBtn := makeIconBtn("bold", sheet, "text_bold", btnSize)
	boldBtn.SetOnClick(func() {
		boldBtn.SetActive(!boldBtn.IsActive())
		statusLabel.SetText("Bold: " + onOff(boldBtn.IsActive()))
	})

	italicBtn := makeIconBtn("italic", sheet, "text_italic", btnSize)
	italicBtn.SetOnClick(func() {
		italicBtn.SetActive(!italicBtn.IsActive())
		statusLabel.SetText("Italic: " + onOff(italicBtn.IsActive()))
	})

	underlineBtn := makeIconBtn("underline", sheet, "text_underline", btnSize)
	underlineBtn.SetOnClick(func() {
		underlineBtn.SetActive(!underlineBtn.IsActive())
		statusLabel.SetText("Underline: " + onOff(underlineBtn.IsActive()))
	})

	toggleBar.AddItem(boldBtn)
	toggleBar.AddItem(italicBtn)
	toggleBar.AddItem(underlineBtn)

	screen.Add(toggleBar)

	// =====================================================================
	// Section 3: Radio group (mutually exclusive tool selection)
	// =====================================================================
	addSectionLabel(screen, font, sizeSmall, "Radio group (mutually exclusive)", 24, 196)

	radioBar := ui.NewToolBar("radio-toolbar")
	radioBar.SetSize(220, 40)
	radioBar.SetPosition(24, 214)

	toolNames := []string{"cursor", "pencil", "paintbrush", "wand"}
	toolLabels := []string{"Cursor", "Pencil", "Brush", "Wand"}

	group := ui.NewToolGroup()
	group.SetOnChange(func(idx int) {
		if idx >= 0 && idx < len(toolLabels) {
			statusLabel.SetText("Tool: " + toolLabels[idx])
		}
	})

	for _, name := range toolNames {
		btn := makeIconBtn("tool-"+name, sheet, name, btnSize)
		group.Add(btn)
		radioBar.AddItem(btn)
	}
	group.SetSelected(0) // cursor selected by default

	screen.Add(radioBar)

	// =====================================================================
	// Section 4: Combined toolbar (mixed modes)
	// =====================================================================
	addSectionLabel(screen, font, sizeSmall, "Combined toolbar (all modes together)", 24, 266)

	comboBar := ui.NewToolBar("combo-toolbar")
	comboBar.SetSize(screenW-48, 40)
	comboBar.SetPosition(24, 284)

	// Action: save
	comboSave := makeIconBtn("c-save", sheet, "disk", btnSize)
	comboSave.SetOnClick(func() { statusLabel.SetText("Combo: Save") })
	comboBar.AddItem(comboSave)

	comboBar.AddSeparator()

	// Radio: tool selection
	comboGroup := ui.NewToolGroup()
	comboGroup.SetOnChange(func(idx int) {
		if idx >= 0 && idx < len(toolLabels) {
			statusLabel.SetText("Combo tool: " + toolLabels[idx])
		}
	})
	for _, name := range toolNames {
		btn := makeIconBtn("c-tool-"+name, sheet, name, btnSize)
		comboGroup.Add(btn)
		comboBar.AddItem(btn)
	}
	comboGroup.SetSelected(0)

	comboBar.AddSeparator()

	// Toggles: bold, italic
	comboBold := makeIconBtn("c-bold", sheet, "text_bold", btnSize)
	comboBold.SetOnClick(func() {
		comboBold.SetActive(!comboBold.IsActive())
		statusLabel.SetText("Combo bold: " + onOff(comboBold.IsActive()))
	})
	comboItalic := makeIconBtn("c-italic", sheet, "text_italic", btnSize)
	comboItalic.SetOnClick(func() {
		comboItalic.SetActive(!comboItalic.IsActive())
		statusLabel.SetText("Combo italic: " + onOff(comboItalic.IsActive()))
	})
	comboBar.AddItem(comboBold)
	comboBar.AddItem(comboItalic)

	screen.Add(comboBar)

	ui.Stage.Add(screen)
	ui.Setup(ui.StageConfig{
		Title:      "WillowUI -- ToolBar Demo",
		Width:      screenW,
		Height:     screenH,
		ClearColor: willow.RGBA(0.08, 0.08, 0.10, 1),
	})
}

// makeIconBtn creates an IconButton with an icon from the spritesheet.
func makeIconBtn(name string, sheet *ebiten.Image, iconName string, size float64) *ui.IconButton {
	btn := ui.NewIconButton(name)
	btn.SetSize(size, size)

	def := icons[iconName]
	sub := sheet.SubImage(image.Rect(def.x, def.y, def.x+16, def.y+16)).(*ebiten.Image)
	btn.SetIconImage(sub)
	btn.SetIconSize(16, 16)

	return btn
}

// loadSheet loads the spritesheet PNG and returns an ebiten.Image.
func loadSheet(path string) *ebiten.Image {
	f, err := os.Open(path)
	if err != nil {
		log.Fatalf("open spritesheet: %v", err)
	}
	defer f.Close()
	decoded, err := png.Decode(f)
	if err != nil {
		log.Fatalf("decode spritesheet: %v", err)
	}
	return ebiten.NewImageFromImage(decoded)
}

func addSectionLabel(screen *ui.Screen, font *willow.FontFamily, fontSize float64, text string, x, y float64) {
	n := willow.NewText("section", text, font)
	n.TextBlock.FontSize = fontSize
	n.TextBlock.Color = willow.RGBA(0.4, 0.5, 0.6, 1)
	n.SetPosition(x, y)
	screen.AddNode(n)
}

func onOff(v bool) string {
	if v {
		return "ON"
	}
	return "OFF"
}
