package wireplumber

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
)

const (
	WirePlumberConfigContent = `wireplumber.profiles = {
  main = {
    monitor.bluez.seat-monitoring = disabled
  }
}
`
)

type ConfigManager struct {
	configDir  string
	configFile string
}

// NewConfigManager creates a new WirePlumber configuration manager
func NewConfigManager() (*ConfigManager, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get user home directory: %w", err)
	}

	configDir := filepath.Join(homeDir, ".config", "wireplumber", "wireplumber.conf.d")
	configFile := filepath.Join(configDir, "99-home-bt-broker.conf")

	return &ConfigManager{
		configDir:  configDir,
		configFile: configFile,
	}, nil
}

// EnsureConfig ensures that the WirePlumber configuration file exists
func (cm *ConfigManager) EnsureConfig() error {
	log.Printf("WirePlumber Config: Ensuring configuration exists at %s", cm.configFile)

	// Check if the config file already exists
	if _, err := os.Stat(cm.configFile); err == nil {
		log.Printf("WirePlumber Config: Configuration file already exists")
		return cm.validateConfigContent()
	}

	// Create the directory if it doesn't exist
	if err := os.MkdirAll(cm.configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Create the config file
	if err := cm.writeConfigFile(); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	log.Printf("WirePlumber Config: Configuration file created successfully")
	return nil
}

// writeConfigFile writes the WirePlumber configuration content to the file
func (cm *ConfigManager) writeConfigFile() error {
	file, err := os.Create(cm.configFile)
	if err != nil {
		return fmt.Errorf("failed to create config file: %w", err)
	}
	defer file.Close()

	_, err = file.WriteString(WirePlumberConfigContent)
	if err != nil {
		return fmt.Errorf("failed to write config content: %w", err)
	}

	return nil
}

// validateConfigContent checks if the existing config file has the correct content
func (cm *ConfigManager) validateConfigContent() error {
	content, err := os.ReadFile(cm.configFile)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	if string(content) != WirePlumberConfigContent {
		log.Printf("WirePlumber Config: Content differs, updating config file")
		return cm.writeConfigFile()
	}

	log.Printf("WirePlumber Config: Configuration file content is correct")
	return nil
}

// RemoveConfig removes the WirePlumber configuration file
func (cm *ConfigManager) RemoveConfig() error {
	if _, err := os.Stat(cm.configFile); os.IsNotExist(err) {
		log.Printf("WirePlumber Config: Configuration file does not exist, nothing to remove")
		return nil
	}

	if err := os.Remove(cm.configFile); err != nil {
		return fmt.Errorf("failed to remove config file: %w", err)
	}

	log.Printf("WirePlumber Config: Configuration file removed successfully")
	return nil
}

// GetConfigPath returns the path to the configuration file
func (cm *ConfigManager) GetConfigPath() string {
	return cm.configFile
}

// ConfigExists checks if the configuration file exists
func (cm *ConfigManager) ConfigExists() bool {
	_, err := os.Stat(cm.configFile)
	return err == nil
}