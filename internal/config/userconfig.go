package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

const DefaultReviewInstructions = "Perform a code review of the included changes"

// UserConfig represents the configuration loaded from ~/.prr-config.json.
type UserConfig struct {
	ReviewInstructionsFile string `json:"reviewInstructionsFile"`
}

// LoadUserConfig loads the user configuration from ~/.prr-config.json.
// If the file does not exist, an empty UserConfig is returned with no error.
func LoadUserConfig() (UserConfig, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return UserConfig{}, nil
	}

	configPath := filepath.Join(homeDir, ".prr-config.json")
	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return UserConfig{}, nil
		}
		return UserConfig{}, err
	}

	var cfg UserConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return UserConfig{}, err
	}

	return cfg, nil
}

// ResolveReviewInstructions returns the review instructions from the config file
// specified in the UserConfig, or the default if no file is specified or the file
// cannot be read.
func ResolveReviewInstructions(cfg UserConfig) string {
	if path := strings.TrimSpace(cfg.ReviewInstructionsFile); path != "" {
		data, err := os.ReadFile(path)
		if err == nil {
			if instructions := strings.TrimSpace(string(data)); instructions != "" {
				return instructions
			}
		}
	}

	return DefaultReviewInstructions
}
