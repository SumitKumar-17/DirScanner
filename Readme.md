# DirScanner

![image](https://github.com/user-attachments/assets/dacfb64e-fb73-4ad2-8a08-f595f9ab18fb)

DirScanner is a CLI tool written in Go that scans a directory and generates a Markdown file containing its tree structure. It supports custom connector styles, pattern-based exclusions, depth limiting, file sizes, and summary statistics.

## Features

- **Markdown tree output** — generates a readable directory tree in Markdown
- **Stdout mode** — print to terminal instead of a file
- **File sizes** — optionally display each file's size next to its name
- **Summary statistics** — append a count of files, directories, and total size
- **Hidden file support** — dotfiles are skipped by default; opt in with a flag
- **Pattern exclusions** — exclude entries via glob patterns or a `.dirignore` file
- **Depth limiting** — restrict how deep the scan traverses
- **Custom connector styles** — change the tree branch symbols

## Installation

```sh
go install github.com/SumitKumar-17/DirScanner@latest
```

## Usage

### Basic usage

Scan a directory and write the tree to a Markdown file:

```sh
dirscanner <directory> <output.md>
```

Print to stdout (omit the output file, or pass `-`):

```sh
dirscanner <directory>
dirscanner <directory> -
```

### Show file sizes

```sh
dirscanner ./myproject structure.md --sizes
```

```
# Directory Structure for ./myproject

├── main.go (3.2 KB)
├── scanner.go (2.8 KB)
└── go.mod (180 B)
```

### Append summary statistics

```sh
dirscanner ./myproject structure.md --stats
```

Adds a summary line at the bottom of the output:

```
**Summary:** 12 files, 4 directories, 48.3 KB total size
```

### Include hidden files (dotfiles)

By default dotfiles and hidden directories are skipped. To include them:

```sh
dirscanner ./myproject structure.md --show-hidden
```

### Exclude files or directories

Using `--exclude` flags (glob patterns):

```sh
dirscanner ./myproject structure.md --exclude "*.txt" --exclude "node_modules"
```

Using a `.dirignore` file placed in the scanned directory:

```
# .dirignore
node_modules
*.log
dist
```

Lines starting with `#` are treated as comments.

### Limit traversal depth

```sh
dirscanner ./myproject structure.md --depth 2
```

### Customize connector symbols

```sh
dirscanner ./myproject structure.md --intermediate "+-- " --last "\`-- " --prefix "    " --branch "|   "
```

### Enable debug logging

```sh
dirscanner ./myproject structure.md --verbose
# or
dirscanner ./myproject structure.md -v
```

## All flags

| Flag | Default | Description |
|------|---------|-------------|
| `--sizes` | `false` | Show file sizes next to filenames |
| `--stats` | `false` | Append summary (files, dirs, total size) to output |
| `--show-hidden` | `false` | Include hidden files and directories (dotfiles) |
| `--exclude` | `[]` | Glob patterns to exclude (repeatable) |
| `--depth` | `-1` | Max traversal depth (`-1` = unlimited) |
| `--intermediate` | `├── ` | Symbol for intermediate tree nodes |
| `--last` | `└── ` | Symbol for the last node in a directory |
| `--prefix` | `    ` | Whitespace prefix for child alignment |
| `--branch` | `│   ` | Branch connector for intermediate nodes |
| `--verbose`, `-v` | `false` | Enable debug logging |

## Contributing

Contributions are welcome! Please open an issue to discuss changes or submit a pull request.

## License

This project is licensed under the MIT License — see the LICENSE file for details.
