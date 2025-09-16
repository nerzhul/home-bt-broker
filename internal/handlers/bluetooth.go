package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/nerzhul/home-bt-broker/internal/bluetooth"
)

// BluetoothHandler handles Bluetooth-related endpoints
type BluetoothHandler struct {
	btManager *bluetooth.BluetoothManager
}

// NewBluetoothHandler creates a new Bluetooth handler
func NewBluetoothHandler() (*BluetoothHandler, error) {
	btManager, err := bluetooth.NewBluetoothManager()
	if err != nil {
		return nil, err
	}

	return &BluetoothHandler{btManager: btManager}, nil
}

// Close closes the Bluetooth manager connection
func (bh *BluetoothHandler) Close() {
	if bh.btManager != nil {
		bh.btManager.Close()
	}
}

// GetAdapters returns all Bluetooth adapters
func (bh *BluetoothHandler) GetAdapters(c echo.Context) error {
	adapters, err := bh.btManager.GetAdapters()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "failed to get adapters: " + err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"adapters": adapters,
	})
}

// GetDevices returns all devices for a specific adapter by MAC address
func (bh *BluetoothHandler) GetDevices(c echo.Context) error {
	adapterMAC := c.Param("adapter")
	if adapterMAC == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "adapter MAC address parameter is required",
		})
	}

	// Resolve MAC address to adapter path
	adapterPath, err := bh.btManager.GetAdapterPathByMAC(adapterMAC)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": "adapter not found: " + err.Error(),
		})
	}

	devices, err := bh.btManager.GetDevices(adapterPath)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "failed to get devices: " + err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"devices": devices,
	})
}

// GetTrustedDevices returns trusted devices for a specific adapter by MAC address
func (bh *BluetoothHandler) GetTrustedDevices(c echo.Context) error {
	adapterMAC := c.Param("adapter")
	if adapterMAC == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "adapter MAC address parameter is required",
		})
	}

	// Resolve MAC address to adapter path
	adapterPath, err := bh.btManager.GetAdapterPathByMAC(adapterMAC)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": "adapter not found: " + err.Error(),
		})
	}

	devices, err := bh.btManager.GetTrustedDevices(adapterPath)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "failed to get trusted devices: " + err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"trusted_devices": devices,
	})
}

// GetConnectedDevices returns connected devices for a specific adapter by MAC address
func (bh *BluetoothHandler) GetConnectedDevices(c echo.Context) error {
	adapterMAC := c.Param("adapter")
	if adapterMAC == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "adapter MAC address parameter is required",
		})
	}

	// Resolve MAC address to adapter path
	adapterPath, err := bh.btManager.GetAdapterPathByMAC(adapterMAC)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": "adapter not found: " + err.Error(),
		})
	}

	devices, err := bh.btManager.GetConnectedDevices(adapterPath)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "failed to get connected devices: " + err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"connected_devices": devices,
	})
}

// ConnectDevice connects to a device by MAC address using adapter MAC
func (bh *BluetoothHandler) ConnectDevice(c echo.Context) error {
	adapterMAC := c.Param("adapter")
	macAddress := c.Param("mac")
	
	if adapterMAC == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "adapter MAC address parameter is required",
		})
	}
	
	if macAddress == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "device MAC address parameter is required",
		})
	}

	// Resolve MAC address to adapter path
	adapterPath, err := bh.btManager.GetAdapterPathByMAC(adapterMAC)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": "adapter not found: " + err.Error(),
		})
	}

	err = bh.btManager.ConnectDevice(adapterPath, macAddress)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "failed to connect device: " + err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "device connection initiated successfully",
	})
}

// TrustDevice trusts a device by MAC address using adapter MAC
func (bh *BluetoothHandler) TrustDevice(c echo.Context) error {
	adapterMAC := c.Param("adapter")
	macAddress := c.Param("mac")
	
	if adapterMAC == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "adapter MAC address parameter is required",
		})
	}
	
	if macAddress == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "device MAC address parameter is required",
		})
	}

	// Resolve MAC address to adapter path
	adapterPath, err := bh.btManager.GetAdapterPathByMAC(adapterMAC)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": "adapter not found: " + err.Error(),
		})
	}

	err = bh.btManager.TrustDevice(adapterPath, macAddress)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "failed to trust device: " + err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "device trusted successfully",
	})
}

// RemoveDevice removes a device by MAC address using adapter MAC
func (bh *BluetoothHandler) RemoveDevice(c echo.Context) error {
	adapterMAC := c.Param("adapter")
	macAddress := c.Param("mac")
	
	if adapterMAC == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "adapter MAC address parameter is required",
		})
	}
	
	if macAddress == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "device MAC address parameter is required",
		})
	}

	// Resolve MAC address to adapter path
	adapterPath, err := bh.btManager.GetAdapterPathByMAC(adapterMAC)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": "adapter not found: " + err.Error(),
		})
	}

	err = bh.btManager.RemoveDevice(adapterPath, macAddress)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "failed to remove device: " + err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "device removed successfully",
	})
}

// PairDevice pairs with a device by MAC address using adapter MAC
func (bh *BluetoothHandler) PairDevice(c echo.Context) error {
	adapterMAC := c.Param("adapter")
	macAddress := c.Param("mac")
	
	if adapterMAC == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "adapter MAC address parameter is required",
		})
	}
	
	if macAddress == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "device MAC address parameter is required",
		})
	}

	// Resolve MAC address to adapter path
	adapterPath, err := bh.btManager.GetAdapterPathByMAC(adapterMAC)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": "adapter not found: " + err.Error(),
		})
	}

	err = bh.btManager.PairDevice(adapterPath, macAddress)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "failed to pair device: " + err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "device pairing initiated successfully",
	})
}