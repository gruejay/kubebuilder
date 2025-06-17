package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type AIConfig struct {
	Provider string `yaml:"provider"` // "openai", "anthropic", "ollama", etc.
	BaseURL  string `yaml:"base_url"`
	Model    string `yaml:"model"`
	APIKey   string `yaml:"api_key,omitempty"` // Optional in config, can use env var
}

type Config struct {
	AI AIConfig `yaml:"ai"`
}

func Load() (*Config, error) {
	configPath, err := getConfigPath()
	if err != nil {
		return getDefaultConfig(), nil // Return default config if no config file
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return getDefaultConfig(), nil // Return default config if file doesn't exist
		}
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Override with environment variables if available
	if apiKey := os.Getenv("KUBEGUIDE_AI_API_KEY"); apiKey != "" {
		config.AI.APIKey = apiKey
	}
	if baseURL := os.Getenv("KUBEGUIDE_AI_BASE_URL"); baseURL != "" {
		config.AI.BaseURL = baseURL
	}
	if model := os.Getenv("KUBEGUIDE_AI_MODEL"); model != "" {
		config.AI.Model = model
	}

	// Auto-detect provider if not set
	if config.AI.Provider == "" || config.AI.Provider == "openai" {
		config.AI.Provider = detectProvider(&config.AI)
	}

	return &config, nil
}

func getConfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	
	configDir := filepath.Join(homeDir, ".config", "kubeguide")
	configFile := filepath.Join(configDir, "config.yaml")
	
	return configFile, nil
}

func getDefaultConfig() *Config {
	// Check for any available API keys and set defaults accordingly
	apiKey := os.Getenv("KUBEGUIDE_AI_API_KEY")
	if apiKey == "" {
		apiKey = os.Getenv("OPENAI_API_KEY")
	}
	if apiKey == "" {
		apiKey = os.Getenv("ANTHROPIC_API_KEY")
	}

	provider := "openai"
	baseURL := "https://api.openai.com/v1"
	model := "gpt-4o-mini"

	// Auto-detect based on API key
	if strings.HasPrefix(apiKey, "sk-ant-") {
		provider = "anthropic"
		baseURL = "https://api.anthropic.com"
		model = "claude-3-haiku-20240307"
	}

	return &Config{
		AI: AIConfig{
			Provider: provider,
			BaseURL:  baseURL,
			Model:    model,
			APIKey:   apiKey,
		},
	}
}

func (c *Config) Save() error {
	configPath, err := getConfigPath()
	if err != nil {
		return err
	}

	// Create config directory if it doesn't exist
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Don't save API key to file for security
	configToSave := *c
	configToSave.AI.APIKey = ""

	data, err := yaml.Marshal(&configToSave)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

func detectProvider(aiConfig *AIConfig) string {
	// Check API key patterns
	if aiConfig.APIKey != "" {
		if strings.HasPrefix(aiConfig.APIKey, "sk-ant-") {
			return "anthropic"
		}
		if strings.HasPrefix(aiConfig.APIKey, "sk-") {
			return "openai"
		}
	}

	// Check base URL patterns
	if strings.Contains(aiConfig.BaseURL, "anthropic.com") {
		return "anthropic"
	}
	if strings.Contains(aiConfig.BaseURL, "openai.com") {
		return "openai"
	}
	if strings.Contains(aiConfig.BaseURL, "localhost") || strings.Contains(aiConfig.BaseURL, "127.0.0.1") {
		return "ollama"
	}

	// Default to openai
	return "openai"
}