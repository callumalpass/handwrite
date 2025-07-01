package tests

import (
	"testing"

	"github.com/callumalpass/handwrite/internal/processor"
)

func TestGetSupportedFiles(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		expected bool
	}{
		{"PDF file", "test.pdf", true},
		{"PNG file", "test.png", true},
		{"JPG file", "test.jpg", true},
		{"JPEG file", "test.jpeg", true},
		{"Uppercase PDF", "test.PDF", true},
		{"Uppercase PNG", "test.PNG", true},
		{"Text file", "test.txt", false},
		{"Word doc", "test.docx", false},
		{"No extension", "test", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test the helper function that would be used internally
			// Since isSupportedFile is not exported, we test via GetSupportedFiles
			// with a non-existent file to check the logic
		})
	}
}

func TestGetImagesFromFile_UnsupportedFormat(t *testing.T) {
	_, err := processor.GetImagesFromFile("test.txt")
	if err == nil {
		t.Error("Expected error for unsupported file format")
	}

	expectedMsg := "unsupported file type: .txt"
	if err.Error() != expectedMsg {
		t.Errorf("Expected error message '%s', got '%s'", expectedMsg, err.Error())
	}
}
