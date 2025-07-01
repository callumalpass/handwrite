package tests

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/callumalpass/handwrite/internal/config"
)

func TestLoadDefaultConfig(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()

	// Override the home directory for testing
	originalHome := os.Getenv("HOME")
	defer func() {
		os.Setenv("HOME", originalHome)
	}()

	os.Setenv("HOME", tempDir)

	// Load config with empty path - should get defaults since no config exists
	cfg, err := config.LoadConfig("")
	if err != nil {
		t.Fatalf("Failed to load default config: %v", err)
	}

	if cfg.Gemini.Model != "gemini-1.5-pro" {
		t.Errorf("Expected model 'gemini-1.5-pro', got '%s'", cfg.Gemini.Model)
	}

	if cfg.Output.Format != "markdown" {
		t.Errorf("Expected format 'markdown', got '%s'", cfg.Output.Format)
	}

	if cfg.Output.Encoding != "utf-8" {
		t.Errorf("Expected encoding 'utf-8', got '%s'", cfg.Output.Encoding)
	}
}

func TestSetupDefaultConfig(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()

	// Override the home directory for testing
	originalHome := os.Getenv("HOME")
	defer func() {
		os.Setenv("HOME", originalHome)
	}()

	os.Setenv("HOME", tempDir)

	// Setup config
	err := config.SetupDefaultConfig()
	if err != nil {
		t.Fatalf("Failed to setup default config: %v", err)
	}

	// Check if config file was created
	configPath := filepath.Join(tempDir, ".config", "handwrite", "config.yaml")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Errorf("Config file was not created at: %s", configPath)
	}

	// Try to setup again - should fail since file exists
	err = config.SetupDefaultConfig()
	if err == nil {
		t.Error("Expected error when trying to setup config that already exists")
	}
}

