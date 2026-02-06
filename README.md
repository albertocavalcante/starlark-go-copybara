# starlark-go-copybara

[![CI](https://github.com/albertocavalcante/starlark-go-copybara/actions/workflows/ci.yml/badge.svg)](https://github.com/albertocavalcante/starlark-go-copybara/actions/workflows/ci.yml)

A clean-room Go implementation of Copybara's Starlark dialect.

## Overview

This project implements Copybara's Starlark dialect in Go, enabling:

- Parsing and evaluation of `copy.bara.sky` files
- Execution of code transformation workflows
- WASM compilation for browser-based tools
- Dry-run simulation for testing workflows

## Status

**Work in progress** - This project is in early development.

## Modules

| Module | Description |
|--------|-------------|
| `core` | Core transformations (move, replace, workflow) |
| `git` | Git origins and destinations |
| `metadata` | Commit message transformations |
| `authoring` | Author handling |
| `folder` | Local folder origins/destinations for testing |

## Installation

```bash
go get github.com/albertocavalcante/starlark-go-copybara
```

## Usage

```go
package main

import (
    "fmt"
    "github.com/albertocavalcante/starlark-go-copybara/copybara"
)

func main() {
    interp := copybara.New()

    config := `
core.workflow(
    name = "default",
    origin = folder.origin(),
    destination = folder.destination(),
    transformations = [
        core.move("src", "lib"),
    ],
)
`

    result, err := interp.Eval("copy.bara.sky", config)
    if err != nil {
        panic(err)
    }

    fmt.Printf("Workflows: %v\n", result.Workflows())
}
```

## Building

```bash
# Build library
go build ./...

# Build WASM
GOOS=js GOARCH=wasm go build -o main.wasm ./wasm/

# Run tests
just test
```

## Development

This project uses:
- [just](https://just.systems/) for task running
- [golangci-lint](https://golangci-lint.run/) v2 for linting
- [lefthook](https://github.com/evilmartians/lefthook) for git hooks
- [gotestsum](https://github.com/gotestyourself/gotestsum) for test output

```bash
# Install tools
go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest
go install github.com/evilmartians/lefthook@latest

# Set up hooks
lefthook install

# Run development workflow
just dev
```

## Reference

This is a clean-room implementation based on:
- [google/copybara](https://github.com/google/copybara) - The original Copybara tool
- [go.starlark.net](https://github.com/google/starlark-go) - Starlark interpreter in Go

## License

MIT License - see [LICENSE](LICENSE)

## Attribution

See [NOTICE](NOTICE) for attribution details.
