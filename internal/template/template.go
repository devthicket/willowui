package template

import (
	"encoding/json"
	"fmt"
	"image"
	"image/gif"
	_ "image/jpeg" // register JPEG decoder for src attribute
	_ "image/png"  // register PNG decoder for src attribute
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/devthicket/willowui/internal/engine"
	"github.com/devthicket/willowui/internal/sg"
	"github.com/devthicket/willowui/internal/theme"
	"github.com/devthicket/willowui/internal/widget"
)

// FactoryContext provides resources needed by component factories.
type FactoryContext struct {
	Fonts           map[string]*sg.FontFamily
	DefaultFont     *sg.FontFamily
	DefaultFontSize float64
	// CustomVariants maps user-defined variant names to their Variant slots.
	// Populated by SetTheme; nil means only built-in names are resolved.
	CustomVariants map[string]widget.Variant

	// themeJSON is the raw JSON of the base theme, stored for inline patch merging.
	themeJSON []byte
	// compiledTheme is the compiled theme set via SetTheme/SetThemeJSON.
	compiledTheme *widget.Theme

	// Custom widget factories and setters registered via RegisterWidget.
	customFactories map[string]componentFactory
	customSetters   map[string]map[string]componentSetter
}

func (fc *FactoryContext) font(name string) *sg.FontFamily {
	if name != "" && fc.Fonts != nil {
		if f, ok := fc.Fonts[name]; ok {
			return f
		}
	}
	return fc.DefaultFont
}

// resolveVariant converts a variant name string to a Variant value.
// Checks custom variants first (from the compiled theme), then falls
// back to built-in names.
func (fc *FactoryContext) resolveVariant(name string) widget.Variant {
	lower := strings.ToLower(name)
	if fc.CustomVariants != nil {
		if v, ok := fc.CustomVariants[lower]; ok {
			return v
		}
	}
	return parseVariant(lower)
}

// componentFactory creates a component from IR attributes and a factory context.
type componentFactory func(name string, attrs []IRAttribute, fc *FactoryContext) (*widget.Component, error)

// componentSetter applies a named attribute value to a component.
type componentSetter func(comp *widget.Component, value any)

// WidgetFactory creates a custom widget component by name.
// The returned component should have UserData set to the typed widget struct
// so that attribute setters can cast it back.
type WidgetFactory func(name string) (*widget.Component, error)

// AttrSetter applies a named attribute value to a custom widget component.
type AttrSetter func(comp *widget.Component, value any)

// TemplateRegistry stores compiled XML templates and instantiates them.
type TemplateRegistry struct {
	templates       map[string]*IRNode
	fc              FactoryContext
	customFactories map[string]componentFactory
	customSetters   map[string]map[string]componentSetter
}

// NewTemplateRegistry creates a new template registry.
func NewTemplateRegistry() *TemplateRegistry {
	return &TemplateRegistry{
		templates: make(map[string]*IRNode),
	}
}

// NewTemplateRegistryWithFont creates a new template registry with a default
// font loaded from raw TTF data and a display font size. This is a convenience
// for the common 3-line boilerplate of NewTemplateRegistry + SetFonts + SetFontSize.
func NewTemplateRegistryWithFont(ttf []byte, size float64) (*TemplateRegistry, error) {
	font, err := sg.NewFontFamilyFromTTF(sg.FontFamilyConfig{Regular: ttf})
	if err != nil {
		return nil, fmt.Errorf("load font: %w", err)
	}
	r := NewTemplateRegistry()
	r.SetFonts(nil, font)
	r.SetFontSize(size)
	return r, nil
}

// RegisterWidget registers a custom widget type for use in XML templates.
// The typeName must not collide with built-in widget names.
// The factory creates the component; setters apply XML attributes by name.
// RegisterWidget must be called before RegisterXML for any template that
// uses the custom widget type.
func (r *TemplateRegistry) RegisterWidget(typeName string, factory WidgetFactory, setters map[string]AttrSetter) {
	if knownComponents[typeName] {
		panic(fmt.Sprintf("RegisterWidget: %q collides with a built-in widget type", typeName))
	}
	if r.customFactories == nil {
		r.customFactories = make(map[string]componentFactory)
		r.customSetters = make(map[string]map[string]componentSetter)
	}
	// Wrap the user-facing WidgetFactory into the internal componentFactory.
	// If the XML element has a name="..." attribute, use that as the widget
	// name; otherwise fall back to the component type name.
	r.customFactories[typeName] = func(name string, attrs []IRAttribute, _ *FactoryContext) (*widget.Component, error) {
		if n := findStaticAttr(attrs, "name"); n != "" {
			name = n
		}
		return factory(name)
	}
	if len(setters) > 0 {
		wrapped := make(map[string]componentSetter, len(setters))
		for k, s := range setters {
			wrapped[k] = componentSetter(s)
		}
		r.customSetters[typeName] = mergeSetters(commonSetters, wrapped)
	} else {
		r.customSetters[typeName] = commonSetters
	}
}

// RegisterXML compiles and registers an XML template by name.
func (r *TemplateRegistry) RegisterXML(name string, xmlData []byte) error {
	node, err := CompileXMLWithTypes(xmlData, r.customTypes())
	if err != nil {
		return fmt.Errorf("template %q: %w", name, err)
	}
	r.templates[name] = node
	return nil
}

// customTypes returns the set of custom widget type names registered on this
// registry, or nil if none.
func (r *TemplateRegistry) customTypes() map[string]bool {
	if len(r.customFactories) == 0 {
		return nil
	}
	m := make(map[string]bool, len(r.customFactories))
	for k := range r.customFactories {
		m[k] = true
	}
	return m
}

// RegisterBinary decodes a .xmlui binary blob and registers the template by name.
func (r *TemplateRegistry) RegisterBinary(name string, binData []byte) error {
	node, err := DecodeIR(binData)
	if err != nil {
		return fmt.Errorf("template %q: %w", name, err)
	}
	r.templates[name] = node
	return nil
}

// Get returns the compiled IR for a named template.
func (r *TemplateRegistry) Get(name string) *IRNode {
	return r.templates[name]
}

// SetFonts configures the font map and default font for instantiation.
func (r *TemplateRegistry) SetFonts(fonts map[string]*sg.FontFamily, defaultFont *sg.FontFamily) {
	r.fc.Fonts = fonts
	r.fc.DefaultFont = defaultFont
}

// SetFontSize configures the default display font size for instantiation.
func (r *TemplateRegistry) SetFontSize(size float64) {
	r.fc.DefaultFontSize = size
}

// SetTheme registers the compiled theme with the registry so that custom
// variant names declared in the theme JSON (e.g. "card", "muted") can be
// resolved when the variant attribute is used in templates.
func (r *TemplateRegistry) SetTheme(t *widget.Theme) {
	if t != nil {
		r.fc.CustomVariants = t.CustomVariants
		r.fc.compiledTheme = t
	}
}

// SetThemeJSON registers a theme from raw JSON bytes. The JSON is stored
// so that inline <Theme> patches in templates can be merged against it.
func (r *TemplateRegistry) SetThemeJSON(data []byte) error {
	t, err := theme.LoadTheme(data)
	if err != nil {
		return err
	}
	r.fc.themeJSON = data
	r.fc.CustomVariants = t.CustomVariants
	r.fc.compiledTheme = t
	return nil
}

// Instantiate creates a live component tree from a named template.
// ctrl may be nil for templates that have no reactive bindings or event handlers.
func (r *TemplateRegistry) Instantiate(name string, ctrl widget.Controller, screen *widget.Screen) (*widget.Component, error) {
	ir := r.templates[name]
	if ir == nil {
		return nil, fmt.Errorf("template %q not found", name)
	}
	var dp DataProvider
	if ctrl != nil {
		dp, _ = ctrl.(DataProvider)
	}
	ctx := &EvalContext{Provider: dp}

	// Propagate custom widget registrations to the factory context.
	r.fc.customFactories = r.customFactories
	r.fc.customSetters = r.customSetters

	fc := &r.fc
	// If the template has an inline <Theme> patch, fork the FactoryContext
	// with a locally compiled theme so the patch is scoped to this instance.
	if len(ir.ThemePatch) > 0 {
		forked, err := r.forkFCWithPatch(ir.ThemePatch)
		if err != nil {
			return nil, fmt.Errorf("template %q: theme patch: %w", name, err)
		}
		fc = forked
	}

	comp, err := instantiateNode(ir, ctx, screen, fc)
	if err != nil {
		return nil, err
	}

	// Apply the patched theme to the root component so descendants inherit it.
	if fc != &r.fc && fc.compiledTheme != nil {
		comp.SetTheme(fc.compiledTheme)
	}
	return comp, nil
}

// InstantiateStatic creates a live component tree from a named template
// without any controller. Use this for templates that have no reactive
// bindings or event handlers.
func (r *TemplateRegistry) InstantiateStatic(name string, screen *widget.Screen) (*widget.Component, error) {
	return r.Instantiate(name, nil, screen)
}

// InstantiateIR creates a live component tree from an IR node directly.
func (r *TemplateRegistry) InstantiateIR(ir *IRNode, ctrl widget.Controller, screen *widget.Screen) (*widget.Component, error) {
	r.fc.customFactories = r.customFactories
	r.fc.customSetters = r.customSetters
	dp, _ := ctrl.(DataProvider)
	ctx := &EvalContext{Provider: dp}
	return instantiateNode(ir, ctx, screen, &r.fc)
}

// forkFCWithPatch creates a shallow copy of r.fc with a theme compiled from
// the base theme JSON merged with the given patch bytes (RFC 7396 JSON Merge Patch).
// SetThemeJSON must have been called first to provide the base theme JSON.
func (r *TemplateRegistry) forkFCWithPatch(patch []byte) (*FactoryContext, error) {
	base := r.fc.themeJSON
	if len(base) == 0 {
		return nil, fmt.Errorf("inline <Theme> patches require SetThemeJSON on the registry")
	}
	merged, err := jsonMergePatch(base, patch)
	if err != nil {
		return nil, err
	}

	t, err := theme.LoadTheme(merged)
	if err != nil {
		return nil, fmt.Errorf("compile merged theme: %w", err)
	}

	return &FactoryContext{
		Fonts:           r.fc.Fonts,
		DefaultFont:     r.fc.DefaultFont,
		DefaultFontSize: r.fc.DefaultFontSize,
		CustomVariants:  t.CustomVariants,
		themeJSON:       merged,
		compiledTheme:   t,
		customFactories: r.fc.customFactories,
		customSetters:   r.fc.customSetters,
	}, nil
}

// jsonMergePatch applies a JSON Merge Patch (RFC 7396) to a base document.
// If base is nil or empty, the patch is returned as-is.
func jsonMergePatch(base, patch []byte) ([]byte, error) {
	if len(base) == 0 {
		return patch, nil
	}
	var baseMap map[string]any
	if err := json.Unmarshal(base, &baseMap); err != nil {
		return nil, fmt.Errorf("parse base theme JSON: %w", err)
	}
	var patchMap map[string]any
	if err := json.Unmarshal(patch, &patchMap); err != nil {
		return nil, fmt.Errorf("parse theme patch JSON: %w", err)
	}
	deepMerge(baseMap, patchMap)
	return json.Marshal(baseMap)
}

// deepMerge merges src into dst recursively (JSON Merge Patch semantics).
func deepMerge(dst, src map[string]any) {
	for k, sv := range src {
		if sv == nil {
			delete(dst, k)
			continue
		}
		srcMap, srcIsMap := sv.(map[string]any)
		dstMap, dstIsMap := dst[k].(map[string]any)
		if srcIsMap && dstIsMap {
			deepMerge(dstMap, srcMap)
		} else {
			dst[k] = sv
		}
	}
}

func instantiateNode(ir *IRNode, ctx *EvalContext, screen *widget.Screen, fc *FactoryContext) (*widget.Component, error) {
	// Create the component via factory — check custom factories first.
	factory := fc.customFactories[ir.ComponentType]
	if factory == nil {
		factory = factories[ir.ComponentType]
	}
	if factory == nil {
		return nil, fmt.Errorf("no factory for %q", ir.ComponentType)
	}
	comp, err := factory(ir.ComponentType, ir.Attributes, fc)
	if err != nil {
		return nil, err
	}

	// Apply static attributes — check custom setters first.
	setters := fc.customSetters[ir.ComponentType]
	if setters == nil {
		setters = attrSetters[ir.ComponentType]
	}
	for _, attr := range ir.Attributes {
		if attr.IsEvent {
			applyEvent(comp, attr, ctx)
			continue
		}
		if attr.Expr != nil {
			applyBinding(comp, attr, ctx, screen, setters)
			continue
		}
		// variant is handled specially to resolve custom theme variant names.
		if attr.Name == "variant" {
			comp.SetVariant(fc.resolveVariant(fmt.Sprint(attr.Static)))
			continue
		}
		// Static attribute
		if setter, ok := setters[attr.Name]; ok {
			setter(comp, attr.Static)
		}
	}

	// Process directives
	for _, dir := range ir.Directives {
		switch dir.Type {
		case DirectiveShow:
			applyShowDirective(comp, dir, ctx, screen)
		case DirectiveIf:
			applyIfDirective(comp, dir, ctx, screen)
		}
	}

	// Recurse children
	isAnchorLayout := comp.Layout == widget.LayoutAnchor
	tl := compAsTwoColumnLayout(comp)
	tb := compAsTabBar(comp)
	rg := compAsRadio(comp)
	dtbl := compAsDataTable(comp)
	acc := compAsAccordion(comp)

	// TwoColumnLayout: collect children in pairs and add via AddRow.
	if tl != nil {
		var pending *widget.Component
		for _, childIR := range ir.Children {
			child, err := instantiateNode(childIR, ctx, screen, fc)
			if err != nil {
				return nil, err
			}
			if pending == nil {
				pending = child
			} else {
				tl.AddRow(pending, child)
				pending = nil
			}
		}
		// Odd child falls through to plain AddChild.
		if pending != nil {
			comp.AddChild(pending)
		}
	} else {
		for _, childIR := range ir.Children {
			// DataTable + Column: add column definition directly.
			if dtbl != nil && childIR.ComponentType == "Column" {
				col := widget.DataTableColumn{
					Key:        findStaticAttr(childIR.Attributes, "key"),
					Header:     findStaticAttr(childIR.Attributes, "header"),
					Tooltip:    findStaticAttr(childIR.Attributes, "tooltip"),
					Weight:     parseFloatAttr(childIR.Attributes, "weight", 1),
					Sortable:   parseBoolAttr(childIR.Attributes, "sortable", false),
					Searchable: parseBoolAttr(childIR.Attributes, "searchable", false),
				}
				if fw := parseFloatAttr(childIR.Attributes, "fixedWidth", 0); fw > 0 {
					col.FixedWidth = fw
				}
				switch findStaticAttr(childIR.Attributes, "sortType") {
				case "numeric":
					col.SortType = widget.SortNumeric
				case "custom":
					col.SortType = widget.SortCustom
				}
				switch findStaticAttr(childIR.Attributes, "textAlign") {
				case "center":
					col.Cell.Align = sg.TextAlignCenter
				case "right":
					col.Cell.Align = sg.TextAlignRight
				}
				if findStaticAttr(childIR.Attributes, "clipMode") == "mask" {
					col.ClipMode = widget.ClipMask
				}
				dtbl.AddColumn(col)
				continue
			}

			// Radio + RadioButton: add option directly without creating a component.
			if rg != nil && childIR.ComponentType == "RadioButton" {
				label := findStaticAttr(childIR.Attributes, "label")
				rg.AddOption(label, fc.DefaultFont, fc.DefaultFontSize)
				continue
			}

			// Accordion + Section: add section with first child as content.
			if acc != nil && childIR.ComponentType == "Section" {
				id := findStaticAttr(childIR.Attributes, "id")
				label := findStaticAttr(childIR.Attributes, "label")
				// Wrap section children in a Panel.
				contentPanel := widget.NewPanel(id + "-content")
				contentPanel.Layout = widget.LayoutVBox
				for _, grandchildIR := range childIR.Children {
					gc, err := instantiateNode(grandchildIR, ctx, screen, fc)
					if err != nil {
						return nil, err
					}
					contentPanel.AddChild(gc)
				}
				acc.AddSection(widget.AccordionSection{
					ID:      id,
					Label:   label,
					Content: &contentPanel.Component,
				})
				continue
			}

			// TabBar + Tab: create a tab page and instantiate Tab's children into it.
			if tb != nil && childIR.ComponentType == "Tab" {
				label := findStaticAttr(childIR.Attributes, "label")
				page, _ := tb.AddTabPage(label, widget.LayoutVBox, 0, widget.Insets{})
				// Apply layout/spacing/padding attributes from Tab element to the page.
				tabSetters := commonSetters
				for _, attr := range childIR.Attributes {
					if attr.Name == "label" || attr.Expr != nil || attr.IsEvent {
						continue
					}
					if setter, ok := tabSetters[attr.Name]; ok {
						setter(page, attr.Static)
					}
				}
				// Instantiate Tab's children into the page.
				for _, grandchildIR := range childIR.Children {
					gc, err := instantiateNode(grandchildIR, ctx, screen, fc)
					if err != nil {
						return nil, err
					}
					page.AddChild(gc)
				}
				continue
			}

			child, err := instantiateNode(childIR, ctx, screen, fc)
			if err != nil {
				return nil, err
			}
			if isAnchorLayout {
				anchor, ox, oy := parseAnchorAttrs(childIR.Attributes)
				comp.AddAnchoredChild(child, anchor, ox, oy)
			} else {
				comp.AddChild(child)
			}
		}
	}

	// Auto-size containers that have children but no explicit size.
	// Supports partial auto-fit: if only width or height is set, compute the other.
	if (comp.Width == 0 || comp.Height == 0) && comp.NumChildren() > 0 {
		autoFitContent(comp)
	}

	return comp, nil
}

func applyEvent(comp *widget.Component, attr IRAttribute, ctx *EvalContext) {
	if ctx.Provider == nil {
		return
	}
	methodName := attr.Static
	handler := func() {
		ctx.Provider.CallMethod(methodName)
	}
	switch attr.Name {
	case "click":
		// Wire through typed component if available (Button has its own onClick).
		if b := compAsButton(comp); b != nil {
			b.SetOnClick(handler)
			return
		}
		if ib := compAsIconButton(comp); ib != nil {
			ib.SetOnClick(handler)
			return
		}
		// Fallback: set onActivate on the Component.
		comp.SetOnActivate(handler)
	case "change":
		if t := compAsToggle(comp); t != nil {
			t.SetOnChange(func(bool) { ctx.Provider.CallMethod(methodName) })
		} else if cb := compAsCheckbox(comp); cb != nil {
			cb.SetOnChange(func(bool) { ctx.Provider.CallMethod(methodName) })
		} else if ti := compAsTextInput(comp); ti != nil {
			ti.SetOnChange(func(string) { ctx.Provider.CallMethod(methodName) })
		} else if mi := compAsMaskedInput(comp); mi != nil {
			mi.SetOnChange(func(string) { ctx.Provider.CallMethod(methodName) })
		} else if sb := compAsSearchBox(comp); sb != nil {
			sb.SetOnChange(func(string) { ctx.Provider.CallMethod(methodName) })
		} else if f := compAsInputField(comp); f != nil {
			f.SetOnChange(func(string) { ctx.Provider.CallMethod(methodName) })
		} else if s := compAsSlider(comp); s != nil {
			s.SetOnChange(func(float64) { ctx.Provider.CallMethod(methodName) })
		} else if ns := compAsNumberStepper(comp); ns != nil {
			ns.SetOnChange(func(float64) { ctx.Provider.CallMethod(methodName) })
		} else if tb := compAsTabBar(comp); tb != nil {
			tb.SetOnChange(func(int) { ctx.Provider.CallMethod(methodName) })
		} else if tbb := compAsToggleButtonBar(comp); tbb != nil {
			tbb.SetOnChange(func(int) { ctx.Provider.CallMethod(methodName) })
		} else if sel := compAsSelect(comp); sel != nil {
			sel.SetOnChange(func(int, widget.SelectOption) { ctx.Provider.CallMethod(methodName) })
		} else if or := compAsOptionRotator(comp); or != nil {
			or.SetOnChange(func(int, string) { ctx.Provider.CallMethod(methodName) })
		} else if rg := compAsRadio(comp); rg != nil {
			rg.SetOnChange(func(int) { ctx.Provider.CallMethod(methodName) })
		} else if tp := compAsTimePicker(comp); tp != nil {
			tp.SetOnTimeChanged(func(int, int, int) { ctx.Provider.CallMethod(methodName) })
		} else if l := compAsList(comp); l != nil {
			l.SetOnChange(func(int) { ctx.Provider.CallMethod(methodName) })
		} else if tl := compAsTileList(comp); tl != nil {
			tl.SetOnChange(func(int) { ctx.Provider.CallMethod(methodName) })
		} else if sl := compAsSortableList(comp); sl != nil {
			sl.SetOnChange(func(int) { ctx.Provider.CallMethod(methodName) })
		} else if st := compAsSortableTreeList(comp); st != nil {
			st.SetOnChange(func(int) { ctx.Provider.CallMethod(methodName) })
		} else if cp := compAsColorPicker(comp); cp != nil {
			cp.SetOnChange(func(sg.Color) { ctx.Provider.CallMethod(methodName) })
		} else if sw := compAsStatWeb(comp); sw != nil {
			sw.SetOnValueChanged(func(int, float64) { ctx.Provider.CallMethod(methodName) })
		} else if ge := compAsGradientEditor(comp); ge != nil {
			ge.SetOnChange(func(widget.Gradient) { ctx.Provider.CallMethod(methodName) })
		} else {
			comp.SetOnEvent("change", handler)
		}
	case "submit":
		if ti := compAsTextInput(comp); ti != nil {
			ti.SetOnSubmit(func(string) { ctx.Provider.CallMethod(methodName) })
		} else if mi := compAsMaskedInput(comp); mi != nil {
			mi.SetOnSubmit(func(string) { ctx.Provider.CallMethod(methodName) })
		} else if sb := compAsSearchBox(comp); sb != nil {
			sb.SetOnSubmit(func(string) { ctx.Provider.CallMethod(methodName) })
		} else if f := compAsInputField(comp); f != nil {
			f.SetOnSubmit(func(string) { ctx.Provider.CallMethod(methodName) })
		} else {
			comp.SetOnEvent("submit", handler)
		}
	case "clear":
		if sb := compAsSearchBox(comp); sb != nil {
			sb.SetOnClear(func() { ctx.Provider.CallMethod(methodName) })
		} else {
			comp.SetOnEvent("clear", handler)
		}
	case "blur":
		if ti := compAsTextInput(comp); ti != nil {
			ti.SetOnBlur(func() { ctx.Provider.CallMethod(methodName) })
		} else if mi := compAsMaskedInput(comp); mi != nil {
			mi.SetOnBlur(func() { ctx.Provider.CallMethod(methodName) })
		} else if sb := compAsSearchBox(comp); sb != nil {
			sb.SetOnBlur(func() { ctx.Provider.CallMethod(methodName) })
		} else if f := compAsInputField(comp); f != nil {
			f.SetOnBlur(func() { ctx.Provider.CallMethod(methodName) })
		} else {
			comp.SetOnEvent("blur", handler)
		}
	case "close":
		if w := compAsWindow(comp); w != nil {
			w.SetOnClose(handler)
		} else if nd := compAsNavDrawer(comp); nd != nil {
			nd.SetOnClose(handler)
		} else if p := compAsPopover(comp); p != nil {
			p.SetOnClose(handler)
		} else {
			comp.SetOnEvent("close", handler)
		}
	case "open":
		if p := compAsPopover(comp); p != nil {
			p.SetOnOpen(handler)
		} else {
			comp.SetOnEvent("open", handler)
		}
	case "commit":
		if cp := compAsColorPicker(comp); cp != nil {
			cp.SetOnCommit(func(sg.Color) { ctx.Provider.CallMethod(methodName) })
		} else {
			comp.SetOnEvent("commit", handler)
		}
	case "select":
		if l := compAsList(comp); l != nil {
			l.SetOnChange(func(int) { ctx.Provider.CallMethod(methodName) })
		} else if tl := compAsTileList(comp); tl != nil {
			tl.SetOnChange(func(int) { ctx.Provider.CallMethod(methodName) })
		} else if sl := compAsSortableList(comp); sl != nil {
			sl.SetOnChange(func(int) { ctx.Provider.CallMethod(methodName) })
		} else if st := compAsSortableTreeList(comp); st != nil {
			st.SetOnChange(func(int) { ctx.Provider.CallMethod(methodName) })
		} else {
			comp.SetOnEvent("select", handler)
		}
	case "reorder":
		if sl := compAsSortableList(comp); sl != nil {
			sl.SetOnReorder(func(int, int) { ctx.Provider.CallMethod(methodName) })
		} else if st := compAsSortableTreeList(comp); st != nil {
			st.SetOnReorder(func(string, string, int) { ctx.Provider.CallMethod(methodName) })
		} else {
			comp.SetOnEvent("reorder", handler)
		}
	case "remove":
		if t := compAsTag(comp); t != nil {
			t.SetOnRemove(handler)
		} else {
			comp.SetOnEvent("remove", handler)
		}
	case "toggle":
		if a := compAsAccordion(comp); a != nil {
			a.SetOnToggle(func(string, bool) { ctx.Provider.CallMethod(methodName) })
		} else if t := compAsTag(comp); t != nil {
			t.SetOnToggle(func(bool) { ctx.Provider.CallMethod(methodName) })
		} else {
			comp.SetOnEvent("toggle", handler)
		}
	case "linkClick":
		if rt := compAsRichText(comp); rt != nil {
			rt.SetOnLinkClick(func(string) { ctx.Provider.CallMethod(methodName) })
		} else {
			comp.SetOnEvent("linkClick", handler)
		}
	case "crop":
		if ic := compAsImageCropper(comp); ic != nil {
			ic.SetOnCropChanged(func(image.Rectangle) { ctx.Provider.CallMethod(methodName) })
		} else {
			comp.SetOnEvent("crop", handler)
		}
	case "capture":
		if ki := compAsKeybindInput(comp); ki != nil {
			ki.SetOnBindingChanged(func(widget.KeyBinding) { ctx.Provider.CallMethod(methodName) })
		} else {
			comp.SetOnEvent("capture", handler)
		}
	default:
		// Unknown event name — register as a custom event so custom widgets
		// can fire it via Component.FireEvent(name).
		comp.SetOnEvent(attr.Name, handler)
	}
}

func applyBinding(comp *widget.Component, attr IRAttribute, ctx *EvalContext, screen *widget.Screen, setters map[string]componentSetter) {
	setter, ok := setters[attr.Name]
	if !ok {
		return
	}

	// Create a watch effect that re-evaluates the expression and applies it
	handle := widget.WatchEffect(func() {
		val, err := EvalExpression(attr.Expr, ctx)
		if err != nil {
			return
		}
		setter(comp, val)
	})

	if screen != nil {
		screen.TrackRef(handle)
	}
}

func applyShowDirective(comp *widget.Component, dir IRDirective, ctx *EvalContext, screen *widget.Screen) {
	handle := widget.WatchEffect(func() {
		val, err := EvalExpression(dir.Expr, ctx)
		if err != nil {
			return
		}
		comp.SetVisible(toBool(val))
	})
	if screen != nil {
		screen.TrackRef(handle)
	}
}

func applyIfDirective(comp *widget.Component, dir IRDirective, ctx *EvalContext, screen *widget.Screen) {
	// ui:if behaves the same as ui:show for now — controls visibility
	applyShowDirective(comp, dir, ctx, screen)
}

// --- Factories ---

var factories = map[string]componentFactory{
	"Component": func(name string, _ []IRAttribute, _ *FactoryContext) (*widget.Component, error) {
		return widget.NewComponent(name), nil
	},
	"Spacer": func(name string, _ []IRAttribute, _ *FactoryContext) (*widget.Component, error) {
		s := widget.NewSpacer(name, 0, 0)
		s.SetUserData(s)
		return &s.Component, nil
	},
	"Panel": func(name string, _ []IRAttribute, _ *FactoryContext) (*widget.Component, error) {
		p := widget.NewPanel(name)
		p.SetUserData(p)
		return &p.Component, nil
	},
	"AnchorLayout": func(name string, _ []IRAttribute, _ *FactoryContext) (*widget.Component, error) {
		al := widget.NewAnchorLayout(name)
		al.SetUserData(al)
		return &al.Component, nil
	},
	"TwoColumnLayout": func(name string, _ []IRAttribute, _ *FactoryContext) (*widget.Component, error) {
		tl := widget.NewTwoColumnLayout(name)
		tl.SetUserData(tl)
		return &tl.Component, nil
	},
	"Label": func(name string, attrs []IRAttribute, fc *FactoryContext) (*widget.Component, error) {
		text := findStaticAttr(attrs, "text")
		fontName := findStaticAttr(attrs, "font")
		fontSize := parseFloatAttr(attrs, "fontSize", fc.DefaultFontSize)
		l := widget.NewLabel(name, text, fc.font(fontName), fontSize)
		l.SetUserData(l)
		return &l.Component, nil
	},
	"Badge": func(name string, attrs []IRAttribute, fc *FactoryContext) (*widget.Component, error) {
		fontName := findStaticAttr(attrs, "font")
		fontSize := parseFloatAttr(attrs, "fontSize", fc.DefaultFontSize)
		b := widget.NewBadge(name, fc.font(fontName), fontSize)
		b.SetUserData(b)
		return &b.Component, nil
	},
	"Button": func(name string, attrs []IRAttribute, fc *FactoryContext) (*widget.Component, error) {
		text := findStaticAttr(attrs, "text")
		fontName := findStaticAttr(attrs, "font")
		fontSize := parseFloatAttr(attrs, "fontSize", fc.DefaultFontSize)
		b := widget.NewButton(name, text, fc.font(fontName), fontSize)
		b.SetUserData(b)
		return &b.Component, nil
	},
	"IconButton": func(name string, _ []IRAttribute, _ *FactoryContext) (*widget.Component, error) {
		b := widget.NewIconButton(name)
		b.SetUserData(b)
		return &b.Component, nil
	},
	"Toggle": func(name string, _ []IRAttribute, _ *FactoryContext) (*widget.Component, error) {
		t := widget.NewToggle(name)
		t.SetUserData(t)
		return &t.Component, nil
	},
	"Checkbox": func(name string, attrs []IRAttribute, fc *FactoryContext) (*widget.Component, error) {
		text := findStaticAttr(attrs, "text")
		fontName := findStaticAttr(attrs, "font")
		fontSize := parseFloatAttr(attrs, "fontSize", fc.DefaultFontSize)
		c := widget.NewCheckbox(name, text, fc.font(fontName), fontSize)
		c.SetUserData(c)
		return &c.Component, nil
	},
	"TextInput": func(name string, attrs []IRAttribute, fc *FactoryContext) (*widget.Component, error) {
		fontSize := parseFloatAttr(attrs, "fontSize", fc.DefaultFontSize)
		ti := widget.NewTextInput(name, fc.font(""), fontSize)
		ti.SetUserData(ti)
		return &ti.Component, nil
	},
	"MaskedInput": func(name string, attrs []IRAttribute, fc *FactoryContext) (*widget.Component, error) {
		fontSize := parseFloatAttr(attrs, "fontSize", fc.DefaultFontSize)
		mi := widget.NewMaskedInput(name, fc.font(""), fontSize)
		mi.SetUserData(mi)
		return &mi.Component, nil
	},
	"TextArea": func(name string, attrs []IRAttribute, fc *FactoryContext) (*widget.Component, error) {
		fontSize := parseFloatAttr(attrs, "fontSize", fc.DefaultFontSize)
		ta := widget.NewTextArea(name, fc.font(""), fontSize)
		ta.SetUserData(ta)
		return &ta.Component, nil
	},
	"Slider": func(name string, _ []IRAttribute, _ *FactoryContext) (*widget.Component, error) {
		s := widget.NewSlider(name)
		s.SetUserData(s)
		return &s.Component, nil
	},
	"ScrollBar": func(name string, _ []IRAttribute, _ *FactoryContext) (*widget.Component, error) {
		s := widget.NewScrollBar(name)
		s.SetUserData(s)
		return &s.Component, nil
	},
	"ProgressBar": func(name string, attrs []IRAttribute, fc *FactoryContext) (*widget.Component, error) {
		p := widget.NewProgressBar(name)
		if parseBoolAttr(attrs, "showLabel", false) {
			p.SetShowLabel(true, fc.DefaultFont, fc.DefaultFontSize)
		}
		p.SetUserData(p)
		return &p.Component, nil
	},
	"MeterBar": func(name string, _ []IRAttribute, _ *FactoryContext) (*widget.Component, error) {
		m := widget.NewMeterBar(name)
		m.SetUserData(m)
		return &m.Component, nil
	},
	"List": func(name string, attrs []IRAttribute, _ *FactoryContext) (*widget.Component, error) {
		h := parseFloatAttr(attrs, "itemHeight", 30)
		l := widget.NewList(name, h)
		l.SetUserData(l)
		return &l.Component, nil
	},
	"TileList": func(name string, attrs []IRAttribute, _ *FactoryContext) (*widget.Component, error) {
		tw := parseFloatAttr(attrs, "tileWidth", 64)
		th := parseFloatAttr(attrs, "tileHeight", 64)
		tl := widget.NewTileList(name, tw, th)
		tl.SetUserData(tl)
		return &tl.Component, nil
	},
	"TreeList": func(name string, attrs []IRAttribute, _ *FactoryContext) (*widget.Component, error) {
		h := parseFloatAttr(attrs, "itemHeight", 30)
		tl := widget.NewTreeList(name, h)
		tl.SetUserData(tl)
		return &tl.Component, nil
	},
	"TabBar": func(name string, attrs []IRAttribute, fc *FactoryContext) (*widget.Component, error) {
		fontSize := parseFloatAttr(attrs, "fontSize", fc.DefaultFontSize)
		tb := widget.NewTabBar(name, fc.font(""), fontSize)
		tb.SetUserData(tb)
		return &tb.Component, nil
	},
	"ToggleButtonBar": func(name string, attrs []IRAttribute, fc *FactoryContext) (*widget.Component, error) {
		fontSize := parseFloatAttr(attrs, "fontSize", fc.DefaultFontSize)
		tbb := widget.NewToggleButtonBar(name, fc.font(""), fontSize)
		if btnsStr := findStaticAttr(attrs, "buttons"); btnsStr != "" {
			for _, s := range strings.Split(btnsStr, ",") {
				s = strings.TrimSpace(s)
				if s != "" {
					tbb.AddButton(s)
				}
			}
		}
		tbb.SetUserData(tbb)
		return &tbb.Component, nil
	},
	"NumberStepper": func(name string, attrs []IRAttribute, fc *FactoryContext) (*widget.Component, error) {
		fontSize := parseFloatAttr(attrs, "fontSize", fc.DefaultFontSize)
		ns := widget.NewNumberStepper(name, fc.font(""), fontSize)
		ns.SetUserData(ns)
		return &ns.Component, nil
	},
"ScrollPanel": func(name string, _ []IRAttribute, _ *FactoryContext) (*widget.Component, error) {
		sp := widget.NewScrollPanel(name)
		sp.SetUserData(sp)
		return &sp.Component, nil
	},
	"NavDrawer": func(name string, attrs []IRAttribute, _ *FactoryContext) (*widget.Component, error) {
		nd := widget.NewNavDrawer(name)
		nd.SetUserData(nd)
		return &nd.Component, nil
	},
	"Window": func(name string, attrs []IRAttribute, fc *FactoryContext) (*widget.Component, error) {
		title := findStaticAttr(attrs, "title")
		fontName := findStaticAttr(attrs, "font")
		fontSize := parseFloatAttr(attrs, "fontSize", fc.DefaultFontSize)
		w := widget.NewWindow(name, title, fc.font(fontName), fontSize)
		w.SetUserData(w)
		return &w.Component, nil
	},
	"RichText": func(name string, attrs []IRAttribute, fc *FactoryContext) (*widget.Component, error) {
		fontSize := parseFloatAttr(attrs, "fontSize", fc.DefaultFontSize)
		rt := widget.NewRichText(name, fc.font(""), fontSize)
		rt.SetUserData(rt)
		return &rt.Component, nil
	},
	"Radio": func(name string, _ []IRAttribute, _ *FactoryContext) (*widget.Component, error) {
		rg := widget.NewRadio(name)
		rg.SetUserData(rg)
		return &rg.Component, nil
	},
	"Tab": func(name string, _ []IRAttribute, _ *FactoryContext) (*widget.Component, error) {
		// Tab is a pseudo-element only valid inside TabBar.
		// If instantiated outside TabBar, it creates a plain component.
		return widget.NewComponent(name), nil
	},
	"RadioButton": func(name string, _ []IRAttribute, _ *FactoryContext) (*widget.Component, error) {
		// RadioButton is a pseudo-element only valid inside Radio.
		// If instantiated outside Radio, it creates a plain component.
		return widget.NewComponent(name), nil
	},
	"Select": func(name string, attrs []IRAttribute, fc *FactoryContext) (*widget.Component, error) {
		var opts []widget.SelectOption
		if optsStr := findStaticAttr(attrs, "options"); optsStr != "" {
			for _, s := range strings.Split(optsStr, ",") {
				s = strings.TrimSpace(s)
				if s != "" {
					opts = append(opts, widget.SelectOption{Label: s})
				}
			}
		}
		if len(opts) == 0 {
			opts = []widget.SelectOption{{Label: "Option 1"}}
		}
		fontSize := parseFloatAttr(attrs, "fontSize", fc.DefaultFontSize)
		s := widget.NewSelect(name, opts, fc.font(""), fontSize)
		s.SetUserData(s)
		return &s.Component, nil
	},
	"SortableList": func(name string, attrs []IRAttribute, _ *FactoryContext) (*widget.Component, error) {
		h := parseFloatAttr(attrs, "itemHeight", 30)
		sl := widget.NewSortableList(name, h)
		sl.SetUserData(sl)
		return &sl.Component, nil
	},
	"SortableTreeList": func(name string, attrs []IRAttribute, fc *FactoryContext) (*widget.Component, error) {
		fontSize := parseFloatAttr(attrs, "fontSize", fc.DefaultFontSize)
		st := widget.NewSortableTreeList(name, fc.font(""), fontSize)
		if v := findStaticAttr(attrs, "allowReparent"); v != "" {
			st.SetAllowReparent(v == "true")
		}
		if v := findStaticAttr(attrs, "allowCrossLevel"); v != "" {
			st.SetAllowCrossLevel(v == "true")
		}
		st.SetUserData(st)
		return &st.Component, nil
	},
	"DragHandle": func(name string, _ []IRAttribute, _ *FactoryContext) (*widget.Component, error) {
		dh := widget.NewDragHandle(name)
		dh.SetUserData(dh)
		return &dh.Component, nil
	},
	"Image": func(name string, _ []IRAttribute, _ *FactoryContext) (*widget.Component, error) {
		im := widget.NewImage(name)
		im.SetUserData(im)
		return &im.Component, nil
	},
	"AnimatedImage": func(name string, _ []IRAttribute, _ *FactoryContext) (*widget.Component, error) {
		ai := widget.NewAnimatedImage(name)
		ai.SetUserData(ai)
		return &ai.Component, nil
	},
	"InputField": func(name string, attrs []IRAttribute, fc *FactoryContext) (*widget.Component, error) {
		fontSize := parseFloatAttr(attrs, "fontSize", fc.DefaultFontSize)
		f := widget.NewInputField(name, fc.font(""), fontSize)
		f.SetUserData(f)
		return &f.Component, nil
	},
	"SearchBox": func(name string, attrs []IRAttribute, fc *FactoryContext) (*widget.Component, error) {
		fontSize := parseFloatAttr(attrs, "fontSize", fc.DefaultFontSize)
		sb := widget.NewSearchBox(name, fc.font(""), fontSize)
		sb.SetUserData(sb)
		return &sb.Component, nil
	},
	"OptionRotator": func(name string, attrs []IRAttribute, fc *FactoryContext) (*widget.Component, error) {
		var opts []string
		if optsStr := findStaticAttr(attrs, "options"); optsStr != "" {
			for _, s := range strings.Split(optsStr, ",") {
				s = strings.TrimSpace(s)
				if s != "" {
					opts = append(opts, s)
				}
			}
		}
		if len(opts) == 0 {
			opts = []string{"Option 1"}
		}
		fontSize := parseFloatAttr(attrs, "fontSize", fc.DefaultFontSize)
		or := widget.NewOptionRotator(name, opts, fc.font(""), fontSize)
		or.SetUserData(or)
		return &or.Component, nil
	},
	"Tag": func(name string, attrs []IRAttribute, fc *FactoryContext) (*widget.Component, error) {
		fontName := findStaticAttr(attrs, "font")
		fontSize := parseFloatAttr(attrs, "fontSize", fc.DefaultFontSize)
		t := widget.NewTag(name, fc.font(fontName), fontSize)
		t.SetUserData(t)
		return &t.Component, nil
	},
	"TagBar": func(name string, attrs []IRAttribute, fc *FactoryContext) (*widget.Component, error) {
		fontName := findStaticAttr(attrs, "font")
		fontSize := parseFloatAttr(attrs, "fontSize", fc.DefaultFontSize)
		tb := widget.NewTagBar(name, fc.font(fontName), fontSize)
		tb.SetUserData(tb)
		return &tb.Component, nil
	},
	"Accordion": func(name string, attrs []IRAttribute, fc *FactoryContext) (*widget.Component, error) {
		acc := widget.NewAccordion(name)
		fontSize := parseFloatAttr(attrs, "fontSize", fc.DefaultFontSize)
		acc.SetFont(fc.DefaultFont, fontSize)
		acc.SetUserData(acc)
		return &acc.Component, nil
	},
	"DataTable": func(name string, attrs []IRAttribute, _ *FactoryContext) (*widget.Component, error) {
		rh := parseFloatAttr(attrs, "rowHeight", 28)
		dt := widget.NewDataTable(name, rh)
		if v := parseFloatAttr(attrs, "headerHeight", 0); v > 0 {
			dt.SetHeaderHeight(v)
		}
		switch findStaticAttr(attrs, "scrollMode") {
		case "static":
			dt.SetScrollMode(widget.ScrollModeStatic)
		}
		switch findStaticAttr(attrs, "selectionMode") {
		case "multi":
			dt.SetSelectionMode(widget.SelectionModeMulti)
		case "none":
			dt.SetSelectionMode(widget.SelectionModeNone)
		}
		if v := parseBoolAttr(attrs, "zebraStriping", false); v {
			dt.SetZebraStriping(true)
		}
		if !parseBoolAttr(attrs, "showHeader", true) {
			dt.SetShowHeader(false)
		}
		if !parseBoolAttr(attrs, "showScrollBar", true) {
			dt.SetShowScrollBar(false)
		}
		if parseBoolAttr(attrs, "showColumnDividers", false) {
			dt.SetShowColumnDividers(true)
		}
		if parseBoolAttr(attrs, "showRowDividers", false) {
			dt.SetShowRowDividers(true)
		}
		if !parseBoolAttr(attrs, "rowClickSelects", true) {
			dt.SetRowClickSelects(false)
		}
		switch findStaticAttr(attrs, "onSortScroll") {
		case "selection":
			dt.SetOnSortScroll(widget.OnSortScrollToSelection)
		case "top":
			dt.SetOnSortScroll(widget.OnSortScrollToTop)
		}
		dt.SetUserData(dt)
		return &dt.Component, nil
	},
	"Column": func(name string, _ []IRAttribute, _ *FactoryContext) (*widget.Component, error) {
		// Column is a pseudo-element only valid inside DataTable.
		return widget.NewComponent(name), nil
	},
	"Section": func(name string, _ []IRAttribute, _ *FactoryContext) (*widget.Component, error) {
		// Section is a pseudo-element only valid inside Accordion.
		return widget.NewComponent(name), nil
	},
	"TimePicker": func(name string, attrs []IRAttribute, fc *FactoryContext) (*widget.Component, error) {
		fontSize := parseFloatAttr(attrs, "fontSize", fc.DefaultFontSize)
		tp := widget.NewTimePicker(name, fc.font(""), fontSize)
		tp.SetUserData(tp)
		return &tp.Component, nil
	},
	"KeybindInput": func(name string, attrs []IRAttribute, fc *FactoryContext) (*widget.Component, error) {
		fontName := findStaticAttr(attrs, "font")
		fontSize := parseFloatAttr(attrs, "fontSize", fc.DefaultFontSize)
		kb := widget.NewKeybindInput(name, fc.font(fontName), fontSize)
		kb.SetUserData(kb)
		return &kb.Component, nil
	},
	"ImageCropper": func(name string, _ []IRAttribute, _ *FactoryContext) (*widget.Component, error) {
		ic := widget.NewImageCropper(name)
		ic.SetUserData(ic)
		return &ic.Component, nil
	},
	"CalendarSelector": func(name string, attrs []IRAttribute, fc *FactoryContext) (*widget.Component, error) {
		fontName := findStaticAttr(attrs, "font")
		fontSize := parseFloatAttr(attrs, "fontSize", fc.DefaultFontSize)
		cs := widget.NewCalendarSelector(name, fc.font(fontName), fontSize)
		cs.SetUserData(cs)
		return &cs.Component, nil
	},
	"ColorPicker": func(name string, attrs []IRAttribute, fc *FactoryContext) (*widget.Component, error) {
		fontSize := parseFloatAttr(attrs, "fontSize", fc.DefaultFontSize)
		cp := widget.NewColorPicker(name, fc.font(""), fontSize)
		cp.SetUserData(cp)
		return &cp.Component, nil
	},
	"Tooltip": func(name string, attrs []IRAttribute, fc *FactoryContext) (*widget.Component, error) {
		tt := widget.NewTooltip(name)
		if v := findStaticAttr(attrs, "text"); v != "" {
			tt.SetText(v, fc.DefaultFont, fc.DefaultFontSize)
		}
		tt.SetUserData(tt)
		return &tt.Component, nil
	},
	"Popover": func(name string, attrs []IRAttribute, fc *FactoryContext) (*widget.Component, error) {
		p := widget.NewPopover(name)
		if v := findStaticAttr(attrs, "title"); v != "" {
			p.SetTitle(v, fc.DefaultFont, fc.DefaultFontSize)
		}
		p.SetUserData(p)
		return &p.Component, nil
	},
	"ToolBar": func(name string, attrs []IRAttribute, _ *FactoryContext) (*widget.Component, error) {
		tb := widget.NewToolBar(name)
		switch findStaticAttr(attrs, "orientation") {
		case "vertical":
			tb.SetOrientation(widget.Vertical)
		}
		switch findStaticAttr(attrs, "overflowMode") {
		case "wrap":
			tb.SetOverflowMode(widget.ToolBarWrap)
		case "scroll":
			tb.SetOverflowMode(widget.ToolBarScroll)
		}
		if parseBoolAttr(attrs, "wrap", false) {
			tb.SetWrap(true)
		}
		tb.SetUserData(tb)
		return &tb.Component, nil
	},
	"StatWeb": func(name string, attrs []IRAttribute, fc *FactoryContext) (*widget.Component, error) {
		fontSize := parseFloatAttr(attrs, "fontSize", fc.DefaultFontSize)
		sw := widget.NewStatWeb(name, fc.font(""), fontSize)
		if parseBoolAttr(attrs, "editable", false) {
			sw.SetEditable(true)
		}
		if !parseBoolAttr(attrs, "fillEnabled", true) {
			sw.SetFillEnabled(false)
		}
		sw.SetUserData(sw)
		return &sw.Component, nil
	},
	"GradientEditor": func(name string, attrs []IRAttribute, fc *FactoryContext) (*widget.Component, error) {
		fontSize := parseFloatAttr(attrs, "fontSize", fc.DefaultFontSize)
		ge := widget.NewGradientEditor(name, fc.font(""), fontSize)
		if !parseBoolAttr(attrs, "showModeSelector", true) {
			ge.SetShowModeSelector(false)
		}
		ge.SetUserData(ge)
		return &ge.Component, nil
	},
	"MenuBar": func(name string, attrs []IRAttribute, fc *FactoryContext) (*widget.Component, error) {
		fontSize := parseFloatAttr(attrs, "fontSize", fc.DefaultFontSize)
		mb := widget.NewMenuBar(name, fc.font(""), fontSize)
		mb.SetUserData(mb)
		return &mb.Component, nil
	},
	"TreeTable": func(name string, attrs []IRAttribute, fc *FactoryContext) (*widget.Component, error) {
		fontName := findStaticAttr(attrs, "font")
		fontSize := parseFloatAttr(attrs, "fontSize", fc.DefaultFontSize)
		tt := widget.NewTreeTable(name, fc.font(fontName), fontSize)
		tt.SetUserData(tt)
		return &tt.Component, nil
	},
}

// --- Attribute setters ---

// Common setter helpers
// sizer is implemented by widget types that need custom SetSize behavior
// (resizing background, border, hit shape, etc.).
type sizer interface {
	SetSize(w, h float64)
}

func setterSize(c *widget.Component, val any) {
	s := fmt.Sprint(val)
	parts := strings.Split(s, ",")
	if len(parts) == 2 {
		w, _ := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
		h, _ := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
		// Dispatch to the typed SetSize if the widget implements sizer.
		if ud := c.UserData(); ud != nil {
			if sz, ok := ud.(sizer); ok {
				sz.SetSize(w, h)
				return
			}
		}
		c.Width = w
		c.Height = h
		c.MarkLayoutDirty()
	}
}

func setterEnabled(c *widget.Component, val any) {
	c.SetEnabled(parseBoolAttrVal(val))
}

func setterVisible(c *widget.Component, val any) {
	c.SetVisible(toBool(val))
}

func setterLayout(c *widget.Component, val any) {
	s := fmt.Sprint(val)
	switch strings.ToLower(s) {
	case "vbox":
		c.Layout = widget.LayoutVBox
	case "hbox":
		c.Layout = widget.LayoutHBox
	case "grid":
		c.Layout = widget.LayoutGrid
	case "flow":
		c.Layout = widget.LayoutFlow
	case "anchor":
		c.Layout = widget.LayoutAnchor
	default:
		c.Layout = widget.LayoutNone
	}
	c.MarkLayoutDirty()
}

func setterSpacing(c *widget.Component, val any) {
	c.Spacing = toFloat(val)
	c.MarkLayoutDirty()
}

func setterX(c *widget.Component, val any) {
	c.OffsetX = toFloat(val)
}

func setterY(c *widget.Component, val any) {
	c.OffsetY = toFloat(val)
}

var commonSetters = map[string]componentSetter{
	"x":       setterX,
	"y":       setterY,
	"enabled": setterEnabled,
	"visible": setterVisible,
	"layout":  setterLayout,
	"spacing": setterSpacing,
	"gridColumns": func(c *widget.Component, val any) {
		c.GridColumns = int(toFloat(val))
		c.MarkLayoutDirty()
	},
	"flow-row-gap": func(c *widget.Component, val any) {
		c.FlowRowGap = toFloat(val)
		c.MarkLayoutDirty()
	},
	"variant": func(c *widget.Component, val any) {
		// For reactive bindings: resolve via the component's effective theme
		// so custom variant names work once the component is in the scene.
		name := strings.ToLower(fmt.Sprint(val))
		c.SetVariant(c.EffectiveTheme().Variant(name))
	},
	"padding": func(c *widget.Component, val any) {
		c.Padding = parseInsets(val)
		c.MarkLayoutDirty()
	},
	"margin": func(c *widget.Component, val any) {
		c.Margin = parseInsets(val)
		c.MarkLayoutDirty()
	},
	"zIndex": func(c *widget.Component, val any) {
		c.SetZIndex(int(toFloat(val)))
	},
	"align": func(c *widget.Component, val any) {
		c.Align = parseAlignment(val)
		c.MarkLayoutDirty()
	},
	"justify": func(c *widget.Component, val any) {
		c.Justify = parseAlignment(val)
		c.MarkLayoutDirty()
	},
	"minWidth": func(c *widget.Component, val any) {
		c.MinWidth = toFloat(val)
		c.MarkLayoutDirty()
	},
	"minHeight": func(c *widget.Component, val any) {
		c.MinHeight = toFloat(val)
		c.MarkLayoutDirty()
	},
	"maxWidth": func(c *widget.Component, val any) {
		c.MaxWidth = toFloat(val)
		c.MarkLayoutDirty()
	},
	"maxHeight": func(c *widget.Component, val any) {
		c.MaxHeight = toFloat(val)
		c.MarkLayoutDirty()
	},
	"fill": func(c *widget.Component, val any) {
		switch strings.ToLower(fmt.Sprint(val)) {
		case "width":
			c.Fill = widget.FillWidth
		case "height":
			c.Fill = widget.FillHeight
		case "both":
			c.Fill = widget.FillBoth
		}
		c.MarkLayoutDirty()
	},
	"grow": func(c *widget.Component, val any) {
		c.Grow = int(toFloat(val))
		c.MarkLayoutDirty()
	},
	"width": func(c *widget.Component, val any) {
		w := toFloat(val)
		if ud := c.UserData(); ud != nil {
			if sz, ok := ud.(sizer); ok {
				sz.SetSize(w, c.Height)
				return
			}
		}
		c.Width = w
		c.MarkLayoutDirty()
	},
	"height": func(c *widget.Component, val any) {
		h := toFloat(val)
		if ud := c.UserData(); ud != nil {
			if sz, ok := ud.(sizer); ok {
				sz.SetSize(c.Width, h)
				return
			}
		}
		c.Height = h
		c.MarkLayoutDirty()
	},
}

func mergeSetters(maps ...map[string]componentSetter) map[string]componentSetter {
	result := make(map[string]componentSetter)
	for _, m := range maps {
		for k, v := range m {
			result[k] = v
		}
	}
	return result
}

var attrSetters = map[string]map[string]componentSetter{
	"Component": commonSetters,
	"Spacer": mergeSetters(commonSetters, map[string]componentSetter{
		"size": setterSize,
	}),
	"Panel": mergeSetters(commonSetters, map[string]componentSetter{
		"size": setterSize,
		"background": func(c *widget.Component, val any) {
			col := parseColor(fmt.Sprint(val))
			if p := compAsPanel(c); p != nil {
				p.SetBackground(col)
			}
		},
		"border": func(c *widget.Component, val any) {
			if p := compAsPanel(c); p != nil {
				parts := strings.SplitN(fmt.Sprint(val), ",", 2)
				if len(parts) == 2 {
					col := parseColor(strings.TrimSpace(parts[0]))
					w, _ := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
					p.SetBorder(col, w)
				}
			}
		},
		"cornerRadius": func(c *widget.Component, val any) {
			if p := compAsPanel(c); p != nil {
				r := toFloat(val)
				p.SetCornerRadii(r, r, r, r)
			}
		},
		"alignItems": func(c *widget.Component, val any) {
			if p := compAsPanel(c); p != nil {
				p.SetAlignment(parseAlignment(val))
			}
		},
		"justifyContent": func(c *widget.Component, val any) {
			if p := compAsPanel(c); p != nil {
				p.SetJustify(parseAlignment(val))
			}
		},
	}),
	"AnchorLayout": mergeSetters(commonSetters, map[string]componentSetter{
		"size": setterSize,
		"background": func(c *widget.Component, val any) {
			col := parseColor(fmt.Sprint(val))
			if al := compAsAnchorLayout(c); al != nil {
				al.SetBackground(col)
			}
		},
		"border": func(c *widget.Component, val any) {
			if al := compAsAnchorLayout(c); al != nil {
				parts := strings.SplitN(fmt.Sprint(val), ",", 2)
				if len(parts) == 2 {
					col := parseColor(strings.TrimSpace(parts[0]))
					w, _ := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
					al.SetBorder(col, w)
				}
			}
		},
		"cornerRadius": func(c *widget.Component, val any) {
			if al := compAsAnchorLayout(c); al != nil {
				r := toFloat(val)
				al.SetCornerRadii(r, r, r, r)
			}
		},
	}),
	"TwoColumnLayout": mergeSetters(commonSetters, map[string]componentSetter{
		"size": setterSize,
		"background": func(c *widget.Component, val any) {
			if tl := compAsTwoColumnLayout(c); tl != nil {
				tl.SetBackground(parseColor(fmt.Sprint(val)))
			}
		},
		"border": func(c *widget.Component, val any) {
			if tl := compAsTwoColumnLayout(c); tl != nil {
				parts := strings.SplitN(fmt.Sprint(val), ",", 2)
				if len(parts) == 2 {
					col := parseColor(strings.TrimSpace(parts[0]))
					w, _ := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
					tl.SetBorder(col, w)
				}
			}
		},
		"gap": func(c *widget.Component, val any) {
			if tl := compAsTwoColumnLayout(c); tl != nil {
				tl.Gap = toFloat(val)
				tl.MarkLayoutDirty()
			}
		},
		"rowSpacing": func(c *widget.Component, val any) {
			if tl := compAsTwoColumnLayout(c); tl != nil {
				tl.RowSpacing = toFloat(val)
				tl.MarkLayoutDirty()
			}
		},
		"leftWidth": func(c *widget.Component, val any) {
			if tl := compAsTwoColumnLayout(c); tl != nil {
				tl.LeftWidth = toFloat(val)
				tl.MarkLayoutDirty()
			}
		},
		"columnRatio": func(c *widget.Component, val any) {
			if tl := compAsTwoColumnLayout(c); tl != nil {
				tl.ColumnRatio = toFloat(val)
				tl.MarkLayoutDirty()
			}
		},
		"leftAlign": func(c *widget.Component, val any) {
			if tl := compAsTwoColumnLayout(c); tl != nil {
				tl.LeftAlign = parseAlignment(val)
				tl.MarkLayoutDirty()
			}
		},
		"rightAlign": func(c *widget.Component, val any) {
			if tl := compAsTwoColumnLayout(c); tl != nil {
				tl.RightAlign = parseAlignment(val)
				tl.MarkLayoutDirty()
			}
		},
	}),
	"Label": mergeSetters(commonSetters, map[string]componentSetter{
		"text": func(c *widget.Component, val any) {
			if l := compAsLabel(c); l != nil {
				l.SetText(fmt.Sprint(val))
			}
		},
		"color": func(c *widget.Component, val any) {
			col := parseColor(fmt.Sprint(val))
			if l := compAsLabel(c); l != nil {
				l.SetColor(col)
			}
		},
		"align": func(c *widget.Component, val any) {
			if l := compAsLabel(c); l != nil {
				switch strings.ToLower(fmt.Sprint(val)) {
				case "center":
					l.SetAlign(sg.TextAlignCenter)
				case "right":
					l.SetAlign(sg.TextAlignRight)
				default:
					l.SetAlign(sg.TextAlignLeft)
				}
			}
		},
		"wrapWidth": func(c *widget.Component, val any) {
			if l := compAsLabel(c); l != nil {
				l.SetWrapWidth(toFloat(val))
			}
		},
		"bold": func(c *widget.Component, val any) {
			if l := compAsLabel(c); l != nil {
				l.SetBold(parseBoolAttrVal(val))
			}
		},
		"italic": func(c *widget.Component, val any) {
			if l := compAsLabel(c); l != nil {
				l.SetItalic(parseBoolAttrVal(val))
			}
		},
		"fontSize": func(c *widget.Component, val any) {
			if l := compAsLabel(c); l != nil {
				l.SetFontSize(toFloat(val))
			}
		},
		"size": setterSize,
	}),
	"Badge": mergeSetters(commonSetters, map[string]componentSetter{
		"text": func(c *widget.Component, val any) {
			if b := compAsBadge(c); b != nil {
				b.SetText(fmt.Sprint(val))
				b.SizeToContent()
			}
		},
		"count": func(c *widget.Component, val any) {
			if b := compAsBadge(c); b != nil {
				b.SetCount(int(toFloat(val)))
				b.SizeToContent()
			}
		},
		"maxCount": func(c *widget.Component, val any) {
			if b := compAsBadge(c); b != nil {
				b.SetMaxCount(int(toFloat(val)))
				b.SizeToContent()
			}
		},
		"dotMode": func(c *widget.Component, val any) {
			if b := compAsBadge(c); b != nil {
				b.SetDotMode(toBool(val))
				b.SizeToContent()
			}
		},
		"kind": func(c *widget.Component, val any) {
			c.SetVariant(parseVariant(val))
		},
		"size": setterSize,
	}),
	"Tag": mergeSetters(commonSetters, map[string]componentSetter{
		"text": func(c *widget.Component, val any) {
			if t := compAsTag(c); t != nil {
				t.SetText(fmt.Sprint(val))
				t.SizeToContent()
			}
		},
		"kind": func(c *widget.Component, val any) {
			c.SetVariant(parseVariant(val))
		},
		"removable": func(c *widget.Component, val any) {
			if t := compAsTag(c); t != nil {
				t.SetRemovable(toBool(val))
				t.SizeToContent()
			}
		},
		"selectable": func(c *widget.Component, val any) {
			if t := compAsTag(c); t != nil {
				t.SetSelectable(toBool(val))
			}
		},
		"selected": func(c *widget.Component, val any) {
			if t := compAsTag(c); t != nil {
				t.SetSelected(toBool(val))
			}
		},
		"size": setterSize,
	}),
	"TagBar": mergeSetters(commonSetters, map[string]componentSetter{
		"placeholder": func(c *widget.Component, val any) {
			if tb := compAsTagBar(c); tb != nil {
				tb.SetPlaceholder(fmt.Sprint(val))
			}
		},
		"size": setterSize,
	}),
	"Accordion": mergeSetters(commonSetters, map[string]componentSetter{
		"size": setterSize,
		"exclusive": func(c *widget.Component, val any) {
			if a := compAsAccordion(c); a != nil {
				a.SetExclusive(toBool(val))
			}
		},
		"animated": func(c *widget.Component, val any) {
			if a := compAsAccordion(c); a != nil {
				a.SetAnimated(toBool(val))
			}
		},
		"expanded": func(c *widget.Component, val any) {
			if a := compAsAccordion(c); a != nil {
				for _, id := range strings.Split(fmt.Sprint(val), ",") {
					id = strings.TrimSpace(id)
					if id != "" {
						a.SetExpanded(id, true)
					}
				}
			}
		},
	}),
	"Button": mergeSetters(commonSetters, map[string]componentSetter{
		"text": func(c *widget.Component, val any) {
			if b := compAsButton(c); b != nil {
				b.SetText(fmt.Sprint(val))
			}
		},
		"autoSize": func(c *widget.Component, val any) {
			if b := compAsButton(c); b != nil {
				b.SetAutoSize(parseBoolAttrVal(val))
			}
		},
		"size": setterSize,
	}),
	"IconButton": mergeSetters(commonSetters, map[string]componentSetter{
		"size": setterSize,
		"icon": func(c *widget.Component, val any) {
			if ib := compAsIconButton(c); ib != nil {
				ib.SetIconKey(fmt.Sprint(val))
			}
		},
	}),
	"Toggle": mergeSetters(commonSetters, map[string]componentSetter{
		"value": func(c *widget.Component, val any) {
			if t := compAsToggle(c); t != nil {
				t.SetValue(toBool(val))
			}
		},
	}),
	"Checkbox": mergeSetters(commonSetters, map[string]componentSetter{
		"checked": func(c *widget.Component, val any) {
			if cb := compAsCheckbox(c); cb != nil {
				cb.SetChecked(toBool(val))
			}
		},
		"text": func(c *widget.Component, val any) {
			if cb := compAsCheckbox(c); cb != nil {
				cb.SetText(fmt.Sprint(val))
			}
		},
	}),
	"TextInput": mergeSetters(commonSetters, map[string]componentSetter{
		"value": func(c *widget.Component, val any) {
			if ti := compAsTextInput(c); ti != nil {
				ti.SetValue(fmt.Sprint(val))
			}
		},
		"placeholder": func(c *widget.Component, val any) {
			if ti := compAsTextInput(c); ti != nil {
				ti.SetPlaceholder(fmt.Sprint(val))
			}
		},
		"maxLength": func(c *widget.Component, val any) {
			if ti := compAsTextInput(c); ti != nil {
				ti.SetMaxLength(int(toFloat(val)))
			}
		},
		"width": func(c *widget.Component, val any) {
			if ti := compAsTextInput(c); ti != nil {
				ti.SetWidth(toFloat(val))
			}
		},
		"size": setterSize,
	}),
	"MaskedInput": mergeSetters(commonSetters, map[string]componentSetter{
		"mask": func(c *widget.Component, val any) {
			if mi := compAsMaskedInput(c); mi != nil {
				mi.SetMask(fmt.Sprint(val))
			}
		},
		"value": func(c *widget.Component, val any) {
			if mi := compAsMaskedInput(c); mi != nil {
				mi.SetValue(fmt.Sprint(val))
			}
		},
		"rawValue": func(c *widget.Component, val any) {
			if mi := compAsMaskedInput(c); mi != nil {
				mi.SetRawValue(fmt.Sprint(val))
			}
		},
		"placeholder": func(c *widget.Component, val any) {
			if mi := compAsMaskedInput(c); mi != nil {
				mi.SetPlaceholder(fmt.Sprint(val))
			}
		},
		"maskPlaceholder": func(c *widget.Component, val any) {
			s := fmt.Sprint(val)
			if len([]rune(s)) > 0 {
				if mi := compAsMaskedInput(c); mi != nil {
					mi.SetMaskPlaceholder([]rune(s)[0])
				}
			}
		},
		"maxLength": func(c *widget.Component, val any) {
			if mi := compAsMaskedInput(c); mi != nil {
				mi.SetMaxLength(int(toFloat(val)))
			}
		},
		"width": func(c *widget.Component, val any) {
			if mi := compAsMaskedInput(c); mi != nil {
				mi.SetWidth(toFloat(val))
			}
		},
		"size": setterSize,
	}),
	"TextArea": mergeSetters(commonSetters, map[string]componentSetter{
		"value": func(c *widget.Component, val any) {
			if ta := compAsTextArea(c); ta != nil {
				ta.SetValue(fmt.Sprint(val))
			}
		},
		"maxLength": func(c *widget.Component, val any) {
			if ta := compAsTextArea(c); ta != nil {
				ta.SetMaxLength(int(toFloat(val)))
			}
		},
		"rows": func(c *widget.Component, val any) {
			if ta := compAsTextArea(c); ta != nil {
				ta.SetRows(int(toFloat(val)))
			}
		},
		"size": setterSize,
	}),
	"Slider": mergeSetters(commonSetters, map[string]componentSetter{
		"value": func(c *widget.Component, val any) {
			if s := compAsSlider(c); s != nil {
				s.SetValue(toFloat(val))
			}
		},
		"range": func(c *widget.Component, val any) {
			if s := compAsSlider(c); s != nil {
				parts := strings.SplitN(fmt.Sprint(val), ",", 2)
				if len(parts) == 2 {
					min, _ := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
					max, _ := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
					s.SetRange(min, max)
				}
			}
		},
		"step": func(c *widget.Component, val any) {
			if s := compAsSlider(c); s != nil {
				s.SetStep(toFloat(val))
			}
		},
		"orientation": func(c *widget.Component, val any) {
			if s := compAsSlider(c); s != nil {
				switch strings.ToLower(fmt.Sprint(val)) {
				case "vertical":
					s.SetOrientation(widget.Vertical)
				default:
					s.SetOrientation(widget.Horizontal)
				}
			}
		},
		"size": setterSize,
	}),
	"ScrollBar": mergeSetters(commonSetters, map[string]componentSetter{
		"value": func(c *widget.Component, val any) {
			if sb := compAsScrollBar(c); sb != nil {
				sb.SetScrollPos(toFloat(val))
			}
		},
		"orientation": func(c *widget.Component, val any) {
			if sb := compAsScrollBar(c); sb != nil {
				switch strings.ToLower(fmt.Sprint(val)) {
				case "horizontal":
					sb.SetOrientation(widget.Horizontal)
				default:
					sb.SetOrientation(widget.Vertical)
				}
			}
		},
		"size": setterSize,
	}),
	"ProgressBar": mergeSetters(commonSetters, map[string]componentSetter{
		"value": func(c *widget.Component, val any) {
			if p := compAsProgressBar(c); p != nil {
				p.SetValue(toFloat(val))
			}
		},
		"range": func(c *widget.Component, val any) {
			if p := compAsProgressBar(c); p != nil {
				parts := strings.SplitN(fmt.Sprint(val), ",", 2)
				if len(parts) == 2 {
					min, _ := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
					max, _ := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
					p.SetRange(min, max)
				}
			}
		},
		"fillColor": func(c *widget.Component, val any) {
			if p := compAsProgressBar(c); p != nil {
				p.SetFillColor(parseColor(fmt.Sprint(val)))
			}
		},
		"size": setterSize,
	}),
	"MeterBar": mergeSetters(commonSetters, map[string]componentSetter{
		"value": func(c *widget.Component, val any) {
			if mb := compAsMeterBar(c); mb != nil {
				mb.SetValue(toFloat(val))
			}
		},
		"range": func(c *widget.Component, val any) {
			if mb := compAsMeterBar(c); mb != nil {
				parts := strings.SplitN(fmt.Sprint(val), ",", 2)
				if len(parts) == 2 {
					min, _ := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
					max, _ := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
					mb.SetRange(min, max)
				}
			}
		},
		"size": setterSize,
	}),
	"List": mergeSetters(commonSetters, map[string]componentSetter{
		"size": setterSize,
		"selected": func(c *widget.Component, val any) {
			if l := compAsList(c); l != nil {
				l.SetSelected(int(toFloat(val)))
			}
		},
	}),
	"TileList": mergeSetters(commonSetters, map[string]componentSetter{
		"size": setterSize,
		"selected": func(c *widget.Component, val any) {
			if tl := compAsTileList(c); tl != nil {
				tl.SetSelected(int(toFloat(val)))
			}
		},
		"columns": func(c *widget.Component, val any) {
			if tl := compAsTileList(c); tl != nil {
				tl.SetColumns(int(toFloat(val)))
			}
		},
	}),
	"TreeList": mergeSetters(commonSetters, map[string]componentSetter{
		"size": setterSize,
		"selectable": func(c *widget.Component, val any) {
			if tl := compAsTreeList(c); tl != nil {
				tl.SetSelectable(parseBoolAttrVal(val))
			}
		},
		"leafOnlySelection": func(c *widget.Component, val any) {
			if tl := compAsTreeList(c); tl != nil {
				tl.SetLeafOnlySelection(parseBoolAttrVal(val))
			}
		},
	}),
	"TabBar": mergeSetters(commonSetters, map[string]componentSetter{
		"size": setterSize,
		"selected": func(c *widget.Component, val any) {
			if tb := compAsTabBar(c); tb != nil {
				tb.SetSelected(int(toFloat(val)))
			}
		},
		"overflowMode": func(c *widget.Component, val any) {
			if tb := compAsTabBar(c); tb != nil {
				s, _ := val.(string)
				switch s {
				case "scroll":
					tb.SetOverflowMode(widget.TabOverflowScroll)
				default:
					tb.SetOverflowMode(widget.TabOverflowClip)
				}
			}
		},
	}),
	"ToggleButtonBar": mergeSetters(commonSetters, map[string]componentSetter{
		"size": setterSize,
		"selected": func(c *widget.Component, val any) {
			if tbb := compAsToggleButtonBar(c); tbb != nil {
				tbb.SetSelected(int(toFloat(val)))
			}
		},
	}),
	"NumberStepper": mergeSetters(commonSetters, map[string]componentSetter{
		"size": setterSize,
		"value": func(c *widget.Component, val any) {
			if ns := compAsNumberStepper(c); ns != nil {
				ns.SetValue(toFloat(val))
			}
		},
		"min": func(c *widget.Component, val any) {
			if ns := compAsNumberStepper(c); ns != nil {
				ns.SetMin(toFloat(val))
			}
		},
		"max": func(c *widget.Component, val any) {
			if ns := compAsNumberStepper(c); ns != nil {
				ns.SetMax(toFloat(val))
			}
		},
		"step": func(c *widget.Component, val any) {
			if ns := compAsNumberStepper(c); ns != nil {
				ns.SetStep(toFloat(val))
			}
		},
		"pageStep": func(c *widget.Component, val any) {
			if ns := compAsNumberStepper(c); ns != nil {
				ns.SetPageStep(toFloat(val))
			}
		},
		"decimals": func(c *widget.Component, val any) {
			if ns := compAsNumberStepper(c); ns != nil {
				ns.SetDecimals(int(toFloat(val)))
			}
		},
	}),
"ScrollPanel": mergeSetters(commonSetters, map[string]componentSetter{
		"size": setterSize,
		"background": func(c *widget.Component, val any) {
			col := parseColor(fmt.Sprint(val))
			if sp := compAsScrollPanel(c); sp != nil {
				sp.SetBackground(col)
			}
		},
		"scrollX": func(c *widget.Component, val any) {
			if sp := compAsScrollPanel(c); sp != nil {
				sp.SetScrollX(toFloat(val))
			}
		},
		"scrollY": func(c *widget.Component, val any) {
			if sp := compAsScrollPanel(c); sp != nil {
				sp.SetScrollY(toFloat(val))
			}
		},
	}),
	"NavDrawer": mergeSetters(commonSetters, map[string]componentSetter{
		"anchor": func(c *widget.Component, val any) {
			if nd := compAsNavDrawer(c); nd != nil {
				switch fmt.Sprint(val) {
				case "right":
					nd.SetAnchor(widget.NavDrawerRight)
				default:
					nd.SetAnchor(widget.NavDrawerLeft)
				}
			}
		},
		"width": func(c *widget.Component, val any) {
			if nd := compAsNavDrawer(c); nd != nil {
				nd.SetWidth(toFloat(val))
			}
		},
		"pinned": func(c *widget.Component, val any) {
			if nd := compAsNavDrawer(c); nd != nil {
				nd.SetPinned(parseBoolAttrVal(val))
			}
		},
		"size": setterSize,
	}),
	"Window": mergeSetters(commonSetters, map[string]componentSetter{
		"title": func(c *widget.Component, val any) {
			if w := compAsWindow(c); w != nil {
				w.SetTitle(fmt.Sprint(val))
			}
		},
		"resizable": func(c *widget.Component, val any) {
			if w := compAsWindow(c); w != nil {
				w.SetResizable(parseBoolAttrVal(val))
			}
		},
		"movable": func(c *widget.Component, val any) {
			if w := compAsWindow(c); w != nil {
				w.SetMovable(parseBoolAttrVal(val))
			}
		},
		"modal": func(c *widget.Component, val any) {
			if w := compAsWindow(c); w != nil {
				w.SetModal(parseBoolAttrVal(val))
			}
		},
		"closeable": func(c *widget.Component, val any) {
			if w := compAsWindow(c); w != nil {
				w.SetCloseable(parseBoolAttrVal(val))
			}
		},
		"minWidth": func(c *widget.Component, val any) {
			if w := compAsWindow(c); w != nil {
				w.SetMinWidth(toFloat(val))
			}
		},
		"minHeight": func(c *widget.Component, val any) {
			if w := compAsWindow(c); w != nil {
				w.SetMinHeight(toFloat(val))
			}
		},
		"escResult": func(c *widget.Component, val any) {
			if w := compAsWindow(c); w != nil {
				w.SetEscResult(fmt.Sprint(val))
			}
		},
		"enterResult": func(c *widget.Component, val any) {
			if w := compAsWindow(c); w != nil {
				w.SetEnterResult(fmt.Sprint(val))
			}
		},
		"size": setterSize,
	}),
	"RichText": mergeSetters(commonSetters, map[string]componentSetter{
		"markup": func(c *widget.Component, val any) {
			if rt := compAsRichText(c); rt != nil {
				rt.SetMarkup(fmt.Sprint(val))
			}
		},
		"wrapWidth": func(c *widget.Component, val any) {
			if rt := compAsRichText(c); rt != nil {
				rt.SetWrapWidth(toFloat(val))
			}
		},
		"color": func(c *widget.Component, val any) {
			col := parseColor(fmt.Sprint(val))
			if rt := compAsRichText(c); rt != nil {
				rt.SetColor(col)
			}
		},
		"align": func(c *widget.Component, val any) {
			if rt := compAsRichText(c); rt != nil {
				switch strings.ToLower(fmt.Sprint(val)) {
				case "center":
					rt.SetAlign(sg.TextAlignCenter)
				case "right":
					rt.SetAlign(sg.TextAlignRight)
				default:
					rt.SetAlign(sg.TextAlignLeft)
				}
			}
		},
		"headingScale": func(c *widget.Component, val any) {
			if rt := compAsRichText(c); rt != nil {
				parts := strings.SplitN(fmt.Sprint(val), ",", 3)
				if len(parts) == 3 {
					h1, _ := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
					h2, _ := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
					h3, _ := strconv.ParseFloat(strings.TrimSpace(parts[2]), 64)
					rt.SetHeadingScale(h1, h2, h3)
				}
			}
		},
	}),
	"Radio": mergeSetters(commonSetters, map[string]componentSetter{
		"selected": func(c *widget.Component, val any) {
			if rg := compAsRadio(c); rg != nil {
				rg.SetSelected(int(toFloat(val)))
			}
		},
		"columns": func(c *widget.Component, val any) {
			if rg := compAsRadio(c); rg != nil {
				rg.SetColumns(int(toFloat(val)))
			}
		},
		"verticalFirst": func(c *widget.Component, val any) {
			if rg := compAsRadio(c); rg != nil {
				rg.SetVerticalFirst(parseBoolAttrVal(val))
			}
		},
	}),
	"RadioButton": commonSetters,
	"Tab":         commonSetters,
	"Select": mergeSetters(commonSetters, map[string]componentSetter{
		"size": setterSize,
		"selected": func(c *widget.Component, val any) {
			if s := compAsSelect(c); s != nil {
				s.SetSelected(int(toFloat(val)))
			}
		},
		"options": func(c *widget.Component, val any) {
			if s := compAsSelect(c); s != nil {
				var opts []widget.SelectOption
				for _, o := range strings.Split(fmt.Sprint(val), ",") {
					o = strings.TrimSpace(o)
					if o != "" {
						opts = append(opts, widget.SelectOption{Label: o})
					}
				}
				if len(opts) > 0 {
					s.SetOptions(opts)
				}
			}
		},
	}),
	"OptionRotator": mergeSetters(commonSetters, map[string]componentSetter{
		"size": setterSize,
		"selected": func(c *widget.Component, val any) {
			if or := compAsOptionRotator(c); or != nil {
				or.SetSelected(int(toFloat(val)))
			}
		},
		"wrap": func(c *widget.Component, val any) {
			if or := compAsOptionRotator(c); or != nil {
				or.SetWrap(parseBoolAttrVal(val))
			}
		},
		"options": func(c *widget.Component, val any) {
			if or := compAsOptionRotator(c); or != nil {
				var opts []string
				for _, o := range strings.Split(fmt.Sprint(val), ",") {
					o = strings.TrimSpace(o)
					if o != "" {
						opts = append(opts, o)
					}
				}
				if len(opts) > 0 {
					or.SetOptions(opts)
				}
			}
		},
	}),
	"SortableList": mergeSetters(commonSetters, map[string]componentSetter{
		"size": setterSize,
		"selected": func(c *widget.Component, val any) {
			if sl := compAsSortableList(c); sl != nil {
				sl.SetSelected(int(toFloat(val)))
			}
		},
		"showHandles": func(c *widget.Component, val any) {
			if sl := compAsSortableList(c); sl != nil {
				sl.SetShowHandles(toBool(val))
			}
		},
		"dragEnabled": func(c *widget.Component, val any) {
			if sl := compAsSortableList(c); sl != nil {
				sl.SetDragEnabled(toBool(val))
			}
		},
		"keyboardReorderEnabled": func(c *widget.Component, val any) {
			if sl := compAsSortableList(c); sl != nil {
				sl.SetKeyboardReorderEnabled(toBool(val))
			}
		},
		"handleSide": func(c *widget.Component, val any) {
			if sl := compAsSortableList(c); sl != nil {
				switch strings.ToLower(fmt.Sprint(val)) {
				case "right":
					sl.SetHandleSide(widget.SortHandleRight)
				default:
					sl.SetHandleSide(widget.SortHandleLeft)
				}
			}
		},
	}),
	"SortableTreeList": mergeSetters(commonSetters, map[string]componentSetter{
		"size": setterSize,
		"selected": func(c *widget.Component, val any) {
			if st := compAsSortableTreeList(c); st != nil {
				st.SetSelected(int(toFloat(val)))
			}
		},
		"allowReparent": func(c *widget.Component, val any) {
			if st := compAsSortableTreeList(c); st != nil {
				st.SetAllowReparent(toBool(val))
			}
		},
		"allowCrossLevel": func(c *widget.Component, val any) {
			if st := compAsSortableTreeList(c); st != nil {
				st.SetAllowCrossLevel(toBool(val))
			}
		},
	}),
	"DragHandle": mergeSetters(commonSetters, map[string]componentSetter{
		"size": setterSize,
		"axis": func(c *widget.Component, val any) {
			if dh := compAsDragHandle(c); dh != nil {
				switch strings.ToLower(fmt.Sprint(val)) {
				case "x":
					dh.SetAxis(widget.DragAxisX)
				case "y":
					dh.SetAxis(widget.DragAxisY)
				case "diagonal":
					dh.SetAxis(widget.DragAxisDiagonal)
				}
			}
		},
		"min": func(c *widget.Component, val any) {
			if dh := compAsDragHandle(c); dh != nil {
				dh.SetMin(toFloat(val))
			}
		},
		"max": func(c *widget.Component, val any) {
			if dh := compAsDragHandle(c); dh != nil {
				dh.SetMax(toFloat(val))
			}
		},
		"gripStyle": func(c *widget.Component, val any) {
			if dh := compAsDragHandle(c); dh != nil {
				switch strings.ToLower(fmt.Sprint(val)) {
				case "dots":
					dh.SetGripStyle(widget.DragGripDots)
				case "lines":
					dh.SetGripStyle(widget.DragGripLines)
				case "none":
					dh.SetGripStyle(widget.DragGripNone)
				}
			}
		},
	}),
	"Image": mergeSetters(commonSetters, map[string]componentSetter{
		"size": setterSize,
		"src": func(c *widget.Component, val any) {
			if im := compAsImage(c); im != nil {
				if img, err := loadImageFile(fmt.Sprint(val)); err == nil {
					im.SetImage(img)
				}
			}
		},
		"atlas": func(c *widget.Component, val any) {
			if im := compAsImage(c); im != nil {
				if t := c.EffectiveTheme(); t != nil {
					sr := t.GetSprite(fmt.Sprint(val))
					if sr.Set {
						im.SetImage(sr.Image)
					}
				}
			}
		},
		"scaleMode": func(c *widget.Component, val any) {
			if im := compAsImage(c); im != nil {
				switch strings.ToLower(fmt.Sprint(val)) {
				case "stretch":
					im.SetScaleMode(widget.ImageScaleStretch)
				case "fit":
					im.SetScaleMode(widget.ImageScaleFit)
				case "fill":
					im.SetScaleMode(widget.ImageScaleFill)
				case "center":
					im.SetScaleMode(widget.ImageScaleCenter)
				case "tile":
					im.SetScaleMode(widget.ImageScaleTile)
				}
			}
		},
		"cornerRadius": func(c *widget.Component, val any) {
			if im := compAsImage(c); im != nil {
				im.SetCornerRadius(toFloat(val))
			}
		},
		"alpha": func(c *widget.Component, val any) {
			if im := compAsImage(c); im != nil {
				im.SetAlpha(float32(toFloat(val)))
			}
		},
	}),
	"AnimatedImage": mergeSetters(commonSetters, map[string]componentSetter{
		"size": setterSize,
		"gif": func(c *widget.Component, val any) {
			if ai := compAsAnimatedImage(c); ai != nil {
				if g, err := loadGIFFile(fmt.Sprint(val)); err == nil {
					ai.LoadGIF(g)
					ai.Play()
				}
			}
		},
		"atlas": func(c *widget.Component, val any) {
			if ai := compAsAnimatedImage(c); ai != nil {
				if t := c.EffectiveTheme(); t != nil {
					sr := t.GetSprite(fmt.Sprint(val))
					if sr.Set {
						b := sr.Image.Bounds()
						ai.SetAtlas(sr.Image, b.Dx(), b.Dy())
					}
				}
			}
		},
		"fps": func(c *widget.Component, val any) {
			if ai := compAsAnimatedImage(c); ai != nil {
				ai.SetFPS(toFloat(val))
			}
		},
		"playMode": func(c *widget.Component, val any) {
			if ai := compAsAnimatedImage(c); ai != nil {
				switch strings.ToLower(fmt.Sprint(val)) {
				case "once":
					ai.SetPlayMode(widget.AnimPlayOnce)
				case "loop":
					ai.SetPlayMode(widget.AnimPlayLoop)
				case "pingpong", "ping-pong":
					ai.SetPlayMode(widget.AnimPlayPingPong)
				}
			}
		},
		"cornerRadius": func(c *widget.Component, val any) {
			if ai := compAsAnimatedImage(c); ai != nil {
				ai.SetCornerRadius(toFloat(val))
			}
		},
	}),
	"InputField": mergeSetters(commonSetters, map[string]componentSetter{
		"label": func(c *widget.Component, val any) {
			if f := compAsInputField(c); f != nil {
				f.SetLabel(fmt.Sprint(val))
			}
		},
		"value": func(c *widget.Component, val any) {
			if f := compAsInputField(c); f != nil {
				f.SetValue(fmt.Sprint(val))
			}
		},
		"placeholder": func(c *widget.Component, val any) {
			if f := compAsInputField(c); f != nil {
				f.SetPlaceholder(fmt.Sprint(val))
			}
		},
		"required": func(c *widget.Component, val any) {
			if f := compAsInputField(c); f != nil {
				f.SetRequired(toBool(val))
			}
		},
		"maxLength": func(c *widget.Component, val any) {
			if f := compAsInputField(c); f != nil {
				f.SetMaxLength(int(toFloat(val)))
			}
		},
		"labelPosition": func(c *widget.Component, val any) {
			if f := compAsInputField(c); f != nil {
				switch strings.ToLower(fmt.Sprint(val)) {
				case "left":
					f.SetLabelPosition(widget.LabelLeft)
				default:
					f.SetLabelPosition(widget.LabelAbove)
				}
			}
		},
		"validationState": func(c *widget.Component, val any) {
			if f := compAsInputField(c); f != nil {
				switch strings.ToLower(fmt.Sprint(val)) {
				case "error":
					f.SetValidationState(widget.ValidationError)
				case "warning":
					f.SetValidationState(widget.ValidationWarning)
				case "success":
					f.SetValidationState(widget.ValidationSuccess)
				default:
					f.SetValidationState(widget.ValidationNone)
				}
			}
		},
		"validationMessage": func(c *widget.Component, val any) {
			if f := compAsInputField(c); f != nil {
				f.SetValidationMessage(fmt.Sprint(val))
			}
		},
		"width": func(c *widget.Component, val any) {
			if f := compAsInputField(c); f != nil {
				f.SetWidth(toFloat(val))
			}
		},
		"size": setterSize,
	}),
	"SearchBox": mergeSetters(commonSetters, map[string]componentSetter{
		"value": func(c *widget.Component, val any) {
			if sb := compAsSearchBox(c); sb != nil {
				sb.SetValue(fmt.Sprint(val))
			}
		},
		"placeholder": func(c *widget.Component, val any) {
			if sb := compAsSearchBox(c); sb != nil {
				sb.SetPlaceholder(fmt.Sprint(val))
			}
		},
		"debounceMs": func(c *widget.Component, val any) {
			if sb := compAsSearchBox(c); sb != nil {
				ms := int(toFloat(val))
				sb.SetDebounce(time.Duration(ms) * time.Millisecond)
			}
		},
		"minQueryLength": func(c *widget.Component, val any) {
			if sb := compAsSearchBox(c); sb != nil {
				sb.SetMinQueryLength(int(toFloat(val)))
			}
		},
		"showSearchIcon": func(c *widget.Component, val any) {
			if sb := compAsSearchBox(c); sb != nil {
				sb.SetShowSearchIcon(toBool(val))
			}
		},
		"showClearButton": func(c *widget.Component, val any) {
			if sb := compAsSearchBox(c); sb != nil {
				sb.SetShowClearButton(toBool(val))
			}
		},
		"width": func(c *widget.Component, val any) {
			if sb := compAsSearchBox(c); sb != nil {
				sb.SetWidth(toFloat(val))
			}
		},
		"size": setterSize,
	}),
	"TimePicker": mergeSetters(commonSetters, map[string]componentSetter{
		"size": setterSize,
		"format": func(c *widget.Component, val any) {
			if tp := compAsTimePicker(c); tp != nil {
				switch fmt.Sprint(val) {
				case "12h":
					tp.SetFormat(widget.TimeFormat12h)
				case "24h":
					tp.SetFormat(widget.TimeFormat24h)
				}
			}
		},
		"showSeconds": func(c *widget.Component, val any) {
			if tp := compAsTimePicker(c); tp != nil {
				tp.SetShowSeconds(toBool(val))
			}
		},
		"time": func(c *widget.Component, val any) {
			if tp := compAsTimePicker(c); tp != nil {
				switch v := val.(type) {
				case time.Time:
					tp.SetTime(v.Hour(), v.Minute(), v.Second())
				case string:
					if t, err := time.Parse("15:04:05", v); err == nil {
						tp.SetTime(t.Hour(), t.Minute(), t.Second())
					} else if t, err := time.Parse("15:04", v); err == nil {
						tp.SetTime(t.Hour(), t.Minute(), 0)
					}
				}
			}
		},
	}),
	"KeybindInput": mergeSetters(commonSetters, map[string]componentSetter{
		"size": setterSize,
	}),
	"ImageCropper": mergeSetters(commonSetters, map[string]componentSetter{
		"size": setterSize,
		"showGrid": func(c *widget.Component, val any) {
			if ic := compAsImageCropper(c); ic != nil {
				ic.SetShowGrid(toBool(val))
			}
		},
	}),
	"CalendarSelector": mergeSetters(commonSetters, map[string]componentSetter{
		"size": setterSize,
		"popupMode": func(c *widget.Component, val any) {
			if cs := compAsCalendarSelector(c); cs != nil {
				cs.SetPopupMode(toBool(val))
			}
		},
		"date": func(c *widget.Component, val any) {
			if cs := compAsCalendarSelector(c); cs != nil {
				switch v := val.(type) {
				case time.Time:
					cs.SetDate(v)
				case string:
					if t, err := time.Parse("2006-01-02", v); err == nil {
						cs.SetDate(t)
					}
				}
			}
		},
	}),
	"ColorPicker": mergeSetters(commonSetters, map[string]componentSetter{
		"size": setterSize,
		"color": func(c *widget.Component, val any) {
			if cp := compAsColorPicker(c); cp != nil {
				switch v := val.(type) {
				case sg.Color:
					cp.SetValue(v)
				case string:
					cp.SetValue(parseColor(v))
				}
			}
		},
		"showAlpha": func(c *widget.Component, val any) {
			if cp := compAsColorPicker(c); cp != nil {
				cp.SetShowAlpha(parseBoolAttrVal(val))
			}
		},
		"defaultMode": func(c *widget.Component, val any) {
			if cp := compAsColorPicker(c); cp != nil {
				switch strings.ToLower(fmt.Sprint(val)) {
				case "rgb":
					cp.SetDefaultMode(widget.ColorModeRGB)
				case "hsv":
					cp.SetDefaultMode(widget.ColorModeHSV)
				case "hsl":
					cp.SetDefaultMode(widget.ColorModeHSL)
				case "float":
					cp.SetDefaultMode(widget.ColorModeFloat)
				default:
					cp.SetDefaultMode(widget.ColorModeHex)
				}
			}
		},
	}),
	"Tooltip": mergeSetters(commonSetters, map[string]componentSetter{
		"size": setterSize,
		"text": func(c *widget.Component, val any) {
			// text is handled in factory; this setter is for reactive bindings
		},
		"anchor": func(c *widget.Component, val any) {
			if tt := compAsTooltip(c); tt != nil {
				switch strings.ToLower(fmt.Sprint(val)) {
				case "above":
					tt.SetAnchor(widget.TooltipAbove)
				case "below":
					tt.SetAnchor(widget.TooltipBelow)
				case "left":
					tt.SetAnchor(widget.TooltipLeft)
				case "right":
					tt.SetAnchor(widget.TooltipRight)
				case "follow", "followmouse":
					tt.SetAnchor(widget.TooltipFollowMouse)
				default:
					tt.SetAnchor(widget.TooltipAuto)
				}
			}
		},
		"showDelay": func(c *widget.Component, val any) {
			if tt := compAsTooltip(c); tt != nil {
				tt.SetShowDelay(int(toFloat(val)))
			}
		},
		"hideDelay": func(c *widget.Component, val any) {
			if tt := compAsTooltip(c); tt != nil {
				tt.SetHideDelay(int(toFloat(val)))
			}
		},
		"fadeIn": func(c *widget.Component, val any) {
			if tt := compAsTooltip(c); tt != nil {
				tt.SetFadeIn(float32(toFloat(val)))
			}
		},
		"fadeOut": func(c *widget.Component, val any) {
			if tt := compAsTooltip(c); tt != nil {
				tt.SetFadeOut(float32(toFloat(val)))
			}
		},
		"clampToScreen": func(c *widget.Component, val any) {
			if tt := compAsTooltip(c); tt != nil {
				tt.SetClampToScreen(parseBoolAttrVal(val))
			}
		},
	}),
	"Popover": mergeSetters(commonSetters, map[string]componentSetter{
		"size": setterSize,
		"title": func(c *widget.Component, val any) {
			// title is handled in factory; this setter is for reactive bindings
		},
		"preferredSide": func(c *widget.Component, val any) {
			if p := compAsPopover(c); p != nil {
				switch strings.ToLower(fmt.Sprint(val)) {
				case "above":
					p.SetPreferredSide(widget.PopoverAbove)
				case "right":
					p.SetPreferredSide(widget.PopoverRight)
				case "left":
					p.SetPreferredSide(widget.PopoverLeft)
				default:
					p.SetPreferredSide(widget.PopoverBelow)
				}
			}
		},
		"contentSize": func(c *widget.Component, val any) {
			if p := compAsPopover(c); p != nil {
				s := fmt.Sprint(val)
				parts := strings.Split(s, ",")
				if len(parts) == 2 {
					w, _ := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
					h, _ := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
					p.SetContentSize(w, h)
				}
			}
		},
		"showCloseButton": func(c *widget.Component, val any) {
			if p := compAsPopover(c); p != nil {
				p.SetShowCloseButton(parseBoolAttrVal(val))
			}
		},
	}),
	"ToolBar": mergeSetters(commonSetters, map[string]componentSetter{
		"size": setterSize,
		"orientation": func(c *widget.Component, val any) {
			if tb := compAsToolBar(c); tb != nil {
				switch strings.ToLower(fmt.Sprint(val)) {
				case "vertical":
					tb.SetOrientation(widget.Vertical)
				default:
					tb.SetOrientation(widget.Horizontal)
				}
			}
		},
		"overflowMode": func(c *widget.Component, val any) {
			if tb := compAsToolBar(c); tb != nil {
				switch strings.ToLower(fmt.Sprint(val)) {
				case "wrap":
					tb.SetOverflowMode(widget.ToolBarWrap)
				case "scroll":
					tb.SetOverflowMode(widget.ToolBarScroll)
				default:
					tb.SetOverflowMode(widget.ToolBarClip)
				}
			}
		},
		"wrap": func(c *widget.Component, val any) {
			if tb := compAsToolBar(c); tb != nil {
				tb.SetWrap(parseBoolAttrVal(val))
			}
		},
	}),
	"StatWeb": mergeSetters(commonSetters, map[string]componentSetter{
		"size": setterSize,
		"editable": func(c *widget.Component, val any) {
			if sw := compAsStatWeb(c); sw != nil {
				sw.SetEditable(parseBoolAttrVal(val))
			}
		},
		"fillEnabled": func(c *widget.Component, val any) {
			if sw := compAsStatWeb(c); sw != nil {
				sw.SetFillEnabled(parseBoolAttrVal(val))
			}
		},
	}),
	"GradientEditor": mergeSetters(commonSetters, map[string]componentSetter{
		"size": setterSize,
		"showModeSelector": func(c *widget.Component, val any) {
			if ge := compAsGradientEditor(c); ge != nil {
				ge.SetShowModeSelector(parseBoolAttrVal(val))
			}
		},
	}),
	"MenuBar": mergeSetters(commonSetters, map[string]componentSetter{
		"size": setterSize,
	}),
}

// --- Component type casting helpers ---
// Each factory stores a back-reference from the Component's node UserData
// so we can recover the typed wrapper.

func compAsLabel(c *widget.Component) *widget.Label {
	if v := c.UserData(); v != nil {
		if l, ok := v.(*widget.Label); ok {
			return l
		}
	}
	return nil
}

func compAsBadge(c *widget.Component) *widget.Badge {
	if v := c.UserData(); v != nil {
		if b, ok := v.(*widget.Badge); ok {
			return b
		}
	}
	return nil
}

func compAsTag(c *widget.Component) *widget.Tag {
	if v := c.UserData(); v != nil {
		if t, ok := v.(*widget.Tag); ok {
			return t
		}
	}
	return nil
}

func compAsTagBar(c *widget.Component) *widget.TagBar {
	if v := c.UserData(); v != nil {
		if tb, ok := v.(*widget.TagBar); ok {
			return tb
		}
	}
	return nil
}

func compAsButton(c *widget.Component) *widget.Button {
	if v := c.UserData(); v != nil {
		if b, ok := v.(*widget.Button); ok {
			return b
		}
	}
	return nil
}

// CompAsLabel is the exported equivalent of compAsLabel, for use by the root
// package tests which cannot call the unexported version.
func CompAsLabel(c *widget.Component) *widget.Label { return compAsLabel(c) }

// CompAsButton is the exported equivalent of compAsButton, for use by the root
// package tests which cannot call the unexported version.
func CompAsButton(c *widget.Component) *widget.Button { return compAsButton(c) }

func compAsIconButton(c *widget.Component) *widget.IconButton {
	if v := c.UserData(); v != nil {
		if b, ok := v.(*widget.IconButton); ok {
			return b
		}
	}
	return nil
}

func compAsPanel(c *widget.Component) *widget.Panel {
	if v := c.UserData(); v != nil {
		if p, ok := v.(*widget.Panel); ok {
			return p
		}
	}
	return nil
}

func compAsToggle(c *widget.Component) *widget.Toggle {
	if v := c.UserData(); v != nil {
		if t, ok := v.(*widget.Toggle); ok {
			return t
		}
	}
	return nil
}

func compAsCheckbox(c *widget.Component) *widget.Checkbox {
	if v := c.UserData(); v != nil {
		if cb, ok := v.(*widget.Checkbox); ok {
			return cb
		}
	}
	return nil
}

func compAsTextInput(c *widget.Component) *widget.TextInput {
	if v := c.UserData(); v != nil {
		if ti, ok := v.(*widget.TextInput); ok {
			return ti
		}
	}
	return nil
}

func compAsMaskedInput(c *widget.Component) *widget.MaskedInput {
	if v := c.UserData(); v != nil {
		if mi, ok := v.(*widget.MaskedInput); ok {
			return mi
		}
	}
	return nil
}

func compAsTextArea(c *widget.Component) *widget.TextArea {
	if v := c.UserData(); v != nil {
		if ta, ok := v.(*widget.TextArea); ok {
			return ta
		}
	}
	return nil
}

func compAsSlider(c *widget.Component) *widget.Slider {
	if v := c.UserData(); v != nil {
		if s, ok := v.(*widget.Slider); ok {
			return s
		}
	}
	return nil
}

func compAsProgressBar(c *widget.Component) *widget.MeterBar {
	if v := c.UserData(); v != nil {
		if p, ok := v.(*widget.MeterBar); ok {
			return p
		}
	}
	return nil
}

func compAsScrollPanel(c *widget.Component) *widget.ScrollPanel {
	if v := c.UserData(); v != nil {
		if sp, ok := v.(*widget.ScrollPanel); ok {
			return sp
		}
	}
	return nil
}

func compAsNavDrawer(c *widget.Component) *widget.NavDrawer {
	if v := c.UserData(); v != nil {
		if nd, ok := v.(*widget.NavDrawer); ok {
			return nd
		}
	}
	return nil
}

func compAsScrollBar(c *widget.Component) *widget.ScrollBar {
	if v := c.UserData(); v != nil {
		if sb, ok := v.(*widget.ScrollBar); ok {
			return sb
		}
	}
	return nil
}

func compAsMeterBar(c *widget.Component) *widget.MeterBar {
	if v := c.UserData(); v != nil {
		if mb, ok := v.(*widget.MeterBar); ok {
			return mb
		}
	}
	return nil
}

func compAsWindow(c *widget.Component) *widget.Window {
	if v := c.UserData(); v != nil {
		if w, ok := v.(*widget.Window); ok {
			return w
		}
	}
	return nil
}

func compAsRichText(c *widget.Component) *widget.RichText {
	if v := c.UserData(); v != nil {
		if rt, ok := v.(*widget.RichText); ok {
			return rt
		}
	}
	return nil
}

func compAsRadio(c *widget.Component) *widget.Radio {
	if v := c.UserData(); v != nil {
		if rg, ok := v.(*widget.Radio); ok {
			return rg
		}
	}
	return nil
}

func compAsTabBar(c *widget.Component) *widget.TabBar {
	if v := c.UserData(); v != nil {
		if tb, ok := v.(*widget.TabBar); ok {
			return tb
		}
	}
	return nil
}

func compAsToggleButtonBar(c *widget.Component) *widget.ToggleButtonBar {
	if v := c.UserData(); v != nil {
		if tbb, ok := v.(*widget.ToggleButtonBar); ok {
			return tbb
		}
	}
	return nil
}

func compAsNumberStepper(c *widget.Component) *widget.NumberStepper {
	if v := c.UserData(); v != nil {
		if ns, ok := v.(*widget.NumberStepper); ok {
			return ns
		}
	}
	return nil
}

func compAsAnchorLayout(c *widget.Component) *widget.AnchorLayout {
	if v := c.UserData(); v != nil {
		if al, ok := v.(*widget.AnchorLayout); ok {
			return al
		}
	}
	return nil
}

func compAsTwoColumnLayout(c *widget.Component) *widget.TwoColumnLayout {
	if v := c.UserData(); v != nil {
		if tl, ok := v.(*widget.TwoColumnLayout); ok {
			return tl
		}
	}
	return nil
}

func compAsList(c *widget.Component) *widget.List {
	if v := c.UserData(); v != nil {
		if l, ok := v.(*widget.List); ok {
			return l
		}
	}
	return nil
}

func compAsTileList(c *widget.Component) *widget.TileList {
	if v := c.UserData(); v != nil {
		if tl, ok := v.(*widget.TileList); ok {
			return tl
		}
	}
	return nil
}

func compAsSelect(c *widget.Component) *widget.Select {
	if v := c.UserData(); v != nil {
		if s, ok := v.(*widget.Select); ok {
			return s
		}
	}
	return nil
}

func compAsOptionRotator(c *widget.Component) *widget.OptionRotator {
	if v := c.UserData(); v != nil {
		if or, ok := v.(*widget.OptionRotator); ok {
			return or
		}
	}
	return nil
}

func compAsAccordion(c *widget.Component) *widget.Accordion {
	if v := c.UserData(); v != nil {
		if a, ok := v.(*widget.Accordion); ok {
			return a
		}
	}
	return nil
}

func compAsTreeList(c *widget.Component) *widget.TreeList {
	if v := c.UserData(); v != nil {
		if tl, ok := v.(*widget.TreeList); ok {
			return tl
		}
	}
	return nil
}

func compAsSortableList(c *widget.Component) *widget.SortableList {
	if v := c.UserData(); v != nil {
		if sl, ok := v.(*widget.SortableList); ok {
			return sl
		}
	}
	return nil
}

func compAsSortableTreeList(c *widget.Component) *widget.SortableTreeList {
	if v := c.UserData(); v != nil {
		if st, ok := v.(*widget.SortableTreeList); ok {
			return st
		}
	}
	return nil
}

func compAsDragHandle(c *widget.Component) *widget.DragHandle {
	if v := c.UserData(); v != nil {
		if dh, ok := v.(*widget.DragHandle); ok {
			return dh
		}
	}
	return nil
}

func compAsImage(c *widget.Component) *widget.Image {
	if v := c.UserData(); v != nil {
		if im, ok := v.(*widget.Image); ok {
			return im
		}
	}
	return nil
}

func compAsAnimatedImage(c *widget.Component) *widget.AnimatedImage {
	if v := c.UserData(); v != nil {
		if ai, ok := v.(*widget.AnimatedImage); ok {
			return ai
		}
	}
	return nil
}

func compAsInputField(c *widget.Component) *widget.InputField {
	if v := c.UserData(); v != nil {
		if f, ok := v.(*widget.InputField); ok {
			return f
		}
	}
	return nil
}

func compAsSearchBox(c *widget.Component) *widget.SearchBox {
	if v := c.UserData(); v != nil {
		if sb, ok := v.(*widget.SearchBox); ok {
			return sb
		}
	}
	return nil
}

func compAsDataTable(c *widget.Component) *widget.DataTable {
	if v := c.UserData(); v != nil {
		if dt, ok := v.(*widget.DataTable); ok {
			return dt
		}
	}
	return nil
}

func compAsTimePicker(c *widget.Component) *widget.TimePicker {
	if v := c.UserData(); v != nil {
		if tp, ok := v.(*widget.TimePicker); ok {
			return tp
		}
	}
	return nil
}

func compAsKeybindInput(c *widget.Component) *widget.KeybindInput {
	if v := c.UserData(); v != nil {
		if kb, ok := v.(*widget.KeybindInput); ok {
			return kb
		}
	}
	return nil
}

func compAsImageCropper(c *widget.Component) *widget.ImageCropper {
	if v := c.UserData(); v != nil {
		if ic, ok := v.(*widget.ImageCropper); ok {
			return ic
		}
	}
	return nil
}

func compAsCalendarSelector(c *widget.Component) *widget.CalendarSelector {
	if v := c.UserData(); v != nil {
		if cs, ok := v.(*widget.CalendarSelector); ok {
			return cs
		}
	}
	return nil
}

func compAsColorPicker(c *widget.Component) *widget.ColorPicker {
	if v := c.UserData(); v != nil {
		if cp, ok := v.(*widget.ColorPicker); ok {
			return cp
		}
	}
	return nil
}

func compAsTooltip(c *widget.Component) *widget.Tooltip {
	if v := c.UserData(); v != nil {
		if tt, ok := v.(*widget.Tooltip); ok {
			return tt
		}
	}
	return nil
}

func compAsPopover(c *widget.Component) *widget.Popover {
	if v := c.UserData(); v != nil {
		if p, ok := v.(*widget.Popover); ok {
			return p
		}
	}
	return nil
}

func compAsToolBar(c *widget.Component) *widget.ToolBar {
	if v := c.UserData(); v != nil {
		if tb, ok := v.(*widget.ToolBar); ok {
			return tb
		}
	}
	return nil
}

func compAsStatWeb(c *widget.Component) *widget.StatWeb {
	if v := c.UserData(); v != nil {
		if sw, ok := v.(*widget.StatWeb); ok {
			return sw
		}
	}
	return nil
}

func compAsGradientEditor(c *widget.Component) *widget.GradientEditor {
	if v := c.UserData(); v != nil {
		if ge, ok := v.(*widget.GradientEditor); ok {
			return ge
		}
	}
	return nil
}

func compAsMenuBar(c *widget.Component) *widget.MenuBar {
	if v := c.UserData(); v != nil {
		if mb, ok := v.(*widget.MenuBar); ok {
			return mb
		}
	}
	return nil
}

// parseAnchorAttrs extracts anchor, anchorOffsetX, and anchorOffsetY from a
// child's IR attributes. Returns defaults (AnchorTopLeft, 0, 0) if absent.
func parseAnchorAttrs(attrs []IRAttribute) (widget.Anchor, float64, float64) {
	anchor := widget.AnchorTopLeft
	var ox, oy float64
	for _, a := range attrs {
		if a.Expr != nil || a.IsEvent {
			continue
		}
		switch a.Name {
		case "anchor":
			anchor = parseAnchorString(a.Static)
		case "anchorOffsetX", "offsetX":
			ox, _ = strconv.ParseFloat(a.Static, 64)
		case "anchorOffsetY", "offsetY":
			oy, _ = strconv.ParseFloat(a.Static, 64)
		}
	}
	return anchor, ox, oy
}

func parseAnchorString(s string) widget.Anchor {
	switch strings.ToLower(strings.ReplaceAll(s, " ", "-")) {
	case "top-left":
		return widget.AnchorTopLeft
	case "top-center":
		return widget.AnchorTopCenter
	case "top-right":
		return widget.AnchorTopRight
	case "middle-left":
		return widget.AnchorMiddleLeft
	case "center":
		return widget.AnchorCenter
	case "middle-right":
		return widget.AnchorMiddleRight
	case "bottom-left":
		return widget.AnchorBottomLeft
	case "bottom-center":
		return widget.AnchorBottomCenter
	case "bottom-right":
		return widget.AnchorBottomRight
	default:
		return widget.AnchorTopLeft
	}
}

// --- Helpers ---

func findStaticAttr(attrs []IRAttribute, name string) string {
	for _, a := range attrs {
		if a.Name == name && a.Expr == nil && !a.IsEvent {
			return a.Static
		}
	}
	return ""
}

func parseFloatAttr(attrs []IRAttribute, name string, def float64) float64 {
	s := findStaticAttr(attrs, name)
	if s == "" {
		return def
	}
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return def
	}
	return f
}

// parseBoolAttr returns the boolean value of a named attribute, or def if not found.
func parseBoolAttr(attrs []IRAttribute, name string, def bool) bool {
	s := findStaticAttr(attrs, name)
	if s == "" {
		return def
	}
	return s == "true" || s == "1"
}

// parseBoolAttrVal handles string attribute values like "true"/"false"
// as well as actual bool and numeric types.
func parseBoolAttrVal(val any) bool {
	if s, ok := val.(string); ok {
		return s == "true" || s == "1"
	}
	return toBool(val)
}

// autoFitContent computes a component's size from its children so that
// container wrappers created by the template (e.g. a ui:show Panel) take
// up the right amount of space in a parent VBox/HBox layout.
// Only computes dimensions that are currently zero (partial auto-fit).
func autoFitContent(c *widget.Component) {
	fitW := c.Width == 0
	fitH := c.Height == 0

	if c.Layout == widget.LayoutHBox {
		var totalW, maxH float64
		for i, child := range c.Children() {
			if child.Width > 0 {
				if i > 0 {
					totalW += c.Spacing
				}
				totalW += child.Width
			}
			if child.Height > maxH {
				maxH = child.Height
			}
		}
		w, h := c.Width, c.Height
		if fitW {
			w = totalW + c.Padding.Horizontal()
		}
		if fitH {
			h = maxH + c.Padding.Vertical()
		}
		applyAutoFitSize(c, w, h)
		return
	}
	var maxW, totalH float64
	for i, child := range c.Children() {
		if child.Width > maxW {
			maxW = child.Width
		}
		totalH += child.Height
		if i > 0 {
			totalH += c.Spacing
		}
	}
	w, h := c.Width, c.Height
	if fitW {
		w = maxW + c.Padding.Horizontal()
	}
	if fitH {
		h = totalH + c.Padding.Vertical()
	}
	applyAutoFitSize(c, w, h)
}

// applyAutoFitSize sets the computed width/height, dispatching through the
// typed SetSize if the widget implements sizer (e.g. Button, Panel).
func applyAutoFitSize(c *widget.Component, w, h float64) {
	if ud := c.UserData(); ud != nil {
		if sz, ok := ud.(sizer); ok {
			sz.SetSize(w, h)
			return
		}
	}
	c.Width = w
	c.Height = h
}

// ParseColor is the exported equivalent of parseColor, for use by the root
// package tests which cannot call the unexported version.
func ParseColor(s string) sg.Color { return parseColor(s) }

func parseColor(s string) sg.Color {
	s = strings.TrimSpace(s)
	if len(s) == 0 {
		return sg.Color{}
	}
	// Support hex colors like #RRGGBB or #RRGGBBAA
	if s[0] == '#' {
		s = s[1:]
		if len(s) == 6 {
			s += "ff"
		}
		if len(s) == 8 {
			r, _ := strconv.ParseUint(s[0:2], 16, 8)
			g, _ := strconv.ParseUint(s[2:4], 16, 8)
			b, _ := strconv.ParseUint(s[4:6], 16, 8)
			a, _ := strconv.ParseUint(s[6:8], 16, 8)
			return sg.RGBA(float64(r)/255, float64(g)/255, float64(b)/255, float64(a)/255)
		}
	}
	return sg.Color{}
}

func parseAlignment(val any) widget.Alignment {
	switch strings.ToLower(fmt.Sprint(val)) {
	case "center":
		return widget.AlignCenter
	case "end":
		return widget.AlignEnd
	case "space-between", "spacebetween":
		return widget.AlignSpaceBetween
	default:
		return widget.AlignStart
	}
}

func parseVariant(val any) widget.Variant {
	switch strings.ToLower(fmt.Sprint(val)) {
	case "secondary":
		return widget.Secondary
	case "accent":
		return widget.Accent
	case "neutral":
		return widget.Neutral
	case "danger":
		return widget.Danger
	case "success":
		return widget.Success
	case "warning":
		return widget.Warning
	case "info":
		return widget.Info
	case "custom1":
		return widget.Custom1
	case "custom2":
		return widget.Custom2
	case "custom3":
		return widget.Custom3
	case "custom4":
		return widget.Custom4
	case "custom5":
		return widget.Custom5
	case "custom6":
		return widget.Custom6
	case "custom7":
		return widget.Custom7
	case "custom8":
		return widget.Custom8
	case "custom9":
		return widget.Custom9
	case "custom10":
		return widget.Custom10
	case "custom11":
		return widget.Custom11
	case "custom12":
		return widget.Custom12
	case "custom13":
		return widget.Custom13
	case "custom14":
		return widget.Custom14
	case "custom15":
		return widget.Custom15
	case "custom16":
		return widget.Custom16
	case "custom17":
		return widget.Custom17
	case "custom18":
		return widget.Custom18
	case "custom19":
		return widget.Custom19
	case "custom20":
		return widget.Custom20
	case "custom21":
		return widget.Custom21
	case "custom22":
		return widget.Custom22
	case "custom23":
		return widget.Custom23
	case "custom24":
		return widget.Custom24
	case "custom25":
		return widget.Custom25
	case "custom26":
		return widget.Custom26
	case "custom27":
		return widget.Custom27
	case "custom28":
		return widget.Custom28
	case "custom29":
		return widget.Custom29
	case "custom30":
		return widget.Custom30
	case "custom31":
		return widget.Custom31
	case "custom32":
		return widget.Custom32
	case "custom33":
		return widget.Custom33
	case "custom34":
		return widget.Custom34
	case "custom35":
		return widget.Custom35
	case "custom36":
		return widget.Custom36
	case "custom37":
		return widget.Custom37
	case "custom38":
		return widget.Custom38
	case "custom39":
		return widget.Custom39
	case "custom40":
		return widget.Custom40
	case "custom41":
		return widget.Custom41
	case "custom42":
		return widget.Custom42
	case "custom43":
		return widget.Custom43
	case "custom44":
		return widget.Custom44
	case "custom45":
		return widget.Custom45
	case "custom46":
		return widget.Custom46
	case "custom47":
		return widget.Custom47
	case "custom48":
		return widget.Custom48
	case "custom49":
		return widget.Custom49
	case "custom50":
		return widget.Custom50
	case "custom51":
		return widget.Custom51
	case "custom52":
		return widget.Custom52
	case "custom53":
		return widget.Custom53
	case "custom54":
		return widget.Custom54
	case "custom55":
		return widget.Custom55
	case "custom56":
		return widget.Custom56
	default:
		return widget.Primary
	}
}

// parseInsets parses a CSS-like insets string into an Insets value.
// Formats: "v" → all sides, "v,h" → vertical/horizontal, "t,r,b,l" → per side.
func parseInsets(val any) widget.Insets {
	parts := strings.Split(strings.TrimSpace(fmt.Sprint(val)), ",")
	switch len(parts) {
	case 1:
		v, _ := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
		return widget.Insets{Top: v, Right: v, Bottom: v, Left: v}
	case 2:
		v, _ := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
		h, _ := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
		return widget.Insets{Top: v, Bottom: v, Right: h, Left: h}
	case 4:
		t, _ := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
		r, _ := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
		b, _ := strconv.ParseFloat(strings.TrimSpace(parts[2]), 64)
		l, _ := strconv.ParseFloat(strings.TrimSpace(parts[3]), 64)
		return widget.Insets{Top: t, Right: r, Bottom: b, Left: l}
	}
	return widget.Insets{}
}

// loadImageFile loads an image file from disk and returns an engine.Image.
func loadImageFile(path string) (engine.Image, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open image %q: %w", path, err)
	}
	defer f.Close()
	decoded, _, err := image.Decode(f)
	if err != nil {
		return nil, fmt.Errorf("decode image %q: %w", path, err)
	}
	return engine.NewImageFromImage(decoded), nil
}

func compAsTreeTable(c *widget.Component) *widget.TreeTable {
	if v := c.UserData(); v != nil {
		if tt, ok := v.(*widget.TreeTable); ok {
			return tt
		}
	}
	return nil
}

// loadGIFFile loads a GIF file from disk and returns a *gif.GIF.
func loadGIFFile(path string) (*gif.GIF, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open gif %q: %w", path, err)
	}
	defer f.Close()
	g, err := gif.DecodeAll(f)
	if err != nil {
		return nil, fmt.Errorf("decode gif %q: %w", path, err)
	}
	return g, nil
}
