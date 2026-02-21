package widget

import "github.com/devthicket/willowui/internal/sg"

// Controller is implemented by per-screen state owners that build UI,
// run per-frame logic, and clean up resources.
type Controller interface {
	// OnCreate is called once when the screen is shown. Build UI here.
	OnCreate(screen *Screen)

	// OnUpdate is called once per frame during the UI update pass.
	OnUpdate(dt float64)

	// OnDestroy is called when the screen is destroyed. Clean up here.
	OnDestroy()
}

// disposable is any resource that can be released.
type disposable interface {
	Stop()
}

// nopController is used when no controller is supplied.
type nopController struct{}

func (nopController) OnCreate(*Screen) {}
func (nopController) OnUpdate(float64) {}
func (nopController) OnDestroy()       {}

// ScreenOption configures a Screen during construction.
type ScreenOption func(*Screen)

// WithController attaches a Controller to the screen.
func WithController(c Controller) ScreenOption {
	return func(s *Screen) { s.controller = c }
}

// WithScene sets the scene on a Screen explicitly.
// Intended for use in tests; in production the scene is set automatically by Stage.Add.
func WithScene(s *sg.Scene) ScreenOption {
	return func(screen *Screen) { screen.scene = s }
}

// Screen bridges a Controller and the underlying willow scene. It owns the
// root component tree, focus manager, and scheduler, and manages the
// controller lifecycle.
type Screen struct {
	root         *Component
	controller   Controller
	scene        *sg.Scene
	scheduler    *Scheduler
	inputManager *InputManager
	focusManager *FocusManager
	refs         []disposable
	created      bool
	visible      bool
}

// NewScreen creates a new Screen. Use WithController to attach a controller.
// The scene is set automatically by Stage.Add; use WithScene in tests.
func NewScreen(opts ...ScreenOption) *Screen {
	s := &Screen{
		root:         NewComponent("screen-root"),
		controller:   nopController{},
		scheduler:    DefaultScheduler,
		inputManager: DefaultInputManager,
		focusManager: DefaultFocusManager,
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

// Add adds a UIElement (Component) to the screen's root.
func (s *Screen) Add(e UIElement) {
	s.root.AddChild(e)
}

// AddNode adds a raw sg.Node to the screen's root.
func (s *Screen) AddNode(n *sg.Node) {
	s.root.AddRawChild(n)
}

// Remove detaches a UIElement from the screen's root.
func (s *Screen) Remove(e UIElement) {
	s.root.RemoveChild(e)
}

// RemoveNode detaches a raw sg.Node from the screen's root.
func (s *Screen) RemoveNode(n *sg.Node) {
	s.root.RemoveRawChild(n)
}

// NumChildren returns the number of direct children attached to the screen root.
func (s *Screen) NumChildren() int {
	return len(s.root.children)
}

// Children returns the direct children of the screen root.
func (s *Screen) Children() []*Component {
	return s.root.Children()
}

// Scheduler returns the scheduler used by this screen.
func (s *Screen) Scheduler() *Scheduler {
	return s.scheduler
}

// InputManager returns the input manager used by this screen.
func (s *Screen) InputManager() *InputManager {
	return s.inputManager
}

// FocusManager returns the focus manager used by this screen.
func (s *Screen) FocusManager() *FocusManager {
	return s.focusManager
}

// Show attaches the root node to the scene and calls OnCreate on first show.
func (s *Screen) Show() {
	if s.visible {
		return
	}
	s.visible = true
	if s.scene != nil {
		s.scene.Root.AddChild(s.root.Node())
	}
	if !s.created {
		s.created = true
		s.controller.OnCreate(s)
	}
}

// Hide detaches the root node from the scene but preserves all state.
func (s *Screen) Hide() {
	if !s.visible {
		return
	}
	s.visible = false
	if s.scene != nil {
		s.scene.Root.RemoveChild(s.root.Node())
	}
}

// Destroy detaches the root node, calls OnDestroy, and disposes all
// tracked refs.
func (s *Screen) Destroy() {
	if s.visible {
		s.Hide()
	}
	if s.created {
		s.controller.OnDestroy()
		s.created = false
	}
	for _, r := range s.refs {
		r.Stop()
	}
	s.refs = s.refs[:0]
}

// Update flushes the scheduler, reads input, runs focus dispatch, fires
// passthrough listeners, and calls the controller's OnUpdate.
func (s *Screen) Update(dt float64) {
	s.scheduler.Flush()
	s.inputManager.Update()
	// Tick overlay managers before the focus manager so they can consume
	// injected keys (arrow nav, Enter, Escape) with higher priority than
	// focus spatial navigation or focused-widget key handling.
	DefaultMenuPopupManager.tick()
	s.focusManager.Update()
	s.inputManager.FireListeners()
	s.controller.OnUpdate(dt)
	if s.root.IsLayoutDirty() {
		s.root.UpdateLayout()
	}
}

// TrackRef registers a disposable resource that will be automatically
// stopped when the screen is destroyed.
func (s *Screen) TrackRef(r disposable) {
	s.refs = append(s.refs, r)
}

// ClearTemplateTree stops all tracked refs and removes all children from
// the screen root. Used by hot reload to tear down the previous component
// tree before re-instantiation.
func (s *Screen) ClearTemplateTree() {
	for _, r := range s.refs {
		r.Stop()
	}
	s.refs = s.refs[:0]
	children := s.root.Children()
	for i := len(children) - 1; i >= 0; i-- {
		s.root.RemoveChild(children[i])
	}
}

// FindByName searches the component tree for a component whose node name
// matches the given name. Returns nil if not found.
func (s *Screen) FindByName(name string) *Component {
	return findByName(s.root, name)
}

func findByName(c *Component, name string) *Component {
	if c.Name() == name {
		return c
	}
	for _, child := range c.children {
		if found := findByName(child, name); found != nil {
			return found
		}
	}
	return nil
}

// Visible returns whether the screen is currently shown.
func (s *Screen) Visible() bool {
	return s.visible
}

// ---------------------------------------------------------------------------
// StageManager
// ---------------------------------------------------------------------------

// StageManager manages a stack of screens. The package-level DefaultStage is
// wired automatically by ui.Setup. Use ui.Stage to access it in applications.
type StageManager struct {
	stack []*Screen
	scene *sg.Scene
}

// DefaultStage is the package-level stage singleton.
var DefaultStage = &StageManager{}

// NewStageManager creates a new StageManager. Primarily used for test isolation;
// in production use ui.Stage.
func NewStageManager() *StageManager {
	return &StageManager{}
}

// SetScene sets the scene used for attaching screen nodes.
// Called by ui.Setup; can also be called in tests.
func (st *StageManager) SetScene(s *sg.Scene) {
	st.scene = s
	for _, screen := range st.stack {
		if screen.scene == nil {
			screen.scene = s
			if screen.visible {
				s.Root.AddChild(screen.root.Node())
			}
		}
	}
}

// Add pushes screen onto the stack and shows it.
func (st *StageManager) Add(screen *Screen) {
	if screen.scene == nil && st.scene != nil {
		screen.scene = st.scene
	}
	st.stack = append(st.stack, screen)
	screen.Show()
}

// Remove destroys and removes a specific screen from the stack.
func (st *StageManager) Remove(screen *Screen) {
	for i, s := range st.stack {
		if s == screen {
			st.stack = append(st.stack[:i], st.stack[i+1:]...)
			screen.Destroy()
			return
		}
	}
}

// Replace destroys the top screen and shows the new screen in its place.
func (st *StageManager) Replace(screen *Screen) {
	if top := st.Top(); top != nil {
		st.stack = st.stack[:len(st.stack)-1]
		top.Destroy()
	}
	if screen.scene == nil && st.scene != nil {
		screen.scene = st.scene
	}
	st.stack = append(st.stack, screen)
	screen.Show()
}

// CloseAll destroys all screens on the stack.
func (st *StageManager) CloseAll() {
	for i := len(st.stack) - 1; i >= 0; i-- {
		st.stack[i].Destroy()
	}
	st.stack = st.stack[:0]
}

// Update calls Update on all visible screens.
func (st *StageManager) Update(dt float64) {
	for _, s := range st.stack {
		if s.visible {
			s.Update(dt)
		}
	}
}

// Top returns the topmost screen, or nil if the stack is empty.
func (st *StageManager) Top() *Screen {
	if len(st.stack) == 0 {
		return nil
	}
	return st.stack[len(st.stack)-1]
}

// Size returns the number of screens on the stack.
func (st *StageManager) Size() int {
	return len(st.stack)
}
