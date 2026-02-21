//go:build ignore

// Generates the three sprite sheet PNGs used by the AnimatedImage example.
// Run: go run gen.go
package main

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
)

const (
	frameSize = 64
	numFrames = 16
)

func main() {
	save("color-strip.png", makeColorStrip(numFrames, frameSize))
	save("pulse-strip.png", makePulseStrip(numFrames, frameSize))
	save("spinner-strip.png", makeSpinnerStrip(numFrames, frameSize))
	fmt.Println("wrote color-strip.png, pulse-strip.png, spinner-strip.png")
}

func save(name string, img image.Image) {
	f, err := os.Create(name)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	if err := png.Encode(f, img); err != nil {
		panic(err)
	}
}

func makeColorStrip(frames, size int) *image.NRGBA {
	img := image.NewNRGBA(image.Rect(0, 0, frames*size, size))
	for i := 0; i < frames; i++ {
		hue := float64(i) / float64(frames) * 360
		r, g, b := hsvToRGB(hue, 0.85, 0.95)
		for y := 0; y < size; y++ {
			for x := 0; x < size; x++ {
				img.SetNRGBA(i*size+x, y, color.NRGBA{r, g, b, 255})
			}
		}
	}
	return img
}

func makePulseStrip(frames, size int) *image.NRGBA {
	img := image.NewNRGBA(image.Rect(0, 0, frames*size, size))
	cx, cy := float64(size)/2, float64(size)/2
	maxR := float64(size)/2 - 6

	for i := 0; i < frames; i++ {
		t := float64(i) / float64(frames)
		radius := maxR * (0.35 + 0.65*t)
		glowR := radius + 6

		for y := 0; y < size; y++ {
			for x := 0; x < size; x++ {
				dx := float64(x) - cx
				dy := float64(y) - cy
				dist := math.Sqrt(dx*dx + dy*dy)
				px := i*size + x

				if dist <= radius {
					f := dist / radius
					r := uint8(40 + 60*f)
					g := uint8(180 + 60*(1-f))
					b := uint8(230 + 25*(1-f))
					img.SetNRGBA(px, y, color.NRGBA{r, g, b, 255})
				} else if dist <= glowR {
					alpha := 1.0 - (dist-radius)/(glowR-radius)
					a := uint8(120 * alpha * alpha)
					img.SetNRGBA(px, y, color.NRGBA{80, 200, 255, a})
				}
			}
		}
	}
	return img
}

func makeSpinnerStrip(frames, size int) *image.NRGBA {
	img := image.NewNRGBA(image.Rect(0, 0, frames*size, size))
	cx, cy := float64(size)/2, float64(size)/2
	outerR := float64(size)/2 - 6
	innerR := outerR - 8

	for i := 0; i < frames; i++ {
		baseAngle := float64(i) / float64(frames) * 2 * math.Pi
		arcLen := math.Pi * 0.75

		for y := 0; y < size; y++ {
			for x := 0; x < size; x++ {
				dx := float64(x) - cx
				dy := float64(y) - cy
				dist := math.Sqrt(dx*dx + dy*dy)
				px := i*size + x

				if dist >= innerR && dist <= outerR {
					img.SetNRGBA(px, y, color.NRGBA{35, 35, 50, 255})
				}

				if dist >= innerR && dist <= outerR {
					angle := math.Atan2(dy, dx)
					if angle < 0 {
						angle += 2 * math.Pi
					}
					rel := angle - baseAngle
					for rel < 0 {
						rel += 2 * math.Pi
					}
					for rel >= 2*math.Pi {
						rel -= 2 * math.Pi
					}
					if rel <= arcLen {
						f := rel / arcLen
						r := uint8(120 + 135*(1-f))
						g := uint8(220 + 35*(1-f))
						b := uint8(140 + 60*(1-f))
						a := uint8(255 - 100*f)
						img.SetNRGBA(px, y, color.NRGBA{r, g, b, a})
					}
				}
			}
		}
	}
	return img
}

func hsvToRGB(h, s, v float64) (uint8, uint8, uint8) {
	h = math.Mod(h, 360)
	c := v * s
	x := c * (1 - math.Abs(math.Mod(h/60, 2)-1))
	m := v - c
	var r, g, b float64
	switch {
	case h < 60:
		r, g, b = c, x, 0
	case h < 120:
		r, g, b = x, c, 0
	case h < 180:
		r, g, b = 0, c, x
	case h < 240:
		r, g, b = 0, x, c
	case h < 300:
		r, g, b = x, 0, c
	default:
		r, g, b = c, 0, x
	}
	return uint8((r + m) * 255), uint8((g + m) * 255), uint8((b + m) * 255)
}
