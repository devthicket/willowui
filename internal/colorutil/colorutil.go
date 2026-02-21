package colorutil

import (
	"fmt"
	"math"
	"strings"

	"github.com/devthicket/willowui/internal/render"
	"github.com/devthicket/willowui/internal/sg"
)

// ── Hex ──────────────────────────────────────────────────────────────────────

// ParseHex parses "#RRGGBB" or "#RRGGBBAA". Returns ok=false on invalid input.
func ParseHex(s string) (sg.Color, bool) {
	s = strings.TrimPrefix(s, "#")
	s = strings.ToUpper(s)

	var r, g, b, a uint8
	switch len(s) {
	case 6:
		n, err := fmt.Sscanf(s, "%02X%02X%02X", &r, &g, &b)
		if err != nil || n != 3 {
			return sg.Color{}, false
		}
		a = 255
	case 8:
		n, err := fmt.Sscanf(s, "%02X%02X%02X%02X", &r, &g, &b, &a)
		if err != nil || n != 4 {
			return sg.Color{}, false
		}
	default:
		return sg.Color{}, false
	}

	return sg.RGBA(float64(r)/255.0, float64(g)/255.0, float64(b)/255.0, float64(a)/255.0), true
}

// FormatHex returns "#RRGGBB" (alpha ignored).
func FormatHex(c sg.Color) string {
	r, g, b, _ := ToRGB255(c)
	return fmt.Sprintf("#%02X%02X%02X", r, g, b)
}

// FormatHexA returns "#RRGGBBAA".
func FormatHexA(c sg.Color) string {
	r, g, b, a := ToRGB255(c)
	return fmt.Sprintf("#%02X%02X%02X%02X", r, g, b, a)
}

// ── RGB 0–255 ─────────────────────────────────────────────────────────────────

// ToRGB255 converts a sg.Color to 0–255 integer components.
// Values are clamped to [0, 255].
func ToRGB255(c sg.Color) (r, g, b, a int) {
	return clampInt(int(math.Round(c.R()*255)), 0, 255),
		clampInt(int(math.Round(c.G()*255)), 0, 255),
		clampInt(int(math.Round(c.B()*255)), 0, 255),
		clampInt(int(math.Round(c.A()*255)), 0, 255)
}

func clampInt(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

// FromRGB255 converts 0–255 integers to a sg.Color.
func FromRGB255(r, g, b, a int) sg.Color {
	return sg.RGBA(float64(r)/255.0, float64(g)/255.0, float64(b)/255.0, float64(a)/255.0)
}

// NormalizeRGB clamps R, G, B to [0, 1] for use with standard color-space
// conversions. Alpha is passed through unchanged. This preserves the hue and
// saturation ratio of overbright tint colors.
func NormalizeRGB(c sg.Color) sg.Color {
	r, g, b := c.R(), c.G(), c.B()
	maxC := math.Max(r, math.Max(g, b))
	if maxC <= 1 {
		return sg.RGBA(
			math.Max(0, r),
			math.Max(0, g),
			math.Max(0, b),
			c.A(),
		)
	}
	// Scale down proportionally so max = 1, preserving hue/saturation.
	return sg.RGBA(r/maxC, g/maxC, b/maxC, c.A())
}

// ── HSV ───────────────────────────────────────────────────────────────────────

// ToHSV converts a sg.Color to HSV. All return values are in [0,1].
func ToHSV(c sg.Color) (h, s, v, a float64) {
	r, g, b := c.R(), c.G(), c.B()
	a = c.A()

	maxC := math.Max(r, math.Max(g, b))
	minC := math.Min(r, math.Min(g, b))
	delta := maxC - minC

	v = maxC

	if delta == 0 {
		h = 0
		s = 0
		return
	}

	s = delta / maxC

	switch maxC {
	case r:
		h = (g - b) / delta
		if h < 0 {
			h += 6
		}
	case g:
		h = (b-r)/delta + 2
	case b:
		h = (r-g)/delta + 4
	}
	h /= 6

	return
}

// FromHSV wraps sg.ColorFromHSV and adds alpha support. All inputs in [0,1].
func FromHSV(h, s, v, a float64) sg.Color {
	c := sg.ColorFromHSV(h, s, v)
	return sg.RGBA(c.R(), c.G(), c.B(), a)
}

// ── HSL ───────────────────────────────────────────────────────────────────────

// ToHSL converts a sg.Color to HSL. All return values are in [0,1].
func ToHSL(c sg.Color) (h, s, l, a float64) {
	r, g, b := c.R(), c.G(), c.B()
	a = c.A()

	maxC := math.Max(r, math.Max(g, b))
	minC := math.Min(r, math.Min(g, b))
	l = (maxC + minC) / 2.0

	delta := maxC - minC
	if delta == 0 {
		h = 0
		s = 0
		return
	}

	if l < 0.5 {
		s = delta / (maxC + minC)
	} else {
		s = delta / (2.0 - maxC - minC)
	}

	switch maxC {
	case r:
		h = (g - b) / delta
		if h < 0 {
			h += 6
		}
	case g:
		h = (b-r)/delta + 2
	case b:
		h = (r-g)/delta + 4
	}
	h /= 6

	return
}

// FromHSL converts HSL (all [0,1]) plus alpha to a sg.Color.
func FromHSL(h, s, l, a float64) sg.Color {
	if s == 0 {
		return sg.RGBA(l, l, l, a)
	}

	var q float64
	if l < 0.5 {
		q = l * (1.0 + s)
	} else {
		q = l + s - l*s
	}
	p := 2.0*l - q

	r := hueToRGB(p, q, h+1.0/3.0)
	g := hueToRGB(p, q, h)
	b := hueToRGB(p, q, h-1.0/3.0)

	return sg.RGBA(r, g, b, a)
}

// ── Gradient utilities ────────────────────────────────────────────────────────

// SampleBilinear returns the bilinearly interpolated color at normalized (u, v)
// where (0,0)=TopLeft, (1,1)=BottomRight.
func SampleBilinear(g render.GradientColors, u, v float64) sg.Color {
	return render.BilinearColor(&g, u, v)
}

// FormatGradientString returns the theme-compatible JSON fill string for g,
// using the most compact form: gradientH/gradientV when applicable, else gradient.
func FormatGradientString(g render.Gradient) string {
	c := g.Colors
	switch g.Mode {
	case render.GradientModeH:
		return fmt.Sprintf("gradientH(%s, %s)", FormatHex(c.TopLeft), FormatHex(c.TopRight))
	case render.GradientModeV:
		return fmt.Sprintf("gradientV(%s, %s)", FormatHex(c.TopLeft), FormatHex(c.BottomLeft))
	default:
		return fmt.Sprintf("gradient(%s, %s, %s, %s)",
			FormatHex(c.TopLeft), FormatHex(c.TopRight),
			FormatHex(c.BottomRight), FormatHex(c.BottomLeft))
	}
}

// DefaultGradient returns a horizontal black→white gradient.
func DefaultGradient() render.Gradient {
	return render.Gradient{
		Mode: render.GradientModeH,
		Colors: render.GradientColors{
			TopLeft:     sg.RGBA(0, 0, 0, 1),
			TopRight:    sg.RGBA(1, 1, 1, 1),
			BottomRight: sg.RGBA(1, 1, 1, 1),
			BottomLeft:  sg.RGBA(0, 0, 0, 1),
		},
	}
}

func hueToRGB(p, q, t float64) float64 {
	if t < 0 {
		t += 1
	}
	if t > 1 {
		t -= 1
	}
	switch {
	case t < 1.0/6.0:
		return p + (q-p)*6.0*t
	case t < 1.0/2.0:
		return q
	case t < 2.0/3.0:
		return p + (q-p)*(2.0/3.0-t)*6.0
	default:
		return p
	}
}
