package handlers

import (
	"database/sql"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/nerzhul/home-bt-broker/internal/database"
	"github.com/nerzhul/home-bt-broker/internal/utils"
)

type BluetoothAliasHandler struct {
	db *sql.DB
}

func NewBluetoothAliasHandler(db *sql.DB) *BluetoothAliasHandler {
	return &BluetoothAliasHandler{db: db}
}

// GetAllAliases returns a map of MAC->alias
// GET /api/v1/bluetooth/aliases
func (h *BluetoothAliasHandler) GetAllAliases(c echo.Context) error {
	aliases, err := database.GetAllAliases(h.db)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]interface{}{"aliases": aliases})
}

// GetAlias returns the alias for a specific MAC
// GET /api/v1/bluetooth/aliases/:mac
func (h *BluetoothAliasHandler) GetAlias(c echo.Context) error {
	mac := c.Param("mac")
	if !utils.IsValidMAC(mac) {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid MAC address format"})
	}
	mac = utils.NormalizeMAC(mac)

	alias, err := database.GetAlias(h.db, mac)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	if alias == "" {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "alias not found"})
	}
	return c.JSON(http.StatusOK, map[string]string{"mac": mac, "alias": alias})
}

// SetAlias sets the alias for a specific MAC
// PUT /api/v1/bluetooth/aliases/:mac
func (h *BluetoothAliasHandler) SetAlias(c echo.Context) error {
	mac := c.Param("mac")
	if !utils.IsValidMAC(mac) {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid MAC address format"})
	}
	mac = utils.NormalizeMAC(mac)
	var req struct {
		Alias string `json:"alias"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}
	if req.Alias == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "alias is required"})
	}
	if err := database.SetAlias(h.db, mac, req.Alias); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]string{"message": "alias set", "mac": mac, "alias": req.Alias})
}

// DeleteAlias deletes the alias for a specific MAC
// DELETE /api/v1/bluetooth/aliases/:mac
func (h *BluetoothAliasHandler) DeleteAlias(c echo.Context) error {
	mac := c.Param("mac")
	if !utils.IsValidMAC(mac) {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid MAC address format"})
	}
	mac = utils.NormalizeMAC(mac)
	if err := database.DeleteAlias(h.db, mac); err != nil {
		if err == sql.ErrNoRows {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "alias not found"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]string{"message": "alias deleted", "mac": mac})
}
