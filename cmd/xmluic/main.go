// xmluic compiles WillowUI assets.
//
// Usage:
//
//	xmluic compile input.xml output.xmlui    # compile XML template to binary
//	xmluic input.xml output.xmlui            # same (backward compatible)
//	xmluic atlas theme.json --out build      # pack nine-slice images into atlas
package main

import (
	"encoding/json"
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"

	ui "github.com/devthicket/willowui"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "compile":
		if len(os.Args) != 4 {
			fmt.Fprintf(os.Stderr, "usage: xmluic compile <input.xml> <output.xmlui>\n")
			os.Exit(1)
		}
		runCompile(os.Args[2], os.Args[3])

	case "atlas":
		runAtlas(os.Args[2:])

	default:
		// Backward compatible: xmluic input.xml output.xmlui
		if len(os.Args) == 3 {
			runCompile(os.Args[1], os.Args[2])
			return
		}
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Fprintf(os.Stderr, `usage:
  xmluic compile <input.xml> <output.xmlui>
  xmluic <input.xml> <output.xmlui>            (backward compatible)
  xmluic atlas <theme.json> --out <dir>
`)
}

func runCompile(xmlPath, outPath string) {
	xmlData, err := os.ReadFile(xmlPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "read %s: %v\n", xmlPath, err)
		os.Exit(1)
	}

	ir, err := ui.CompileXML(xmlData)
	if err != nil {
		fmt.Fprintf(os.Stderr, "compile %s: %v\n", xmlPath, err)
		os.Exit(1)
	}

	binData, err := ui.EncodeIR(ir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "encode: %v\n", err)
		os.Exit(1)
	}

	if err := os.WriteFile(outPath, binData, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "write %s: %v\n", outPath, err)
		os.Exit(1)
	}

	fmt.Printf("compiled %s -> %s (%d bytes)\n", xmlPath, outPath, len(binData))
}

func runAtlas(args []string) {
	if len(args) < 1 {
		fmt.Fprintf(os.Stderr, "usage: xmluic atlas <theme.json> --out <dir>\n")
		os.Exit(1)
	}

	themePath := args[0]
	outDir := "."

	// Parse --out flag.
	for i := 1; i < len(args); i++ {
		if args[i] == "--out" && i+1 < len(args) {
			outDir = args[i+1]
			i++
		}
	}

	// Read theme JSON.
	data, err := os.ReadFile(themePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "read %s: %v\n", themePath, err)
		os.Exit(1)
	}

	// Collect image paths.
	imagePaths, err := ui.CollectThemeImages(data)
	if err != nil {
		fmt.Fprintf(os.Stderr, "collect images: %v\n", err)
		os.Exit(1)
	}

	if len(imagePaths) == 0 {
		fmt.Fprintln(os.Stderr, "no nine-slice images found in theme")
		os.Exit(1)
	}

	// Load theme (loads + packs images).
	theme, err := ui.LoadThemeFromFile(themePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "load theme: %v\n", err)
		os.Exit(1)
	}

	if theme.Atlas == nil {
		fmt.Fprintln(os.Stderr, "no atlas produced (no images)")
		os.Exit(1)
	}

	// Ensure output directory exists.
	if err := os.MkdirAll(outDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "mkdir %s: %v\n", outDir, err)
		os.Exit(1)
	}

	// Write atlas page images.
	pages := theme.Atlas.Pages
	pageFiles := make([]string, len(pages))
	for i, page := range pages {
		name := fmt.Sprintf("atlas-%d.png", i)
		pageFiles[i] = name
		pngPath := filepath.Join(outDir, name)

		f, err := os.Create(pngPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "create %s: %v\n", pngPath, err)
			os.Exit(1)
		}
		// Read pixels back from the GPU into a standard RGBA image for encoding.
		bounds := page.Bounds()
		rgba := image.NewRGBA(bounds)
		page.ReadPixels(rgba.Pix)
		if err := png.Encode(f, rgba); err != nil {
			f.Close()
			fmt.Fprintf(os.Stderr, "encode %s: %v\n", pngPath, err)
			os.Exit(1)
		}
		f.Close()
		fmt.Printf("wrote %s\n", pngPath)
	}

	// Build TexturePacker-compatible JSON.
	atlasJSON := buildAtlasJSON(theme, imagePaths, pageFiles)
	jsonPath := filepath.Join(outDir, "atlas.json")
	if err := os.WriteFile(jsonPath, atlasJSON, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "write %s: %v\n", jsonPath, err)
		os.Exit(1)
	}
	fmt.Printf("wrote %s\n", jsonPath)
	fmt.Printf("atlas: %d image(s), %d page(s)\n", len(imagePaths), len(pages))
}

// buildAtlasJSON generates a TexturePacker-compatible JSON atlas descriptor.
func buildAtlasJSON(theme *ui.Theme, imagePaths []string, pageFiles []string) []byte {
	type frame struct {
		X int `json:"x"`
		Y int `json:"y"`
		W int `json:"w"`
		H int `json:"h"`
	}
	type spriteEntry struct {
		Frame            frame `json:"frame"`
		Rotated          bool  `json:"rotated"`
		Trimmed          bool  `json:"trimmed"`
		SpriteSourceSize frame `json:"spriteSourceSize"`
		SourceSize       struct {
			W int `json:"w"`
			H int `json:"h"`
		} `json:"sourceSize"`
	}

	// For single-page atlas, use hash format.
	// For multi-page, use array format.
	if len(pageFiles) <= 1 {
		frames := make(map[string]spriteEntry)
		for _, name := range imagePaths {
			r := theme.Atlas.Region(name)
			entry := spriteEntry{
				Frame:            frame{X: int(r.X), Y: int(r.Y), W: int(r.Width), H: int(r.Height)},
				Rotated:          r.Rotated,
				SpriteSourceSize: frame{X: int(r.OffsetX), Y: int(r.OffsetY), W: int(r.Width), H: int(r.Height)},
			}
			entry.SourceSize.W = int(r.OriginalW)
			entry.SourceSize.H = int(r.OriginalH)
			frames[name] = entry
		}

		type hashAtlas struct {
			Frames map[string]spriteEntry `json:"frames"`
			Meta   struct {
				Image string `json:"image"`
			} `json:"meta"`
		}
		out := hashAtlas{Frames: frames}
		if len(pageFiles) > 0 {
			out.Meta.Image = pageFiles[0]
		}
		data, _ := json.MarshalIndent(out, "", "    ")
		return data
	}

	// Multi-page: group by page.
	type texturePage struct {
		Image  string                 `json:"image"`
		Frames map[string]spriteEntry `json:"frames"`
	}
	pages := make([]texturePage, len(pageFiles))
	for i, pf := range pageFiles {
		pages[i] = texturePage{Image: pf, Frames: make(map[string]spriteEntry)}
	}
	for _, name := range imagePaths {
		r := theme.Atlas.Region(name)
		pageIdx := int(r.Page)
		if pageIdx >= len(pages) {
			continue
		}
		entry := spriteEntry{
			Frame:            frame{X: int(r.X), Y: int(r.Y), W: int(r.Width), H: int(r.Height)},
			Rotated:          r.Rotated,
			SpriteSourceSize: frame{X: int(r.OffsetX), Y: int(r.OffsetY), W: int(r.Width), H: int(r.Height)},
		}
		entry.SourceSize.W = int(r.OriginalW)
		entry.SourceSize.H = int(r.OriginalH)
		pages[pageIdx].Frames[name] = entry
	}

	type arrayAtlas struct {
		Textures []texturePage `json:"textures"`
	}
	data, _ := json.MarshalIndent(arrayAtlas{Textures: pages}, "", "    ")
	return data
}
