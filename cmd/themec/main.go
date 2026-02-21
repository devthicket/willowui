// themec is the WillowUI theme compiler and coverage tool.
//
// Usage:
//
//	themec <theme-dir> -o <output.theme>      compile a theme directory
//	themec <theme-dir>                         compile (writes <theme-dir>.theme)
//	themec coverage <target.json>              check a single theme
//	themec coverage <dir/>                     check all .json in a directory
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"image"
	_ "image/png"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/devthicket/willowui/internal/theme"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "coverage" {
		if err := runCoverage(os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "themec coverage: %v\n", err)
			os.Exit(1)
		}
		return
	}

	outFlag := flag.String("o", "", "output .theme file path (default: <dir>.theme)")
	glyphsFlag := flag.String("glyphs", "", "path to default-glyphs.png (default: embedded)")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: themec [flags] <theme-dir>\n")
		fmt.Fprintf(os.Stderr, "       themec coverage <target.json|dir>\n\n")
		fmt.Fprintf(os.Stderr, "Compiles a theme directory into a single .theme binary (WUIT format).\n\n")
		fmt.Fprintf(os.Stderr, "The theme directory must contain a theme.json file.\n")
		fmt.Fprintf(os.Stderr, "Source images referenced by nine-grids and sprites are resolved\n")
		fmt.Fprintf(os.Stderr, "relative to the theme directory.\n\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	if flag.NArg() < 1 {
		flag.Usage()
		os.Exit(1)
	}

	themeDir := flag.Arg(0)
	outPath := *outFlag
	if outPath == "" {
		outPath = strings.TrimRight(themeDir, string(filepath.Separator)) + ".theme"
	}

	if err := runCompile(themeDir, outPath, *glyphsFlag); err != nil {
		fmt.Fprintf(os.Stderr, "themec: %v\n", err)
		os.Exit(1)
	}
}

// ---------------------------------------------------------------------------
// coverage subcommand
// ---------------------------------------------------------------------------

// knownWidgets is the canonical list of widget component names recognized by
// the theme compiler (from componentPropertyMaps in themecompile.go).
// knownWidgets matches the keys in componentPropertyMaps (themecompile.go).
var knownWidgets = []string{
	"accordion", "badge", "button", "calendarSelector", "checkbox",
	"colorPicker", "dataTable", "dragHandle", "iconButton", "image",
	"imageCropper", "inventory", "label", "list", "meterBar",
	"menuPopup", "optionRotator", "panel", "popover", "propertyInspector",
	"radio", "richText", "richTextEditor", "scrollBar", "select",
	"slider", "sortableList", "statWeb", "tabs", "textArea", "textInput",
	"tileList", "timePicker", "toggle", "toggleButtonBar", "toolBar",
	"tooltip", "treeList", "window",
}

var dollarRefRe = regexp.MustCompile(`\$([a-zA-Z][a-zA-Z0-9]*)`)

func runCoverage(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: themec coverage <target.json|dir>")
	}

	targetPath := args[0]

	var targets []string
	info, err := os.Stat(targetPath)
	if err != nil {
		return err
	}
	if info.IsDir() {
		entries, err := os.ReadDir(targetPath)
		if err != nil {
			return err
		}
		for _, e := range entries {
			if !e.IsDir() && strings.HasSuffix(e.Name(), ".json") {
				targets = append(targets, filepath.Join(targetPath, e.Name()))
			}
		}
		if len(targets) == 0 {
			return fmt.Errorf("no .json files found in %s", targetPath)
		}
	} else {
		targets = []string{targetPath}
	}

	exitCode := 0
	for _, t := range targets {
		issues, err := checkTheme(t)
		if err != nil {
			return fmt.Errorf("%s: %w", t, err)
		}
		name := filepath.Base(t)
		if len(issues) == 0 {
			fmt.Printf("%s: ok\n", name)
			continue
		}
		exitCode = 1
		fmt.Printf("\n%s:\n", name)
		for _, issue := range issues {
			fmt.Printf("  %s\n", issue)
		}
	}
	fmt.Println()
	if exitCode != 0 {
		os.Exit(exitCode)
	}
	return nil
}

func checkTheme(path string) ([]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, err
	}

	var issues []string

	// 1. Collect defined color tokens.
	definedColors := make(map[string]bool)
	if colorsObj, ok := raw["colors"].(map[string]any); ok {
		for name := range colorsObj {
			definedColors[name] = true
		}
	}

	// 2. Collect all $references used in component values.
	componentSource := raw
	if comps, ok := raw["components"].(map[string]any); ok {
		componentSource = comps
	}

	usedRefs := make(map[string][]string) // token -> list of paths where used
	collectRefs("", componentSource, usedRefs)
	// Also check the global defaults section.
	if globals, ok := componentSource["_"].(map[string]any); ok {
		collectRefs("_", globals, usedRefs)
	}

	// 3. Report unresolved $references.
	var unresolvedTokens []string
	for token := range usedRefs {
		if !definedColors[token] {
			unresolvedTokens = append(unresolvedTokens, token)
		}
	}
	sort.Strings(unresolvedTokens)
	for _, token := range unresolvedTokens {
		paths := usedRefs[token]
		sort.Strings(paths)
		if len(paths) <= 3 {
			issues = append(issues, fmt.Sprintf("UNRESOLVED $%s (used in %s)", token, strings.Join(paths, ", ")))
		} else {
			issues = append(issues, fmt.Sprintf("UNRESOLVED $%s (used in %d places)", token, len(paths)))
		}
	}

	// 4. Report missing widget coverage.
	var missingWidgets []string
	for _, w := range knownWidgets {
		if _, ok := componentSource[w]; !ok {
			missingWidgets = append(missingWidgets, w)
		}
	}
	if len(missingWidgets) > 0 {
		issues = append(issues, fmt.Sprintf("MISSING WIDGETS (%d): %s",
			len(missingWidgets), strings.Join(missingWidgets, ", ")))
	}

	return issues, nil
}

// collectRefs walks a JSON tree and finds all $tokenName references in string values.
func collectRefs(prefix string, obj map[string]any, out map[string][]string) {
	for k, v := range obj {
		path := k
		if prefix != "" {
			path = prefix + "." + k
		}
		switch val := v.(type) {
		case string:
			for _, match := range dollarRefRe.FindAllStringSubmatch(val, -1) {
				out[match[1]] = append(out[match[1]], path)
			}
		case map[string]any:
			collectRefs(path, val, out)
		}
	}
}

// ---------------------------------------------------------------------------
// compile subcommand (original behavior)
// ---------------------------------------------------------------------------

func runCompile(themeDir, outPath, glyphsPath string) error {
	// Read theme.json.
	jsonPath := filepath.Join(themeDir, "theme.json")
	themeJSON, err := os.ReadFile(jsonPath)
	if err != nil {
		return fmt.Errorf("read theme.json: %w", err)
	}

	// Validate JSON.
	var raw map[string]any
	if err := json.Unmarshal(themeJSON, &raw); err != nil {
		return fmt.Errorf("parse theme.json: %w", err)
	}

	// Load default glyphs spritesheet.
	var glyphsPNG []byte
	if glyphsPath != "" {
		glyphsPNG, err = os.ReadFile(glyphsPath)
		if err != nil {
			return fmt.Errorf("read glyphs: %w", err)
		}
	} else {
		// Try to find embedded glyphs at the standard path relative to repo.
		candidates := []string{
			filepath.Join(themeDir, "..", "assets", "icons", "default-glyphs.png"),
			"assets/icons/default-glyphs.png",
		}
		for _, p := range candidates {
			data, err := os.ReadFile(p)
			if err == nil {
				glyphsPNG = data
				break
			}
		}
		// Glyphs are optional — atlas will just not include default icons.
	}

	// Collect images for atlas packing.
	input := &theme.ThemeAtlasInput{
		DefaultGlyphsPNG: glyphsPNG,
		NineGridImages:   make(map[string]image.Image),
		SpriteImages:     make(map[string]image.Image),
	}

	// Load nine-grid source images.
	if gridsRaw, ok := raw["nine-grids"].(map[string]any); ok {
		srcCache := make(map[string]image.Image)
		for name, entry := range gridsRaw {
			obj, ok := entry.(map[string]any)
			if !ok {
				continue
			}
			src, _ := obj["source"].(string)
			if src == "" {
				return fmt.Errorf("nine-grids[%q]: missing source", name)
			}
			img, err := loadImageCached(filepath.Join(themeDir, src), srcCache)
			if err != nil {
				return fmt.Errorf("nine-grids[%q]: %w", name, err)
			}
			// Use the source filename as the atlas key (same as themecompile).
			input.NineGridImages[src] = img
		}
	}

	// Load sprite source images and extract sub-regions.
	if spritesRaw, ok := raw["sprites"].(map[string]any); ok {
		srcCache := make(map[string]image.Image)
		for name, entry := range spritesRaw {
			obj, ok := entry.(map[string]any)
			if !ok {
				continue
			}
			src, _ := obj["src"].(string)
			if src == "" {
				return fmt.Errorf("sprites[%q]: missing src", name)
			}

			srcImg, err := loadImageCached(filepath.Join(themeDir, src), srcCache)
			if err != nil {
				return fmt.Errorf("sprites[%q]: %w", name, err)
			}

			// Extract sub-region.
			x := intFromAny(obj["x"])
			y := intFromAny(obj["y"])
			w := intFromAny(obj["w"])
			h := intFromAny(obj["h"])
			if w == 0 || h == 0 {
				return fmt.Errorf("sprites[%q]: missing or zero w/h", name)
			}

			sub := extractSubImage(srcImg, x, y, w, h)
			input.SpriteImages[name] = sub
		}
	}

	// Pack atlas.
	var atlasJSON, atlasPNG []byte
	atlasOut, err := theme.CompileThemeAtlas(input)
	if err != nil {
		return fmt.Errorf("atlas pack: %w", err)
	}
	if atlasOut != nil {
		atlasJSON = atlasOut.AtlasJSON
		atlasPNG = atlasOut.AtlasPNG
	}

	// Encode WUIT binary.
	binary, err := theme.EncodeThemeBinary(themeJSON, atlasJSON, atlasPNG)
	if err != nil {
		return fmt.Errorf("encode binary: %w", err)
	}

	// Write output.
	if err := os.WriteFile(outPath, binary, 0644); err != nil {
		return fmt.Errorf("write output: %w", err)
	}

	aw, ah := atlasImgSize(atlasPNG)
	fmt.Printf("wrote %s (%d bytes, atlas %dx%d)\n", outPath, len(binary), aw, ah)

	return nil
}

func loadImageCached(path string, cache map[string]image.Image) (image.Image, error) {
	if img, ok := cache[path]; ok {
		return img, nil
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	img, _, err := image.Decode(f)
	if err != nil {
		return nil, err
	}
	cache[path] = img
	return img, nil
}

func extractSubImage(src image.Image, x, y, w, h int) image.Image {
	type subImager interface {
		SubImage(r image.Rectangle) image.Image
	}
	if si, ok := src.(subImager); ok {
		return si.SubImage(image.Rect(x, y, x+w, y+h))
	}
	// Fallback: copy pixels.
	dst := image.NewNRGBA(image.Rect(0, 0, w, h))
	for dy := 0; dy < h; dy++ {
		for dx := 0; dx < w; dx++ {
			dst.Set(dx, dy, src.At(x+dx, y+dy))
		}
	}
	return dst
}

func intFromAny(v any) int {
	switch n := v.(type) {
	case float64:
		return int(n)
	case int:
		return n
	}
	return 0
}

func atlasImgSize(pngData []byte) (int, int) {
	if len(pngData) == 0 {
		return 0, 0
	}
	cfg, _, err := image.DecodeConfig(bytes.NewReader(pngData))
	if err != nil {
		return 0, 0
	}
	return cfg.Width, cfg.Height
}
