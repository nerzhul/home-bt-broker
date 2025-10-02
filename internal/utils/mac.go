package utils

import (
	"regexp"
	"strings"
)

// NormalizeMAC converts MAC address to uppercase with colons
func NormalizeMAC(mac string) string {
	// Replace dashes with colons and convert to uppercase
	normalized := strings.ReplaceAll(mac, "-", ":")
	return strings.ToUpper(normalized)
}

// IsValidMAC validates MAC address format
func IsValidMAC(mac string) bool {
	// MAC address pattern: XX:XX:XX:XX:XX:XX or XX-XX-XX-XX-XX-XX (but not mixed)
	colonPattern := `^([0-9A-Fa-f]{2}:){5}([0-9A-Fa-f]{2})$`
	dashPattern := `^([0-9A-Fa-f]{2}-){5}([0-9A-Fa-f]{2})$`

	colonMatch, _ := regexp.MatchString(colonPattern, mac)
	dashMatch, _ := regexp.MatchString(dashPattern, mac)

	return colonMatch || dashMatch
}
