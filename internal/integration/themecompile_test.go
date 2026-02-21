package integration

import (
	"bytes"
	"image"
	"image/png"
	"math"
	"os"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/devthicket/willow"
	ui "github.com/devthicket/willowui"
	"github.com/devthicket/willowui/internal/core"
	interntheme "github.com/devthicket/willowui/internal/theme"
)

// ---------------------------------------------------------------------------
// Color parsing tests
// ---------------------------------------------------------------------------

func TestParseColor_Hex6(t *testing.T) {
	c, err := ui.ParseColor("#3A7AFE")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := willow.RGBA(0x3A/255.0, 0x7A/255.0, 0xFE/255.0, 1)
	assertColorApprox(t, c, want)
}

func TestParseColor_Hex8(t *testing.T) {
	c, err := ui.ParseColor("#3A7AFE80")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := willow.RGBA(0x3A/255.0, 0x7A/255.0, 0xFE/255.0, 0x80/255.0)
	assertColorApprox(t, c, want)
}

func TestParseColor_Hex3(t *testing.T) {
	c, err := ui.ParseColor("#38F")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// #38F → #3388FF
	want := willow.RGBA(0x33/255.0, 0x88/255.0, 0xFF/255.0, 1)
	assertColorApprox(t, c, want)
}

func TestParseColor_RGBA(t *testing.T) {
	c, err := ui.ParseColor("rgba(66, 133, 244, 0.35)")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := willow.RGBA(66.0/255, 133.0/255, 244.0/255, 0.35)
	assertColorApprox(t, c, want)
}

func TestParseColor_RGB(t *testing.T) {
	c, err := ui.ParseColor("rgb(66, 133, 244)")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := willow.RGBA(66.0/255, 133.0/255, 244.0/255, 1)
	assertColorApprox(t, c, want)
}

func TestParseColor_Named(t *testing.T) {
	tests := []struct {
		input string
		want  willow.Color
	}{
		{"white", willow.RGBA(1, 1, 1, 1)},
		{"black", willow.RGBA(0, 0, 0, 1)},
		{"transparent", willow.RGBA(0, 0, 0, 0)},
		{"WHITE", willow.RGBA(1, 1, 1, 1)},
	}
	for _, tt := range tests {
		c, err := ui.ParseColor(tt.input)
		if err != nil {
			t.Errorf("ParseColor(%q): unexpected error: %v", tt.input, err)
			continue
		}
		assertColorApprox(t, c, tt.want)
	}
}

func TestParseColor_Invalid(t *testing.T) {
	_, err := ui.ParseColor("not-a-color")
	if err == nil {
		t.Fatal("expected error for invalid color")
	}
	if !strings.Contains(err.Error(), "invalid color format") {
		t.Errorf("error should mention 'invalid color format', got: %v", err)
	}
}

// ---------------------------------------------------------------------------
// State fallback tests
// ---------------------------------------------------------------------------

func TestCompile_MissingHover_FallsToDefault(t *testing.T) {
	js := `{
		"components": {
			"label": {
				"primary": {
					"textColor": { "default": "#FF0000" }
				}
			}
		}
	}`
	theme, err := ui.LoadTheme([]byte(js))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Hover should fall back to default.
	defColor := theme.Label.Primary.TextColor.Resolve(ui.StateDefault)
	hoverColor := theme.Label.Primary.TextColor.Resolve(ui.StateHover)
	assertColorApprox(t, hoverColor, defColor)
}

func TestCompile_MissingActive_FallsToHoverThenDefault(t *testing.T) {
	js := `{
		"components": {
			"label": {
				"primary": {
					"textColor": {
						"default": "#FF0000",
						"hover": "#00FF00"
					}
				}
			}
		}
	}`
	theme, err := ui.LoadTheme([]byte(js))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Active fallback chain: hover, default. Since hover is defined, active = hover.
	hoverColor := theme.Label.Primary.TextColor.Resolve(ui.StateHover)
	activeColor := theme.Label.Primary.TextColor.Resolve(ui.StateActive)
	assertColorApprox(t, activeColor, hoverColor)
}

func TestCompile_MissingDisabled_FallsToDefault(t *testing.T) {
	js := `{
		"components": {
			"label": {
				"primary": {
					"textColor": { "default": "#AABBCC" }
				}
			}
		}
	}`
	theme, err := ui.LoadTheme([]byte(js))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Disabled falls back to default (no auto-dimming).
	defColor := theme.Label.Primary.TextColor.Resolve(ui.StateDefault)
	disabledColor := theme.Label.Primary.TextColor.Resolve(ui.StateDisabled)
	assertColorApprox(t, disabledColor, defColor)
}

func TestCompile_MissingFocus_FallsToActiveThenHover(t *testing.T) {
	js := `{
		"components": {
			"label": {
				"primary": {
					"textColor": {
						"default": "#FF0000",
						"hover": "#00FF00",
						"active": "#0000FF"
					}
				}
			}
		}
	}`
	theme, err := ui.LoadTheme([]byte(js))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Focus fallback chain: active, hover, default. Active is defined.
	activeColor := theme.Label.Primary.TextColor.Resolve(ui.StateActive)
	focusColor := theme.Label.Primary.TextColor.Resolve(ui.StateFocus)
	assertColorApprox(t, focusColor, activeColor)
}

func TestCompile_AllStatesExplicit(t *testing.T) {
	js := `{
		"components": {
			"label": {
				"primary": {
					"textColor": {
						"default":       "#FF0000",
						"hover":         "#00FF00",
						"active":        "#0000FF",
						"disabled":      "#AAAAAA",
						"focus":         "#FFFF00",
						"focusHover":    "#FF00FF",
						"focusActive":   "#00FFFF",
						"focusDisabled": "#888888"
					}
				}
			}
		}
	}`
	theme, err := ui.LoadTheme([]byte(js))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Verify each state has its own explicit value.
	prop := &theme.Label.Primary.TextColor
	assertColorHex(t, prop.Resolve(ui.StateDefault), "#FF0000")
	assertColorHex(t, prop.Resolve(ui.StateHover), "#00FF00")
	assertColorHex(t, prop.Resolve(ui.StateActive), "#0000FF")
	assertColorHex(t, prop.Resolve(ui.StateDisabled), "#AAAAAA")
	assertColorHex(t, prop.Resolve(ui.StateFocus), "#FFFF00")
	assertColorHex(t, prop.Resolve(ui.StateFocusHover), "#FF00FF")
	assertColorHex(t, prop.Resolve(ui.StateFocusActive), "#00FFFF")
	assertColorHex(t, prop.Resolve(ui.StateFocusDisabled), "#888888")
}

// ---------------------------------------------------------------------------
// Group fallback tests
// ---------------------------------------------------------------------------

func TestCompile_MissingSecondary_FallsToPrimary(t *testing.T) {
	js := `{
		"components": {
			"label": {
				"primary": {
					"textColor": { "default": "#FF0000" }
				}
			}
		}
	}`
	theme, err := ui.LoadTheme([]byte(js))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Secondary not defined — Group(Secondary) should fall back to Primary.
	pri := theme.Label.Group(ui.Primary)
	sec := theme.Label.Group(ui.Secondary)
	if pri != sec {
		t.Error("missing secondary should fall back to primary group pointer")
	}
}

func TestCompile_MissingAccent_FallsToPrimary(t *testing.T) {
	js := `{
		"components": {
			"button": {
				"primary": {
					"backgroundColor": { "default": "#3A7AFE" },
					"textColor": { "default": "#FFFFFF" }
				}
			}
		}
	}`
	theme, err := ui.LoadTheme([]byte(js))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	pri := theme.Button.Group(ui.Primary)
	acc := theme.Button.Group(ui.Accent)
	if pri != acc {
		t.Error("missing accent should fall back to primary group pointer")
	}
}

func TestCompile_ExplicitAccent_UsesOwnValues(t *testing.T) {
	js := `{
		"components": {
			"button": {
				"primary": {
					"backgroundColor": { "default": "#3A7AFE" },
					"textColor": { "default": "#FFFFFF" }
				},
				"accent": {
					"backgroundColor": { "default": "#D43A3A" },
					"textColor": { "default": "#FFFFFF" }
				}
			}
		}
	}`
	theme, err := ui.LoadTheme([]byte(js))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	accBg := theme.Button.Group(ui.Accent).Background.Resolve(ui.StateDefault)
	if accBg.Type != ui.BgSolid {
		t.Fatal("accent background should be BgSolid")
	}
	wantAccent, _ := ui.ParseColor("#D43A3A")
	assertColorApprox(t, accBg.Color, wantAccent)

	// Primary should still be its own value.
	priBg := theme.Button.Group(ui.Primary).Background.Resolve(ui.StateDefault)
	wantPri, _ := ui.ParseColor("#3A7AFE")
	assertColorApprox(t, priBg.Color, wantPri)
}

func TestCompile_MissingComponentType_Allowed(t *testing.T) {
	// Theme with only button — no toggle section. Should compile fine.
	js := `{
		"components": {
			"button": {
				"primary": {
					"backgroundColor": { "default": "#3A7AFE" },
					"textColor": { "default": "#FFFFFF" }
				}
			}
		}
	}`
	theme, err := ui.LoadTheme([]byte(js))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Toggle config should be zero-value (no toggle section in JSON).
	zero := willow.Color{}
	if theme.Toggle.Primary.TrackColor[ui.StateDefault] != zero {
		t.Error("toggle should be zero-value when not in JSON")
	}
}

// ---------------------------------------------------------------------------
// Validation tests
// ---------------------------------------------------------------------------

func TestCompile_MissingPrimaryGroup_Error(t *testing.T) {
	js := `{
		"components": {
			"button": {
				"accent": {
					"backgroundColor": { "default": "#D43A3A" },
					"textColor": { "default": "#FFFFFF" }
				}
			}
		}
	}`
	_, err := ui.LoadTheme([]byte(js))
	if err == nil {
		t.Fatal("expected error for missing primary group")
	}
	if !strings.Contains(err.Error(), "missing required group \"primary\"") {
		t.Errorf("error should mention missing primary, got: %v", err)
	}
}

func TestCompile_MissingDefaultState_Allowed(t *testing.T) {
	// "default" state is optional — missing it compiles cleanly with a zero/transparent fallback.
	js := `{
		"components": {
			"label": {
				"primary": {
					"textColor": { "hover": "#00FF00" }
				}
			}
		}
	}`
	_, err := ui.LoadTheme([]byte(js))
	if err != nil {
		t.Fatalf("expected no error for missing default state, got: %v", err)
	}
}

func TestCompile_InvalidColor_ErrorWithPath(t *testing.T) {
	js := `{
		"components": {
			"button": {
				"primary": {
					"backgroundColor": { "default": "not-a-color" },
					"textColor": { "default": "#FFFFFF" }
				}
			}
		}
	}`
	_, err := ui.LoadTheme([]byte(js))
	if err == nil {
		t.Fatal("expected error for invalid color")
	}
	errStr := err.Error()
	if !strings.Contains(errStr, "button.primary.backgroundColor.default") {
		t.Errorf("error should include JSON path, got: %v", errStr)
	}
	if !strings.Contains(errStr, "invalid color format") {
		t.Errorf("error should mention invalid color format, got: %v", errStr)
	}
}

func TestCompile_MultipleErrors(t *testing.T) {
	js := `{
		"components": {
			"button": {
				"primary": {
					"backgroundColor": { "default": "bad1" },
					"textColor": { "default": "bad2" }
				}
			}
		}
	}`
	_, err := ui.LoadTheme([]byte(js))
	if err == nil {
		t.Fatal("expected errors")
	}
	errStr := err.Error()
	if !strings.Contains(errStr, "backgroundColor") || !strings.Contains(errStr, "textColor") {
		t.Errorf("should report errors for both properties, got: %v", errStr)
	}
}

func TestCompile_EmptyJSON_NoError(t *testing.T) {
	// Empty object — no component sections. This is valid (empty theme).
	js := `{}`
	theme, err := ui.LoadTheme([]byte(js))
	if err != nil {
		t.Fatalf("empty JSON should compile: %v", err)
	}
	if theme == nil {
		t.Fatal("theme should not be nil")
	}
}

// ---------------------------------------------------------------------------
// Integration tests
// ---------------------------------------------------------------------------

func TestCompile_FullTheme(t *testing.T) {
	js := `{
		"colors": {
			"primaryBlue": "#3A7AFE",
			"primaryHover": "#4E94FF",
			"primaryActive": "#3369CC",
			"accentRed": "#D43A3A",
			"fontDefault": "#EEEEEE",
			"fontDisabled": "rgba(238, 238, 238, 0.4)",
			"disabledBg": "rgba(115, 115, 122, 0.6)",
			"surfaceDark": "#262629",
			"borderDefault": "#4D4D54",
			"focusBlue": "#59A6FF",
			"trackGray": "#666670"
		},
		"components": {
			"_": {
				"borderWidth": 1
			},
			"button": {
				"_": {
					"cornerRadius": 4
				},
				"primary": {
					"backgroundColor": {
						"default": "$primaryBlue",
						"hover": "$primaryHover",
						"active": "$primaryActive",
						"disabled": "$disabledBg"
					},
					"textColor": {
						"default": "$fontDefault",
						"disabled": "$fontDisabled"
					},
					"borderColor": {
						"default": "$borderDefault",
						"focus": "$focusBlue"
					},
					"padding": { "top": 8, "right": 16, "bottom": 8, "left": 16 }
				},
				"accent": {
					"backgroundColor": {
						"default": "$accentRed",
						"hover": "#E04848",
						"active": "#B02828"
					},
					"textColor": {
						"default": "#FFFFFF"
					}
				}
			},
			"toggle": {
				"primary": {
					"trackColor": {
						"default": "$primaryBlue",
						"disabled": "$disabledBg"
					},
					"thumbColor": {
						"default": "$fontDefault",
						"disabled": "$fontDisabled"
					},
					"cornerRadius": 12
				}
			},
			"panel": {
				"primary": {
					"backgroundColor": {
						"default": "$surfaceDark"
					}
				}
			},
			"window": {
				"primary": {
					"backgroundColor": {
						"default": "$surfaceDark"
					},
					"titleBackgroundColor": {
						"default": "$primaryBlue"
					},
					"titleTextColor": {
						"default": "$fontDefault"
					},
					"resizeHandleColor": {
						"default": "$trackGray"
					}
				}
			},
			"label": {
				"primary": {
					"textColor": {
						"default": "$fontDefault"
					}
				}
			}
		}
	}`

	theme, err := ui.LoadTheme([]byte(js))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify button primary background ($reference resolved).
	bg := theme.Button.Primary.Background.Resolve(ui.StateDefault)
	if bg.Type != ui.BgSolid {
		t.Fatal("button primary bg should be BgSolid")
	}
	wantBg, _ := ui.ParseColor("#3A7AFE")
	assertColorApprox(t, bg.Color, wantBg)

	// Verify button padding.
	if theme.Button.Primary.Padding != (ui.Insets{Top: 8, Right: 16, Bottom: 8, Left: 16}) {
		t.Errorf("button padding = %v", theme.Button.Primary.Padding)
	}

	// Verify button cornerRadius inherited from button._.
	if theme.Button.Primary.CornerRadius != 4 {
		t.Errorf("button cornerRadius = %v, want 4", theme.Button.Primary.CornerRadius)
	}

	// Verify button borderWidth inherited from components._.
	if theme.Button.Primary.BorderWidth != 1 {
		t.Errorf("button borderWidth = %v, want 1 (from components._)", theme.Button.Primary.BorderWidth)
	}

	// Verify accent variant exists and differs from primary.
	accBg := theme.Button.Group(ui.Accent).Background.Resolve(ui.StateDefault)
	wantAcc, _ := ui.ParseColor("#D43A3A")
	assertColorApprox(t, accBg.Color, wantAcc)

	// Accent should also inherit cornerRadius from button._.
	if theme.Button.Group(ui.Accent).CornerRadius != 4 {
		t.Errorf("accent cornerRadius = %v, want 4 (from button._)", theme.Button.Group(ui.Accent).CornerRadius)
	}

	// Verify panel background ($reference).
	panelBg := theme.Panel.Primary.Background.Resolve(ui.StateDefault)
	wantPanel, _ := ui.ParseColor("#262629")
	assertColorApprox(t, panelBg.Color, wantPanel)

	// Verify window titleBackground ($reference).
	titleBg := theme.Window.Primary.TitleBackground.Resolve(ui.StateDefault)
	if titleBg.Type != ui.BgSolid {
		t.Fatal("window title bg should be BgSolid")
	}
	assertColorApprox(t, titleBg.Color, wantBg)

	// Verify label textColor ($reference).
	labelText := theme.Label.Primary.TextColor.Resolve(ui.StateDefault)
	wantFont, _ := ui.ParseColor("#EEEEEE")
	assertColorApprox(t, labelText, wantFont)

	// Verify toggle cornerRadius round-trips through JSON.
	if theme.Toggle.Primary.CornerRadius != 12 {
		t.Errorf("toggle cornerRadius = %v, want 12", theme.Toggle.Primary.CornerRadius)
	}
}

func TestCompileFromFile(t *testing.T) {
	// Write a temp file and load it.
	dir := t.TempDir()
	path := dir + "/test.json"
	js := `{
		"components": {
			"label": {
				"primary": {
					"textColor": { "default": "#FF0000" }
				}
			}
		}
	}`
	if err := writeTestFile(path, js); err != nil {
		t.Fatal(err)
	}
	theme, err := ui.LoadThemeFromFile(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	wantColor, _ := ui.ParseColor("#FF0000")
	assertColorApprox(t, theme.Label.Primary.TextColor.Resolve(ui.StateDefault), wantColor)
}

func TestCompileFromFS(t *testing.T) {
	js := `{
		"components": {
			"label": {
				"primary": {
					"textColor": { "default": "#00FF00" }
				}
			}
		}
	}`
	fsys := fstest.MapFS{
		"themes/dark.json": &fstest.MapFile{Data: []byte(js)},
	}
	theme, err := ui.LoadThemeFromFS(fsys, "themes/dark.json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	wantColor, _ := ui.ParseColor("#00FF00")
	assertColorApprox(t, theme.Label.Primary.TextColor.Resolve(ui.StateDefault), wantColor)
}

func TestCompile_ResultIsUsable(t *testing.T) {
	js := `{
		"components": {
			"button": {
				"primary": {
					"backgroundColor": { "default": "#FF0000" },
					"textColor": { "default": "#FFFFFF" },
					"padding": { "top": 4, "right": 8, "bottom": 4, "left": 8 }
				}
			},
			"label": {
				"primary": {
					"textColor": { "default": "#EEEEEE" }
				}
			}
		}
	}`
	theme, err := ui.LoadTheme([]byte(js))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Use the compiled theme with a component.
	btn := ui.NewButton("test-btn", "Test", newTestFont(), 16)
	btn.SetTheme(theme)

	// Verify the button picked up the theme's background.
	wantBg, _ := ui.ParseColor("#FF0000")
	if !colorApproxEqual(btn.BgNode().Color(), wantBg) {
		t.Errorf("button bg = %v, want ~%v", btn.BgNode().Color(), wantBg)
	}
}

// ---------------------------------------------------------------------------
// ValidateTheme tests
// ---------------------------------------------------------------------------

func TestValidateTheme_AllPresent(t *testing.T) {
	js := `{
		"components": {
			"button": {
				"primary": {
					"backgroundColor": { "default": "#3A7AFE" },
					"textColor": { "default": "#FFFFFF" }
				}
			},
			"label": {
				"primary": {
					"textColor": { "default": "#FFFFFF" }
				}
			}
		}
	}`
	theme, err := ui.LoadTheme([]byte(js))
	if err != nil {
		t.Fatalf("compile: %v", err)
	}
	if err := ui.ValidateTheme(theme, "button", "label"); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestValidateTheme_Missing(t *testing.T) {
	js := `{
		"components": {
			"button": {
				"primary": {
					"backgroundColor": { "default": "#3A7AFE" },
					"textColor": { "default": "#FFFFFF" }
				}
			}
		}
	}`
	theme, err := ui.LoadTheme([]byte(js))
	if err != nil {
		t.Fatalf("compile: %v", err)
	}
	err = ui.ValidateTheme(theme, "button", "window")
	if err == nil {
		t.Fatal("expected error for missing window config")
	}
	if !strings.Contains(err.Error(), "window") {
		t.Errorf("error should mention window, got: %v", err)
	}
}

// ---------------------------------------------------------------------------
// Unknown top-level key warning test
// ---------------------------------------------------------------------------

func TestCompile_UnknownTopLevelKey_Accepted(t *testing.T) {
	// Unknown top-level keys are silently accepted for forward compatibility.
	js := `{
		"variables": { "primary": "#3A7AFE" },
		"$schema": "https://example.com/theme.json",
		"components": {
			"label": {
				"primary": {
					"textColor": { "default": "#FFFFFF" }
				}
			}
		}
	}`
	theme, err := ui.LoadTheme([]byte(js))
	if err != nil {
		t.Fatalf("unknown top-level keys should not cause error: %v", err)
	}
	// Label should still compile correctly.
	wantColor, _ := ui.ParseColor("#FFFFFF")
	assertColorApprox(t, theme.Label.Primary.TextColor.Resolve(ui.StateDefault), wantColor)
}

// ---------------------------------------------------------------------------
// Nine-slice background tests
// ---------------------------------------------------------------------------

// testPNG16 returns a minimal 16x16 PNG image as bytes.
func testPNG16() []byte {
	img := image.NewRGBA(image.Rect(0, 0, 16, 16))
	// Fill with some color so it's not empty.
	for y := 0; y < 16; y++ {
		for x := 0; x < 16; x++ {
			img.Set(x, y, image.White)
		}
	}
	var buf bytes.Buffer
	_ = png.Encode(&buf, img)
	return buf.Bytes()
}

func TestCompile_NineGridBg(t *testing.T) {
	pngData := testPNG16()
	fsys := fstest.MapFS{
		"theme.json": &fstest.MapFile{Data: []byte(`{
			"nine-grids": {
				"btnGrid": {
					"source": "button_bg.png",
					"innerRegion": { "x": 4, "y": 4, "width": 8, "height": 8 },
					"auto-slice": { "top": 4, "right": 4, "bottom": 4, "left": 4 }
				}
			},
			"components": {
				"button": {
					"primary": {
						"backgroundGrid": { "default": "btnGrid" },
						"textColor": { "default": "#FFFFFF" }
					}
				}
			}
		}`)},
		"button_bg.png": &fstest.MapFile{Data: pngData},
	}
	theme, err := ui.LoadThemeFromFS(fsys, "theme.json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	bg := theme.Button.Primary.Background.Resolve(ui.StateDefault)
	if bg.Type != ui.BgNineSlice {
		t.Fatalf("expected BgNineSlice, got %d", bg.Type)
	}
	if bg.Slice == nil {
		t.Fatal("Slice should not be nil")
	}
	if bg.Slice.Insets.Top != 4 || bg.Slice.Insets.Left != 4 {
		t.Errorf("insets = %v, want {4,4,4,4}", bg.Slice.Insets)
	}
	// Region should be resolved (non-zero width/height from atlas packing).
	if bg.Slice.Region.Width != 16 || bg.Slice.Region.Height != 16 {
		t.Errorf("Region = %dx%d, want 16x16", bg.Slice.Region.Width, bg.Slice.Region.Height)
	}
	if theme.Atlas == nil {
		t.Error("theme.Atlas should not be nil when nine-grid images are used")
	}
	// Check innerRegion.
	if bg.Slice.InnerRegion.X != 4 || bg.Slice.InnerRegion.Y != 4 {
		t.Errorf("InnerRegion = %v, want x=4 y=4", bg.Slice.InnerRegion)
	}
}

func TestCompile_NineGridBg_GridWinsOverColor(t *testing.T) {
	pngData := testPNG16()
	fsys := fstest.MapFS{
		"theme.json": &fstest.MapFile{Data: []byte(`{
			"nine-grids": {
				"btnGrid": {
					"source": "btn.png",
					"innerRegion": { "x": 4, "y": 4, "width": 8, "height": 8 },
					"auto-slice": { "top": 4, "right": 4, "bottom": 4, "left": 4 }
				}
			},
			"components": {
				"button": {
					"primary": {
						"backgroundColor": { "default": "#FF0000" },
						"backgroundGrid": { "default": "btnGrid" },
						"textColor": { "default": "#FFFFFF" }
					}
				}
			}
		}`)},
		"btn.png": &fstest.MapFile{Data: pngData},
	}
	theme, err := ui.LoadThemeFromFS(fsys, "theme.json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Grid should win over color.
	defBg := theme.Button.Primary.Background.Resolve(ui.StateDefault)
	if defBg.Type != ui.BgNineSlice {
		t.Errorf("default should be BgNineSlice (grid wins), got %d", defBg.Type)
	}
}

func TestCompile_NineGridBg_RawBytes_RejectsImages(t *testing.T) {
	js := `{
		"nine-grids": {
			"btnGrid": {
				"source": "button_bg.png",
				"innerRegion": { "x": 4, "y": 4, "width": 8, "height": 8 },
				"auto-slice": { "top": 4, "right": 4, "bottom": 4, "left": 4 }
			}
		},
		"components": {
			"button": {
				"primary": {
					"backgroundGrid": { "default": "btnGrid" },
					"textColor": { "default": "#FFFFFF" }
				}
			}
		}
	}`
	_, err := ui.LoadTheme([]byte(js))
	if err == nil {
		t.Fatal("expected error for nine-grid in LoadTheme (raw bytes)")
	}
	if !strings.Contains(err.Error(), "LoadThemeFromFile") {
		t.Errorf("error should mention LoadThemeFromFile, got: %v", err)
	}
}

func TestCompile_NineGridBg_MissingImage(t *testing.T) {
	fsys := fstest.MapFS{
		"theme.json": &fstest.MapFile{Data: []byte(`{
			"nine-grids": {
				"btnGrid": {
					"source": "nonexistent.png",
					"innerRegion": { "x": 4, "y": 4, "width": 8, "height": 8 },
					"auto-slice": { "top": 4, "right": 4, "bottom": 4, "left": 4 }
				}
			},
			"components": {
				"button": {
					"primary": {
						"backgroundGrid": { "default": "btnGrid" },
						"textColor": { "default": "#FFFFFF" }
					}
				}
			}
		}`)},
	}
	_, err := ui.LoadThemeFromFS(fsys, "theme.json")
	if err == nil {
		t.Fatal("expected error for missing image file")
	}
	if !strings.Contains(err.Error(), "nonexistent.png") {
		t.Errorf("error should mention the missing file, got: %v", err)
	}
}

func TestCompile_NineGridBg_InvalidInsets(t *testing.T) {
	js := `{
		"nine-grids": {
			"badGrid": {
				"source": "btn.png",
				"innerRegion": { "x": 0, "y": 0, "width": 16, "height": 16 },
				"auto-slice": { "top": 0, "right": 0, "bottom": 0, "left": 0 }
			}
		},
		"components": {
			"button": {
				"primary": {
					"backgroundGrid": { "default": "badGrid" },
					"textColor": { "default": "#FFFFFF" }
				}
			}
		}
	}`
	_, err := ui.LoadTheme([]byte(js))
	if err == nil {
		t.Fatal("expected error for zero insets")
	}
	if !strings.Contains(err.Error(), "insets must have at least one positive value") {
		t.Errorf("error should mention insets, got: %v", err)
	}
}

func TestCompile_NineGridBg_FallbackChain(t *testing.T) {
	pngData := testPNG16()
	fsys := fstest.MapFS{
		"theme.json": &fstest.MapFile{Data: []byte(`{
			"nine-grids": {
				"btnGrid": {
					"source": "btn.png",
					"innerRegion": { "x": 4, "y": 4, "width": 8, "height": 8 },
					"auto-slice": { "top": 4, "right": 4, "bottom": 4, "left": 4 }
				}
			},
			"components": {
				"button": {
					"primary": {
						"backgroundGrid": { "default": "btnGrid" },
						"textColor": { "default": "#FFFFFF" }
					}
				}
			}
		}`)},
		"btn.png": &fstest.MapFile{Data: pngData},
	}
	theme, err := ui.LoadThemeFromFS(fsys, "theme.json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Active (not defined) should fall back to hover → default (nine-slice).
	activeBg := theme.Button.Primary.Background.Resolve(ui.StateActive)
	if activeBg.Type != ui.BgNineSlice {
		t.Errorf("active should fall back to BgNineSlice, got %d", activeBg.Type)
	}
}

func TestCompile_NineGridBg_UnknownGridKey(t *testing.T) {
	js := `{
		"components": {
			"panel": {
				"primary": {
					"backgroundGrid": { "default": "nonexistent" }
				}
			}
		}
	}`
	_, err := ui.LoadTheme([]byte(js))
	if err == nil {
		t.Fatal("expected error for unknown nine-grid key")
	}
	if !strings.Contains(err.Error(), "unknown nine-grid key") {
		t.Errorf("error should mention unknown grid key, got: %v", err)
	}
}

func TestCompile_NineGridBg_ManualSlices(t *testing.T) {
	pngData := testPNG16()
	fsys := fstest.MapFS{
		"theme.json": &fstest.MapFile{Data: []byte(`{
			"nine-grids": {
				"panelGrid": {
					"source": "panel.png",
					"region": { "x": 0, "y": 0, "width": 16, "height": 16 },
					"innerRegion": { "x": 4, "y": 4, "width": 8, "height": 8 },
					"slices": {
						"topLeftCorner": { "x": 0, "y": 0, "width": 4, "height": 4 },
						"topEdge": { "x": 4, "y": 0, "width": 8, "height": 4 },
						"topRightCorner": { "x": 12, "y": 0, "width": 4, "height": 4 },
						"leftEdge": { "x": 0, "y": 4, "width": 4, "height": 8 },
						"center": { "x": 4, "y": 4, "width": 8, "height": 8 },
						"rightEdge": { "x": 12, "y": 4, "width": 4, "height": 8 },
						"bottomLeftCorner": { "x": 0, "y": 12, "width": 4, "height": 4 },
						"bottomEdge": { "x": 4, "y": 12, "width": 8, "height": 4 },
						"bottomRightCorner": { "x": 12, "y": 12, "width": 4, "height": 4 }
					}
				}
			},
			"components": {
				"panel": {
					"primary": {
						"backgroundGrid": { "default": "panelGrid" }
					}
				}
			}
		}`)},
		"panel.png": &fstest.MapFile{Data: pngData},
	}
	theme, err := ui.LoadThemeFromFS(fsys, "theme.json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	bg := theme.Panel.Primary.Background.Resolve(ui.StateDefault)
	if bg.Type != ui.BgNineSlice {
		t.Fatalf("expected BgNineSlice, got %d", bg.Type)
	}
	// Insets should be derived from corner dimensions.
	if bg.Slice.Insets.Left != 4 || bg.Slice.Insets.Top != 4 {
		t.Errorf("insets = %v, want Left=4 Top=4", bg.Slice.Insets)
	}
	if bg.Slice.Insets.Right != 4 || bg.Slice.Insets.Bottom != 4 {
		t.Errorf("insets = %v, want Right=4 Bottom=4", bg.Slice.Insets)
	}
}

func TestCompile_NineGridBg_MissingSlicesAndAutoSlice(t *testing.T) {
	js := `{
		"nine-grids": {
			"badGrid": {
				"source": "btn.png",
				"innerRegion": { "x": 4, "y": 4, "width": 8, "height": 8 }
			}
		}
	}`
	_, err := ui.LoadTheme([]byte(js))
	if err == nil {
		t.Fatal("expected error for missing slices and auto-slice")
	}
	if !strings.Contains(err.Error(), "requires either \"slices\" or \"auto-slice\"") {
		t.Errorf("error should mention missing slices, got: %v", err)
	}
}

// ---------------------------------------------------------------------------
// Panel border + borderWidth compile tests
// ---------------------------------------------------------------------------

func TestCompile_PanelBorderAndBorderWidth(t *testing.T) {
	js := `{
		"components": {
			"panel": {
				"primary": {
					"backgroundColor": { "default": "#262629" },
					"borderColor": {
						"default": "#4D4D54",
						"focus": "#59A6FF"
					},
					"borderWidth": 2
				}
			}
		}
	}`
	theme, err := ui.LoadTheme([]byte(js))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Check border color.
	borderDef := theme.Panel.Primary.Border.Resolve(ui.StateDefault)
	wantBorder, _ := ui.ParseColor("#4D4D54")
	assertColorApprox(t, borderDef, wantBorder)

	borderFocus := theme.Panel.Primary.Border.Resolve(ui.StateFocus)
	wantFocus, _ := ui.ParseColor("#59A6FF")
	assertColorApprox(t, borderFocus, wantFocus)

	// Check borderWidth.
	if theme.Panel.Primary.BorderWidth != 2 {
		t.Errorf("borderWidth = %f, want 2", theme.Panel.Primary.BorderWidth)
	}
}

func TestCompile_PanelBorderWidth_InheritsToVariant(t *testing.T) {
	js := `{
		"components": {
			"panel": {
				"primary": {
					"backgroundColor": { "default": "#262629" },
					"borderColor": { "default": "#4D4D54" },
					"borderWidth": 3
				},
				"accent": {
					"backgroundColor": { "default": "#FF0000" }
				}
			}
		}
	}`
	theme, err := ui.LoadTheme([]byte(js))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Accent variant should fall back to primary's borderWidth (3).
	accGroup := theme.Panel.Group(ui.Accent)
	if accGroup.BorderWidth != 3 {
		t.Errorf("accent borderWidth = %f, want 3 (inherited from primary)", accGroup.BorderWidth)
	}
	// Accent border should fall back to primary's border.
	wantBorder, _ := ui.ParseColor("#4D4D54")
	assertColorApprox(t, accGroup.Border.Resolve(ui.StateDefault), wantBorder)
}

// ---------------------------------------------------------------------------
// Colors section tests
// ---------------------------------------------------------------------------

func TestCompile_ColorsSection_BasicReference(t *testing.T) {
	js := `{
		"colors": {
			"myRed": "#FF0000",
			"myGreen": "#00FF00"
		},
		"components": {
			"label": {
				"primary": {
					"textColor": {
						"default": "$myRed",
						"hover": "$myGreen"
					}
				}
			}
		}
	}`
	theme, err := ui.LoadTheme([]byte(js))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	wantRed, _ := ui.ParseColor("#FF0000")
	wantGreen, _ := ui.ParseColor("#00FF00")
	assertColorApprox(t, theme.Label.Primary.TextColor.Resolve(ui.StateDefault), wantRed)
	assertColorApprox(t, theme.Label.Primary.TextColor.Resolve(ui.StateHover), wantGreen)
}

func TestCompile_ColorsSection_UndefinedReference_Error(t *testing.T) {
	js := `{
		"colors": {
			"myRed": "#FF0000"
		},
		"components": {
			"label": {
				"primary": {
					"textColor": {
						"default": "$noSuchColor"
					}
				}
			}
		}
	}`
	_, err := ui.LoadTheme([]byte(js))
	if err == nil {
		t.Fatal("expected error for undefined $reference")
	}
	if !strings.Contains(err.Error(), "undefined color reference") {
		t.Errorf("error should mention undefined reference, got: %v", err)
	}
}

func TestCompile_ColorsSection_InvalidColorValue_Error(t *testing.T) {
	js := `{
		"colors": {
			"badColor": "not-a-color"
		},
		"components": {
			"label": {
				"primary": {
					"textColor": { "default": "#FFFFFF" }
				}
			}
		}
	}`
	_, err := ui.LoadTheme([]byte(js))
	if err == nil {
		t.Fatal("expected error for invalid color in colors section")
	}
	if !strings.Contains(err.Error(), "colors.badColor") {
		t.Errorf("error should mention colors.badColor, got: %v", err)
	}
}

func TestCompile_ColorsSection_BackgroundReference(t *testing.T) {
	js := `{
		"colors": {
			"blue": "#3A7AFE"
		},
		"components": {
			"button": {
				"primary": {
					"backgroundColor": {
						"default": "$blue"
					},
					"textColor": { "default": "#FFFFFF" }
				}
			}
		}
	}`
	theme, err := ui.LoadTheme([]byte(js))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	bg := theme.Button.Primary.Background.Resolve(ui.StateDefault)
	if bg.Type != ui.BgSolid {
		t.Fatal("expected BgSolid")
	}
	wantBlue, _ := ui.ParseColor("#3A7AFE")
	assertColorApprox(t, bg.Color, wantBlue)
}

// ---------------------------------------------------------------------------
// Components wrapper tests
// ---------------------------------------------------------------------------

func TestCompile_ComponentsWrapper_Required(t *testing.T) {
	// New format: components are under "components" key.
	js := `{
		"components": {
			"label": {
				"primary": {
					"textColor": { "default": "#FF0000" }
				}
			}
		}
	}`
	theme, err := ui.LoadTheme([]byte(js))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	wantRed, _ := ui.ParseColor("#FF0000")
	assertColorApprox(t, theme.Label.Primary.TextColor.Resolve(ui.StateDefault), wantRed)
}

func TestCompile_OldFormat_StillWorks(t *testing.T) {
	// Old format: components at top level (backward compat).
	js := `{
		"label": {
			"primary": {
				"textColor": { "default": "#FF0000" }
			}
		}
	}`
	theme, err := ui.LoadTheme([]byte(js))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	wantRed, _ := ui.ParseColor("#FF0000")
	assertColorApprox(t, theme.Label.Primary.TextColor.Resolve(ui.StateDefault), wantRed)
}

// ---------------------------------------------------------------------------
// Underscore defaults tests
// ---------------------------------------------------------------------------

func TestCompile_GlobalUnderscoreDefaults(t *testing.T) {
	js := `{
		"components": {
			"_": {
				"borderWidth": 2,
				"cornerRadius": 6
			},
			"button": {
				"primary": {
					"backgroundColor": { "default": "#3A7AFE" },
					"textColor": { "default": "#FFFFFF" }
				}
			},
			"panel": {
				"primary": {
					"backgroundColor": { "default": "#262629" }
				}
			}
		}
	}`
	theme, err := ui.LoadTheme([]byte(js))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Button should inherit borderWidth and cornerRadius from components._.
	if theme.Button.Primary.BorderWidth != 2 {
		t.Errorf("button borderWidth = %v, want 2", theme.Button.Primary.BorderWidth)
	}
	if theme.Button.Primary.CornerRadius != 6 {
		t.Errorf("button cornerRadius = %v, want 6", theme.Button.Primary.CornerRadius)
	}
	// Panel should inherit too.
	if theme.Panel.Primary.BorderWidth != 2 {
		t.Errorf("panel borderWidth = %v, want 2", theme.Panel.Primary.BorderWidth)
	}
	if theme.Panel.Primary.CornerRadius != 6 {
		t.Errorf("panel cornerRadius = %v, want 6", theme.Panel.Primary.CornerRadius)
	}
}

func TestCompile_ComponentUnderscoreOverridesGlobal(t *testing.T) {
	js := `{
		"components": {
			"_": {
				"cornerRadius": 4
			},
			"button": {
				"_": {
					"cornerRadius": 8
				},
				"primary": {
					"backgroundColor": { "default": "#3A7AFE" },
					"textColor": { "default": "#FFFFFF" }
				}
			}
		}
	}`
	theme, err := ui.LoadTheme([]byte(js))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// button._ should override components._.
	if theme.Button.Primary.CornerRadius != 8 {
		t.Errorf("button cornerRadius = %v, want 8 (from button._)", theme.Button.Primary.CornerRadius)
	}
}

func TestCompile_VariantOverridesUnderscore(t *testing.T) {
	js := `{
		"components": {
			"button": {
				"_": {
					"cornerRadius": 4
				},
				"primary": {
					"backgroundColor": { "default": "#3A7AFE" },
					"textColor": { "default": "#FFFFFF" },
					"cornerRadius": 12
				}
			}
		}
	}`
	theme, err := ui.LoadTheme([]byte(js))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Explicit variant value should override button._.
	if theme.Button.Primary.CornerRadius != 12 {
		t.Errorf("button cornerRadius = %v, want 12 (explicit in primary)", theme.Button.Primary.CornerRadius)
	}
}

// ---------------------------------------------------------------------------
// Unset tests
// ---------------------------------------------------------------------------

func TestCompile_Unset_NullClearsInherited(t *testing.T) {
	js := `{
		"components": {
			"_": {
				"borderWidth": 2
			},
			"button": {
				"primary": {
					"backgroundColor": { "default": "#3A7AFE" },
					"textColor": { "default": "#FFFFFF" },
					"borderWidth": null
				}
			}
		}
	}`
	theme, err := ui.LoadTheme([]byte(js))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// borderWidth was unset via null, so it should be zero (not inherited from _).
	if theme.Button.Primary.BorderWidth != 0 {
		t.Errorf("button borderWidth = %v, want 0 (unset via null)", theme.Button.Primary.BorderWidth)
	}
}

func TestCompile_Unset_NilString(t *testing.T) {
	js := `{
		"components": {
			"_": {
				"borderWidth": 2
			},
			"button": {
				"primary": {
					"backgroundColor": { "default": "#3A7AFE" },
					"textColor": { "default": "#FFFFFF" },
					"borderWidth": "nil"
				}
			}
		}
	}`
	theme, err := ui.LoadTheme([]byte(js))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if theme.Button.Primary.BorderWidth != 0 {
		t.Errorf("button borderWidth = %v, want 0 (unset via \"nil\")", theme.Button.Primary.BorderWidth)
	}
}

func TestCompile_Unset_NoneString(t *testing.T) {
	js := `{
		"components": {
			"_": {
				"borderWidth": 2
			},
			"button": {
				"primary": {
					"backgroundColor": { "default": "#3A7AFE" },
					"textColor": { "default": "#FFFFFF" },
					"borderWidth": "none"
				}
			}
		}
	}`
	theme, err := ui.LoadTheme([]byte(js))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if theme.Button.Primary.BorderWidth != 0 {
		t.Errorf("button borderWidth = %v, want 0 (unset via \"none\")", theme.Button.Primary.BorderWidth)
	}
}

func TestCompile_Unset_EmptyString(t *testing.T) {
	js := `{
		"components": {
			"_": {
				"borderWidth": 2
			},
			"button": {
				"primary": {
					"backgroundColor": { "default": "#3A7AFE" },
					"textColor": { "default": "#FFFFFF" },
					"borderWidth": ""
				}
			}
		}
	}`
	theme, err := ui.LoadTheme([]byte(js))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if theme.Button.Primary.BorderWidth != 0 {
		t.Errorf("button borderWidth = %v, want 0 (unset via \"\")", theme.Button.Primary.BorderWidth)
	}
}

func TestCompile_Unset_PaddingNull(t *testing.T) {
	js := `{
		"components": {
			"_": {
				"padding": { "top": 8, "right": 8, "bottom": 8, "left": 8 }
			},
			"button": {
				"primary": {
					"backgroundColor": { "default": "#3A7AFE" },
					"textColor": { "default": "#FFFFFF" },
					"padding": null
				}
			}
		}
	}`
	theme, err := ui.LoadTheme([]byte(js))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Padding was unset via null — should not inherit from _.
	if theme.Button.Primary.Padding != (ui.Insets{}) {
		t.Errorf("button padding = %v, want zero (unset via null)", theme.Button.Primary.Padding)
	}
}

// ---------------------------------------------------------------------------
// Property rename tests
// ---------------------------------------------------------------------------

func TestCompile_PropertyRename_BackgroundColor(t *testing.T) {
	// "backgroundColor" is the canonical property name for backgrounds.
	js := `{
		"components": {
			"button": {
				"primary": {
					"backgroundColor": { "default": "#3A7AFE" },
					"textColor": { "default": "#FFFFFF" }
				}
			}
		}
	}`
	theme, err := ui.LoadTheme([]byte(js))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	bg := theme.Button.Primary.Background.Resolve(ui.StateDefault)
	wantBg, _ := ui.ParseColor("#3A7AFE")
	assertColorApprox(t, bg.Color, wantBg)
}

func TestCompile_PropertyRename_BorderColor(t *testing.T) {
	js := `{
		"components": {
			"panel": {
				"primary": {
					"backgroundColor": { "default": "#262629" },
					"borderColor": { "default": "#4D4D54" }
				}
			}
		}
	}`
	theme, err := ui.LoadTheme([]byte(js))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	wantBorder, _ := ui.ParseColor("#4D4D54")
	assertColorApprox(t, theme.Panel.Primary.Border.Resolve(ui.StateDefault), wantBorder)
}

func TestCompile_PropertyRename_TitleBackgroundColor(t *testing.T) {
	js := `{
		"components": {
			"window": {
				"primary": {
					"backgroundColor": { "default": "#262629" },
					"titleBackgroundColor": { "default": "#3A7AFE" },
					"titleTextColor": { "default": "#FFFFFF" },
					"resizeHandleColor": { "default": "#666670" }
				}
			}
		}
	}`
	theme, err := ui.LoadTheme([]byte(js))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	titleBg := theme.Window.Primary.TitleBackground.Resolve(ui.StateDefault)
	if titleBg.Type != ui.BgSolid {
		t.Fatal("expected BgSolid for titleBackground")
	}
	wantBg, _ := ui.ParseColor("#3A7AFE")
	assertColorApprox(t, titleBg.Color, wantBg)
}

func TestCompile_PropertyRename_BarBackgroundColor(t *testing.T) {
	js := `{
		"components": {
			"tabs": {
				"primary": {
					"barBackgroundColor": { "default": "#3A7AFE" },
					"selectedTabColor": { "default": "#FFFFFF" },
					"unselectedTabColor": { "default": "#AAAAAA" }
				}
			}
		}
	}`
	theme, err := ui.LoadTheme([]byte(js))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	barBg := theme.Tabs.Primary.BarBackground.Resolve(ui.StateDefault)
	if barBg.Type != ui.BgSolid {
		t.Fatal("expected BgSolid for barBackground")
	}
	wantBg, _ := ui.ParseColor("#3A7AFE")
	assertColorApprox(t, barBg.Color, wantBg)
}

// ---------------------------------------------------------------------------
// CollectThemeImages tests
// ---------------------------------------------------------------------------

func TestCollectThemeImages_NineGrids(t *testing.T) {
	js := `{
		"nine-grids": {
			"grid1": {
				"source": "assets/grid1.png",
				"innerRegion": { "x": 4, "y": 4, "width": 8, "height": 8 },
				"auto-slice": { "top": 4, "right": 4, "bottom": 4, "left": 4 }
			},
			"grid2": {
				"source": "assets/grid2.png",
				"innerRegion": { "x": 8, "y": 8, "width": 32, "height": 32 },
				"auto-slice": { "top": 8, "right": 8, "bottom": 8, "left": 8 }
			}
		},
		"components": {}
	}`
	paths, err := ui.CollectThemeImages([]byte(js))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(paths) != 2 {
		t.Fatalf("expected 2 paths, got %d", len(paths))
	}
	has := map[string]bool{paths[0]: true, paths[1]: true}
	if !has["assets/grid1.png"] || !has["assets/grid2.png"] {
		t.Errorf("unexpected paths: %v", paths)
	}
}

func TestCollectThemeImages_NoNineGrids(t *testing.T) {
	js := `{ "components": {} }`
	paths, err := ui.CollectThemeImages([]byte(js))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(paths) != 0 {
		t.Errorf("expected 0 paths, got %d", len(paths))
	}
}

func TestCollectThemeImages_DedupsSources(t *testing.T) {
	js := `{
		"nine-grids": {
			"grid1": {
				"source": "shared.png",
				"innerRegion": { "x": 0, "y": 0, "width": 16, "height": 16 },
				"auto-slice": { "top": 4, "right": 4, "bottom": 4, "left": 4 }
			},
			"grid2": {
				"source": "shared.png",
				"innerRegion": { "x": 0, "y": 0, "width": 16, "height": 16 },
				"auto-slice": { "top": 4, "right": 4, "bottom": 4, "left": 4 }
			}
		},
		"components": {}
	}`
	paths, err := ui.CollectThemeImages([]byte(js))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(paths) != 1 {
		t.Errorf("expected 1 deduplicated path, got %d: %v", len(paths), paths)
	}
}

func TestCollectThemeImages_MissingSource(t *testing.T) {
	js := `{
		"nine-grids": {
			"grid1": {
				"innerRegion": { "x": 0, "y": 0, "width": 16, "height": 16 },
				"auto-slice": { "top": 4, "right": 4, "bottom": 4, "left": 4 }
			}
		},
		"components": {}
	}`
	_, err := ui.CollectThemeImages([]byte(js))
	if err == nil {
		t.Fatal("expected error for missing source")
	}
	if !strings.Contains(err.Error(), "missing source") {
		t.Errorf("error should mention missing source, got: %v", err)
	}
}

// ---------------------------------------------------------------------------
// Gradient background tests
// ---------------------------------------------------------------------------

func TestCompile_GradientBackground(t *testing.T) {
	js := `{
		"components": {
			"panel": {
				"primary": {
					"backgroundColor": {
						"default": "gradient(#FF0000, #00FF00, #0000FF, #FFFF00)"
					}
				}
			}
		}
	}`
	theme, err := ui.LoadTheme([]byte(js))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	bg := theme.Panel.Primary.Background.Resolve(ui.StateDefault)
	if bg.Type != ui.BgGradient {
		t.Fatalf("expected BgGradient, got %d", bg.Type)
	}
	if bg.Gradient == nil {
		t.Fatal("Gradient should not be nil")
	}
	wantTL, _ := ui.ParseColor("#FF0000")
	wantTR, _ := ui.ParseColor("#00FF00")
	wantBR, _ := ui.ParseColor("#0000FF")
	wantBL, _ := ui.ParseColor("#FFFF00")
	assertColorApprox(t, bg.Gradient.TopLeft, wantTL)
	assertColorApprox(t, bg.Gradient.TopRight, wantTR)
	assertColorApprox(t, bg.Gradient.BottomRight, wantBR)
	assertColorApprox(t, bg.Gradient.BottomLeft, wantBL)
}

func TestCompile_GradientV(t *testing.T) {
	js := `{
		"components": {
			"panel": {
				"primary": {
					"backgroundColor": {
						"default": "gradientV(#FF0000, #0000FF)"
					}
				}
			}
		}
	}`
	theme, err := ui.LoadTheme([]byte(js))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	bg := theme.Panel.Primary.Background.Resolve(ui.StateDefault)
	if bg.Type != ui.BgGradient {
		t.Fatalf("expected BgGradient, got %d", bg.Type)
	}
	wantRed, _ := ui.ParseColor("#FF0000")
	wantBlue, _ := ui.ParseColor("#0000FF")
	// TL = TR = top color (red)
	assertColorApprox(t, bg.Gradient.TopLeft, wantRed)
	assertColorApprox(t, bg.Gradient.TopRight, wantRed)
	// BL = BR = bottom color (blue)
	assertColorApprox(t, bg.Gradient.BottomLeft, wantBlue)
	assertColorApprox(t, bg.Gradient.BottomRight, wantBlue)
}

func TestCompile_GradientH(t *testing.T) {
	js := `{
		"components": {
			"panel": {
				"primary": {
					"backgroundColor": {
						"default": "gradientH(#FF0000, #0000FF)"
					}
				}
			}
		}
	}`
	theme, err := ui.LoadTheme([]byte(js))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	bg := theme.Panel.Primary.Background.Resolve(ui.StateDefault)
	if bg.Type != ui.BgGradient {
		t.Fatalf("expected BgGradient, got %d", bg.Type)
	}
	wantRed, _ := ui.ParseColor("#FF0000")
	wantBlue, _ := ui.ParseColor("#0000FF")
	// TL = BL = left color (red)
	assertColorApprox(t, bg.Gradient.TopLeft, wantRed)
	assertColorApprox(t, bg.Gradient.BottomLeft, wantRed)
	// TR = BR = right color (blue)
	assertColorApprox(t, bg.Gradient.TopRight, wantBlue)
	assertColorApprox(t, bg.Gradient.BottomRight, wantBlue)
}

func TestCompile_GradientWithReferences(t *testing.T) {
	js := `{
		"colors": {
			"topColor": "#FF0000",
			"botColor": "#0000FF"
		},
		"components": {
			"button": {
				"primary": {
					"backgroundColor": {
						"default": "gradientV($topColor, $botColor)"
					},
					"textColor": { "default": "#FFFFFF" }
				}
			}
		}
	}`
	theme, err := ui.LoadTheme([]byte(js))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	bg := theme.Button.Primary.Background.Resolve(ui.StateDefault)
	if bg.Type != ui.BgGradient {
		t.Fatalf("expected BgGradient, got %d", bg.Type)
	}
	wantRed, _ := ui.ParseColor("#FF0000")
	wantBlue, _ := ui.ParseColor("#0000FF")
	assertColorApprox(t, bg.Gradient.TopLeft, wantRed)
	assertColorApprox(t, bg.Gradient.BottomRight, wantBlue)
}

func TestCompile_GradientWithRGBA(t *testing.T) {
	js := `{
		"components": {
			"panel": {
				"primary": {
					"backgroundColor": {
						"default": "gradient(rgba(255, 0, 0, 0.5), rgba(0, 255, 0, 1.0), rgba(0, 0, 255, 0.8), rgba(255, 255, 0, 1.0))"
					}
				}
			}
		}
	}`
	theme, err := ui.LoadTheme([]byte(js))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	bg := theme.Panel.Primary.Background.Resolve(ui.StateDefault)
	if bg.Type != ui.BgGradient {
		t.Fatalf("expected BgGradient, got %d", bg.Type)
	}
	// Check that alpha values are parsed correctly.
	if math.Abs(bg.Gradient.TopLeft.A()-0.5) > 1e-3 {
		t.Errorf("TL alpha = %f, want 0.5", bg.Gradient.TopLeft.A())
	}
	if math.Abs(bg.Gradient.BottomRight.A()-0.8) > 1e-3 {
		t.Errorf("BR alpha = %f, want 0.8", bg.Gradient.BottomRight.A())
	}
}

func TestCompile_GradientInColorProperty_Error(t *testing.T) {
	js := `{
		"components": {
			"label": {
				"primary": {
					"textColor": {
						"default": "gradientV(#FF0000, #0000FF)"
					}
				}
			}
		}
	}`
	_, err := ui.LoadTheme([]byte(js))
	if err == nil {
		t.Fatal("expected error for gradient in color property")
	}
	if !strings.Contains(err.Error(), "gradients are not supported for color properties") {
		t.Errorf("error should mention gradients not supported for color properties, got: %v", err)
	}
}

func TestParseGradient_InvalidFormat(t *testing.T) {
	tests := []string{
		"gradient(#FF0000)",                   // too few args
		"gradient(#FF0000, #00FF00)",          // still too few
		"gradient(#FF0000, #00FF00, #0000FF)", // 3 args, need 4
		"gradientV(#FF0000)",                  // too few for V
		"gradientH(#FF0000)",                  // too few for H
		"gradientX(#FF0000, #0000FF)",         // unknown prefix
		"not-a-gradient",                      // not gradient at all
	}
	for _, s := range tests {
		_, err := interntheme.ParseGradient(s, nil)
		if err == nil {
			t.Errorf("expected error for %q", s)
		}
	}
}

// ---------------------------------------------------------------------------
// Per-state float property compilation tests
// ---------------------------------------------------------------------------

func TestCompileFloatProperty_BareNumber(t *testing.T) {
	fp, errs := interntheme.CompileFloatProperty("test.offsetY", float64(3))
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	// All states should have the same value.
	for s := core.ComponentState(0); s < core.StateCount; s++ {
		if fp.Resolve(s) != 3 {
			t.Errorf("state %d: got %f, want 3", s, fp.Resolve(s))
		}
	}
}

func TestCompileFloatProperty_PerState(t *testing.T) {
	data := map[string]any{
		"default": float64(0),
		"hover":   float64(-1),
		"active":  float64(2),
	}
	fp, errs := interntheme.CompileFloatProperty("test.offsetY", data)
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	if fp.Resolve(ui.StateDefault) != 0 {
		t.Errorf("default: got %f, want 0", fp.Resolve(ui.StateDefault))
	}
	if fp.Resolve(ui.StateHover) != -1 {
		t.Errorf("hover: got %f, want -1", fp.Resolve(ui.StateHover))
	}
	if fp.Resolve(ui.StateActive) != 2 {
		t.Errorf("active: got %f, want 2", fp.Resolve(ui.StateActive))
	}
	// Disabled should fallback to default (0).
	if fp.Resolve(ui.StateDisabled) != 0 {
		t.Errorf("disabled: got %f, want 0", fp.Resolve(ui.StateDisabled))
	}
}

func TestCompileFloatProperty_MissingDefault(t *testing.T) {
	// "default" state is optional — missing it compiles cleanly with a NaN/zero fallback.
	data := map[string]any{
		"hover": float64(-1),
	}
	_, errs := interntheme.CompileFloatProperty("test.offsetY", data)
	if len(errs) != 0 {
		t.Fatalf("expected no error for missing default state, got: %v", errs)
	}
}

func TestCompileFloatProperty_InvalidType(t *testing.T) {
	_, errs := interntheme.CompileFloatProperty("test.offsetY", "not a number")
	if len(errs) == 0 {
		t.Fatal("expected error for invalid type")
	}
}

func TestCompileFloatProperty_InvalidStateValue(t *testing.T) {
	data := map[string]any{
		"default": float64(0),
		"hover":   "not-a-number",
	}
	_, errs := interntheme.CompileFloatProperty("test.offsetY", data)
	if len(errs) == 0 {
		t.Fatal("expected error for non-numeric state value")
	}
}

func TestCompileTheme_ButtonOffsetY(t *testing.T) {
	jsonStr := `{
		"components": {
			"button": {
				"primary": {
					"backgroundColor": { "default": "#333333" },
					"textColor": { "default": "#FFFFFF" },
					"offsetY": { "default": 0, "hover": -1, "active": 2 }
				}
			}
		}
	}`
	theme, err := ui.LoadTheme([]byte(jsonStr))
	if err != nil {
		t.Fatalf("LoadTheme failed: %v", err)
	}
	group := theme.Button.Group(ui.Primary)
	if group.OffsetY.Resolve(ui.StateDefault) != 0 {
		t.Errorf("default offsetY = %f, want 0", group.OffsetY.Resolve(ui.StateDefault))
	}
	if group.OffsetY.Resolve(ui.StateHover) != -1 {
		t.Errorf("hover offsetY = %f, want -1", group.OffsetY.Resolve(ui.StateHover))
	}
	if group.OffsetY.Resolve(ui.StateActive) != 2 {
		t.Errorf("active offsetY = %f, want 2", group.OffsetY.Resolve(ui.StateActive))
	}
}

func TestCompileTheme_ButtonOffsetBareNumber(t *testing.T) {
	jsonStr := `{
		"components": {
			"button": {
				"primary": {
					"backgroundColor": { "default": "#333333" },
					"textColor": { "default": "#FFFFFF" },
					"offsetX": 5
				}
			}
		}
	}`
	theme, err := ui.LoadTheme([]byte(jsonStr))
	if err != nil {
		t.Fatalf("LoadTheme failed: %v", err)
	}
	group := theme.Button.Group(ui.Primary)
	for s := core.ComponentState(0); s < core.StateCount; s++ {
		if group.OffsetX.Resolve(s) != 5 {
			t.Errorf("state %d offsetX = %f, want 5", s, group.OffsetX.Resolve(s))
		}
	}
}

func TestCompileTheme_ButtonOffsetUnderscoreCascade(t *testing.T) {
	jsonStr := `{
		"components": {
			"button": {
				"_": {
					"offsetY": { "default": 0, "hover": -2, "active": 3 }
				},
				"primary": {
					"backgroundColor": { "default": "#333333" },
					"textColor": { "default": "#FFFFFF" }
				}
			}
		}
	}`
	theme, err := ui.LoadTheme([]byte(jsonStr))
	if err != nil {
		t.Fatalf("LoadTheme failed: %v", err)
	}
	group := theme.Button.Group(ui.Primary)
	if group.OffsetY.Resolve(ui.StateHover) != -2 {
		t.Errorf("hover offsetY = %f, want -2", group.OffsetY.Resolve(ui.StateHover))
	}
	if group.OffsetY.Resolve(ui.StateActive) != 3 {
		t.Errorf("active offsetY = %f, want 3", group.OffsetY.Resolve(ui.StateActive))
	}
}

func TestCompileTheme_ButtonOffsetVariantOverride(t *testing.T) {
	jsonStr := `{
		"components": {
			"button": {
				"primary": {
					"backgroundColor": { "default": "#333333" },
					"textColor": { "default": "#FFFFFF" },
					"offsetY": { "default": 0, "hover": -1 }
				},
				"accent": {
					"backgroundColor": { "default": "#555555" },
					"textColor": { "default": "#FFFFFF" },
					"offsetY": { "default": 0, "hover": -3 }
				}
			}
		}
	}`
	theme, err := ui.LoadTheme([]byte(jsonStr))
	if err != nil {
		t.Fatalf("LoadTheme failed: %v", err)
	}
	priGroup := theme.Button.Group(ui.Primary)
	if priGroup.OffsetY.Resolve(ui.StateHover) != -1 {
		t.Errorf("primary hover offsetY = %f, want -1", priGroup.OffsetY.Resolve(ui.StateHover))
	}
	accGroup := theme.Button.Group(ui.Accent)
	if accGroup.OffsetY.Resolve(ui.StateHover) != -3 {
		t.Errorf("accent hover offsetY = %f, want -3", accGroup.OffsetY.Resolve(ui.StateHover))
	}
}

func TestCompileTheme_ButtonTextOffset(t *testing.T) {
	jsonStr := `{
		"components": {
			"button": {
				"primary": {
					"backgroundColor": { "default": "#333333" },
					"textColor": { "default": "#FFFFFF" },
					"textOffsetY": { "default": 0, "hover": -1, "active": 1 }
				}
			}
		}
	}`
	theme, err := ui.LoadTheme([]byte(jsonStr))
	if err != nil {
		t.Fatalf("LoadTheme failed: %v", err)
	}
	group := theme.Button.Group(ui.Primary)
	if group.TextOffsetY.Resolve(ui.StateDefault) != 0 {
		t.Errorf("default textOffsetY = %f, want 0", group.TextOffsetY.Resolve(ui.StateDefault))
	}
	if group.TextOffsetY.Resolve(ui.StateHover) != -1 {
		t.Errorf("hover textOffsetY = %f, want -1", group.TextOffsetY.Resolve(ui.StateHover))
	}
	if group.TextOffsetY.Resolve(ui.StateActive) != 1 {
		t.Errorf("active textOffsetY = %f, want 1", group.TextOffsetY.Resolve(ui.StateActive))
	}
}

func TestCompileTheme_ButtonTextOffsetBareNumber(t *testing.T) {
	jsonStr := `{
		"components": {
			"button": {
				"primary": {
					"backgroundColor": { "default": "#333333" },
					"textColor": { "default": "#FFFFFF" },
					"textOffsetX": 3
				}
			}
		}
	}`
	theme, err := ui.LoadTheme([]byte(jsonStr))
	if err != nil {
		t.Fatalf("LoadTheme failed: %v", err)
	}
	group := theme.Button.Group(ui.Primary)
	for s := core.ComponentState(0); s < core.StateCount; s++ {
		if group.TextOffsetX.Resolve(s) != 3 {
			t.Errorf("state %d textOffsetX = %f, want 3", s, group.TextOffsetX.Resolve(s))
		}
	}
}

func TestCompileTheme_ButtonTextOffsetVariantOverride(t *testing.T) {
	jsonStr := `{
		"components": {
			"button": {
				"primary": {
					"backgroundColor": { "default": "#333333" },
					"textColor": { "default": "#FFFFFF" },
					"textOffsetY": { "default": 0, "hover": -1 }
				},
				"accent": {
					"backgroundColor": { "default": "#555555" },
					"textColor": { "default": "#FFFFFF" },
					"textOffsetY": { "default": 0, "hover": -4 }
				}
			}
		}
	}`
	theme, err := ui.LoadTheme([]byte(jsonStr))
	if err != nil {
		t.Fatalf("LoadTheme failed: %v", err)
	}
	priGroup := theme.Button.Group(ui.Primary)
	if priGroup.TextOffsetY.Resolve(ui.StateHover) != -1 {
		t.Errorf("primary hover textOffsetY = %f, want -1", priGroup.TextOffsetY.Resolve(ui.StateHover))
	}
	accGroup := theme.Button.Group(ui.Accent)
	if accGroup.TextOffsetY.Resolve(ui.StateHover) != -4 {
		t.Errorf("accent hover textOffsetY = %f, want -4", accGroup.TextOffsetY.Resolve(ui.StateHover))
	}
}

// ---------------------------------------------------------------------------
// Helpers (local to this file)
// ---------------------------------------------------------------------------

func assertColorApprox(t *testing.T, got, want willow.Color) {
	t.Helper()
	if !colorApproxEqual(got, want) {
		t.Errorf("color = {R:%.4f G:%.4f B:%.4f A:%.4f}, want {R:%.4f G:%.4f B:%.4f A:%.4f}",
			got.R(), got.G(), got.B(), got.A(), want.R(), want.G(), want.B(), want.A())
	}
}

func assertColorHex(t *testing.T, got willow.Color, hex string) {
	t.Helper()
	want, err := ui.ParseColor(hex)
	if err != nil {
		t.Fatalf("bad test hex %q: %v", hex, err)
	}
	assertColorApprox(t, got, want)
}

func colorApproxEqual(a, b willow.Color) bool {
	const eps = 1.0 / 512
	return math.Abs(a.R()-b.R()) < eps &&
		math.Abs(a.G()-b.G()) < eps &&
		math.Abs(a.B()-b.B()) < eps &&
		math.Abs(a.A()-b.A()) < eps
}

func writeTestFile(path, content string) error {
	return os.WriteFile(path, []byte(content), 0644)
}

// Prevent unused import errors for interntheme if only ResolveFloatFallbacks is called
var _ = interntheme.ResolveFloatFallbacks
