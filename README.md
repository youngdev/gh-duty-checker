# gh-duty-checker

`gh-duty-checker` is a command-line tool for checking vehicle duty/tax information from Ghana's official customs portal. It allows you to quickly find assessed tax values for used vehicles based on make, model, and year.

## Installation

There are two ways to install `gh-duty-checker`.

### Option 1: Using `go install` (Recommended)

To install `gh-duty-checker`, you need to have Go installed on your system. Then, you can use `go install` to install it directly from the GitHub repository:

```sh
go install github.com/youngdev/gh-duty-checker@latest
```

This will download, compile, and install the binary to your Go bin directory (`$GOPATH/bin`).

### Option 2: Manual Installation from Binaries

You can also download pre-compiled binaries from the `bin/` directory in this repository or from the [Releases](https://github.com/youngdev/gh-duty-checker/releases) page. Download the appropriate binary for your operating system (Darwin for macOS, Windows), make it executable, and move it to a directory in your `PATH`.

For macOS:
```sh
# For AMD64
chmod +x gh-duty-checker-darwin-amd64-v1.0.0
# For ARM64
chmod +x gh-duty-checker-darwin-arm64-v1.0.0
```

For Windows, you can run the `.exe` file directly.

## Usage

The primary use of this tool is to search for vehicles and view their tax assessment details.

### 1. Basic Search: Find a vehicle by make and model

This is the most common use case. It will list all vehicles matching the specified make and model.

```sh
gh-duty-checker -make "TOYOTA" -model "CAMRY"
```

### 2. Search with a Specific Year

You can narrow down your search by specifying the year of manufacture.

```sh
gh-duty-checker -make "TOYOTA" -model "CAMRY" -year "2019"
```

### 3. Search within a Date Range

If you want to see vehicles assessed within a specific period, you can use the `-start-date` and `-end-date` flags. The date format is `YYYY-MM-DD`.

```sh
gh-duty-checker -make "HONDA" -model "ACCORD" -start-date "2023-01-01" -end-date "2023-03-31"
```

### 4. Get Detailed Tax Breakdown for a Specific Vehicle

After running a search, the results will include a `No.` for each vehicle. Use this number with the `-detail` flag to get a complete breakdown of the taxes and duties.

First, run a search:
```sh
gh-duty-checker -make "KIA" -model "SPORTAGE" -year 2020
```

From the output, pick a `No.` (e.g., `1`) and run:
```sh
gh-duty-checker -make "KIA" -model "SPORTAGE" -year 2020 -detail 1
```

### 5. Debugging Network Requests

If you are facing issues or want to inspect the raw requests and responses, you can use the `-debug` flag.

```sh
gh-duty-checker -make "TOYOTA" -model "RAV4" -debug
``` 