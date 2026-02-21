package markup

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/devthicket/willowui/internal/sg"
)

// ParseColor parses a color string in any supported format:
// hex 6-digit (#3A7AFE), hex 8-digit (#3A7AFE80), hex 3-digit (#38F),
// rgba(r,g,b,a), rgb(r,g,b), or named (white, black, transparent).
func ParseColor(s string) (sg.Color, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return sg.Color{}, fmt.Errorf("empty color string")
	}

	// Named colors.
	switch strings.ToLower(s) {
	case "white":
		return sg.RGBA(1, 1, 1, 1), nil
	case "black":
		return sg.RGBA(0, 0, 0, 1), nil
	case "transparent":
		return sg.RGBA(0, 0, 0, 0), nil
	}

	// Hex formats.
	if strings.HasPrefix(s, "#") {
		return parseHexColor(s)
	}

	// rgba() / rgb().
	lower := strings.ToLower(s)
	if strings.HasPrefix(lower, "rgba(") && strings.HasSuffix(lower, ")") {
		return parseRGBA(s[5 : len(s)-1])
	}
	if strings.HasPrefix(lower, "rgb(") && strings.HasSuffix(lower, ")") {
		return parseRGB(s[4 : len(s)-1])
	}

	return sg.Color{}, fmt.Errorf("invalid color format %q", s)
}

func parseHexColor(s string) (sg.Color, error) {
	hex := s[1:]
	switch len(hex) {
	case 3:
		// Expand #RGB → #RRGGBB
		r, err := strconv.ParseUint(string([]byte{hex[0], hex[0]}), 16, 8)
		if err != nil {
			return sg.Color{}, fmt.Errorf("invalid color format %q", s)
		}
		g, err := strconv.ParseUint(string([]byte{hex[1], hex[1]}), 16, 8)
		if err != nil {
			return sg.Color{}, fmt.Errorf("invalid color format %q", s)
		}
		b, err := strconv.ParseUint(string([]byte{hex[2], hex[2]}), 16, 8)
		if err != nil {
			return sg.Color{}, fmt.Errorf("invalid color format %q", s)
		}
		return sg.RGBA(float64(r)/255, float64(g)/255, float64(b)/255, 1), nil

	case 6:
		return parseHex6(hex, s)

	case 8:
		c, err := parseHex6(hex[:6], s)
		if err != nil {
			return sg.Color{}, err
		}
		a, err := strconv.ParseUint(hex[6:8], 16, 8)
		if err != nil {
			return sg.Color{}, fmt.Errorf("invalid color format %q", s)
		}
		return sg.RGBA(c.R(), c.G(), c.B(), float64(a)/255), nil

	default:
		return sg.Color{}, fmt.Errorf("invalid color format %q", s)
	}
}

func parseHex6(hex, orig string) (sg.Color, error) {
	r, err := strconv.ParseUint(hex[0:2], 16, 8)
	if err != nil {
		return sg.Color{}, fmt.Errorf("invalid color format %q", orig)
	}
	g, err := strconv.ParseUint(hex[2:4], 16, 8)
	if err != nil {
		return sg.Color{}, fmt.Errorf("invalid color format %q", orig)
	}
	b, err := strconv.ParseUint(hex[4:6], 16, 8)
	if err != nil {
		return sg.Color{}, fmt.Errorf("invalid color format %q", orig)
	}
	return sg.RGBA(float64(r)/255, float64(g)/255, float64(b)/255, 1), nil
}

func parseRGBA(inner string) (sg.Color, error) {
	parts := strings.Split(inner, ",")
	if len(parts) != 4 {
		return sg.Color{}, fmt.Errorf("invalid rgba() format: expected 4 values, got %d", len(parts))
	}
	r, err := strconv.Atoi(strings.TrimSpace(parts[0]))
	if err != nil || r < 0 || r > 255 {
		return sg.Color{}, fmt.Errorf("invalid rgba() red value: %s", strings.TrimSpace(parts[0]))
	}
	g, err := strconv.Atoi(strings.TrimSpace(parts[1]))
	if err != nil || g < 0 || g > 255 {
		return sg.Color{}, fmt.Errorf("invalid rgba() green value: %s", strings.TrimSpace(parts[1]))
	}
	b, err := strconv.Atoi(strings.TrimSpace(parts[2]))
	if err != nil || b < 0 || b > 255 {
		return sg.Color{}, fmt.Errorf("invalid rgba() blue value: %s", strings.TrimSpace(parts[2]))
	}
	a, err := strconv.ParseFloat(strings.TrimSpace(parts[3]), 64)
	if err != nil || a < 0 || a > 1 {
		return sg.Color{}, fmt.Errorf("invalid rgba() alpha value: %s", strings.TrimSpace(parts[3]))
	}
	return sg.RGBA(float64(r)/255, float64(g)/255, float64(b)/255, a), nil
}

func parseRGB(inner string) (sg.Color, error) {
	parts := strings.Split(inner, ",")
	if len(parts) != 3 {
		return sg.Color{}, fmt.Errorf("invalid rgb() format: expected 3 values, got %d", len(parts))
	}
	r, err := strconv.Atoi(strings.TrimSpace(parts[0]))
	if err != nil || r < 0 || r > 255 {
		return sg.Color{}, fmt.Errorf("invalid rgb() red value: %s", strings.TrimSpace(parts[0]))
	}
	g, err := strconv.Atoi(strings.TrimSpace(parts[1]))
	if err != nil || g < 0 || g > 255 {
		return sg.Color{}, fmt.Errorf("invalid rgb() green value: %s", strings.TrimSpace(parts[1]))
	}
	b, err := strconv.Atoi(strings.TrimSpace(parts[2]))
	if err != nil || b < 0 || b > 255 {
		return sg.Color{}, fmt.Errorf("invalid rgb() blue value: %s", strings.TrimSpace(parts[2]))
	}
	return sg.RGBA(float64(r)/255, float64(g)/255, float64(b)/255, 1), nil
}
