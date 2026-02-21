package theme

import (
	"math"
	"testing"

	"github.com/devthicket/willowui/internal/core"
	"github.com/devthicket/willowui/internal/sg"
)

// ---------------------------------------------------------------------------
// ColorProperty.Resolve
// ---------------------------------------------------------------------------

func TestColorProperty_Resolve(t *testing.T) {
	red := sg.RGBA(1, 0, 0, 1)
	blue := sg.RGBA(0, 0, 1, 1)

	var p ColorProperty
	p[core.StateDefault] = red
	p[core.StateHover] = blue

	if got := p.Resolve(core.StateDefault); got != red {
		t.Errorf("Resolve(Default) = %v, want %v", got, red)
	}
	if got := p.Resolve(core.StateHover); got != blue {
		t.Errorf("Resolve(Hover) = %v, want %v", got, blue)
	}
	// Unset state returns zero color.
	if got := p.Resolve(core.StateDisabled); got != (sg.Color{}) {
		t.Errorf("Resolve(Disabled) = %v, want zero", got)
	}
}

// ---------------------------------------------------------------------------
// BackgroundProperty.Resolve
// ---------------------------------------------------------------------------

func TestBackgroundProperty_Resolve(t *testing.T) {
	c := sg.RGBA(0.5, 0.5, 0.5, 1)
	bg := core.SolidBackground(c)

	var p BackgroundProperty
	p[core.StateDefault] = bg

	got := p.Resolve(core.StateDefault)
	if got.Type != core.BgSolid {
		t.Errorf("Resolve(Default).Type = %v, want BgSolid", got.Type)
	}
	if got.Color != c {
		t.Errorf("Resolve(Default).Color = %v, want %v", got.Color, c)
	}

	// Unset state returns BgNone.
	got2 := p.Resolve(core.StateHover)
	if got2.Type != core.BgNone {
		t.Errorf("Resolve(Hover).Type = %v, want BgNone", got2.Type)
	}
}

// ---------------------------------------------------------------------------
// FloatProperty.Resolve
// ---------------------------------------------------------------------------

func TestFloatProperty_Resolve(t *testing.T) {
	var p FloatProperty
	p[core.StateDefault] = 1.5
	p[core.StateHover] = 2.5

	if got := p.Resolve(core.StateDefault); got != 1.5 {
		t.Errorf("Resolve(Default) = %v, want 1.5", got)
	}
	if got := p.Resolve(core.StateHover); got != 2.5 {
		t.Errorf("Resolve(Hover) = %v, want 2.5", got)
	}
	if got := p.Resolve(core.StateDisabled); got != 0 {
		t.Errorf("Resolve(Disabled) = %v, want 0", got)
	}
}

// ---------------------------------------------------------------------------
// NewColorPropUniform
// ---------------------------------------------------------------------------

func TestNewColorPropUniform(t *testing.T) {
	c := sg.RGBA(0.2, 0.4, 0.6, 1)
	p := NewColorPropUniform(c)

	for s := core.ComponentState(0); s < core.StateCount; s++ {
		if p[s] != c {
			t.Errorf("state %d = %v, want %v", s, p[s], c)
		}
	}
}

// ---------------------------------------------------------------------------
// NewColorPropStates + fallback resolution
// ---------------------------------------------------------------------------

func TestNewColorPropStates_Fallbacks(t *testing.T) {
	red := sg.RGBA(1, 0, 0, 1)
	blue := sg.RGBA(0, 0, 1, 1)

	p := NewColorPropStates(map[core.ComponentState]sg.Color{
		core.StateDefault: red,
		core.StateHover:   blue,
	})

	// Default is set explicitly.
	if p[core.StateDefault] != red {
		t.Errorf("Default = %v, want red", p[core.StateDefault])
	}
	// Hover is set explicitly.
	if p[core.StateHover] != blue {
		t.Errorf("Hover = %v, want blue", p[core.StateHover])
	}
	// Active falls back: Active -> Hover -> Default; Hover is set -> blue.
	if p[core.StateActive] != blue {
		t.Errorf("Active = %v, want blue (via Hover fallback)", p[core.StateActive])
	}
	// Disabled falls back to Default -> red.
	if p[core.StateDisabled] != red {
		t.Errorf("Disabled = %v, want red (via Default fallback)", p[core.StateDisabled])
	}
}

func TestNewColorPropStates_OnlyDefault(t *testing.T) {
	green := sg.RGBA(0, 1, 0, 1)
	p := NewColorPropStates(map[core.ComponentState]sg.Color{
		core.StateDefault: green,
	})

	// Every state should fall back to default.
	for s := core.ComponentState(0); s < core.StateCount; s++ {
		if p[s] != green {
			t.Errorf("state %d = %v, want green", s, p[s])
		}
	}
}

func TestNewColorPropStates_Empty(t *testing.T) {
	p := NewColorPropStates(map[core.ComponentState]sg.Color{})

	// All states remain zero since nothing was set and no fallback chain
	// resolves to a non-zero color.
	zero := sg.Color{}
	for s := core.ComponentState(0); s < core.StateCount; s++ {
		if p[s] != zero {
			t.Errorf("state %d = %v, want zero", s, p[s])
		}
	}
}

// ---------------------------------------------------------------------------
// ResolveColorFallbacks explicit
// ---------------------------------------------------------------------------

func TestResolveColorFallbacks(t *testing.T) {
	white := sg.RGBA(1, 1, 1, 1)
	var p ColorProperty
	p[core.StateDefault] = white
	// All others are zero.

	ResolveColorFallbacks(&p)

	// All states should resolve to white via fallback chains.
	for s := core.ComponentState(0); s < core.StateCount; s++ {
		if p[s] != white {
			t.Errorf("after fallback, state %d = %v, want white", s, p[s])
		}
	}
}

// ---------------------------------------------------------------------------
// NewSolidBgPropUniform
// ---------------------------------------------------------------------------

func TestNewSolidBgPropUniform(t *testing.T) {
	c := sg.RGBA(0.1, 0.2, 0.3, 1)
	p := NewSolidBgPropUniform(c)

	for s := core.ComponentState(0); s < core.StateCount; s++ {
		if p[s].Type != core.BgSolid {
			t.Errorf("state %d Type = %v, want BgSolid", s, p[s].Type)
		}
		if p[s].Color != c {
			t.Errorf("state %d Color = %v, want %v", s, p[s].Color, c)
		}
	}
}

// ---------------------------------------------------------------------------
// NewSolidBgPropStates + fallback
// ---------------------------------------------------------------------------

func TestNewSolidBgPropStates_Fallbacks(t *testing.T) {
	red := sg.RGBA(1, 0, 0, 1)
	p := NewSolidBgPropStates(map[core.ComponentState]sg.Color{
		core.StateDefault: red,
	})

	// All states should fall back to default's solid background.
	for s := core.ComponentState(0); s < core.StateCount; s++ {
		if p[s].Type != core.BgSolid {
			t.Errorf("state %d Type = %v, want BgSolid", s, p[s].Type)
		}
		if p[s].Color != red {
			t.Errorf("state %d Color = %v, want red", s, p[s].Color)
		}
	}
}

func TestNewSolidBgPropStates_MultipleStates(t *testing.T) {
	red := sg.RGBA(1, 0, 0, 1)
	blue := sg.RGBA(0, 0, 1, 1)
	p := NewSolidBgPropStates(map[core.ComponentState]sg.Color{
		core.StateDefault: red,
		core.StateHover:   blue,
	})

	if p[core.StateDefault].Color != red {
		t.Errorf("Default Color = %v, want red", p[core.StateDefault].Color)
	}
	if p[core.StateHover].Color != blue {
		t.Errorf("Hover Color = %v, want blue", p[core.StateHover].Color)
	}
	// Active -> Hover -> blue
	if p[core.StateActive].Color != blue {
		t.Errorf("Active Color = %v, want blue (via Hover)", p[core.StateActive].Color)
	}
}

func TestNewSolidBgPropStates_Empty(t *testing.T) {
	p := NewSolidBgPropStates(map[core.ComponentState]sg.Color{})

	for s := core.ComponentState(0); s < core.StateCount; s++ {
		if p[s].Type != core.BgNone {
			t.Errorf("state %d Type = %v, want BgNone", s, p[s].Type)
		}
	}
}

// ---------------------------------------------------------------------------
// ResolveBgFallbacks explicit
// ---------------------------------------------------------------------------

func TestResolveBgFallbacks(t *testing.T) {
	c := sg.RGBA(0.5, 0.5, 0.5, 1)
	var p BackgroundProperty
	p[core.StateDefault] = core.SolidBackground(c)

	ResolveBgFallbacks(&p)

	for s := core.ComponentState(0); s < core.StateCount; s++ {
		if p[s].Type != core.BgSolid {
			t.Errorf("state %d Type = %v, want BgSolid", s, p[s].Type)
		}
	}
}

// ---------------------------------------------------------------------------
// NewFloatPropUniform
// ---------------------------------------------------------------------------

func TestNewFloatPropUniform(t *testing.T) {
	p := NewFloatPropUniform(3.14)

	for s := core.ComponentState(0); s < core.StateCount; s++ {
		if p[s] != 3.14 {
			t.Errorf("state %d = %v, want 3.14", s, p[s])
		}
	}
}

func TestNewFloatPropUniform_Zero(t *testing.T) {
	p := NewFloatPropUniform(0)

	for s := core.ComponentState(0); s < core.StateCount; s++ {
		if p[s] != 0 {
			t.Errorf("state %d = %v, want 0", s, p[s])
		}
	}
}

// ---------------------------------------------------------------------------
// NewFloatPropStates + fallback
// ---------------------------------------------------------------------------

func TestNewFloatPropStates_Fallbacks(t *testing.T) {
	p := NewFloatPropStates(map[core.ComponentState]float64{
		core.StateDefault: 10,
		core.StateHover:   20,
	})

	if p[core.StateDefault] != 10 {
		t.Errorf("Default = %v, want 10", p[core.StateDefault])
	}
	if p[core.StateHover] != 20 {
		t.Errorf("Hover = %v, want 20", p[core.StateHover])
	}
	// Active -> Hover -> 20
	if p[core.StateActive] != 20 {
		t.Errorf("Active = %v, want 20 (via Hover)", p[core.StateActive])
	}
	// Disabled -> Default -> 10
	if p[core.StateDisabled] != 10 {
		t.Errorf("Disabled = %v, want 10 (via Default)", p[core.StateDisabled])
	}
}

func TestNewFloatPropStates_EmptyMapFallsToZero(t *testing.T) {
	p := NewFloatPropStates(map[core.ComponentState]float64{})

	for s := core.ComponentState(0); s < core.StateCount; s++ {
		if math.IsNaN(p[s]) {
			t.Errorf("state %d is still NaN, want 0", s)
		}
		if p[s] != 0 {
			t.Errorf("state %d = %v, want 0", s, p[s])
		}
	}
}

func TestNewFloatPropStates_OnlyDefault(t *testing.T) {
	p := NewFloatPropStates(map[core.ComponentState]float64{
		core.StateDefault: 42,
	})

	for s := core.ComponentState(0); s < core.StateCount; s++ {
		if p[s] != 42 {
			t.Errorf("state %d = %v, want 42", s, p[s])
		}
	}
}

// ---------------------------------------------------------------------------
// ResolveFloatFallbacks explicit
// ---------------------------------------------------------------------------

func TestResolveFloatFallbacks_NaNBecomesZero(t *testing.T) {
	var p FloatProperty
	for i := range p {
		p[i] = math.NaN()
	}

	ResolveFloatFallbacks(&p)

	for s := core.ComponentState(0); s < core.StateCount; s++ {
		if math.IsNaN(p[s]) {
			t.Errorf("state %d still NaN after fallback", s)
		}
		if p[s] != 0 {
			t.Errorf("state %d = %v, want 0", s, p[s])
		}
	}
}

func TestResolveFloatFallbacks_PartialSet(t *testing.T) {
	var p FloatProperty
	for i := range p {
		p[i] = math.NaN()
	}
	p[core.StateDefault] = 5
	p[core.StateHover] = 15

	ResolveFloatFallbacks(&p)

	if p[core.StateDefault] != 5 {
		t.Errorf("Default = %v, want 5", p[core.StateDefault])
	}
	if p[core.StateHover] != 15 {
		t.Errorf("Hover = %v, want 15", p[core.StateHover])
	}
	// Active -> Hover -> 15
	if p[core.StateActive] != 15 {
		t.Errorf("Active = %v, want 15", p[core.StateActive])
	}

	// All states should be resolved (no NaN).
	for s := core.ComponentState(0); s < core.StateCount; s++ {
		if math.IsNaN(p[s]) {
			t.Errorf("state %d still NaN", s)
		}
	}
}

// ---------------------------------------------------------------------------
// Config[G].Group
// ---------------------------------------------------------------------------

func TestConfig_Group_Primary(t *testing.T) {
	cfg := Config[int]{Primary: 42}

	got := cfg.Group(Primary)
	if *got != 42 {
		t.Errorf("Group(Primary) = %d, want 42", *got)
	}
}

func TestConfig_Group_FallbackToPrimary(t *testing.T) {
	cfg := Config[int]{Primary: 99}

	// Secondary is not set (nil pointer), should fall back to Primary.
	got := cfg.Group(Secondary)
	if *got != 99 {
		t.Errorf("Group(Secondary) = %d, want 99 (fallback to Primary)", *got)
	}
}

func TestConfig_Group_SpecificVariant(t *testing.T) {
	val := 77
	cfg := Config[int]{Primary: 42}
	cfg.Variants[Secondary-1] = &val

	got := cfg.Group(Secondary)
	if *got != 77 {
		t.Errorf("Group(Secondary) = %d, want 77", *got)
	}

	// Primary is still 42.
	got2 := cfg.Group(Primary)
	if *got2 != 42 {
		t.Errorf("Group(Primary) = %d, want 42", *got2)
	}
}

func TestConfig_Group_MultipleVariants(t *testing.T) {
	dangerVal := 100
	successVal := 200
	cfg := Config[int]{Primary: 1}
	cfg.Variants[Danger-1] = &dangerVal
	cfg.Variants[Success-1] = &successVal

	if *cfg.Group(Danger) != 100 {
		t.Errorf("Group(Danger) = %d, want 100", *cfg.Group(Danger))
	}
	if *cfg.Group(Success) != 200 {
		t.Errorf("Group(Success) = %d, want 200", *cfg.Group(Success))
	}
	// Unset variant falls back.
	if *cfg.Group(Warning) != 1 {
		t.Errorf("Group(Warning) = %d, want 1 (fallback)", *cfg.Group(Warning))
	}
}

func TestConfig_Group_ZeroVariantIsPrimary(t *testing.T) {
	cfg := Config[string]{Primary: "hello"}

	// Variant(0) is Primary.
	got := cfg.Group(Variant(0))
	if *got != "hello" {
		t.Errorf("Group(0) = %q, want %q", *got, "hello")
	}
}

func TestConfig_Group_OutOfRangeVariant(t *testing.T) {
	cfg := Config[int]{Primary: 55}

	// Out of range variant falls back to Primary.
	got := cfg.Group(Variant(VariantCount + 10))
	if *got != 55 {
		t.Errorf("Group(out-of-range) = %d, want 55 (fallback)", *got)
	}
}
