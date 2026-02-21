package main

import (
	"fmt"
	"image"
	_ "image/gif"
	_ "image/png"
	"os"
	"strconv"
	"strings"
)

func main() {
	if len(os.Args) < 4 {
		fmt.Fprintf(os.Stderr, "Usage: pixelchecker <image> <x> <y> [x2 y2 ...]\n")
		fmt.Fprintf(os.Stderr, "  pixelchecker screenshot.png 160 120\n")
		fmt.Fprintf(os.Stderr, "  pixelchecker screenshot.png 160 120 20 20\n")
		os.Exit(1)
	}

	f, err := os.Open(os.Args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "open: %v\n", err)
		os.Exit(1)
	}
	defer f.Close()

	img, format, err := image.Decode(f)
	if err != nil {
		fmt.Fprintf(os.Stderr, "decode: %v\n", err)
		os.Exit(1)
	}

	bounds := img.Bounds()
	fmt.Printf("image: %s (%s) %dx%d\n", os.Args[1], format, bounds.Dx(), bounds.Dy())

	args := os.Args[2:]
	if len(args)%2 != 0 {
		fmt.Fprintf(os.Stderr, "error: coordinates must be in x y pairs\n")
		os.Exit(1)
	}

	for i := 0; i < len(args); i += 2 {
		x, err1 := strconv.Atoi(strings.TrimSpace(args[i]))
		y, err2 := strconv.Atoi(strings.TrimSpace(args[i+1]))
		if err1 != nil || err2 != nil {
			fmt.Fprintf(os.Stderr, "invalid coordinate: %s %s\n", args[i], args[i+1])
			continue
		}

		if x < bounds.Min.X || x >= bounds.Max.X || y < bounds.Min.Y || y >= bounds.Max.Y {
			fmt.Printf("(%d, %d): out of bounds\n", x, y)
			continue
		}

		r, g, b, a := img.At(x, y).RGBA()
		r8, g8, b8, a8 := r>>8, g>>8, b>>8, a>>8
		fmt.Printf("(%d, %d): rgba(%d, %d, %d, %d) #%02X%02X%02X\n", x, y, r8, g8, b8, a8, r8, g8, b8)
	}
}
