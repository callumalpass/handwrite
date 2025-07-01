package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

type Config struct {
	Gemini   GeminiConfig   `mapstructure:"gemini" yaml:"gemini"`
	Template TemplateConfig `mapstructure:"template" yaml:"template"`
	Output   OutputConfig   `mapstructure:"output" yaml:"output"`
}

type GeminiConfig struct {
	Model  string `mapstructure:"model" yaml:"model"`
	Prompt string `mapstructure:"prompt" yaml:"prompt"`
}

type TemplateConfig struct {
	Path      string                 `mapstructure:"path" yaml:"path"`
	Variables map[string]interface{} `mapstructure:"variables" yaml:"variables"`
}

type OutputConfig struct {
	Format   string `mapstructure:"format" yaml:"format"`
	Encoding string `mapstructure:"encoding" yaml:"encoding"`
}

func GetDefaultConfigPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(homeDir, ".config", "handwrite", "config.yaml")
}

func SetupDefaultConfig() error {
	configPath := GetDefaultConfigPath()
	
	if _, err := os.Stat(configPath); err == nil {
		return fmt.Errorf("config file already exists at %s", configPath)
	}

	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	defaultConfig := `gemini:
  model: "gemini-1.5-pro"
  prompt: |
    Extract the handwritten text from this image.
    - Use $ for LaTeX, not` + "```" + `latex.
    - Transcribe the text exactly as it appears.
    - The output must be only the transcribed Markdown, with no additional commentary.

template:
  path: "templates/note_template.md"
  variables: {}

output:
  format: "markdown"
  encoding: "utf-8"
`

	if err := os.WriteFile(configPath, []byte(defaultConfig), 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

func LoadConfig(configPath string) (*Config, error) {
	// Load .env file if it exists
	_ = godotenv.Load()

	viper.SetConfigType("yaml")
	
	if configPath != "" {
		viper.SetConfigFile(configPath)
	} else {
		defaultPath := GetDefaultConfigPath()
		if _, err := os.Stat(defaultPath); err == nil {
			viper.SetConfigFile(defaultPath)
		} else {
			// Use default config if no file exists
			return getDefaultConfig(), nil
		}
	}

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			return getDefaultConfig(), nil
		}
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &config, nil
}

func getDefaultConfig() *Config {
	return &Config{
		Gemini: GeminiConfig{
			Model: "gemini-1.5-pro",
			Prompt: `Extract the handwritten text from this image.
- Use $ for LaTeX, not ` + "```" + `latex.
- Transcribe the text exactly as it appears.
- The output must be only the transcribed Markdown, with no additional commentary.`,
		},
		Template: TemplateConfig{
			Path:      "templates/note_template.md",
			Variables: make(map[string]interface{}),
		},
		Output: OutputConfig{
			Format:   "markdown",
			Encoding: "utf-8",
		},
	}
}

func GetGeminiAPIKey() (string, error) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		return "", fmt.Errorf("GEMINI_API_KEY not found in environment variables")
	}
	return apiKey, nil
}