# starlark-go-copybara

[![CI](https://github.com/albertocavalcante/starlark-go-copybara/actions/workflows/ci.yml/badge.svg)](https://github.com/albertocavalcante/starlark-go-copybara/actions/workflows/ci.yml)

A clean-room Go implementation of Copybara's Starlark dialect.

## Features

- Parse and evaluate `copy.bara.sky` configuration files
- Execute code transformation workflows
- WASM compilation for browser-based tools
- Dry-run simulation for testing workflows

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
    authoring = authoring.pass_thru("Default Author <author@example.com>"),
    transformations = [
        core.move("src", "lib"),
        core.replace(
            before = "old_name",
            after = "new_name",
        ),
        metadata.squash_notes(),
    ],
)
`

    result, err := interp.Eval("copy.bara.sky", config)
    if err != nil {
        panic(err)
    }

    for _, wf := range result.Workflows() {
        fmt.Printf("Workflow: %s\n", wf.Name())
    }
}
```

## Modules

| Module | Description |
|--------|-------------|
| `core` | Workflows and transformations (move, copy, replace, remove, verify_match) |
| `git` | Git origins and destinations (including GitHub) |
| `metadata` | Commit message transformations |
| `authoring` | Author handling modes |
| `folder` | Local folder origins/destinations for testing |

## Status

**Work in progress** - Core modules are implemented. See [CONTRIBUTING.md](CONTRIBUTING.md) for development status.

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for build instructions, development workflow, and guidelines.

## Reference

This is a clean-room implementation based on:
- [google/copybara](https://github.com/google/copybara) - The original Copybara tool
- [go.starlark.net](https://github.com/google/starlark-go) - Starlark interpreter in Go

## License

MIT License - see [LICENSE](LICENSE)

See [NOTICE](NOTICE) for attribution details.
