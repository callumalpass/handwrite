package template

import (
	"fmt"
	"os"
	"path/filepath"
	"text/template"
	"time"
)

type Data struct {
	Content            string
	Filename           string
	AbsolutePDFPath    string
	RelativePDFPath    string
	SourcePathAbsolute string
	SourcePathRelative string
	DatetimeProcessed  string
	PageCount          int
	ModelUsed          string
	CustomVariables    map[string]interface{}
}

func RenderTemplate(templatePath, outputPath string, data Data) error {
	// Check if template exists
	if _, err := os.Stat(templatePath); os.IsNotExist(err) {
		return fmt.Errorf("template file not found: %s", templatePath)
	}

	// Read template file
	tmplContent, err := os.ReadFile(templatePath)
	if err != nil {
		return fmt.Errorf("failed to read template file: %w", err)
	}

	// Parse template
	tmpl, err := template.New(filepath.Base(templatePath)).Parse(string(tmplContent))
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	// Create output directory if it doesn't exist
	outputDir := filepath.Dir(outputPath)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Create output file
	outputFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer outputFile.Close()

	// Execute template
	if err := tmpl.Execute(outputFile, data); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	return nil
}

func CreateTemplateData(content, filename, inputPath, outputDir string, pageCount int, modelUsed string, customVars map[string]interface{}) Data {
	absoluteInputPath, _ := filepath.Abs(inputPath)
	relativeInputPath, _ := filepath.Rel(outputDir, absoluteInputPath)

	return Data{
		Content:            content,
		Filename:           filename,
		AbsolutePDFPath:    absoluteInputPath,
		RelativePDFPath:    relativeInputPath,
		SourcePathAbsolute: absoluteInputPath,
		SourcePathRelative: relativeInputPath,
		DatetimeProcessed:  time.Now().Format(time.RFC3339),
		PageCount:          pageCount,
		ModelUsed:          modelUsed,
		CustomVariables:    customVars,
	}
}
