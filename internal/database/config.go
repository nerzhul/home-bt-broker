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