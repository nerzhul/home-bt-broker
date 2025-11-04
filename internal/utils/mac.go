package utils

import (
	"regexp"
	"strings"
)

// MAC address regex: XX:XX:XX:XX:XX:XX or XX-XX-XX-XX-XX-XX
var macRegex = regexp.MustCompile(`^([0-9A-Fa-f]{2}[:-]){5}([0-9A-Fa-f]{2})$`)

// IsValidMAC checks if a string is a valid MAC address
func IsValidMAC(mac string) bool {
	return macRegex.MatchString(mac)
}

// NormalizeMAC normalizes a MAC address to uppercase with colons
// e.g., aa:bb:cc:dd:ee:ff -> AA:BB:CC:DD:EE:FF
// e.g., AA-BB-CC-DD-EE-FF -> AA:BB:CC:DD:EE:FF
func NormalizeMAC(mac string) string {
	mac = strings.ToUpper(mac)
	mac = strings.ReplaceAll(mac, "-", ":")
	return mac
}
