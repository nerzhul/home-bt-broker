package database

import (
	"database/sql"
	"fmt"

	"github.com/nerzhul/home-bt-broker/internal/utils"
)

// SetAlias creates or updates an alias for a device MAC
func SetAlias(db *sql.DB, mac string, alias string) error {
	if !utils.IsValidMAC(mac) {
		return fmt.Errorf("invalid MAC address format")
	}
	mac = utils.NormalizeMAC(mac)
	if alias == "" {
		return fmt.Errorf("alias cannot be empty")
	}
	query := `INSERT INTO bluetooth_aliases (mac, alias, updated_at) VALUES (?, ?, CURRENT_TIMESTAMP)
	ON CONFLICT(mac) DO UPDATE SET alias=excluded.alias, updated_at=CURRENT_TIMESTAMP`
	_, err := db.Exec(query, mac, alias)
	if err != nil {
		return fmt.Errorf("failed to set alias: %w", err)
	}
	return nil
}

// GetAlias retrieves alias for a MAC, returns empty string if not found
func GetAlias(db *sql.DB, mac string) (string, error) {
	if !utils.IsValidMAC(mac) {
		return "", fmt.Errorf("invalid MAC address format")
	}
	mac = utils.NormalizeMAC(mac)
	var alias string
	err := db.QueryRow(`SELECT alias FROM bluetooth_aliases WHERE mac = ?`, mac).Scan(&alias)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", nil
		}
		return "", fmt.Errorf("failed to get alias: %w", err)
	}
	return alias, nil
}

// DeleteAlias removes alias for a MAC
func DeleteAlias(db *sql.DB, mac string) error {
	if !utils.IsValidMAC(mac) {
		return fmt.Errorf("invalid MAC address format")
	}
	mac = utils.NormalizeMAC(mac)
	res, err := db.Exec(`DELETE FROM bluetooth_aliases WHERE mac = ?`, mac)
	if err != nil {
		return fmt.Errorf("failed to delete alias: %w", err)
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// GetAllAliases returns a map of mac->alias
func GetAllAliases(db *sql.DB) (map[string]string, error) {
	rows, err := db.Query(`SELECT mac, alias FROM bluetooth_aliases`)
	if err != nil {
		return nil, fmt.Errorf("failed to list aliases: %w", err)
	}
	defer rows.Close()
	aliases := make(map[string]string)
	for rows.Next() {
		var mac, alias string
		if err := rows.Scan(&mac, &alias); err != nil {
			return nil, fmt.Errorf("failed to scan alias: %w", err)
		}
		aliases[mac] = alias
	}
	return aliases, nil
}
