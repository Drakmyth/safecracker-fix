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

## Testing

### Unit tests

Run the unit tests with:

```powershell
go test ./...
```

Unit tests do not require a copy of Safecracker.

### Integration tests

Integration tests verify the patcher against actual game
executables. They require local fixtures that are not included
in this repository.

For the v1.01 Steam target, copy an original, unmodified
Safecracker.exe to:

```
internal/safecracker/testdata/v101_steam/Safecracker.exe
```

The fixture must match the supported executable's SHA-256 hash.
Do not use an executable that has already been patched.

Run the integration tests with:

```powershell
go test -tags=integration -count=1 ./...
```

This command runs both unit and integration tests. Integration
tests fail if a required fixture is missing or invalid.

The tests verify the original executable, apply the patch in
memory, compare the result against the known-good patched
SHA-256 hash, and confirm that reversing the patch restores
the original executable exactly.

Integration tests never modify fixture files.

### Continuous integration

CI runs the unit tests and checks that integration-tagged
tests compile and pass static analysis. Actual integration
tests must be run locally with the required game fixtures.



## Disclaimer

This is an unofficial community project. It is not affiliated with the game's developers or publishers. Game files are not distributed with this project.

## License

This project is licensed under the [MIT License](LICENSE).
