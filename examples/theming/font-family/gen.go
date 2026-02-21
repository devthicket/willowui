//go:build ignore

// gen.go bakes the Lato TTF files into a .fontbundle for this example.
//
// Run:  go generate ./examples/theming/font-family/
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
	output := filepath.Join(dir, "..", "..", "_assets", "fonts", "lato.fontbundle")

	// fontgen is a separate Go module under the willow repo.
	fontgenDir := filepath.Join(dir, "..", "..", "..", "..", "willow", "cmd", "fontgen")
	cmd := exec.Command("go", "run", ".",
		"--regular", filepath.Join(input, "Lato-Regular.ttf"),
		"--bold", filepath.Join(input, "Lato-Bold.ttf"),
		"--italic", filepath.Join(input, "Lato-Italic.ttf"),
		"--bolditalic", filepath.Join(input, "Lato-BoldItalic.ttf"),
		"--sizes=64,256",
		"--charset=ascii",
		"--width=4096",
		"--auto",
		"-o", output,
	)
	cmd.Dir = fontgenDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	log.Printf("baking Lato font bundle -> %s", output)
	if err := cmd.Run(); err != nil {
		log.Fatalf("fontgen failed: %v", err)
	}
	log.Println("done")
}
