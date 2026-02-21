// Conditional Watch — reactive demo.
// A dual-server monitor that showcases WatchEffect's dynamic dependency
// tracking. Both servers update every frame via a simulation controller.
// The "Active Monitor" panel uses a single WatchEffect that reads from
// Server A or Server B depending on the selection: when Server A is active
// it subscribes to A's refs; switch to Server B and the watch automatically
// drops A's refs and picks up B's. No manual unsubscribe needed — the
// scheduler manages subscription bookkeeping on every re-run.
package main

import (
	"fmt"
	"math"

	"github.com/devthicket/willow"
	ui "github.com/devthicket/willowui"
)

const (
	screenW  = 800
	screenH  = 480
	colLeft  = 40.0
	colRight = 450.0
)

type serverRefs struct {
	cpu *ui.Ref[float64]
	mem *ui.Ref[float64]
	net *ui.Ref[float64]
}

type simController struct {
	a, b  serverRefs
	frame int
}

func (c *simController) OnCreate(s *ui.Screen) {}
func (c *simController) OnDestroy()            {}

func (c *simController) OnUpdate(_ float64) {
	c.frame++
	f := float64(c.frame)

	// Server A: low-load machine — all stats stay in the 0–50% range, slow waves.
	c.a.cpu.Set(clamp01(0.25 + 0.20*math.Sin(f*0.022)))
	c.a.mem.Set(clamp01(0.30 + 0.12*math.Sin(f*0.009)))
	c.a.net.Set(clamp01(math.Abs(math.Sin(f*0.038)) * 0.48))

	// Server B: overloaded machine — all stats stay in the 60–100% range, fast churn.
	c.b.cpu.Set(clamp01(0.80 + 0.19*math.Sin(f*0.14)))
	c.b.mem.Set(clamp01(0.78 + 0.18*math.Sin(f*0.11+1.0)))
	c.b.net.Set(clamp01(0.62 + math.Abs(math.Sin(f*0.18+2.0))*0.37))
}

func clamp01(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

func main() {
	font := ui.MustLoadDefaultFont()

	const (
		sizeLarge = 26.0
		sizeSmall = 16.0
	)

	// ── Reactive state ────────────────────────────────────────────────────────
	srvA := serverRefs{
		cpu: ui.NewRef(0.50),
		mem: ui.NewRef(0.50),
		net: ui.NewRef(0.30),
	}
	srvB := serverRefs{
		cpu: ui.NewRef(0.72),
		mem: ui.NewRef(0.68),
		net: ui.NewRef(0.45),
	}

	// activeRef drives the conditional dependency switch.
	activeRef := ui.NewRef(0) // 0 = Server A, 1 = Server B

	ctrl := &simController{a: srvA, b: srvB}
	screen := ui.NewScreen(ui.WithController(ctrl))

	// ── Title ─────────────────────────────────────────────────────────────────
	title := ui.NewLabel("title", "Reactive: Conditional Watch", font, sizeLarge)
	title.SetColor(willow.RGBA(1, 1, 1, 1))
	title.SetPosition(24, 14)
	screen.Add(title)

	div := ui.NewDivider("div-title", screenW)
	div.SetPosition(0, 46)
	screen.AddNode(div)

	// ── Left: both servers live feed (always updating) ────────────────────────
	addSection(screen, font, sizeSmall, "ALL SERVERS  (both always live)", colLeft, 62)
	addServerBlock(screen, font, sizeSmall, "Server A", colLeft, 84, srvA)
	addServerBlock(screen, font, sizeSmall, "Server B", colLeft, 210, srvB)

	divSel := ui.NewDivider("div-sel", 380)
	divSel.SetPosition(colLeft, 340)
	screen.AddNode(divSel)

	addSection(screen, font, sizeSmall, "SELECT SERVER", colLeft, 354)

	tbb := ui.NewToggleButtonBar("srv-sel", font, sizeSmall)
	tbb.AddButton("Server A")
	tbb.AddButton("Server B")
	tbb.BindSelected(activeRef)
	tbb.SetSize(220, 32)
	tbb.SetPosition(colLeft, 374)
	screen.Add(tbb)

	// ── Right: active monitor ─────────────────────────────────────────────────
	addSection(screen, font, sizeSmall, "ACTIVE MONITOR", colRight, 62)

	activeNameLbl := ui.NewLabel("act-name", "Server A", font, 20.0)
	activeNameLbl.SetColor(willow.RGBA(0.50, 1.0, 0.60, 1))
	activeNameLbl.SetPosition(colRight, 84)
	screen.Add(activeNameLbl)

	// nil ref = no binding; the WatchEffect below drives these directly.
	actCpuBar, actCpuLbl := addStatRow(screen, font, sizeSmall, "act-cpu", "CPU", colRight, 112, nil)
	actMemBar, actMemLbl := addStatRow(screen, font, sizeSmall, "act-mem", "MEM", colRight, 136, nil)
	actNetBar, actNetLbl := addStatRow(screen, font, sizeSmall, "act-net", "NET", colRight, 160, nil)

	divRight := ui.NewDivider("div-right", 330)
	divRight.SetPosition(colRight, 192)
	screen.AddNode(divRight)

	addSection(screen, font, sizeSmall, "WATCH DEPENDENCIES", colRight, 206)

	depsLbl := ui.NewLabel("deps", "Subscribed to:  A.cpu   A.mem   A.net", font, sizeSmall)
	depsLbl.SetColor(willow.RGBA(0.70, 0.60, 1.0, 1))
	depsLbl.SetPosition(colRight, 228)
	screen.Add(depsLbl)

	noteLbl := ui.NewLabel("note",
		"Switch server → watch re-runs, drops the\nprevious server's refs, and subscribes to\nthe new server's refs automatically.",
		font, sizeSmall)
	noteLbl.SetColor(willow.RGBA(0.38, 0.44, 0.50, 1))
	noteLbl.SetPosition(colRight, 254)
	screen.Add(noteLbl)

	// ── Conditional WatchEffect: core demonstration ───────────────────────────
	// This WatchEffect's dependency set changes at runtime. On each run it
	// reads activeRef (always tracked). If srv==0 it reads srvA.{cpu,mem,net};
	// if srv==1 it reads srvB.{cpu,mem,net}. The scheduler automatically
	// unsubscribes from the previous branch's refs between runs.
	//
	// Bars and labels are set directly — no proxy Ref in between — so the
	// active monitor is always in sync with the left panel in the same flush.
	ui.WatchEffect(func() {
		var src serverRefs
		if activeRef.Get() == 0 { // always tracked — triggers re-run on switch
			src = srvA // srvA.* tracked only while activeRef == 0
		} else {
			src = srvB // srvB.* tracked only while activeRef == 1
		}
		cpu, mem, net := src.cpu.Get(), src.mem.Get(), src.net.Get()
		actCpuBar.SetValue(cpu)
		actCpuLbl.SetText(fmt.Sprintf("%d%%", int(cpu*100)))
		actMemBar.SetValue(mem)
		actMemLbl.SetText(fmt.Sprintf("%d%%", int(mem*100)))
		actNetBar.SetValue(net)
		actNetLbl.SetText(fmt.Sprintf("%d%%", int(net*100)))
	})

	// Update name label and dependency readout whenever the selection changes.
	ui.WatchValue(activeRef, func(_, srv int) {
		names := [2]string{"Server A", "Server B"}
		deps := [2]string{
			"Subscribed to:  A.cpu   A.mem   A.net",
			"Subscribed to:  B.cpu   B.mem   B.net",
		}
		activeNameLbl.SetText(names[srv])
		depsLbl.SetText(deps[srv])
	})

	ui.Stage.Add(screen)
	ui.Setup(ui.StageConfig{
		Title:      "Reactive — Conditional Watch",
		Width:      screenW,
		Height:     screenH,
		ClearColor: willow.RGBA(0.08, 0.08, 0.10, 1),
	})
}

func addSection(screen *ui.Screen, font *willow.FontFamily, size float64, text string, x, y float64) {
	l := ui.NewSectionLabel("sec", text, font, size)
	l.SetPosition(x, y)
	screen.Add(l)
}

// addServerBlock renders a named server panel with CPU, MEM, and NET stat rows.
func addServerBlock(screen *ui.Screen, font *willow.FontFamily, fontSize float64, name string, x, y float64, refs serverRefs) {
	hdr := ui.NewLabel("srv-hdr", name, font, fontSize)
	hdr.SetColor(willow.RGBA(0.85, 0.92, 0.98, 1))
	hdr.SetPosition(x, y)
	screen.Add(hdr)

	prefix := name[len(name)-1:] // "A" or "B"
	addStatRow(screen, font, fontSize, prefix+"-cpu", "CPU", x, y+20, refs.cpu)
	addStatRow(screen, font, fontSize, prefix+"-mem", "MEM", x, y+44, refs.mem)
	addStatRow(screen, font, fontSize, prefix+"-net", "NET", x, y+68, refs.net)
}

// addStatRow renders a metric row: stat label, progress bar, and live percentage.
// If ref is non-nil the bar binds to ref and a WatchEffect keeps the value label
// in sync automatically. Pass nil to drive the bar and label from a WatchEffect
// in the caller (avoids a proxy-ref hop when combined with conditional watching).
func addStatRow(screen *ui.Screen, font *willow.FontFamily, fontSize float64, id, stat string, x, y float64, ref *ui.Ref[float64]) (*ui.ProgressBar, *ui.Label) {
	statLbl := ui.NewSectionLabel("lbl-"+id, stat, font, fontSize)
	statLbl.SetPosition(x, y)
	screen.Add(statLbl)

	bar := ui.NewProgressBar("bar-" + id)
	bar.SetSize(160, 10)
	bar.SetPosition(x+36, y+3)
	screen.Add(bar)

	valLbl := ui.NewLabel("val-"+id, "0%", font, fontSize)
	valLbl.SetColor(willow.RGBA(0.85, 0.92, 0.98, 1))
	valLbl.SetPosition(x+204, y)
	screen.Add(valLbl)

	if ref != nil {
		bar.BindValue(ref)
		ui.WatchEffect(func() {
			valLbl.SetText(fmt.Sprintf("%d%%", int(ref.Get()*100)))
		})
	}

	return bar, valLbl
}
