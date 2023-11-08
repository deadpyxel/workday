
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

Install workday with go

```bash
go install github.com/deadpyxel/workday@latest
```

And the you can just start to use it:
```bash
workday
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

## Acknowledgements

 - Gopher's Public Discord
 - [cobra-cli](https://github.com/spf13/cobra-cli)
 - [Cobra Docs](https://github.com/spf13/cobra)
 - [Viper](https://github.com/spf13/viper)

## License

[MIT](https://choosealicense.com/licenses/mit/)
