package handlers

import (
	"database/sql"
	"net/http"
	"regexp"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/nerzhul/home-bt-broker/internal/database"
)

type BluetoothAudioHandler struct {
	db *sql.DB
}

type BluetoothAudioDevice struct {
	MAC     string `json:"mac"`
	Enabled bool   `json:"enabled"`
}

func NewBluetoothAudioHandler(db *sql.DB) *BluetoothAudioHandler {
	return &BluetoothAudioHandler{db: db}
}

// isValidMAC validates MAC address format
func isValidMAC(mac string) bool {
	// MAC address pattern: XX:XX:XX:XX:XX:XX or XX-XX-XX-XX-XX-XX
	pattern := `^([0-9A-Fa-f]{2}[:-]){5}([0-9A-Fa-f]{2})$`
	matched, _ := regexp.MatchString(pattern, mac)
	return matched
}

// normalizeMAC converts MAC address to uppercase with colons
func normalizeMAC(mac string) string {
	// Replace dashes with colons and convert to uppercase
	normalized := strings.ReplaceAll(mac, "-", ":")
	return strings.ToUpper(normalized)
}

// GetBluetoothAudioDevices returns all Bluetooth devices marked as audio-capable
// GET /api/v1/bluetooth/audio-devices
func (h *BluetoothAudioHandler) GetBluetoothAudioDevices(c echo.Context) error {
	devices, err := database.GetAllBluetoothAudioDevices(h.db)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	var result []BluetoothAudioDevice
	for mac, enabled := range devices {
		result = append(result, BluetoothAudioDevice{
			MAC:     mac,
			Enabled: enabled,
		})
	}

	return c.JSON(http.StatusOK, result)
}

// GetBluetoothAudioDevice checks if a specific Bluetooth device is marked as audio-capable
// GET /api/v1/bluetooth/audio-devices/:mac
func (h *BluetoothAudioHandler) GetBluetoothAudioDevice(c echo.Context) error {
	mac := c.Param("mac")
	if !isValidMAC(mac) {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid MAC address format"})
	}

	mac = normalizeMAC(mac)

	enabled, err := database.GetBluetoothAudioDevice(h.db, mac)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	device := BluetoothAudioDevice{
		MAC:     mac,
		Enabled: enabled,
	}

	return c.JSON(http.StatusOK, device)
}

// SetBluetoothAudioDevice marks a Bluetooth device as audio-capable or not
// PUT /api/v1/bluetooth/audio-devices/:mac
func (h *BluetoothAudioHandler) SetBluetoothAudioDevice(c echo.Context) error {
	mac := c.Param("mac")
	if !isValidMAC(mac) {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid MAC address format"})
	}

	mac = normalizeMAC(mac)

	var req struct {
		Enabled bool `json:"enabled"`
	}

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}

	err := database.SetBluetoothAudioDevice(h.db, mac, req.Enabled)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	device := BluetoothAudioDevice{
		MAC:     mac,
		Enabled: req.Enabled,
	}

	return c.JSON(http.StatusOK, device)
}

// RemoveBluetoothAudioDevice removes a Bluetooth device from audio-capable list
// DELETE /api/v1/bluetooth/audio-devices/:mac
func (h *BluetoothAudioHandler) RemoveBluetoothAudioDevice(c echo.Context) error {
	mac := c.Param("mac")
	if !isValidMAC(mac) {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid MAC address format"})
	}

	mac = normalizeMAC(mac)

	err := database.RemoveBluetoothAudioDevice(h.db, mac)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "bluetooth audio device removed successfully"})
}
