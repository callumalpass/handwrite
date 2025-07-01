package tests

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/callumalpass/handwrite/internal/template"
)

func TestCreateTemplateData(t *testing.T) {
	content := "Test content"
	filename := "test.pdf"
	inputPath := "/path/to/test.pdf"
	outputDir := "/output"
	pageCount := 3
	modelUsed := "gemini-1.5-pro"
	customVars := map[string]interface{}{
		"author": "Test Author",
	}

	data := template.CreateTemplateData(content, filename, inputPath, outputDir, pageCount, modelUsed, customVars)

	if data.Content != content {
		t.Errorf("Expected content '%s', got '%s'", content, data.Content)
	}

	if data.Filename != filename {
		t.Errorf("Expected filename '%s', got '%s'", filename, data.Filename)
	}

	if data.PageCount != pageCount {
		t.Errorf("Expected page count %d, got %d", pageCount, data.PageCount)
	}

	if data.ModelUsed != modelUsed {
		t.Errorf("Expected model '%s', got '%s'", modelUsed, data.ModelUsed)
	}

	if data.CustomVariables["author"] != "Test Author" {
		t.Error("Custom variables not properly set")
	}

	if data.DatetimeProcessed == "" {
		t.Error("DatetimeProcessed should not be empty")
	}
}

func TestRenderTemplate(t *testing.T) {
	// Create a temporary template file
	tempDir := t.TempDir()
	templatePath := filepath.Join(tempDir, "test_template.md")
	outputPath := filepath.Join(tempDir, "output.md")

	templateContent := `# {{.Filename}}

{{.Content}}

Pages: {{.PageCount}}
Model: {{.ModelUsed}}`

	err := os.WriteFile(templatePath, []byte(templateContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create template file: %v", err)
	}

	// Create test data
	data := template.Data{
		Content:   "Test handwritten content",
		Filename:  "test.pdf",
		PageCount: 2,
		ModelUsed: "gemini-1.5-pro",
	}

	// Render template
	err = template.RenderTemplate(templatePath, outputPath, data)
	if err != nil {
		t.Fatalf("Failed to render template: %v", err)
	}

	// Check output file exists
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Error("Output file was not created")
	}

	// Check output content
	outputContent, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	output := string(outputContent)
	if !strings.Contains(output, "# test.pdf") {
		t.Error("Output does not contain expected title")
	}

	if !strings.Contains(output, "Test handwritten content") {
		t.Error("Output does not contain expected content")
	}

	if !strings.Contains(output, "Pages: 2") {
		t.Error("Output does not contain expected page count")
	}
}

func TestRenderTemplate_TemplateNotFound(t *testing.T) {
	tempDir := t.TempDir()
	templatePath := filepath.Join(tempDir, "nonexistent.md")
	outputPath := filepath.Join(tempDir, "output.md")

	data := template.Data{}

	err := template.RenderTemplate(templatePath, outputPath, data)
	if err == nil {
		t.Error("Expected error for non-existent template file")
	}

	if !strings.Contains(err.Error(), "template file not found") {
		t.Errorf("Expected 'template file not found' error, got: %v", err)
	}
}
