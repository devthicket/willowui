//go:build ignore

// gen.go bakes the Lato and Noto Serif TTF files into .fontbundle archives.
//
// Run:  go generate ./examples/widgets/richtext-mixed-fonts/
package main

import (
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

func main() {
	_, src, _, _ := runtime.Caller(0)
	dir := filepath.Dir(src)
	input := filepath.Join(dir, "..", "..", "_assets", "fonts", "input")
	fontsDir := filepath.Join(dir, "..", "..", "_assets", "fonts")

	// fontgen is a separate Go module under the willow repo.
	fontgenDir := filepath.Join(dir, "..", "..", "..", "..", "willow", "cmd", "fontgen")

	// Bake Lato.
	bake(fontgenDir, "Lato", filepath.Join(fontsDir, "lato.fontbundle"),
		filepath.Join(input, "Lato-Regular.ttf"),
		filepath.Join(input, "Lato-Bold.ttf"),
		filepath.Join(input, "Lato-Italic.ttf"),
		filepath.Join(input, "Lato-BoldItalic.ttf"),
	)

	// Bake Noto Serif.
	bake(fontgenDir, "Noto Serif", filepath.Join(fontsDir, "notoserif.fontbundle"),
		filepath.Join(input, "NotoSerif-Regular.ttf"),
		filepath.Join(input, "NotoSerif-Bold.ttf"),
		filepath.Join(input, "NotoSerif-Italic.ttf"),
		filepath.Join(input, "NotoSerif-BoldItalic.ttf"),
	)
}

func bake(fontgenDir, name, output, regular, bold, italic, bolditalic string) {
	cmd := exec.Command("go", "run", ".",
		"--regular", regular,
		"--bold", bold,
		"--italic", italic,
		"--bolditalic", bolditalic,
		"--sizes=256",
		"--charset=ascii",
		"--auto",
		"-o", output,
	)
	cmd.Dir = fontgenDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	log.Printf("baking %s -> %s", name, output)
	if err := cmd.Run(); err != nil {
		log.Fatalf("fontgen %s failed: %v", name, err)
	}
	log.Println("done")
}
