package integration

import (
	"testing"

	"github.com/devthicket/willow"
	ui "github.com/devthicket/willowui"
)

// --- Construction ---

func TestNewDragHandleCreates(t *testing.T) {
	resetScheduler()
	dh := ui.NewDragHandle("dh")
	defer dh.Dispose()

	if dh.Node() == nil {
		t.Fatal("Node() should not be nil")
	}
	if dh.Name() != "dh" {
		t.Errorf("Name() = %q, want %q", dh.Name(), "dh")
	}
}

func TestDragHandleDefaultGripDots(t *testing.T) {
	resetScheduler()
	dh := ui.NewDragHandle("dh")
	defer dh.Dispose()

	if dh.GripStyle() != ui.DragGripDots {
		t.Errorf("default grip style = %d, want DragGripDots", dh.GripStyle())
	}
	if len(dh.GripNodes()) == 0 {
		t.Error("default grip should have dot nodes")
	}
}

func TestDragHandleGripNoneHasNoNodes(t *testing.T) {
	resetScheduler()
	dh := ui.NewDragHandle("dh")
	defer dh.Dispose()

	dh.SetGripStyle(ui.DragGripNone)
	if len(dh.GripNodes()) != 0 {
		t.Errorf("GripNone should produce 0 grip nodes, got %d", len(dh.GripNodes()))
	}
}

func TestDragHandleGripLines(t *testing.T) {
	resetScheduler()
	dh := ui.NewDragHandle("dh")
	defer dh.Dispose()

	dh.SetGripStyle(ui.DragGripLines)
	if len(dh.GripNodes()) == 0 {
		t.Error("GripLines should have line nodes")
	}
}

// --- Axis ---

func TestDragHandleSetAxis(t *testing.T) {
	resetScheduler()
	dh := ui.NewDragHandle("dh")
	defer dh.Dispose()

	dh.SetAxis(ui.DragAxisX)
	if dh.Axis() != ui.DragAxisX {
		t.Errorf("Axis() = %d, want DragAxisX", dh.Axis())
	}
	dh.SetAxis(ui.DragAxisDiagonal)
	if dh.Axis() != ui.DragAxisDiagonal {
		t.Errorf("Axis() = %d, want DragAxisDiagonal", dh.Axis())
	}
}

// --- Delegate mode (no target) ---

func TestDragHandleDelegateModeCallbacks(t *testing.T) {
	resetScheduler()
	dh := ui.NewDragHandle("dh")
	defer dh.Dispose()
	dh.SetAxis(ui.DragAxisY)
	dh.SetSize(60, 8)

	startCalled := false
	var lastDelta float64
	var endValue float64

	dh.SetOnDragStart(func() { startCalled = true })
	dh.SetOnDrag(func(delta float64) { lastDelta = delta })
	dh.SetOnDragEnd(func(v float64) { endValue = v })

	// Simulate drag via node callbacks.
	node := dh.Node()
	node.GetOnDragStart()(willow.DragContext{
		Node:    node,
		GlobalX: 100, GlobalY: 200,
		StartX: 100, StartY: 200,
	})

	if !startCalled {
		t.Error("OnDragStart should have fired")
	}

	node.GetOnDrag()(willow.DragContext{
		Node:    node,
		GlobalX: 100, GlobalY: 250,
		StartX: 100, StartY: 200,
	})

	if lastDelta != 50 {
		t.Errorf("OnDrag delta = %f, want 50", lastDelta)
	}

	node.GetOnDragEnd()(willow.DragContext{
		Node:    node,
		GlobalX: 100, GlobalY: 250,
		StartX: 100, StartY: 200,
	})

	if endValue != 50 {
		t.Errorf("OnDragEnd value = %f, want 50", endValue)
	}
}

// --- Resize mode ---

func TestDragHandleResizesTargetY(t *testing.T) {
	resetScheduler()
	panel := ui.NewPanel("panel")
	panel.SetSize(200, 100)
	defer panel.Dispose()

	dh := ui.NewDragHandle("dh")
	defer dh.Dispose()
	dh.SetAxis(ui.DragAxisY)
	dh.SetTarget(&panel.Component)

	node := dh.Node()
	node.GetOnDragStart()(willow.DragContext{
		Node: node, GlobalX: 100, GlobalY: 100, StartX: 100, StartY: 100,
	})
	node.GetOnDrag()(willow.DragContext{
		Node: node, GlobalX: 100, GlobalY: 160, StartX: 100, StartY: 100,
	})

	if panel.Height != 160 {
		t.Errorf("panel.Height = %f, want 160", panel.Height)
	}
	if panel.Width != 200 {
		t.Errorf("panel.Width should remain 200, got %f", panel.Width)
	}
}

func TestDragHandleResizesTargetX(t *testing.T) {
	resetScheduler()
	panel := ui.NewPanel("panel")
	panel.SetSize(200, 100)
	defer panel.Dispose()

	dh := ui.NewDragHandle("dh")
	defer dh.Dispose()
	dh.SetAxis(ui.DragAxisX)
	dh.SetTarget(&panel.Component)

	node := dh.Node()
	node.GetOnDragStart()(willow.DragContext{
		Node: node, GlobalX: 200, GlobalY: 50, StartX: 200, StartY: 50,
	})
	node.GetOnDrag()(willow.DragContext{
		Node: node, GlobalX: 300, GlobalY: 50, StartX: 200, StartY: 50,
	})

	if panel.Width != 300 {
		t.Errorf("panel.Width = %f, want 300", panel.Width)
	}
	if panel.Height != 100 {
		t.Errorf("panel.Height should remain 100, got %f", panel.Height)
	}
}

func TestDragHandleResizesDiagonal(t *testing.T) {
	resetScheduler()
	panel := ui.NewPanel("panel")
	panel.SetSize(200, 100)
	defer panel.Dispose()

	dh := ui.NewDragHandle("dh")
	defer dh.Dispose()
	dh.SetAxis(ui.DragAxisDiagonal)
	dh.SetTarget(&panel.Component)

	node := dh.Node()
	node.GetOnDragStart()(willow.DragContext{
		Node: node, GlobalX: 200, GlobalY: 100, StartX: 200, StartY: 100,
	})
	node.GetOnDrag()(willow.DragContext{
		Node: node, GlobalX: 250, GlobalY: 130, StartX: 200, StartY: 100,
	})

	if panel.Width != 250 {
		t.Errorf("panel.Width = %f, want 250", panel.Width)
	}
	if panel.Height != 130 {
		t.Errorf("panel.Height = %f, want 130", panel.Height)
	}
}

// --- Clamping ---

func TestDragHandleClampMin(t *testing.T) {
	resetScheduler()
	panel := ui.NewPanel("panel")
	panel.SetSize(200, 100)
	defer panel.Dispose()

	dh := ui.NewDragHandle("dh")
	defer dh.Dispose()
	dh.SetAxis(ui.DragAxisY)
	dh.SetTarget(&panel.Component)
	dh.SetMin(50)

	node := dh.Node()
	node.GetOnDragStart()(willow.DragContext{
		Node: node, GlobalX: 100, GlobalY: 100, StartX: 100, StartY: 100,
	})
	// Drag upward to shrink below min.
	node.GetOnDrag()(willow.DragContext{
		Node: node, GlobalX: 100, GlobalY: 30, StartX: 100, StartY: 100,
	})

	if panel.Height != 50 {
		t.Errorf("panel.Height = %f, want 50 (clamped to min)", panel.Height)
	}
}

func TestDragHandleClampMax(t *testing.T) {
	resetScheduler()
	panel := ui.NewPanel("panel")
	panel.SetSize(200, 100)
	defer panel.Dispose()

	dh := ui.NewDragHandle("dh")
	defer dh.Dispose()
	dh.SetAxis(ui.DragAxisY)
	dh.SetTarget(&panel.Component)
	dh.SetMax(300)

	node := dh.Node()
	node.GetOnDragStart()(willow.DragContext{
		Node: node, GlobalX: 100, GlobalY: 100, StartX: 100, StartY: 100,
	})
	node.GetOnDrag()(willow.DragContext{
		Node: node, GlobalX: 100, GlobalY: 500, StartX: 100, StartY: 100,
	})

	if panel.Height != 300 {
		t.Errorf("panel.Height = %f, want 300 (clamped to max)", panel.Height)
	}
}

func TestDragHandleClampDecreasing(t *testing.T) {
	resetScheduler()
	panel := ui.NewPanel("panel")
	panel.SetSize(200, 200)
	defer panel.Dispose()

	dh := ui.NewDragHandle("dh")
	defer dh.Dispose()
	dh.SetAxis(ui.DragAxisY)
	dh.SetTarget(&panel.Component)
	dh.SetMin(80)
	dh.SetMax(500)

	node := dh.Node()
	node.GetOnDragStart()(willow.DragContext{
		Node: node, GlobalX: 100, GlobalY: 200, StartX: 100, StartY: 200,
	})
	// Drag up to decrease.
	node.GetOnDrag()(willow.DragContext{
		Node: node, GlobalX: 100, GlobalY: 10, StartX: 100, StartY: 200,
	})

	if panel.Height != 80 {
		t.Errorf("panel.Height = %f, want 80 (clamped to min)", panel.Height)
	}
}

// --- OnDragEnd final value in resize mode ---

func TestDragHandleOnDragEndReportsFinalValue(t *testing.T) {
	resetScheduler()
	panel := ui.NewPanel("panel")
	panel.SetSize(200, 100)
	defer panel.Dispose()

	dh := ui.NewDragHandle("dh")
	defer dh.Dispose()
	dh.SetAxis(ui.DragAxisY)
	dh.SetTarget(&panel.Component)
	dh.SetMin(50)
	dh.SetMax(300)

	var finalVal float64
	dh.SetOnDragEnd(func(v float64) { finalVal = v })

	node := dh.Node()
	node.GetOnDragStart()(willow.DragContext{
		Node: node, GlobalX: 100, GlobalY: 100, StartX: 100, StartY: 100,
	})
	node.GetOnDrag()(willow.DragContext{
		Node: node, GlobalX: 100, GlobalY: 250, StartX: 100, StartY: 100,
	})
	node.GetOnDragEnd()(willow.DragContext{
		Node: node, GlobalX: 100, GlobalY: 250, StartX: 100, StartY: 100,
	})

	if finalVal != 250 {
		t.Errorf("OnDragEnd value = %f, want 250", finalVal)
	}
}

// --- XML ---

func TestDragHandleXMLInstantiation(t *testing.T) {
	resetScheduler()
	xmlData := []byte(`<DragHandle name="dh" axis="y" min="50" max="400" gripStyle="lines" width="200" height="8"/>`)
	reg := ui.NewTemplateRegistry()
	reg.SetFonts(nil, newTestFont())
	reg.SetFontSize(16)
	if err := reg.RegisterXML("test", xmlData); err != nil {
		t.Fatalf("RegisterXML: %v", err)
	}
	comp, err := reg.Instantiate("test", nil, nil)
	if err != nil {
		t.Fatalf("Instantiate: %v", err)
	}
	dh, ok := comp.UserData().(*ui.DragHandle)
	if !ok {
		t.Fatal("UserData should be *DragHandle")
	}
	if dh.Axis() != ui.DragAxisY {
		t.Errorf("Axis() = %d, want DragAxisY", dh.Axis())
	}
	if dh.Min() != 50 {
		t.Errorf("Min() = %f, want 50", dh.Min())
	}
	if dh.Max() != 400 {
		t.Errorf("Max() = %f, want 400", dh.Max())
	}
	if dh.GripStyle() != ui.DragGripLines {
		t.Errorf("GripStyle() = %d, want DragGripLines", dh.GripStyle())
	}
}

// --- Theme ---

func TestDragHandleThemeApplied(t *testing.T) {
	resetScheduler()
	dh := ui.NewDragHandle("dh")
	defer dh.Dispose()

	// Just verify UpdateVisuals doesn't panic and grip nodes exist.
	dh.UpdateVisuals()
	if len(dh.GripNodes()) == 0 {
		t.Error("grip nodes should exist after UpdateVisuals")
	}
}
