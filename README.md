# WillowUI

[![CI](https://github.com/devthicket/willowui/actions/workflows/ci.yml/badge.svg)](https://github.com/devthicket/willowui/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/devthicket/willowui.svg)](https://pkg.go.dev/github.com/devthicket/willowui)
[![Go Report Card](https://goreportcard.com/badge/github.com/devthicket/willowui)](https://goreportcard.com/report/github.com/devthicket/willowui)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

A reactive UI toolkit for Go, built on [Willow](https://github.com/devthicket/willow) and [Ebitengine](https://ebitengine.org). WillowUI provides a complete widget library, reactive state management, JSON theming, and XML templating for building desktop and game UIs.

> **[Read the full documentation at devthicket.org/willow-ui](https://www.devthicket.org/willow-ui)**
>
> Guides, API reference, theming cookbook, and interactive examples.

**Status:** Actively developed. API may change before `v1.0.0`.

<p align="center">
  <img src="https://www.devthicket.org/willow-ui/gif/button.gif" width="380" alt="Buttons">
  <img src="https://www.devthicket.org/willow-ui/gif/text-input.gif" width="380" alt="Text Input">
</p>
<p align="center">
  <img src="https://www.devthicket.org/willow-ui/gif/sortable-tree-list.gif" width="380" alt="Sortable Tree List">
  <img src="https://www.devthicket.org/willow-ui/gif/calendar.gif" width="380" alt="Calendar">
</p>

## Features

- **50+ widgets** -- buttons, text inputs, sliders, lists, trees, data tables, menus, color pickers, modals, toasts, and more
- **Reactive state** -- `Ref`, `Computed`, `WatchEffect`, reactive `Array` and `Record` types that drive automatic UI updates
- **JSON theming** -- style every widget property from a single JSON file; ships with 7 built-in themes
- **XML templates** -- declare layouts in XML with reactive bindings, conditionals, and event handlers
- **Layout system** -- VBox, HBox, Anchor, Flow, and TwoColumn layouts with padding, spacing, and alignment
- **Rich text** -- inline markup for bold, italic, color, and size within a single text widget
- **Focus management** -- tab navigation, focus rings, and keyboard-driven interaction
- **Screen management** -- push, pop, and replace screens with a built-in stage manager
- **Visual test runner** -- automated screenshot testing via JSON scripts and input injection

## Quick Start

```bash
go get github.com/devthicket/willowui
```

```go
package main

import (
	"github.com/devthicket/willow"
	ui "github.com/devthicket/willowui"
)

func main() {
	font := ui.MustLoadDefaultFont()
	clicks := ui.NewRef(0)

	btn := ui.NewButton("inc", "+1", font, 16)
	btn.SetOnClick(ui.Increment(clicks, 1))

	label := ui.NewLabel("count", "0", font, 36)
	formatted, _ := ui.BindFormatterf(clicks, "Count: %d") // app-lifetime: handle not stopped
	label.BindText(formatted)

	row := ui.NewHBox("row")
	row.Spacing = 16
	row.SetPosition(20, 15)
	row.AddChild(btn)
	row.AddChild(label)

	ui.Setup(ui.StageConfig{
		Title:      "Counter",
		Width:      400,
		Height:     70,
		ClearColor: willow.RGBA(0.08, 0.08, 0.10, 1),
	}, row)
}
```

## Widgets

| Category | Widgets |
|---|---|
| Input | Button, IconButton, TextInput, TextArea, MaskedInput, KeybindInput, SearchBox |
| Selection | Checkbox, RadioButton, Toggle, OptionRotator, Select |
| Range | Slider, NumberStepper, ProgressBar, MeterBar, ScrollBar |
| Lists | List, TreeList, TileList, SortableList, SortableTreeList, DataTable, TreeTable |
| Containers | Panel, Window, ScrollPanel, NavDrawer, Accordion, Popover |
| Navigation | TabBar, ToggleButtonBar, ToolBar, MenuBar |
| Display | Label, RichText, Badge, Tag, TagBar, Tooltip, Toast, Image, AnimatedImage |
| Specialty | ColorPicker, GradientEditor, ImageCropper, TimePicker, CalendarSelector, StatWeb, DragHandle, InputField |

<details>
<summary>Widget gallery (click to expand)</summary>

<br>

| | |
|---|---|
| ![Accordion](https://www.devthicket.org/willow-ui/gif/accordion.gif) | ![Calendar](https://www.devthicket.org/willow-ui/gif/calendar.gif) |
| ![Checkbox](https://www.devthicket.org/willow-ui/gif/checkbox.gif) | ![Color Picker](https://www.devthicket.org/willow-ui/gif/color-picker.gif) |
| ![Data Table](https://www.devthicket.org/willow-ui/gif/data-table.gif) | ![List](https://www.devthicket.org/willow-ui/gif/list.gif) |
| ![Masked Input](https://www.devthicket.org/willow-ui/gif/masked-input.gif) | ![Menu Bar](https://www.devthicket.org/willow-ui/gif/menu-bar.gif) |
| ![Number Stepper](https://www.devthicket.org/willow-ui/gif/number-stepper.gif) | ![Option Rotator](https://www.devthicket.org/willow-ui/gif/option-rotator.gif) |
| ![Progress Bar](https://www.devthicket.org/willow-ui/gif/progress-bar.gif) | ![Radio](https://www.devthicket.org/willow-ui/gif/radio.gif) |
| ![Scroll Panel](https://www.devthicket.org/willow-ui/gif/scroll-panel.gif) | ![Select](https://www.devthicket.org/willow-ui/gif/select.gif) |
| ![Slider](https://www.devthicket.org/willow-ui/gif/slider.gif) | ![Sortable List](https://www.devthicket.org/willow-ui/gif/sortable-list.gif) |
| ![Sortable Tree List](https://www.devthicket.org/willow-ui/gif/sortable-tree-list.gif) | ![Stat Web](https://www.devthicket.org/willow-ui/gif/stat-web.gif) |
| ![Tab Bar](https://www.devthicket.org/willow-ui/gif/tab-bar.gif) | ![Tag Bar](https://www.devthicket.org/willow-ui/gif/tag-bar.gif) |
| ![Text Area](https://www.devthicket.org/willow-ui/gif/text-area.gif) | ![Text Input](https://www.devthicket.org/willow-ui/gif/text-input.gif) |
| ![Tile List](https://www.devthicket.org/willow-ui/gif/tile-list.gif) | ![Time Picker](https://www.devthicket.org/willow-ui/gif/time-picker.gif) |
| ![Toggle](https://www.devthicket.org/willow-ui/gif/toggle.gif) | ![Toggle Button Bar](https://www.devthicket.org/willow-ui/gif/toggle-button-bar.gif) |
| ![Tooltip](https://www.devthicket.org/willow-ui/gif/tooltip.gif) | ![Tree List](https://www.devthicket.org/willow-ui/gif/tree-list.gif) |
| ![Window](https://www.devthicket.org/willow-ui/gif/window.gif) | ![Button](https://www.devthicket.org/willow-ui/gif/button.gif) |

</details>

## Theming

Themes are plain JSON. Load one and every widget picks it up:

```go
theme, err := ui.LoadThemeFromFile("themes/dark.json")
ui.SetTheme(theme)
```

Ships with 7 built-in themes: dark, forest, jrpg, macos, neon, windows, and debug. Create your own -- each widget type has its own section with full control over colors, corners, padding, and more. See the [theme docs](https://www.devthicket.org/willow-ui/docs) for the full schema.

## XML Templates

Define layouts declaratively and bind reactive state:

```xml
<VBox padding="16" spacing="8">
    <Label text="Hello, {{name}}" fontSize="24" />
    <Button text="Click me" on:click="handleClick" />
</VBox>
```

Templates compile to a binary format for fast startup, and support hot reload during development.

## Examples

The `examples/` directory has 60+ runnable demos:

```bash
go run ./examples/widgets/buttons/
go run ./examples/reactive/counter/
go run ./examples/templating/xml-basic/
go run ./examples/theming/theme-gallery/
```

## Built with

- **Go** 1.24+
- **[Willow](https://github.com/devthicket/willow)** -- scene graph rendering engine
- **[Ebitengine](https://ebitengine.org)** v2.9+ -- GPU backend

## Contributing

Contributions are welcome. Please open an issue first for major changes to discuss the design. For bug fixes and small improvements, open a pull request directly.

```bash
go build ./...
go test ./...
go vet ./...
```

## License

[MIT](LICENSE)

---

<p align="center">
  <b><a href="https://www.devthicket.org/willow-ui">devthicket.org/willow-ui</a></b> -- Full documentation, guides, and interactive examples
</p>
