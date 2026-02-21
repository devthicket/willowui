package widget

import (
	"image"
	"image/draw"
	"image/gif"
	"math"

	"github.com/devthicket/willowui/internal/engine"
	"github.com/devthicket/willowui/internal/sg"
)

// AnimPlayMode controls how an AnimatedImage loops.
type AnimPlayMode int

const (
	AnimPlayOnce     AnimPlayMode = iota // play once, stop at last frame
	AnimPlayLoop                         // restart from frame 0
	AnimPlayPingPong                     // reverse at each end
)

// AnimatedImage extends Image to play back a frame-strip sprite animation,
// cycling through regions at a configurable FPS.
type AnimatedImage struct {
	Component

	scaleMode ImageScaleMode
	tint      sg.Color

	// Frame storage — exactly one mode is active at a time.
	frameImages  []engine.Image     // atlas/gif mode: individual images
	frameRegions []sg.TextureRegion // region mode: texture regions
	frameDelays  []float64          // per-frame delays in seconds (nil = use fps)

	// imgNode is a willow sprite child of the component node.
	imgNode *sg.Node

	// Playback state.
	fps       float64
	playMode  AnimPlayMode
	playing   bool
	frame     int
	timer     float64 // accumulated time in seconds
	direction int     // +1 or -1 for ping-pong

	// Callbacks.
	onFrameChanged func(frame int)
	onComplete     func()

	// imgW/imgH track the pixel-space size applied to imgNode.
	imgW, imgH float64
	imgX, imgY float64
}

// NewAnimatedImage creates an AnimatedImage widget with no frames set.
func NewAnimatedImage(name string) *AnimatedImage {
	a := &AnimatedImage{
		tint:      sg.RGBA(1, 1, 1, 1),
		fps:       12,
		direction: 1,
	}
	initComponent(&a.Component, name)
	a.initBackground(name)

	a.imgNode = sg.NewSprite(name+"-img", sg.TextureRegion{})
	a.imgNode.SetVisible(false)
	a.node.AddChild(a.imgNode)

	a.onThemeChange = func() { a.UpdateVisuals() }

	// Auto-update: advance animation each frame.
	a.node.OnUpdate = func(dt float64) {
		a.Update(dt)
	}

	a.SetSize(64, 64)
	return a
}

// --- Frame sources ---

// SetAtlas auto-slices an atlas image into frames of the given size,
// scanning left-to-right, top-to-bottom. Each frame is extracted via
// SubImage so the sprite node renders one frame at a time.
func (a *AnimatedImage) SetAtlas(img engine.Image, frameWidth, frameHeight int) {
	if img == nil || frameWidth <= 0 || frameHeight <= 0 {
		return
	}
	bounds := img.Bounds()
	atlasW := bounds.Dx()
	atlasH := bounds.Dy()

	a.frameImages = nil
	a.frameRegions = nil
	a.frameDelays = nil

	for y := 0; y+frameHeight <= atlasH; y += frameHeight {
		for x := 0; x+frameWidth <= atlasW; x += frameWidth {
			sub := img.SubImage(image.Rect(
				bounds.Min.X+x, bounds.Min.Y+y,
				bounds.Min.X+x+frameWidth, bounds.Min.Y+y+frameHeight,
			)).(engine.Image)
			a.frameImages = append(a.frameImages, sub)
		}
	}

	a.frame = 0
	a.timer = 0
	a.direction = 1
	a.applyFrame()
}

// SetFrames sets an explicit list of texture regions as animation frames.
func (a *AnimatedImage) SetFrames(frames []sg.TextureRegion) {
	a.frameRegions = frames
	a.frameImages = nil
	a.frameDelays = nil
	a.frame = 0
	a.timer = 0
	a.direction = 1
	a.imgNode.SetCustomImage(nil)
	a.applyFrame()
}

// LoadGIF loads frames from a decoded GIF, using its built-in per-frame
// delays and compositing each frame onto a full canvas.
func (a *AnimatedImage) LoadGIF(g *gif.GIF) {
	if g == nil || len(g.Image) == 0 {
		return
	}

	bounds := image.Rect(0, 0, g.Config.Width, g.Config.Height)
	canvas := image.NewRGBA(bounds)
	var frames []engine.Image
	var delays []float64

	for i, frame := range g.Image {
		// Composite this frame onto the canvas.
		disposal := byte(gif.DisposalNone)
		if i < len(g.Disposal) {
			disposal = g.Disposal[i]
		}

		draw.Draw(canvas, frame.Bounds(), frame, frame.Bounds().Min, draw.Over)

		// Snapshot the composited canvas into an ebiten image.
		snap := image.NewRGBA(bounds)
		copy(snap.Pix, canvas.Pix)
		frames = append(frames, engine.NewImageFromImage(snap))

		// GIF delays are in 100ths of a second.
		delay := 0.1 // default 100ms
		if i < len(g.Delay) && g.Delay[i] > 0 {
			delay = float64(g.Delay[i]) / 100.0
		}
		delays = append(delays, delay)

		// Handle disposal for next frame.
		switch disposal {
		case gif.DisposalBackground:
			draw.Draw(canvas, frame.Bounds(), image.Transparent, image.Point{}, draw.Src)
		case gif.DisposalPrevious:
			// Simplified: treat as none (most GIFs don't use this).
		}
	}

	a.frameImages = frames
	a.frameRegions = nil
	a.frameDelays = delays
	a.frame = 0
	a.timer = 0
	a.direction = 1

	// Use GIF loop count: 0 = loop forever, >0 = play N times.
	if g.LoopCount == 0 {
		a.playMode = AnimPlayLoop
	} else {
		a.playMode = AnimPlayOnce
	}

	a.applyFrame()
}

// --- Playback ---

// SetFPS sets the playback speed in frames per second.
func (a *AnimatedImage) SetFPS(fps float64) {
	if fps <= 0 {
		fps = 1
	}
	a.fps = fps
}

// SetPlayMode sets the loop behavior.
func (a *AnimatedImage) SetPlayMode(mode AnimPlayMode) {
	a.playMode = mode
}

// Play starts or resumes playback.
func (a *AnimatedImage) Play() {
	a.playing = true
}

// Pause halts playback without resetting the frame index.
func (a *AnimatedImage) Pause() {
	a.playing = false
}

// Stop halts playback and resets to frame 0.
func (a *AnimatedImage) Stop() {
	a.playing = false
	a.frame = 0
	a.timer = 0
	a.direction = 1
	a.applyFrame()
}

// SetFrame jumps to a specific frame index.
func (a *AnimatedImage) SetFrame(n int) {
	count := a.FrameCount()
	if count == 0 {
		return
	}
	if n < 0 {
		n = 0
	}
	if n >= count {
		n = count - 1
	}
	old := a.frame
	a.frame = n
	a.applyFrame()
	if old != n && a.onFrameChanged != nil {
		a.onFrameChanged(n)
	}
}

// CurrentFrame returns the current frame index.
func (a *AnimatedImage) CurrentFrame() int { return a.frame }

// IsPlaying returns true if the animation is currently playing.
func (a *AnimatedImage) IsPlaying() bool { return a.playing }

// FrameCount returns the number of frames in the animation.
func (a *AnimatedImage) FrameCount() int {
	if len(a.frameImages) > 0 {
		return len(a.frameImages)
	}
	return len(a.frameRegions)
}

// --- Callbacks ---

// SetOnFrameChanged sets a callback fired each time the frame index changes.
func (a *AnimatedImage) SetOnFrameChanged(fn func(frame int)) {
	a.onFrameChanged = fn
}

// SetOnComplete sets a callback fired when a non-looping animation finishes.
func (a *AnimatedImage) SetOnComplete(fn func()) {
	a.onComplete = fn
}

// --- Display ---

// SetScaleMode sets how the current frame is scaled within the widget bounds.
func (a *AnimatedImage) SetScaleMode(mode ImageScaleMode) {
	a.scaleMode = mode
	a.rebuildLayout()
}

// SetTint sets the color multiplied over the image.
func (a *AnimatedImage) SetTint(c sg.Color) {
	a.tint = c
	a.imgNode.SetColor(c)
	a.MarkDrawDirty()
}

// SetCornerRadius sets the corner rounding. -1 means full pill.
func (a *AnimatedImage) SetCornerRadius(r float64) {
	cr := r
	if cr < 0 {
		cr = math.Min(a.Width, a.Height) / 2
	}
	a.applyCornerRadius(cr)
	a.MarkDrawDirty()
}

// SetSize sets the widget dimensions.
func (a *AnimatedImage) SetSize(w, h float64) {
	a.Width = w
	a.Height = h
	a.resizeBackground(w, h)
	a.node.HitShape = sg.HitRect{X: 0, Y: 0, Width: w, Height: h}
	a.rebuildLayout()
	a.MarkLayoutDirty()
}

// UpdateVisuals applies theme colors based on current state.
func (a *AnimatedImage) UpdateVisuals() {
	a.state = computeState(a.enabled, a.focused, a.hovered, a.pressed)
	group := a.EffectiveTheme().AnimatedImage.Group(a.Variant())

	bg := group.Background.Resolve(a.state)
	a.applyBackground(bg)

	cr := resolveCornerRadius(group.CornerRadius, math.Min(a.Width, a.Height))
	a.applyCornerRadius(cr)

	a.MarkDrawDirty()
}

// --- Internal ---

// Update advances the animation by dt seconds. Called automatically via OnUpdate.
func (a *AnimatedImage) Update(dt float64) {
	if !a.playing || a.FrameCount() <= 1 {
		return
	}

	a.timer += dt

	for {
		frameDur := a.currentFrameDuration()
		if a.timer < frameDur {
			break
		}
		a.timer -= frameDur
		a.advanceFrame()
		if !a.playing {
			break
		}
	}
}

// currentFrameDuration returns the delay for the current frame in seconds.
func (a *AnimatedImage) currentFrameDuration() float64 {
	if a.frameDelays != nil && a.frame < len(a.frameDelays) {
		return a.frameDelays[a.frame]
	}
	return 1.0 / a.fps
}

// advanceFrame moves to the next frame according to the play mode.
func (a *AnimatedImage) advanceFrame() {
	n := a.FrameCount()
	if n <= 1 {
		return
	}

	old := a.frame
	next := a.frame + a.direction

	switch a.playMode {
	case AnimPlayOnce:
		if next >= n {
			a.frame = n - 1
			a.playing = false
			a.applyFrame()
			if old != a.frame && a.onFrameChanged != nil {
				a.onFrameChanged(a.frame)
			}
			if a.onComplete != nil {
				a.onComplete()
			}
			return
		}
		a.frame = next

	case AnimPlayLoop:
		if next >= n {
			a.frame = 0
		} else {
			a.frame = next
		}

	case AnimPlayPingPong:
		if next >= n {
			a.direction = -1
			a.frame = n - 2
			if a.frame < 0 {
				a.frame = 0
			}
		} else if next < 0 {
			a.direction = 1
			a.frame = 1
			if a.frame >= n {
				a.frame = 0
			}
		} else {
			a.frame = next
		}
	}

	if old != a.frame {
		a.applyFrame()
		if a.onFrameChanged != nil {
			a.onFrameChanged(a.frame)
		}
	}
}

// applyFrame sets the imgNode to display the current frame.
func (a *AnimatedImage) applyFrame() {
	count := a.FrameCount()
	if count == 0 {
		a.imgNode.SetVisible(false)
		return
	}
	a.imgNode.SetVisible(true)

	if len(a.frameImages) > 0 {
		// Atlas mode: swap the custom image to the sub-image for this frame.
		a.imgNode.SetCustomImage(a.frameImages[a.frame])
		a.imgNode.SetTextureRegion(sg.TextureRegion{})
	} else {
		// Region mode: swap the texture region.
		a.imgNode.SetCustomImage(nil)
		a.imgNode.SetTextureRegion(a.frameRegions[a.frame])
	}

	a.rebuildLayout()
}

// nativeDimensions returns the native pixel size of the current frame.
func (a *AnimatedImage) nativeDimensions() (w, h float64) {
	if len(a.frameImages) > 0 && a.frame < len(a.frameImages) {
		b := a.frameImages[a.frame].Bounds()
		return float64(b.Dx()), float64(b.Dy())
	}
	if len(a.frameRegions) > 0 && a.frame < len(a.frameRegions) {
		fr := a.frameRegions[a.frame]
		return float64(fr.Width), float64(fr.Height)
	}
	return 0, 0
}

// rebuildLayout repositions and scales imgNode for the current frame.
func (a *AnimatedImage) rebuildLayout() {
	imgW, imgH := a.nativeDimensions()
	if imgW <= 0 || imgH <= 0 {
		a.MarkDrawDirty()
		return
	}

	widW := a.Width
	widH := a.Height

	var applyW, applyH, applyX, applyY float64

	switch a.scaleMode {
	case ImageScaleStretch:
		applyW, applyH = widW, widH
		applyX, applyY = 0, 0

	case ImageScaleFit:
		s := math.Min(widW/imgW, widH/imgH)
		applyW = imgW * s
		applyH = imgH * s
		applyX = (widW - applyW) / 2
		applyY = (widH - applyH) / 2

	case ImageScaleFill:
		s := math.Max(widW/imgW, widH/imgH)
		applyW = imgW * s
		applyH = imgH * s
		applyX = (widW - applyW) / 2
		applyY = (widH - applyH) / 2

	case ImageScaleCenter:
		applyW, applyH = imgW, imgH
		applyX = (widW - imgW) / 2
		applyY = (widH - imgH) / 2

	default:
		applyW, applyH = widW, widH
		applyX, applyY = 0, 0
	}

	a.imgW, a.imgH = applyW, applyH
	a.imgX, a.imgY = applyX, applyY
	a.imgNode.SetScale(applyW/imgW, applyH/imgH)
	a.imgNode.SetPosition(applyX, applyY)
	a.MarkDrawDirty()
}
