// TileList - reactive demo.
// A spell book browser: a ToggleButtonBar filters by school, a TileList grid
// shows matching spells with TileList.BindSelected, and a detail panel on the
// right is driven entirely by Computed and WatchEffect from the selection Ref.
// ProgressBars display power and mana cost reactively via BindValue.
package main

import (
	"fmt"

	"github.com/devthicket/willow"
	ui "github.com/devthicket/willowui"
)

const (
	screenW = 960
	screenH = 560
	tileW   = 150.0
	tileH   = 80.0
	listX   = 20.0
	listY   = 100.0
	listW   = 480.0
	listH   = 440.0
	detailX = 520.0
	detailY = 58.0
	detailW = 420.0
)

// Spell represents an ability in the spell book.
type Spell struct {
	Name   string
	School string
	Power  int // 1–100
	Cost   int // 1–100
	Desc   string
}

var allSpells = []Spell{
	{"Fireball", "Fire", 85, 60, "Launches a blazing orb that explodes on impact, hitting all nearby foes."},
	{"Flame Shield", "Fire", 40, 35, "Wraps the caster in searing flames that burn attackers on contact."},
	{"Ember Spray", "Fire", 30, 20, "A short-range cone of scattered embers. Low cost, low commitment."},
	{"Fire Nova", "Fire", 95, 80, "A massive detonation centered on the caster. Devastating but dangerous."},
	{"Blizzard", "Ice", 90, 75, "Rains shards of ice over a wide area, slowing and damaging foes."},
	{"Frost Bolt", "Ice", 55, 40, "A piercing shard of ice with good single-target punch."},
	{"Ice Wall", "Ice", 20, 30, "Raises a thick barrier of ice. Useful for blocking movement."},
	{"Cryo Pulse", "Ice", 70, 55, "A freezing shockwave that slows all enemies it passes through."},
	{"Chain Lightning", "Lightning", 80, 65, "An arc of electricity that jumps between up to four targets."},
	{"Static Field", "Lightning", 35, 25, "Electrifies the ground briefly, zapping anyone who steps on it."},
	{"Thunder Strike", "Lightning", 90, 70, "A focused bolt called down from above. High accuracy, high impact."},
	{"Ball Lightning", "Lightning", 75, 60, "A slow-moving orb that shocks nearby foes as it drifts forward."},
	{"Shadow Step", "Dark", 10, 45, "Instantly teleport behind a target. Zero damage, maximum utility."},
	{"Soul Drain", "Dark", 60, 50, "Siphons life force from an enemy, healing the caster slightly."},
	{"Void Rift", "Dark", 95, 90, "Tears open a dimensional rift dealing catastrophic damage to all inside."},
	{"Curse", "Dark", 25, 30, "Weakens an enemy's defenses for several seconds. Simple and effective."},
}

var schoolColors = map[string]willow.Color{
	"Fire":      willow.RGBA(1.00, 0.35, 0.10, 1),
	"Ice":       willow.RGBA(0.30, 0.75, 1.00, 1),
	"Lightning": willow.RGBA(1.00, 0.95, 0.20, 1),
	"Dark":      willow.RGBA(0.60, 0.20, 0.90, 1),
}

func schoolColor(s string) willow.Color {
	if c, ok := schoolColors[s]; ok {
		return c
	}
	return willow.RGBA(0.6, 0.6, 0.6, 1)
}

func filterSpells(schoolIdx int) []Spell {
	if schoolIdx == 0 {
		return allSpells
	}
	schools := []string{"", "Fire", "Ice", "Lightning", "Dark"}
	target := schools[schoolIdx]
	var out []Spell
	for _, s := range allSpells {
		if s.School == target {
			out = append(out, s)
		}
	}
	return out
}

func main() {
	font := ui.MustLoadDefaultFont()

	const (
		sizeLarge  = 22.0
		sizeMedium = 16.0
		sizeSmall  = 16.0
	)

	screen := ui.NewScreen()

	title := willow.NewText("title", "Reactive: Spell Book Browser", font)
	title.TextBlock.FontSize = sizeLarge
	title.TextBlock.Color = willow.RGBA(1, 1, 1, 1)
	title.SetPosition(24, 14)
	screen.AddNode(title)

	// ── School filter ToggleButtonBar ─────────────────────────────────────────
	schoolRef := ui.NewRef(0)

	tbb := ui.NewToggleButtonBar("school-bar", font, sizeSmall)
	tbb.AddButton("All")
	tbb.AddButton("Fire")
	tbb.AddButton("Ice")
	tbb.AddButton("Lightning")
	tbb.AddButton("Dark")
	tbb.SetSize(listW, 32)
	tbb.SetPosition(listX, 58)
	tbb.BindSelected(schoolRef)
	screen.Add(tbb)

	// ── Spell tile list ───────────────────────────────────────────────────────
	selectedRef := ui.NewRef(-1)

	// tileNodes tracks the data-driven sub-nodes for each pooled tile component
	// so SetUpdateItem can update them in place without rebuilding the node tree.
	type spellNodes struct {
		bar    *willow.Node
		name   *willow.Node
		school *willow.Node
	}
	tileNodes := make(map[*ui.Component]*spellNodes)

	tileList := ui.NewTileList("spells", tileW, tileH)
	tileList.SetSize(listW, listH)
	tileList.SetColumns(3)
	tileList.SetSelectable(true)
	tileList.SetPosition(listX, listY)
	tileList.BindSelected(selectedRef)
	screen.Add(tileList)

	tileList.SetRenderItem(func(idx int, data any) *ui.Component {
		spell := data.(Spell)
		sc := schoolColor(spell.School)

		panel := ui.NewPanel("sp")
		panel.SetSize(tileW, tileH)

		// Colored school indicator on the left edge.
		bar := willow.NewSprite("bar", willow.TextureRegion{})
		bar.SetPosition(0, 0)
		bar.SetScale(5, tileH)
		bar.SetColor(sc)
		panel.AddRawChild(bar)

		// Spell name.
		nameTxt := willow.NewText("nm", spell.Name, font)
		nameTxt.TextBlock.FontSize = 16
		nameTxt.TextBlock.Color = willow.RGBA(1, 1, 1, 1)
		nameTxt.SetPosition(12, 16)
		panel.AddRawChild(nameTxt)

		// School label (colored).
		schoolTxt := willow.NewText("sc", spell.School, font)
		schoolTxt.TextBlock.FontSize = 15
		schoolTxt.TextBlock.Color = sc
		schoolTxt.SetPosition(12, 50)
		panel.AddRawChild(schoolTxt)

		comp := &panel.Component
		tileNodes[comp] = &spellNodes{bar: bar, name: nameTxt, school: schoolTxt}
		return comp
	})

	// SetUpdateItem lets the tile list refresh existing components in place
	// instead of disposing and recreating them on every SetItems call.
	tileList.SetUpdateItem(func(idx int, data any, comp *ui.Component) {
		spell := data.(Spell)
		sc := schoolColor(spell.School)
		if n, ok := tileNodes[comp]; ok {
			n.bar.SetColor(sc)
			n.name.SetContent(spell.Name)
			n.school.SetContent(spell.School)
			n.school.SetTextColor(sc)
		}
	})

	// ── Detail panel ──────────────────────────────────────────────────────────
	var currentSpells []Spell

	detailNameLbl := ui.NewLabel("d-name", "Select a spell", font, 18)
	detailNameLbl.SetColor(willow.RGBA(1, 1, 1, 1))
	detailNameLbl.SetPosition(detailX, detailY)
	screen.Add(detailNameLbl)

	detailSchoolLbl := ui.NewLabel("d-school", "", font, sizeSmall)
	detailSchoolLbl.SetPosition(detailX, detailY+26)
	screen.Add(detailSchoolLbl)

	addDivider(screen, detailX, detailY+46, detailW)

	addFieldLabel(screen, font, sizeSmall, "Power:", detailX, detailY+58)
	powerRef := ui.NewRef(0.0)
	powerBar := ui.NewProgressBar("power-bar")
	powerBar.SetSize(detailW-70, 14)
	powerBar.SetPosition(detailX+70, detailY+60)
	powerBar.BindValue(powerRef)
	screen.Add(powerBar)

	powerValLbl := ui.NewLabel("power-val", "", font, sizeSmall)
	powerValLbl.SetColor(willow.RGBA(0.7, 0.8, 0.5, 1))
	powerValLbl.SetPosition(detailX+detailW-34, detailY+61)
	screen.Add(powerValLbl)

	addFieldLabel(screen, font, sizeSmall, "Cost:", detailX, detailY+84)
	costRef := ui.NewRef(0.0)
	costBar := ui.NewProgressBar("cost-bar")
	costBar.SetSize(detailW-70, 14)
	costBar.SetPosition(detailX+70, detailY+86)
	costBar.BindValue(costRef)
	screen.Add(costBar)

	costValLbl := ui.NewLabel("cost-val", "", font, sizeSmall)
	costValLbl.SetColor(willow.RGBA(0.7, 0.8, 0.5, 1))
	costValLbl.SetPosition(detailX+detailW-34, detailY+87)
	screen.Add(costValLbl)

	addDivider(screen, detailX, detailY+110, detailW)

	detailDescLbl := ui.NewLabel("d-desc", "", font, sizeSmall)
	detailDescLbl.SetColor(willow.RGBA(0.72, 0.78, 0.84, 1))
	detailDescLbl.SetWrapWidth(detailW)
	detailDescLbl.SetPosition(detailX, detailY+122)
	screen.Add(detailDescLbl)

	// ── Reactive wiring ───────────────────────────────────────────────────────

	// reloadSpells rebuilds the tile list for the given school index.
	// Called on initial setup and whenever the school tab changes.
	reloadSpells := func(schoolIdx int) {
		currentSpells = filterSpells(schoolIdx)
		items := make([]ui.ListItem, len(currentSpells))
		for i, s := range currentSpells {
			items[i] = ui.ListItem{Data: s}
		}
		tileList.SetItems(items)
		selectedRef.Set(-1)
	}

	// Load all spells on startup, then re-filter whenever the school tab changes.
	reloadSpells(0)
	tbb.SetOnChange(func(idx int) {
		reloadSpells(idx)
	})

	// Update detail panel when selection changes.
	ui.WatchEffect(func() {
		idx := selectedRef.Get()
		if idx < 0 || idx >= len(currentSpells) {
			detailNameLbl.SetText("Select a spell")
			detailSchoolLbl.SetText("")
			detailDescLbl.SetText("")
			powerRef.Set(0)
			costRef.Set(0)
			powerValLbl.SetText("")
			costValLbl.SetText("")
			return
		}
		s := currentSpells[idx]
		detailNameLbl.SetText(s.Name)
		detailSchoolLbl.SetText(s.School)
		detailSchoolLbl.SetColor(schoolColor(s.School))
		detailDescLbl.SetText(s.Desc)
		powerRef.Set(float64(s.Power) / 100.0)
		costRef.Set(float64(s.Cost) / 100.0)
		powerValLbl.SetText(fmt.Sprintf("%d", s.Power))
		costValLbl.SetText(fmt.Sprintf("%d", s.Cost))
	})

	ui.Stage.Add(screen)
	ui.Setup(ui.StageConfig{
		Title:      "Reactive — Spell Book Browser",
		Width:      screenW,
		Height:     screenH,
		ClearColor: willow.RGBA(0.08, 0.08, 0.10, 1),
	})
}

func addDivider(screen *ui.Screen, x, y, width float64) {
	d := willow.NewSprite("div", willow.TextureRegion{})
	d.SetPosition(x, y)
	d.SetScale(width, 1)
	d.SetColor(willow.RGBA(0.22, 0.27, 0.32, 1))
	screen.AddNode(d)
}

func addFieldLabel(screen *ui.Screen, font *willow.FontFamily, fontSize float64, text string, x, y float64) {
	n := willow.NewText("fl", text, font)
	n.TextBlock.FontSize = fontSize
	n.TextBlock.Color = willow.RGBA(0.46, 0.55, 0.65, 1)
	n.SetPosition(x, y)
	screen.AddNode(n)
}
