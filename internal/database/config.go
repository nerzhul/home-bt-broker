package database

import (
	"database/sql"
	"fmt"
)

type Config struct {
	Key   string `json:"key" db:"config_key"`
	Value string `json:"value" db:"config_value"`
}

// GetConfig retrieves a configuration value by key
func GetConfig(db *sql.DB, key string) (*Config, error) {
	config := &Config{}
	query := `SELECT config_key, config_value FROM config WHERE config_key = ?`
	
	err := db.QueryRow(query, key).Scan(&config.Key, &config.Value)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("config key '%s' not found", key)
		}
		return nil, fmt.Errorf("failed to get config: %w", err)
	}
	
	return config, nil
}

// SetConfig creates or updates a configuration entry
func SetConfig(db *sql.DB, key, value string) error {
	query := `INSERT OR REPLACE INTO config (config_key, config_value) VALUES (?, ?)`
	
	_, err := db.Exec(query, key, value)
	if err != nil {
		return fmt.Errorf("failed to set config: %w", err)
	}
	
	return nil
}

// DeleteConfig removes a configuration entry
func DeleteConfig(db *sql.DB, key string) error {
	query := `DELETE FROM config WHERE config_key = ?`
	
	result, err := db.Exec(query, key)
	if err != nil {
		return fmt.Errorf("failed to delete config: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return fmt.Errorf("config key '%s' not found", key)
	}
	
	return nil
}

// ConfigExists checks if a configuration key exists
func ConfigExists(db *sql.DB, key string) (bool, error) {
	query := `SELECT 1 FROM config WHERE config_key = ?`
	
	var exists int
	err := db.QueryRow(query, key).Scan(&exists)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, fmt.Errorf("failed to check config existence: %w", err)
	}
	
	return true, nil
}

// SetBluetoothAudioDevice marks a Bluetooth device MAC as audio-capable
func SetBluetoothAudioDevice(db *sql.DB, mac string, enabled bool) error {
	key := fmt.Sprintf("bluetooth_audio_device_%s", mac)
	value := "false"
	if enabled {
		value = "true"
	}
	
	return SetConfig(db, key, value)
}

// GetBluetoothAudioDevice checks if a Bluetooth device MAC is marked as audio-capable
func GetBluetoothAudioDevice(db *sql.DB, mac string) (bool, error) {
	key := fmt.Sprintf("bluetooth_audio_device_%s", mac)
	
	config, err := GetConfig(db, key)
	if err != nil {
		// If key doesn't exist, device is not marked as audio-capable
		return false, nil
	}
	
	return config.Value == "true", nil
}

// GetAllBluetoothAudioDevices returns all Bluetooth devices marked as audio-capable
func GetAllBluetoothAudioDevices(db *sql.DB) (map[string]bool, error) {
	query := `SELECT config_key, config_value FROM config WHERE config_key LIKE 'bluetooth_audio_device_%'`
	
	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query bluetooth audio devices: %w", err)
	}
	defer rows.Close()
	
	devices := make(map[string]bool)
	
	for rows.Next() {
		var key, value string
		if err := rows.Scan(&key, &value); err != nil {
			return nil, fmt.Errorf("failed to scan bluetooth audio device: %w", err)
		}
		
		// Extract MAC from key (remove "bluetooth_audio_device_" prefix)
		mac := key[23:] // len("bluetooth_audio_device_") = 23
		devices[mac] = value == "true"
	}
	
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}
	
	return devices, nil
}

// RemoveBluetoothAudioDevice removes a Bluetooth device from audio-capable list
func RemoveBluetoothAudioDevice(db *sql.DB, mac string) error {
	key := fmt.Sprintf("bluetooth_audio_device_%s", mac)
	return DeleteConfig(db, key)
}