package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"golang.org/x/oauth2"
)

type Config struct {
	ClientID     string        `json:"client_id"`
	ClientSecret string        `json:"client_secret"`
	Token        *oauth2.Token `json:"token,omitempty"`
}

// GetConfigPath returns the path to the TUIs global config file path
func GetConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	// Credentials will get configured at $HOME/.spotify-tui/config.json
	configDir := filepath.Join(home, ".spotify-tui")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return "", err
	}
	return filepath.Join(configDir, "config.json"), nil
}

// LoadConfig returns the parsed JSON data from $HOME/.spotify-tui/config.json
func LoadConfig() (*Config, error) {
	path, err := GetConfigPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{}, nil
		}
		return nil, err
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

// Save writes auth info to $HOME/.spotify-tui/config.json
func (c *Config) Save() error {
	path, err := GetConfigPath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0600)
}

// IsConfigured checks to see if a clientID and clientSecret have been set
func (c *Config) IsConfigured() bool {
	return c.ClientID != "" && c.ClientSecret != ""
}

// PromptForCredentials prompts the user for credentails on the command line
func (c *Config) PromptForCredentials() error {
	if c.ClientID == "" {
		fmt.Print("Enter Spotify Client ID: ")
		fmt.Scanln(&c.ClientID)
	}
	if c.ClientSecret == "" {
		fmt.Print("Enter Spotify Client Secret: ")
		fmt.Scanln(&c.ClientSecret)
	}
	return c.Save()
}
