# Architecture

This document describes the internal architecture of gosh.

## Overview

gosh is a POSIX-like interactive shell. It reads user input, parses it into commands, and executes them either as builtin functions or external processes.

## Execution Flow

```
User Input
    │
    ▼
parseArgs()          Parse raw input into tokens (handles quotes, escapes)
    │
    ▼
splitPipeline()      Split tokens on "|" into command segments
    │
    ├── Single segment ──► extractRedirect() ──► dispatch()
    │
    └── Multiple segments ──► extractRedirect() on last segment
                              ──► runPipeline()
```

## Package Structure

### `cmd/gosh/`

Minimal entry point. Creates a `Shell` instance and calls `Run()`.

### `internal/shell/`

All shell logic lives here, organized by concern:

| File | Responsibility |
|------|---------------|
| `shell.go` | `Shell` struct, REPL loop, history persistence |
| `parse.go` | Tokenizing input: quote handling, escape sequences, pipeline splitting |
| `builtins.go` | Builtin command implementations (`echo`, `cd`, `pwd`, `type`, `history`) |
| `exec.go` | Command dispatch, external process execution, pipeline orchestration |
| `redirect.go` | Parsing redirection operators and opening output files |
| `path.go` | Searching `PATH` for executables |
| `complete.go` | Tab completion for command names |

## Key Design Decisions

### State Management

The `Shell` struct holds all mutable state (command history, history offset). Builtins that need shell state (like `history`) are methods on `Shell`. Stateless builtins (like `echo`, `pwd`) are plain functions.

### Pipeline Execution

Pipelines use OS-level pipes (`os.Pipe()`). External commands are started with `cmd.Start()` for true concurrency. Builtin commands in pipelines run in goroutines so they can write to pipes without blocking the main loop.

### I/O Redirection

Redirection is extracted from parsed tokens before command dispatch. The `redirect` struct carries file paths and append flags. `resolveStreams()` opens files and returns `io.Writer` interfaces, keeping command implementations stream-agnostic.

### Tab Completion

The completer implements the `readline.AutoCompleter` interface. It matches against both builtin names and executables found in PATH. Double-tab shows all matches when there's no unique completion.

## Dependencies

- [`github.com/chzyer/readline`](https://github.com/chzyer/readline) - Readline library for interactive input, line editing, and history navigation.
