package integration

import (
	"math"
	"testing"

	ui "github.com/devthicket/willowui"
)

// ---------------------------------------------------------------------------
// Slider
// ---------------------------------------------------------------------------

func TestNewSliderDefaults(t *testing.T) {
	resetScheduler()
	s := ui.NewSlider("slider")
	defer s.Dispose()

	if s.Node() == nil {
		t.Fatal("Node() should not be nil")
	}
	if s.Value() != 0 {
		t.Errorf("initial value should be 0, got %f", s.Value())
	}
	if s.BgNode() == nil {
		t.Fatal("bgNode should not be nil")
	}
	if s.ThumbNode() == nil {
		t.Fatal("thumbComp.node should not be nil")
	}
}

func TestSliderSetValue(t *testing.T) {
	resetScheduler()
	s := ui.NewSlider("slider")
	defer s.Dispose()

	s.SetValue(0.5)
	if s.Value() != 0.5 {
		t.Errorf("Value() = %f, want 0.5", s.Value())
	}
}

func TestSliderClampValue(t *testing.T) {
	resetScheduler()
	s := ui.NewSlider("slider")
	defer s.Dispose()

	s.SetRange(0, 100)
	s.SetValue(150)
	if s.Value() != 100 {
		t.Errorf("Value() = %f, want 100 (clamped)", s.Value())
	}

	s.SetValue(-10)
	if s.Value() != 0 {
		t.Errorf("Value() = %f, want 0 (clamped)", s.Value())
	}
}

func TestSliderStepSnapping(t *testing.T) {
	resetScheduler()
	s := ui.NewSlider("slider")
	defer s.Dispose()

	s.SetRange(0, 100)
	s.SetStep(10)
	s.SetValue(23)
	if s.Value() != 20 {
		t.Errorf("Value() = %f, want 20 (snapped to step 10)", s.Value())
	}

	s.SetValue(27)
	if s.Value() != 30 {
		t.Errorf("Value() = %f, want 30 (snapped to step 10)", s.Value())
	}
}

func TestSliderOnChange(t *testing.T) {
	resetScheduler()
	s := ui.NewSlider("slider")
	defer s.Dispose()

	var got float64
	s.SetOnChange(func(v float64) { got = v })

	s.SetValue(0.7)
	if got != 0.7 {
		t.Errorf("onChange got %f, want 0.7", got)
	}
}

func TestSliderBindValue(t *testing.T) {
	resetScheduler()
	s := ui.NewSlider("slider")
	defer s.Dispose()

	ref := ui.NewRef(0.5)
	s.BindValue(ref)

	if s.Value() != 0.5 {
		t.Errorf("binding should sync initial value, got %f", s.Value())
	}

	ref.Set(0.8)
	ui.DefaultScheduler.Flush()
	if s.Value() != 0.8 {
		t.Errorf("reactive update should set value, got %f", s.Value())
	}
}

func TestSliderSetSize(t *testing.T) {
	resetScheduler()
	s := ui.NewSlider("slider")
	defer s.Dispose()

	s.SetSize(300, 20)
	if s.Width != 300 || s.Height != 20 {
		t.Errorf("size = %fx%f, want 300x20", s.Width, s.Height)
	}
}

func TestSliderVerticalOrientation(t *testing.T) {
	resetScheduler()
	s := ui.NewSlider("slider")
	defer s.Dispose()

	s.SetOrientation(ui.Vertical)
	if s.GetOrientation() != ui.Vertical {
		t.Error("orientation should be Vertical")
	}
}

// ---------------------------------------------------------------------------
// ScrollBar
// ---------------------------------------------------------------------------

func TestNewScrollBarDefaults(t *testing.T) {
	resetScheduler()
	sb := ui.NewScrollBar("scrollbar")
	defer sb.Dispose()

	if sb.Node() == nil {
		t.Fatal("Node() should not be nil")
	}
	if sb.BgNode() == nil {
		t.Fatal("bgNode should not be nil")
	}
	if sb.ThumbNode() == nil {
		t.Fatal("thumbComp.node should not be nil")
	}
}

func TestScrollBarSetContentSize(t *testing.T) {
	resetScheduler()
	sb := ui.NewScrollBar("scrollbar")
	defer sb.Dispose()

	sb.SetSize(20, 200)
	sb.SetContentSize(1000, 200)

	// Thumb height should be proportional: (200/1000) * 200 = 40
	expectedH := (200.0 / 1000.0) * 200.0
	if math.Abs(sb.ThumbHeight()-expectedH) > 1 {
		t.Errorf("thumb height = %f, want ~%f", sb.ThumbHeight(), expectedH)
	}
}

func TestScrollBarThumbMinSize(t *testing.T) {
	resetScheduler()
	sb := ui.NewScrollBar("scrollbar")
	defer sb.Dispose()

	sb.SetSize(20, 200)
	sb.SetContentSize(100000, 200)

	// With huge content, thumb should be at minimum size (20).
	if sb.ThumbHeight() < 20 {
		t.Errorf("thumb height = %f, should be at least 20", sb.ThumbHeight())
	}
}

func TestScrollBarBindScrollPos(t *testing.T) {
	resetScheduler()
	sb := ui.NewScrollBar("scrollbar")
	defer sb.Dispose()

	sb.SetSize(20, 200)
	sb.SetContentSize(500, 200)

	ref := ui.NewRef(100.0)
	sb.BindScrollPos(ref)

	if sb.ScrollPos() != 100 {
		t.Errorf("binding should sync initial value, got %f", sb.ScrollPos())
	}

	ref.Set(200.0)
	ui.DefaultScheduler.Flush()
	if sb.ScrollPos() != 200 {
		t.Errorf("reactive update should set scroll pos, got %f", sb.ScrollPos())
	}
}

func TestScrollBarSetScrollPos(t *testing.T) {
	resetScheduler()
	sb := ui.NewScrollBar("scrollbar")
	defer sb.Dispose()

	sb.SetSize(20, 200)
	sb.SetContentSize(500, 200)

	sb.SetScrollPos(100)
	if sb.ScrollPos() != 100 {
		t.Errorf("ScrollPos() = %f, want 100", sb.ScrollPos())
	}

	// Clamp above max.
	sb.SetScrollPos(400)
	if sb.ScrollPos() != 300 {
		t.Errorf("ScrollPos() = %f, want 300 (clamped)", sb.ScrollPos())
	}

	// Clamp below 0.
	sb.SetScrollPos(-10)
	if sb.ScrollPos() != 0 {
		t.Errorf("ScrollPos() = %f, want 0 (clamped)", sb.ScrollPos())
	}
}

// ---------------------------------------------------------------------------
// MeterBar (ProgressBar alias)
// ---------------------------------------------------------------------------

func TestNewProgressBarDefaults(t *testing.T) {
	resetScheduler()
	pb := ui.NewProgressBar("progress")
	defer pb.Dispose()

	if pb.Node() == nil {
		t.Fatal("Node() should not be nil")
	}
	if pb.Value() != 0 {
		t.Errorf("initial value should be 0, got %f", pb.Value())
	}
	if pb.BgNode() == nil {
		t.Fatal("bgNode should not be nil")
	}
	if pb.FillNode() == nil {
		t.Fatal("fillComp.node should not be nil")
	}
}

func TestProgressBarSetValue(t *testing.T) {
	resetScheduler()
	pb := ui.NewProgressBar("progress")
	defer pb.Dispose()

	pb.SetSize(200, 20)
	pb.SetValue(0.5)
	if pb.Value() != 0.5 {
		t.Errorf("Value() = %f, want 0.5", pb.Value())
	}

	// Fill width should be 50% of total width.
	if math.Abs(pb.FillWidth()-100) > 1 {
		t.Errorf("fill width = %f, want ~100", pb.FillWidth())
	}
}

func TestProgressBarClampValue(t *testing.T) {
	resetScheduler()
	pb := ui.NewProgressBar("progress")
	defer pb.Dispose()

	pb.SetValue(1.5)
	if pb.Value() != 1.0 {
		t.Errorf("Value() = %f, want 1.0 (clamped)", pb.Value())
	}

	pb.SetValue(-0.5)
	if pb.Value() != 0 {
		t.Errorf("Value() = %f, want 0 (clamped)", pb.Value())
	}
}

func TestProgressBarBindValue(t *testing.T) {
	resetScheduler()
	pb := ui.NewProgressBar("progress")
	defer pb.Dispose()

	ref := ui.NewRef(0.3)
	pb.BindValue(ref)

	if pb.Value() != 0.3 {
		t.Errorf("binding should sync initial value, got %f", pb.Value())
	}

	ref.Set(0.7)
	ui.DefaultScheduler.Flush()
	if pb.Value() != 0.7 {
		t.Errorf("reactive update should set value, got %f", pb.Value())
	}
}

func TestProgressBarSetSize(t *testing.T) {
	resetScheduler()
	pb := ui.NewProgressBar("progress")
	defer pb.Dispose()

	pb.SetSize(300, 25)
	if pb.Width != 300 || pb.Height != 25 {
		t.Errorf("size = %fx%f, want 300x25", pb.Width, pb.Height)
	}
}

func TestProgressBarShowLabel(t *testing.T) {
	resetScheduler()
	font := newTestFont()
	pb := ui.NewProgressBar("progress")
	defer pb.Dispose()

	pb.SetShowLabel(true, font, 12)
	if pb.LabelComp() == nil {
		t.Fatal("label should not be nil after SetShowLabel(true)")
	}

	pb.SetShowLabel(false, nil, 0)
	// label node should be hidden.
}

// ---------------------------------------------------------------------------
// MeterBar (range mode)
// ---------------------------------------------------------------------------

func TestNewMeterBarDefaults(t *testing.T) {
	resetScheduler()
	mb := ui.NewMeterBar("meter")
	defer mb.Dispose()

	if mb.Node() == nil {
		t.Fatal("Node() should not be nil")
	}
}

func TestMeterBarSetRange(t *testing.T) {
	resetScheduler()
	mb := ui.NewMeterBar("meter")
	defer mb.Dispose()

	mb.SetSize(200, 20)
	mb.SetRange(0, 200)
	mb.SetValue(100)

	// Should map to 0.5 internally.
	if math.Abs(mb.Value()-0.5) > 0.001 {
		t.Errorf("Value() = %f, want 0.5", mb.Value())
	}
}

func TestMeterBarSetRangeEdgeCases(t *testing.T) {
	resetScheduler()
	mb := ui.NewMeterBar("meter")
	defer mb.Dispose()

	mb.SetRange(10, 10) // min == max
	mb.SetValue(10)
	if mb.Value() != 0 {
		t.Errorf("Value() = %f, want 0 when min==max", mb.Value())
	}
}

func TestMeterBarClampRawValue(t *testing.T) {
	resetScheduler()
	mb := ui.NewMeterBar("meter")
	defer mb.Dispose()

	mb.SetRange(0, 100)
	mb.SetValue(150)
	if math.Abs(mb.Value()-1.0) > 0.001 {
		t.Errorf("Value() = %f, want 1.0 (clamped)", mb.Value())
	}
}
