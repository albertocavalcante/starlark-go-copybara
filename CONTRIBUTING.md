# Contributing to starlark-go-copybara

Thank you for your interest in contributing! This document covers how to build, develop, and submit changes.

## Prerequisites

- **Go 1.24+** (uses range-over-int, slices package)
- **just** - Task runner ([installation](https://just.systems/))
- **golangci-lint v2** - Linter ([installation](https://golangci-lint.run/))
- **lefthook** - Git hooks manager (optional)

## Quick Start

```bash
# Clone the repository
git clone https://github.com/albertocavalcante/starlark-go-copybara.git
cd starlark-go-copybara

# Install development tools
go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest
go install github.com/evilmartians/lefthook@latest

# Set up git hooks (optional but recommended)
lefthook install

# Run tests
just test
```

## Development Commands

Run `just --list` to see all available commands. Common ones:

| Command | Description |
|---------|-------------|
| `just test` | Run all tests |
| `just test-v` | Run tests with verbose output |
| `just test-cover` | Run tests with coverage report |
| `just lint` | Run golangci-lint |
| `just fmt` | Format code with gofmt |
| `just dev` | Format, lint, and test (development workflow) |
| `just ci` | Run full CI checks locally |
| `just build-wasm` | Build WASM target |

## Project Structure

```
starlark-go-copybara/
├── copybara/      # Main interpreter and public API
├── core/          # core.* module (workflow, move, copy, replace, etc.)
├── git/           # git.* module (origin, destination, github)
├── metadata/      # metadata.* module (commit message transforms)
├── authoring/     # authoring.* module (author handling)
├── folder/        # folder.* module (local testing)
├── types/         # Core types (Path, Change, OriginRef, etc.)
├── transform/     # Transformation interface and context
├── eval/          # Starlark evaluator
├── analysis/      # Configuration introspection
├── vcs/           # VCS abstraction interfaces
└── wasm/          # WASM entry point
```

## Building

```bash
# Build all packages
just build

# Build WASM target
just build-wasm
# Output: main.wasm
```

## Testing

```bash
# Run all tests
just test

# Run with race detector
just test-race

# Generate coverage report
just test-cover
# Opens coverage.html

# Run benchmarks
just bench
```

## Code Style

This project follows standard Go conventions with additional guidelines:

- **Go 1.24+ idioms**: Use `range n` for integers, `slices` package, `any` instead of `interface{}`
- **No unnecessary abstractions**: Keep code simple and direct
- **Comprehensive tests**: Aim for high coverage, especially for Starlark built-ins
- **Clear error messages**: Include context in error returns

The linter configuration is in `.golangci.toml`. Run `just lint` before committing.

## Adding a New Built-in Function

1. Add the function to the appropriate module (e.g., `core/module.go`)
2. Implement the function with proper Starlark value handling
3. Add tests in `*_test.go`
4. Update the module's `starlarkstruct.Module` members

Example pattern:

```go
func myBuiltin(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
    var name string
    if err := starlark.UnpackArgs(b.Name(), args, kwargs, "name", &name); err != nil {
        return nil, err
    }
    return starlark.String(name), nil
}
```

## Adding a New Type

Types that are exposed to Starlark must implement `starlark.Value`:

```go
type MyType struct {
    // fields
}

func (m *MyType) String() string        { return "MyType(...)" }
func (m *MyType) Type() string          { return "MyType" }
func (m *MyType) Freeze()               { /* freeze mutable fields */ }
func (m *MyType) Truth() starlark.Bool  { return starlark.True }
func (m *MyType) Hash() (uint32, error) { return 0, fmt.Errorf("unhashable: MyType") }

// Compile-time interface check
var _ starlark.Value = (*MyType)(nil)
```

## Submitting Changes

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/my-feature`)
3. Make your changes
4. Run `just dev` to format, lint, and test
5. Commit with a clear message
6. Push and open a pull request

## Clean-Room Implementation

This is a **clean-room implementation**. Do not copy code from google/copybara. Instead:

- Read the Copybara documentation to understand behavior
- Implement the functionality from scratch
- Test against expected behavior, not implementation details

## License

By contributing, you agree that your contributions will be licensed under the MIT License.
