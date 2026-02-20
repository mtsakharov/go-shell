# Contributing to gosh

Thanks for your interest in contributing! This document outlines guidelines for contributing to the project.

## Getting Started

1. Fork the repository
2. Clone your fork:
   ```sh
   git clone https://github.com/<your-username>/go-shell.git
   cd go-shell
   ```
3. Create a feature branch:
   ```sh
   git checkout -b feature/your-feature
   ```
4. Make your changes and verify they build:
   ```sh
   make build
   make lint
   ```

## Development Workflow

### Building

```sh
make build    # Compile the binary
make run      # Build and run interactively
```

### Testing

```sh
make test     # Run all tests
make lint     # Run go vet
```

### Code Style

- Follow standard Go conventions and [Effective Go](https://go.dev/doc/effective-go)
- Run `gofmt` or `goimports` before committing
- Keep functions focused and small
- Add comments for exported symbols

### Project Layout

- `cmd/gosh/` - Entry point only. Keep it minimal.
- `internal/shell/` - All shell logic. This is intentionally `internal` to prevent external imports.

## Submitting Changes

1. Ensure your code compiles cleanly: `make build`
2. Run linting: `make lint`
3. Commit with a clear message describing the change
4. Push your branch and open a pull request against `main`

### Commit Messages

Use clear, imperative-mood commit messages:

```
Add glob expansion support
Fix pipe cleanup on early exit
Refactor history into separate file
```

### Pull Request Guidelines

- Keep PRs focused on a single concern
- Describe what the change does and why
- Reference any related issues

## Ideas for Contributions

- Add new builtins (e.g., `export`, `unset`, `alias`)
- Implement glob expansion (`*`, `?`)
- Add environment variable expansion (`$VAR`, `${VAR}`)
- Signal handling (`SIGINT`, `SIGTSTP`)
- Job control (background processes with `&`)
- Add unit tests for parsing, redirection, and builtins
- Improve error messages

## Reporting Issues

When reporting bugs, please include:

- Steps to reproduce the issue
- Expected vs actual behavior
- Go version (`go version`)
- OS and architecture
