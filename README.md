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

## Prerequisites for AI Feature

To use the AI feature, you need a Gemini API key:

1. Get your API key from [Google AI Studio](https://aistudio.google.com/app/apikey)
2. Set it as an environment variable:
   ```sh
   export GEMINI_API_KEY="your-api-key-here"
   ```

## Usage

The primary use of this tool is to search for vehicles and view their tax assessment details. You can use it with traditional flags or with the new AI-powered natural language interface.

### 1. Basic Search for a Common Car

This searches for an older, high-volume car. The `-list` flag displays a summary table of matching vehicles, and `-tax-list` shows a detailed tax breakdown for the first result found. The `-assessment=3y` flag widens the search to the last 3 years to increase the chances of a match.

```sh
gh-duty-checker -make="Toyota" -model="Camry" -year=2019 -list -tax-list -assessment=3y
```

### 2. Search for a Popular SUV (Case-Insensitive)

This example demonstrates that the tool correctly handles lowercase make names. It searches for a recent popular SUV.

```sh
gh-duty-checker -make="toyota" -model="RAV4" -year=2022 -list -assessment=2y
```

### 3. Get Only the Tax Breakdown for a Luxury Car

If you only want the detailed tax breakdown without seeing the summary list, you can omit the `-list` flag.

```sh
gh-duty-checker -make="Mercedes-Benz" -model="E-Class" -year=2021 -tax-list -assessment=2y
```

### 4. Search for a Vehicle with a Multi-Word Name

This example shows that spaces in both the make ("Land Rover") and model ("Range Rover") are handled correctly.

```sh
gh-duty-checker -make="Land Rover" -model="Range Rover" -year=2022 -list -tax-list -assessment=2y
```

### 5. Search within a Narrow Date Range

This example uses a shorter assessment window (`90d`). The tool will either show the results or correctly report that no data was found, which is expected behavior for narrow date ranges.

```sh
gh-duty-checker -make="Honda" -model="Civic" -year=2023 -list -assessment=90d
```

### 6. Using AI to Interpret Natural Language Queries

The new AI feature allows you to ask questions in natural language. The AI will automatically:
- Extract the vehicle make and model from your query
- Search for duty information for the past 10 years
- Provide a comprehensive analysis of the results

```sh
# Example 1: Ask about luxury vehicles
gh-duty-checker -ai "what are the duty rates people paying for Lamborghini Urus"

# Example 2: Simple query
gh-duty-checker -ai "How much duty for BMW X5?"

# Example 3: Conversational query
gh-duty-checker -ai "I'm thinking of importing a Mercedes G-Wagon, what are the typical duties?"
```

The AI will:
1. Identify the make and model from your question
2. Search external.unipassghana.com for the past 10 years (from current year)
3. Analyze all the results and provide insights on:
   - Duty rate trends over the years
   - Notable patterns or changes
   - Exchange rate impacts
   - Total tax summaries

### All Available Flags

```
Usage of gh-duty-checker:
  -ai string
    	Use AI to interpret natural language query (e.g., 'what are the duty rates people paying for Lamborghini Urus')
  -assessment string
    	Assessment date range (e.g., 4d, 2w, 3m, 1y). '1d' for today only. (default "3m")
  -debug
    	Enable debug logging to print request details.
  -list
    	Display the list of matching vehicles and their total tax.
  -make string
    	Make of the car (e.g., 'Tesla')
  -model string
    	Model of the car (e.g., 'Model X')
  -tax-list
    	Display the detailed tax breakdown for the most recent vehicle found.
  -year string
    	Year of manufacture (default "2024")
```

### Note on AI Usage

When using the `-ai` flag:
- The AI uses Google's Gemini 2.5 Pro model
- It automatically searches for the past 10 years of data
- Each year search covers a 5-year assessment period
- Results are analyzed and presented in a conversational format
- Make sure your GEMINI_API_KEY environment variable is set

### TODO
- Figure out a way to fetch all for a day without make & model selection?