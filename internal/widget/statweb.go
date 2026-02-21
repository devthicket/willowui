package widget

import (
	"math"

	"github.com/devthicket/willowui/internal/engine"
	"github.com/devthicket/willowui/internal/sg"
)

// StatAxis defines a single spoke on a StatWeb.
type StatAxis struct {
	Name  string
	Min   float64
	Max   float64
	Value float64
}

// StatWeb is an editable polygon stat display (spider/radar chart) with named
// axes and optional draggable handles for attribute editing.
type StatWeb struct {
	Component

	font           *sg.FontFamily
	fontSize       float64
	axes           []StatAxis
	editable       bool
	onValueChanged func(index int, value float64)

	// Visual nodes.
	gridMesh   *sg.Node   // concentric polygon grid lines
	spokeMesh  *sg.Node   // spoke lines from center to tips
	polyFill   *sg.Node   // filled polygon area
	polyStroke *sg.Node   // polygon outline
	handles    []*sg.Node // draggable handle circles per axis
	labels     []*sg.Node // text labels per axis

	// Drag state.
	dragging     bool
	dragIndex    int
	dragOriginCX float64 // center X in world coords at drag start
	dragOriginCY float64 // center Y in world coords at drag start
}

// NewStatWeb creates a StatWeb with the given name, font source, and font size.
func NewStatWeb(name string, source *sg.FontFamily, fontSize float64) *StatWeb {
	var font *sg.FontFamily
	if source != nil {
		font = source
	}
	s := &StatWeb{
		font:     font,
		fontSize: fontSize,
	}
	initComponent(&s.Component, name)
	s.initBackground(name)
	s.initBorder(name)
	s.SetSize(300, 300)

	s.onVisualStateChange = func() { s.UpdateVisuals() }
	s.onThemeChange = func() { s.UpdateVisuals() }

	s.UpdateVisuals()
	return s
}

// SetAxes defines all spokes at once.
func (s *StatWeb) SetAxes(axes []StatAxis) {
	if len(axes) > 8 {
		axes = axes[:8]
	}
	s.axes = make([]StatAxis, len(axes))
	copy(s.axes, axes)
	// Clamp values.
	for i := range s.axes {
		s.axes[i].Value = clampAxis(s.axes[i])
	}
	s.rebuildNodes()
	s.UpdateVisuals()
}

// Axes returns a copy of the current axes.
func (s *StatWeb) Axes() []StatAxis {
	out := make([]StatAxis, len(s.axes))
	copy(out, s.axes)
	return out
}

// SetValues bulk-sets values by index (parallel to axes).
func (s *StatWeb) SetValues(values []float64) {
	for i := range s.axes {
		if i < len(values) {
			s.axes[i].Value = clampAxis(s.axes[i])
			s.axes[i].Value = clampVal(values[i], s.axes[i].Min, s.axes[i].Max)
		}
	}
	s.updateGeometry()
}

// Values returns a copy of all axis values.
func (s *StatWeb) Values() []float64 {
	out := make([]float64, len(s.axes))
	for i, a := range s.axes {
		out[i] = a.Value
	}
	return out
}

// SetValue sets a single axis value by index.
func (s *StatWeb) SetValue(index int, value float64) {
	if index < 0 || index >= len(s.axes) {
		return
	}
	old := s.axes[index].Value
	s.axes[index].Value = clampVal(value, s.axes[index].Min, s.axes[index].Max)
	s.updateGeometry()
	if s.axes[index].Value != old && s.onValueChanged != nil {
		s.onValueChanged(index, s.axes[index].Value)
	}
}

// Value returns a single axis value by index.
func (s *StatWeb) Value(index int) float64 {
	if index < 0 || index >= len(s.axes) {
		return 0
	}
	return s.axes[index].Value
}

// SetFillEnabled shows or hides the semi-transparent polygon fill.
func (s *StatWeb) SetFillEnabled(v bool) {
	if s.polyFill != nil {
		s.polyFill.SetVisible(v)
	}
}

// SetEditable enables or disables draggable axis handles.
func (s *StatWeb) SetEditable(v bool) {
	s.editable = v
	for _, h := range s.handles {
		h.SetVisible(v)
	}
}

// IsEditable returns whether the stat web is in editable mode.
func (s *StatWeb) IsEditable() bool {
	return s.editable
}

// SetOnValueChanged sets the callback for when a handle is dragged.
func (s *StatWeb) SetOnValueChanged(fn func(index int, value float64)) {
	s.onValueChanged = fn
}

// SetSize sets the widget dimensions.
func (s *StatWeb) SetSize(w, h float64) {
	s.Width = w
	s.Height = h
	s.resizeBackground(w, h)
	s.resizeBorder(w, h)
	s.node.HitShape = sg.HitRect{X: 0, Y: 0, Width: w, Height: h}
	s.rebuildNodes()
	s.MarkLayoutDirty()
}

// Default fallback colors used when no theme is loaded.
var (
	swDefaultGrid   = sg.RGBA(0.3, 0.35, 0.4, 0.4)
	swDefaultSpoke  = sg.RGBA(0.35, 0.4, 0.45, 0.5)
	swDefaultFill   = sg.RGBA(0.35, 0.55, 0.85, 0.25)
	swDefaultStroke = sg.RGBA(0.4, 0.65, 0.95, 0.8)
	swDefaultHandle = sg.RGBA(0.9, 0.9, 0.95, 1)
	swDefaultLabel  = sg.RGBA(0.7, 0.75, 0.8, 1)
)

// colorOrDefault returns c if it has any alpha, otherwise returns def.
func colorOrDefault(c, def sg.Color) sg.Color {
	if c.A() > 0 {
		return c
	}
	return def
}

// UpdateVisuals applies theme colors based on current state.
func (s *StatWeb) UpdateVisuals() {
	s.state = computeState(s.enabled, s.focused, s.hovered, s.pressed)
	group := s.EffectiveTheme().StatWeb.Group(s.Variant())

	bg := group.Background.Resolve(s.state)
	s.applyBackground(bg)

	if s.gridMesh != nil {
		s.gridMesh.SetColor(colorOrDefault(group.GridColor.Resolve(s.state), swDefaultGrid))
	}
	if s.spokeMesh != nil {
		s.spokeMesh.SetColor(colorOrDefault(group.SpokeColor.Resolve(s.state), swDefaultSpoke))
	}
	if s.polyFill != nil {
		s.polyFill.SetColor(colorOrDefault(group.PolygonFill.Resolve(s.state), swDefaultFill))
	}
	if s.polyStroke != nil {
		s.polyStroke.SetColor(colorOrDefault(group.PolygonStroke.Resolve(s.state), swDefaultStroke))
	}

	handleColor := colorOrDefault(group.HandleColor.Resolve(s.state), swDefaultHandle)
	for _, h := range s.handles {
		h.SetColor(handleColor)
	}

	labelColor := colorOrDefault(group.LabelColor.Resolve(s.state), swDefaultLabel)
	labelFontSize := group.LabelFontSize
	if labelFontSize <= 0 {
		labelFontSize = s.fontSize
	}
	for _, l := range s.labels {
		l.TextBlock.Color = labelColor
		l.TextBlock.FontSize = labelFontSize
	}

	s.MarkDrawDirty()
}

// ---------------------------------------------------------------------------
// Internal geometry building
// ---------------------------------------------------------------------------

func (s *StatWeb) rebuildNodes() {
	// Remove old visual nodes.
	s.removeOldNodes()

	n := len(s.axes)
	if n < 3 {
		return
	}

	group := s.EffectiveTheme().StatWeb.Group(s.Variant())
	cx, cy := s.Width/2, s.Height/2
	labelOffset := group.LabelOffset
	if labelOffset <= 0 {
		labelOffset = 16
	}
	radius := math.Min(cx, cy) - labelOffset - 4

	if radius < 10 {
		radius = 10
	}

	// Grid levels (concentric polygons).
	gridLevels := group.GridLevels
	if gridLevels <= 0 {
		gridLevels = 4
	}
	gridW := group.SpokeWidth
	if gridW <= 0 {
		gridW = 1
	}

	// Build grid mesh.
	var gVerts []engine.Vertex
	var gInds []uint16
	for level := 1; level <= gridLevels; level++ {
		r := radius * float64(level) / float64(gridLevels)
		v, idx := buildPolygonRing(cx, cy, r, n, gridW)
		offset := uint16(len(gVerts))
		for i := range idx {
			idx[i] += offset
		}
		gVerts = append(gVerts, v...)
		gInds = append(gInds, idx...)
	}
	s.gridMesh = sg.NewMesh(s.node.Name+"-grid", sg.WhitePixel, gVerts, gInds)
	s.gridMesh.SetColor(colorOrDefault(group.GridColor.Resolve(s.state), swDefaultGrid))
	s.node.AddChild(s.gridMesh)

	// Build spoke mesh (lines from center to each tip).
	spokeW := group.SpokeWidth
	if spokeW <= 0 {
		spokeW = 1
	}
	var sVerts []engine.Vertex
	var sInds []uint16
	for i := 0; i < n; i++ {
		angle := spokeAngle(i, n)
		tx := cx + radius*math.Cos(angle)
		ty := cy + radius*math.Sin(angle)
		v, idx := buildLine(cx, cy, tx, ty, spokeW)
		offset := uint16(len(sVerts))
		for j := range idx {
			idx[j] += offset
		}
		sVerts = append(sVerts, v...)
		sInds = append(sInds, idx...)
	}
	s.spokeMesh = sg.NewMesh(s.node.Name+"-spokes", sg.WhitePixel, sVerts, sInds)
	s.spokeMesh.SetColor(colorOrDefault(group.SpokeColor.Resolve(s.state), swDefaultSpoke))
	s.node.AddChild(s.spokeMesh)

	// Build filled polygon.
	fillVerts, fillInds := s.buildValuePolygon(cx, cy, radius, n)
	s.polyFill = sg.NewMesh(s.node.Name+"-fill", sg.WhitePixel, fillVerts, fillInds)
	s.polyFill.SetColor(colorOrDefault(group.PolygonFill.Resolve(s.state), swDefaultFill))
	s.node.AddChild(s.polyFill)

	// Build polygon outline.
	strokeW := group.PolygonStrokeWidth
	if strokeW <= 0 {
		strokeW = 2
	}
	strokeVerts, strokeInds := s.buildValueRing(cx, cy, radius, n, strokeW)
	s.polyStroke = sg.NewMesh(s.node.Name+"-stroke", sg.WhitePixel, strokeVerts, strokeInds)
	s.polyStroke.SetColor(colorOrDefault(group.PolygonStroke.Resolve(s.state), swDefaultStroke))
	s.node.AddChild(s.polyStroke)

	// Build handles and labels.
	handleRadius := group.HandleRadius
	if handleRadius <= 0 {
		handleRadius = 6
	}
	s.handles = make([]*sg.Node, n)
	s.labels = make([]*sg.Node, n)
	handleColor := colorOrDefault(group.HandleColor.Resolve(s.state), swDefaultHandle)
	labelColor := colorOrDefault(group.LabelColor.Resolve(s.state), swDefaultLabel)
	labelFontSize := group.LabelFontSize
	if labelFontSize <= 0 {
		labelFontSize = s.fontSize
	}

	for i := 0; i < n; i++ {
		angle := spokeAngle(i, n)
		frac := s.axisFraction(i)
		hx := cx + radius*frac*math.Cos(angle)
		hy := cy + radius*frac*math.Sin(angle)

		// Handle: small filled circle.
		hVerts, hInds := buildCircleFill(0, 0, handleRadius, 12)
		handle := sg.NewMesh(s.node.Name+"-handle-"+s.axes[i].Name, sg.WhitePixel, hVerts, hInds)
		handle.SetPosition(hx, hy)
		handle.SetColor(handleColor)
		handle.SetVisible(s.editable)
		handle.Interactable = true
		// Expand hit shape beyond visual radius for easier grabbing.
		hitR := math.Max(handleRadius, 10)
		handle.HitShape = sg.HitRect{X: -hitR, Y: -hitR, Width: hitR * 2, Height: hitR * 2}

		// Wire drag events.
		idx := i
		handle.OnDragStart(func(ctx sg.DragContext) {
			if !s.enabled || !s.editable {
				return
			}
			s.dragging = true
			s.dragIndex = idx
			// Compute center in world coords from the drag context.
			// GlobalXY - LocalXY gives the handle's world origin; subtract
			// the handle's local position (relative to the component root)
			// to get the component's world origin, then add center offset.
			handleWorldX := ctx.GlobalX - ctx.LocalX
			handleWorldY := ctx.GlobalY - ctx.LocalY
			handleLocalX := handle.X()
			handleLocalY := handle.Y()
			s.dragOriginCX = handleWorldX - handleLocalX + cx
			s.dragOriginCY = handleWorldY - handleLocalY + cy
		})
		handle.OnDrag(func(ctx sg.DragContext) {
			if !s.dragging || s.dragIndex != idx {
				return
			}
			s.handleDrag(ctx, idx, radius)
		})
		handle.OnDragEnd(func(_ sg.DragContext) {
			if s.dragIndex == idx {
				s.dragging = false
			}
		})

		s.node.AddChild(handle)
		s.handles[i] = handle

		// Label at spoke tip.
		lx := cx + (radius+labelOffset)*math.Cos(angle)
		ly := cy + (radius+labelOffset)*math.Sin(angle)
		label := sg.NewText(s.node.Name+"-label-"+s.axes[i].Name, s.axes[i].Name, s.font)
		label.TextBlock.FontSize = labelFontSize
		label.TextBlock.Color = labelColor
		label.TextBlock.Align = sg.TextAlignCenter
		label.SetPosition(lx, ly-labelFontSize/2)
		s.node.AddChild(label)
		s.labels[i] = label
	}
}

func (s *StatWeb) removeOldNodes() {
	if s.gridMesh != nil {
		s.node.RemoveChild(s.gridMesh)
		s.gridMesh = nil
	}
	if s.spokeMesh != nil {
		s.node.RemoveChild(s.spokeMesh)
		s.spokeMesh = nil
	}
	if s.polyFill != nil {
		s.node.RemoveChild(s.polyFill)
		s.polyFill = nil
	}
	if s.polyStroke != nil {
		s.node.RemoveChild(s.polyStroke)
		s.polyStroke = nil
	}
	for _, h := range s.handles {
		s.node.RemoveChild(h)
	}
	s.handles = nil
	for _, l := range s.labels {
		s.node.RemoveChild(l)
	}
	s.labels = nil
}

func (s *StatWeb) updateGeometry() {
	n := len(s.axes)
	if n < 3 || s.polyFill == nil {
		return
	}

	group := s.EffectiveTheme().StatWeb.Group(s.Variant())
	cx, cy := s.Width/2, s.Height/2
	labelOffset := group.LabelOffset
	if labelOffset <= 0 {
		labelOffset = 16
	}
	radius := math.Min(cx, cy) - labelOffset - 4
	if radius < 10 {
		radius = 10
	}

	handleRadius := group.HandleRadius
	if handleRadius <= 0 {
		handleRadius = 6
	}
	strokeW := group.PolygonStrokeWidth
	if strokeW <= 0 {
		strokeW = 2
	}

	// Update filled polygon.
	fillVerts, fillInds := s.buildValuePolygon(cx, cy, radius, n)
	s.polyFill.SetMeshVertices(fillVerts)
	s.polyFill.SetMeshIndices(fillInds)
	s.polyFill.InvalidateMeshAABB()

	// Update outline.
	strokeVerts, strokeInds := s.buildValueRing(cx, cy, radius, n, strokeW)
	s.polyStroke.SetMeshVertices(strokeVerts)
	s.polyStroke.SetMeshIndices(strokeInds)
	s.polyStroke.InvalidateMeshAABB()

	// Update handle positions.
	for i := 0; i < n; i++ {
		angle := spokeAngle(i, n)
		frac := s.axisFraction(i)
		hx := cx + radius*frac*math.Cos(angle)
		hy := cy + radius*frac*math.Sin(angle)
		s.handles[i].SetPosition(hx, hy)
	}

	s.MarkDrawDirty()
}

func (s *StatWeb) handleDrag(ctx sg.DragContext, index int, radius float64) {
	angle := spokeAngle(index, len(s.axes))
	// Vector from center to pointer in world space.
	dx := ctx.GlobalX - s.dragOriginCX
	dy := ctx.GlobalY - s.dragOriginCY
	// Project onto spoke direction.
	proj := dx*math.Cos(angle) + dy*math.Sin(angle)
	frac := proj / radius
	if frac < 0 {
		frac = 0
	}
	if frac > 1 {
		frac = 1
	}
	a := s.axes[index]
	newVal := a.Min + frac*(a.Max-a.Min)
	newVal = clampVal(newVal, a.Min, a.Max)
	old := s.axes[index].Value
	s.axes[index].Value = newVal
	s.updateGeometry()
	if newVal != old && s.onValueChanged != nil {
		s.onValueChanged(index, newVal)
	}
}

func (s *StatWeb) axisFraction(i int) float64 {
	a := s.axes[i]
	if a.Max <= a.Min {
		return 0
	}
	return (a.Value - a.Min) / (a.Max - a.Min)
}

func (s *StatWeb) buildValuePolygon(cx, cy, radius float64, n int) ([]engine.Vertex, []uint16) {
	// Fan triangulation from center.
	verts := make([]engine.Vertex, n+1)
	verts[0] = vertex(cx, cy)
	for i := 0; i < n; i++ {
		angle := spokeAngle(i, n)
		frac := s.axisFraction(i)
		verts[i+1] = vertex(
			cx+radius*frac*math.Cos(angle),
			cy+radius*frac*math.Sin(angle),
		)
	}
	inds := make([]uint16, 0, n*3)
	for i := 0; i < n; i++ {
		inds = append(inds, 0, uint16(i+1), uint16((i+1)%n+1))
	}
	return verts, inds
}

func (s *StatWeb) buildValueRing(cx, cy, radius float64, n int, width float64) ([]engine.Vertex, []uint16) {
	pts := make([][2]float64, n)
	for i := 0; i < n; i++ {
		angle := spokeAngle(i, n)
		frac := s.axisFraction(i)
		pts[i] = [2]float64{
			cx + radius*frac*math.Cos(angle),
			cy + radius*frac*math.Sin(angle),
		}
	}
	return buildClosedPolyline(pts, width)
}

// ---------------------------------------------------------------------------
// Geometry helpers
// ---------------------------------------------------------------------------

// spokeAngle returns the angle for spoke i of n spokes (starting from top, clockwise).
func spokeAngle(i, n int) float64 {
	return -math.Pi/2 + 2*math.Pi*float64(i)/float64(n)
}

func clampAxis(a StatAxis) float64 {
	return clampVal(a.Value, a.Min, a.Max)
}

func clampVal(v, min, max float64) float64 {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}

func vertex(x, y float64) engine.Vertex {
	return engine.Vertex{
		DstX:   float32(x),
		DstY:   float32(y),
		SrcX:   0,
		SrcY:   0,
		ColorR: 1,
		ColorG: 1,
		ColorB: 1,
		ColorA: 1,
	}
}

// buildPolygonRing builds a closed polyline ring (regular polygon outline).
func buildPolygonRing(cx, cy, radius float64, sides int, width float64) ([]engine.Vertex, []uint16) {
	pts := make([][2]float64, sides)
	for i := 0; i < sides; i++ {
		angle := spokeAngle(i, sides)
		pts[i] = [2]float64{cx + radius*math.Cos(angle), cy + radius*math.Sin(angle)}
	}
	return buildClosedPolyline(pts, width)
}

// buildClosedPolyline builds a closed polygon outline from a set of points.
// Uses miter joins so line width stays consistent at corners.
func buildClosedPolyline(pts [][2]float64, width float64) ([]engine.Vertex, []uint16) {
	n := len(pts)
	if n < 2 {
		return nil, nil
	}
	hw := width / 2
	verts := make([]engine.Vertex, n*2)
	for i := 0; i < n; i++ {
		prev := (i - 1 + n) % n
		next := (i + 1) % n

		// Edge normals for the two edges meeting at this vertex.
		dx0 := pts[i][0] - pts[prev][0]
		dy0 := pts[i][1] - pts[prev][1]
		len0 := math.Sqrt(dx0*dx0 + dy0*dy0)
		if len0 < 1e-9 {
			len0 = 1
		}
		nx0 := -dy0 / len0
		ny0 := dx0 / len0

		dx1 := pts[next][0] - pts[i][0]
		dy1 := pts[next][1] - pts[i][1]
		len1 := math.Sqrt(dx1*dx1 + dy1*dy1)
		if len1 < 1e-9 {
			len1 = 1
		}
		nx1 := -dy1 / len1
		ny1 := dx1 / len1

		// Miter direction: average of the two edge normals.
		mx := nx0 + nx1
		my := ny0 + ny1
		mLen := math.Sqrt(mx*mx + my*my)
		if mLen < 1e-9 {
			mx, my = nx0, ny0
		} else {
			mx /= mLen
			my /= mLen
		}

		// Scale miter to maintain correct width: hw / dot(miter, edgeNormal).
		dot := mx*nx1 + my*ny1
		if dot < 0.1 {
			dot = 0.1 // clamp to avoid extreme spikes at very sharp angles
		}
		miterHW := hw / dot

		verts[i*2] = vertex(pts[i][0]+mx*miterHW, pts[i][1]+my*miterHW)
		verts[i*2+1] = vertex(pts[i][0]-mx*miterHW, pts[i][1]-my*miterHW)
	}
	inds := make([]uint16, 0, n*6)
	for i := 0; i < n; i++ {
		a := uint16(i * 2)
		b := uint16(i*2 + 1)
		c := uint16(((i + 1) % n) * 2)
		d := uint16(((i+1)%n)*2 + 1)
		inds = append(inds, a, b, c, b, d, c)
	}
	return verts, inds
}

// buildLine builds a single line segment as a thin quad.
func buildLine(x1, y1, x2, y2, width float64) ([]engine.Vertex, []uint16) {
	dx := x2 - x1
	dy := y2 - y1
	length := math.Sqrt(dx*dx + dy*dy)
	if length < 1e-9 {
		return nil, nil
	}
	hw := width / 2
	nx := -dy / length * hw
	ny := dx / length * hw
	verts := []engine.Vertex{
		vertex(x1+nx, y1+ny),
		vertex(x1-nx, y1-ny),
		vertex(x2+nx, y2+ny),
		vertex(x2-nx, y2-ny),
	}
	inds := []uint16{0, 1, 2, 1, 3, 2}
	return verts, inds
}

// buildCircleFill builds a filled circle as a triangle fan.
func buildCircleFill(cx, cy, radius float64, segments int) ([]engine.Vertex, []uint16) {
	verts := make([]engine.Vertex, segments+1)
	verts[0] = vertex(cx, cy)
	for i := 0; i < segments; i++ {
		a := 2 * math.Pi * float64(i) / float64(segments)
		verts[i+1] = vertex(cx+radius*math.Cos(a), cy+radius*math.Sin(a))
	}
	inds := make([]uint16, 0, segments*3)
	for i := 0; i < segments; i++ {
		inds = append(inds, 0, uint16(i+1), uint16((i+1)%segments+1))
	}
	return verts, inds
}
