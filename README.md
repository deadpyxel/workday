[![Go Reference](https://pkg.go.dev/badge/github.com/deadpyxel/workday.svg)](https://pkg.go.dev/github.com/deadpyxel/workday)
[![Go Report Card](https://goreportcard.com/badge/github.com/deadpyxel/workday)](https://goreportcard.com/report/github.com/deadpyxel/workday)
[![GitHub release](https://img.shields.io/github/release/deadpyxel/workday.svg)](https://github.com/deadpyxel/workday/releases)
[![GitHub Workflow Status](https://img.shields.io/github/actions/workflow/status/deadpyxel/workday/go-ci.yml?branch=main)](https://github.com/deadpyxel/workday/actions)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

# Workday

A simple CLI written in go to help with my day to day activity tracking at work.
## Features

> Disclaimer: The goals of this tool are aligned to my workflow and processes

- Simple command structure
- Plain text storage (a simple JSON)
- Fully CLI Based
- Very small footprint (In memory, CPU and codebase)
- Cross platform
- Configurable using config files

## Installation

### Using Go Install

Install workday with go

```bash
go install github.com/deadpyxel/workday@latest
```

### Using Pre-built Binaries

Download the latest release for your platform from the [GitHub releases page](https://github.com/deadpyxel/workday/releases).

Available platforms:
- Linux (amd64, 386, arm64, arm)
- macOS (amd64, arm64)
- Windows (amd64, 386)

### Usage

After installation, you can start using workday:
```bash
workday
```

Check the version:
```bash
workday version
```

## Configuration

Workday allows you to configure some options using a YAML configuration file. By default, it will search for the file under your `$HOME/.config/workday/config.yaml`, but you can pass the configuration file path with the `--config` flag. An example of a valid config file can be seen below.

```yaml
journalPath: "/path/to/your/journal.json"
```

## Running Tests

To run tests, run the following command

```bash
go test -cover -v ./...
```
If you want to run the benchmarks:

```bash
go test -bench=. -v ./...
```

## Run Locally

Clone the project

```bash
git clone https://github.com/deadpyxel/workday.git
```

Go to the project directory

```bash
cd workday
```

Build the project locally

```bash
go build -o bin/
```

Run the app

```bash
./bin/workday
```

## Releases

Releases are automated using [GoReleaser](https://goreleaser.com/) and GitHub Actions. When a new tag is pushed (format: `v*`), the release workflow will:

1. Build binaries for all supported platforms
2. Create checksums for all artifacts
3. Generate release notes from commits
4. Publish the release on GitHub

To create a new release:
```bash
git tag v1.0.0
git push origin v1.0.0
```

## Acknowledgements

 - Gopher's Public Discord
 - [cobra-cli](https://github.com/spf13/cobra-cli)
 - [Cobra Docs](https://github.com/spf13/cobra)
 - [Viper](https://github.com/spf13/viper)
 - [GoReleaser](https://goreleaser.com/)

## License

[MIT](https://choosealicense.com/licenses/mit/)
