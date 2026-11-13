# Safecracker Fix Utility

[![CI](https://github.com/Drakmyth/safecracker-fix/actions/workflows/ci.yml/badge.svg)](https://github.com/Drakmyth/safecracker-fix/actions/workflows/ci.yml)

A Windows compatibility utility for *Safecracker: The Ultimate Puzzle Adventure* (2006).

## Overview

Safecracker Fix Utility aims to make the game easier to run on modern Windows systems by addressing performance and graphics compatibility issues.

## Planned features

- Automatic detection of supported game installations
- Reversible performance patch for floating-point timer precision issues
- dgVoodoo2 installation and configuration
- Native Windows graphical interface
- Backup and restoration of modified game files

## Project status

This project is in early development and is not ready for general use.

## Development

The application is written in Go.

### Build

```powershell
go build -o bin/safecracker-fix.exe ./cmd/safecracker-fix
```

### Run

```powershell
go run ./cmd/safecracker-fix
```

### Test

```powershell
go test ./...
```

## Disclaimer

This is an unofficial community project. It is not affiliated with the game's developers or publishers. Game files are not distributed with this project.

## License

This project is licensed under the [MIT License](LICENSE).
