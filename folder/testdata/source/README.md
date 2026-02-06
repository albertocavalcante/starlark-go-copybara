# Example Source

This directory contains example source files for testing the folder module.

## Files

- `main.go` - Main application entry point
- `util.go` - Utility functions
- `config/app.yaml` - Application configuration

## Usage

Use this directory as the origin for testing folder workflows:

```starlark
core.workflow(
    name = "test",
    origin = folder.origin(path = "testdata/source"),
    destination = folder.destination(path = "/tmp/output"),
    transformations = [],
)
```
