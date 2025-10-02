package handlers

import (
	"database/sql"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/nerzhul/home-bt-broker/internal/database"
	"github.com/nerzhul/home-bt-broker/internal/pipewire"
	"github.com/nerzhul/home-bt-broker/internal/utils"
)

type BluetoothAudioHandler struct {
	db *sql.DB
}

type BluetoothAudioDevice struct {
	MAC     string `json:"mac"`
	Enabled bool   `json:"enabled"`
}

type CombinedAudioResponse struct {
	Devices []string `json:"devices"`
}

func NewBluetoothAudioHandler(db *sql.DB) *BluetoothAudioHandler {
	return &BluetoothAudioHandler{db: db}
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
	if !utils.IsValidMAC(mac) {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid MAC address format"})
	}

	mac = utils.NormalizeMAC(mac)

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
	if !utils.IsValidMAC(mac) {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid MAC address format"})
	}

	mac = utils.NormalizeMAC(mac)

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
	if !utils.IsValidMAC(mac) {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid MAC address format"})
	}

	mac = utils.NormalizeMAC(mac)

	err := database.RemoveBluetoothAudioDevice(h.db, mac)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "bluetooth audio device removed successfully"})
}

// GetCombinedAudioConfig returns the combined audio configuration
// GET /api/v1/bluetooth/combined-audio
func (h *BluetoothAudioHandler) GetCombinedAudioConfig(c echo.Context) error {
	config, err := database.GetCombinedAudioConfig(h.db)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	response := CombinedAudioResponse{
		Devices: config.Devices,
	}

	return c.JSON(http.StatusOK, response)
}

// AddDeviceToCombined adds a device to the combined audio configuration
// POST /api/v1/bluetooth/combined-audio/devices/:mac
func (h *BluetoothAudioHandler) AddDeviceToCombined(c echo.Context) error {
	mac := c.Param("mac")
	if !utils.IsValidMAC(mac) {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid MAC address format"})
	}

	mac = utils.NormalizeMAC(mac)

	// Check if device is marked as audio-capable
	isAudio, err := database.GetBluetoothAudioDevice(h.db, mac)
	if err != nil || !isAudio {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "device must be marked as audio-capable first"})
	}

	err = database.AddDeviceToCombined(h.db, mac)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	// Update PipeWire stream rules configuration
	combinedConfig, err := database.GetCombinedAudioConfig(h.db)
	if err == nil {
		if err := pipewire.UpdateCombinedSinks(combinedConfig.Devices); err != nil {
			// Log the error but don't fail the API call since the device was added to DB
			c.Echo().Logger.Errorf("Failed to update PipeWire stream rules: %v", err)
		}
	} else {
		c.Echo().Logger.Errorf("Failed to get combined audio config for PipeWire update: %v", err)
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "device added to combined audio"})
}

// RemoveDeviceFromCombined removes a device from the combined audio configuration
// DELETE /api/v1/bluetooth/combined-audio/devices/:mac
func (h *BluetoothAudioHandler) RemoveDeviceFromCombined(c echo.Context) error {
	mac := c.Param("mac")
	if !utils.IsValidMAC(mac) {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid MAC address format"})
	}

	mac = utils.NormalizeMAC(mac)

	err := database.RemoveDeviceFromCombined(h.db, mac)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	// Update PipeWire stream rules configuration
	combinedConfig, err := database.GetCombinedAudioConfig(h.db)
	if err == nil {
		if len(combinedConfig.Devices) == 0 {
			// Clear stream rules if no devices left
			if err := pipewire.ClearStreamRules(); err != nil {
				c.Echo().Logger.Errorf("Failed to clear PipeWire stream rules: %v", err)
			}
		} else {
			// Update with remaining devices
			if err := pipewire.UpdateCombinedSinks(combinedConfig.Devices); err != nil {
				c.Echo().Logger.Errorf("Failed to update PipeWire stream rules: %v", err)
			}
		}
	} else {
		c.Echo().Logger.Errorf("Failed to get combined audio config for PipeWire update: %v", err)
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "device removed from combined audio"})
}
