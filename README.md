# Handwrite - Go Edition

A high-performance command-line tool written in Go to convert handwritten notes from PDF files into organized Markdown documents using Google's Gemini AI.

## Features

- **PDF to Markdown:** Converts each page of a PDF into an image and uses OCR to extract handwritten text
- **Gemini OCR:** Uses Google's Gemini models for text recognition with configurable prompts
- **Concurrent Processing:** Leverages Go's goroutines for fast parallel processing
- **Customizable Templates:** Uses Go's text/template package for flexible output formatting
- **Progress Tracking:** Real-time progress bars for batch operations
- **Batch Processing:** Process entire directories of files recursively
- **Obsidian Compatible:** Generates Markdown with YAML frontmatter for note-taking apps

## Installation

### Prerequisites

- Go 1.21 or higher
- Google Gemini API key

### From Source

```bash
git clone https://github.com/callumalpass/handwrite.git
cd handwrite
go build -o handwrite .
```

### Install Globally

```bash
go install github.com/callumalpass/handwrite@latest
```

## Configuration

### Environment Variables

Set your Gemini API key as an environment variable:

```bash
export GEMINI_API_KEY="YOUR_API_KEY_HERE"
```

Alternatively, create a `.env` file in your working directory:

```
GEMINI_API_KEY="YOUR_API_KEY_HERE"
```

### Configuration File

Create a configuration file:

```bash
handwrite config setup
```

This creates `~/.config/handwrite/config.yaml` with the following structure:

```yaml
gemini:
  model: "gemini-1.5-pro"
  prompt: |
    Extract the handwritten text from this image.
    - Use $ for LaTeX, not ```latex.
    - Transcribe the text exactly as it appears.
    - The output must be only the transcribed Markdown, with no additional commentary.

template:
  path: "templates/note_template.md"
  variables: {}

output:
  format: "markdown"
  encoding: "utf-8"
```

## Usage

### Basic Commands

```bash
# Process a single PDF or image file
handwrite process /path/to/your/note.pdf /path/to/output/directory

# Process all PDF and image files in a directory
handwrite process /path/to/input/directory /path/to/output/directory

# Use a custom configuration file
handwrite process /path/to/your/note.pdf /path/to/output/directory --config /path/to/config.yaml

# Control concurrency (default: 4 workers)
handwrite process /path/to/directory /path/to/output --workers 8
```

### Supported File Formats

- PDF files (`.pdf`) - Processed directly by Gemini
- Image files (`.png`, `.jpg`, `.jpeg`)

### Template System

The default template includes:

- **Content:** The full transcribed text from the input file
- **Filename:** The name of the source file
- **RelativePDFPath:** Relative path to the source file from output directory
- **AbsolutePDFPath:** Absolute path to the source file
- **DatetimeProcessed:** ISO timestamp when the file was processed
- **PageCount:** Number of pages/images processed
- **ModelUsed:** Name of the Gemini model used
- **CustomVariables:** Any custom variables defined in the config file

### Custom Templates

Create a custom template using Go's text/template syntax:

```markdown
---
title: {{.Filename}}
created: {{.DatetimeProcessed}}
pages: {{.PageCount}}
model: {{.ModelUsed}}
source: "[[{{.RelativePDFPath}}]]"
---

# {{.Filename}}

{{.Content}}

---
*Processed on {{.DatetimeProcessed}} using {{.ModelUsed}}*
```

## Development

### Build from Source

```bash
git clone https://github.com/callumalpass/handwrite.git
cd handwrite
go mod download
go build -o handwrite .
```

### Run Tests

```bash
go test ./tests/...
```

### Project Structure

```
handwrite/
├── main.go                 # Entry point
├── cmd/                    # CLI commands
│   ├── root.go            # Root command setup
│   ├── process.go         # Process command implementation
│   └── config.go          # Config command implementation
├── internal/              # Internal packages
│   ├── config/           # Configuration management
│   ├── processor/        # PDF/image processing
│   ├── gemini/          # Gemini API client
│   └── template/        # Template rendering
├── templates/            # Default templates
├── tests/               # Test files
└── go.mod              # Go module definition
```


## License

This project is licensed under the MIT License.

## Acknowledgements

This Go rewrite improves upon the original Python version, providing better performance and easier distribution while maintaining full feature compatibility.
