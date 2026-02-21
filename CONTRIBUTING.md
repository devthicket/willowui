# Contributing to WillowUI

Contributions are welcome. This guide covers how to get started, run tests, and submit changes.

## Getting Started

```bash
git clone https://github.com/devthicket/willowui.git
cd willowui
go build ./...
```

WillowUI requires **Go 1.24+** and depends on [Ebitengine](https://ebitengine.org), which needs platform graphics libraries. On Linux:

```bash
sudo apt-get install libasound2-dev libgl1-mesa-dev libxcursor-dev libxi-dev libxinerama-dev libxrandr-dev libxxf86vm-dev
```

macOS and Windows require no extra system packages.

## Running Tests

```bash
go build ./...   # compile everything
go vet ./...     # static analysis
go test ./...    # run all tests
```

Integration tests live in `internal/integration/` and test the public API via `ui "github.com/devthicket/willowui"`. New widget tests should go there.

## Running Examples

```bash
go run ./examples/widgets/buttons/
go run ./examples/reactive/counter/
go run ./examples/theming/theme-gallery/
```

## Code Style

- Follow standard Go conventions (`gofmt`, `go vet`)
- Every exported symbol in `willowui.go` must have a doc comment
- Use printable ASCII only (U+0020-U+007E) in widget text -- see the font safety section in the README
- Keep the public API in the root package; implementation goes in `internal/`
- Tests import the public API: `ui "github.com/devthicket/willowui"`

## Submitting Changes

1. **Bug fixes and small improvements** -- open a pull request directly
2. **Major changes or new features** -- open an issue first to discuss the design
3. Keep commits focused and messages concise
4. Make sure `go build ./...`, `go vet ./...`, and `go test ./...` all pass before submitting

## Project Structure

| Directory | Purpose |
|---|---|
| `willowui.go` | Public API surface (re-exports from internal packages) |
| `internal/` | All implementation packages |
| `internal/integration/` | Integration tests against the public API |
| `examples/` | Runnable demos organized by category |
| `assets/` | Embedded resources (fonts, icons) |

## License

By contributing, you agree that your contributions will be licensed under the [MIT License](LICENSE).
