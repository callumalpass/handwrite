# Handwriting Processor

A command-line tool to convert handwritten notes from PDF files into organized Markdown documents using Google's Gemini AI.

## Features

- **PDF to Markdown:** Converts each page of a PDF into an image and uses OCR to extract handwritten text.
- **Gemini OCR:** Uses Google's Gemini models for text recognition.
- **Customizable Templates:** Uses Jinja2 templates to format the output, allowing you to define the structure of your notes.
- **Metadata:** Includes metadata in the output file, such as the source PDF, processing date, and AI model used.
- **Obsidian Compatible:** Generates Markdown with YAML frontmatter for use in Obsidian and other note-taking apps.

## Installation

Install from the repository:

```bash
pip install git+https://github.com/callumalpass/handwrite.git
```

For local development:

```bash
git clone https://github.com/callumalpass/handwrite.git
cd handwrite
pip install -e ".[dev]"
```

## Configuration

### Environment Variables

Set your Gemini API key as an environment variable:

```bash
export GEMINI_API_KEY="YOUR_API_KEY_HERE"
```

Alternatively, create a `.env` file in your project directory:

```
GEMINI_API_KEY="YOUR_API_KEY_HERE"
```

### Configuration File

Create a configuration file to customize the tool's behavior:

```bash
handwrite config --setup
```

This creates `~/.config/handwrite/config.yaml` with the following structure:

```yaml
gemini:
  model: gemini-1.5-pro
  prompt: |
    Extract the handwritten text from this image.
    - Use $ for LaTeX, not ```latex.
    - Transcribe the text exactly as it appears.
    - The output must be only the transcribed Markdown, with no additional commentary.
template:
  path: /path/to/template.md
  variables:
    custom_var: "custom_value"
output:
  format: markdown
  encoding: utf-8
```

## Usage

Once installed, you can run the tool from any directory:

```bash
# Process a single PDF or image file
handwrite process /path/to/your/note.pdf /path/to/output/directory

# Process all PDF and image files in a directory
handwrite process /path/to/input/directory /path/to/output/directory

# Use a custom configuration file
handwrite process /path/to/your/note.pdf /path/to/output/directory --config /path/to/config.yaml
```

### Commands

#### Process Command

```bash
handwrite process <input_path> <output_dir> [--config CONFIG_FILE]
```

**Arguments:**
-   `input_path`: Path to a PDF/image file or directory containing files to process
-   `output_dir`: Directory where output Markdown files will be saved
-   `--config`: (Optional) Path to a custom configuration file

**Supported file formats:**
- PDF files (`.pdf`)
- Image files (`.png`, `.jpg`, `.jpeg`)

#### Config Command

```bash
handwrite config --setup
```

Creates the default configuration file at `~/.config/handwrite/config.yaml`.

## Customization

### Template Variables

You can customize the output template. The template uses Jinja2 syntax and has access to these variables:

-   `content`: The full transcribed text from the input file
-   `filename`: The name of the source file
-   `relative_pdf_path`: Relative path to the source file from output directory
-   `absolute_pdf_path`: Absolute path to the source file
-   `datetime_processed`: ISO timestamp when the file was processed
-   `page_count`: Number of pages/images processed
-   `model_used`: Name of the Gemini model used
-   Any custom variables defined in the config file

### Custom Templates

Create a custom template and reference it in your config:

```yaml
template:
  path: /path/to/your/custom_template.md
  variables:
    author: "Your Name"
    project: "My Project"
```

### Batch Processing

Process multiple files by pointing to a directory:

```bash
handwrite process /path/to/pdfs/ /path/to/output/
```

This will:
- Recursively find all PDF and image files
- Process each file individually
- Create corresponding Markdown files in the output directory
- Provide progress feedback and error reporting

## Development

Install development dependencies:

```bash
pip install -e ".[dev]"
```

Run tests:

```bash
python -m pytest tests/
```

Run linting:

```bash
flake8 handwriting_processor tests
black --check handwriting_processor tests
```

## Troubleshooting

### Common Issues

1. **API Key not found**: Ensure `GEMINI_API_KEY` is set in your environment or `.env` file
2. **No text extracted**: Check if the image quality is sufficient and text is clearly visible
3. **Template errors**: Verify your custom template syntax and variable names
4. **File not found**: Ensure input paths exist and are accessible

### Logging

The tool provides detailed logging at INFO level by default. You can redirect output to capture logs:

```bash
handwrite process input.pdf output/ 2>&1 | tee processing.log
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License.

## Acknowledgements

This project was inspired by the work of Tejas Raskar on [noted.md](https://github.com/tejas-raskar/noted.md).