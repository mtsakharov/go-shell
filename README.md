# gosh - Go Shell

A lightweight, interactive POSIX-like shell written in Go. Built as part of the [CodeCrafters "Build Your Own Shell"](https://app.codecrafters.io/courses/shell/overview) challenge, then refactored into a standalone project.

## Features

- **Interactive REPL** with readline support (line editing, history navigation)
- **Builtin commands**: `echo`, `exit`, `type`, `pwd`, `cd`, `history`
- **External command execution** via PATH lookup
- **Pipelines**: chain commands with `|`
- **I/O redirection**: `>`, `>>`, `2>`, `2>>` (stdout and stderr)
- **Quote handling**: single quotes, double quotes with escape sequences
- **Tab completion** for builtins and executables
- **Persistent command history** via `HISTFILE` environment variable

## Requirements

- Go 1.25+

## Installation

```sh
# Clone the repository
git clone https://github.com/mtsakharov/go-shell.git
cd go-shell

# Build
make build

# Or install directly
go install ./cmd/gosh
```

## Usage

```sh
# Run the shell
./gosh

# Or with history persistence
HISTFILE=~/.gosh_history ./gosh
```

### Builtin Commands

| Command | Description |
|---------|-------------|
| `echo [args...]` | Print arguments to stdout |
| `exit` | Exit the shell |
| `type <command>` | Show whether a command is a builtin or its path |
| `pwd` | Print the current working directory |
| `cd [dir]` | Change directory (defaults to `$HOME`) |
| `history [n]` | Show command history (last `n` entries) |
| `history -r <file>` | Read history from file |
| `history -w <file>` | Write history to file |
| `history -a <file>` | Append new history entries to file |

### Pipelines

```sh
$ echo hello | cat
hello
$ ls -la | grep go | head -5
```

### Redirection

```sh
$ echo hello > output.txt       # truncate stdout to file
$ echo world >> output.txt      # append stdout to file
$ cmd 2> errors.log             # redirect stderr to file
$ cmd 2>> errors.log            # append stderr to file
```

## Project Structure

```
go-shell/
├── cmd/
│   └── gosh/
│       └── main.go             # Entry point
├── internal/
│   └── shell/
│       ├── shell.go            # Shell struct, REPL loop, history
│       ├── builtins.go         # Builtin command implementations
│       ├── exec.go             # Command dispatch and pipeline execution
│       ├── parse.go            # Argument parsing and pipeline splitting
│       ├── path.go             # PATH lookup utilities
│       ├── redirect.go         # I/O redirection handling
│       └── complete.go         # Tab completion
├── docs/
│   └── architecture.md         # Architecture documentation
├── Makefile
├── go.mod
├── go.sum
├── README.md
└── CONTRIBUTING.md
```

## Development

```sh
make build   # Build the binary
make run     # Build and run
make test    # Run tests
make lint    # Run go vet
make clean   # Remove build artifacts
```

## License

MIT
