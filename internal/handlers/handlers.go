package handlers

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/nerzhul/home-bt-broker/internal/database"
)

// AuthMiddleware vérifie l'authentification HTTP Basic (user/pass)
func AuthMiddleware(db database.DatabaseInterface) echo.MiddlewareFunc {
       return func(next echo.HandlerFunc) echo.HandlerFunc {
	       return func(c echo.Context) error {
		       username, password, ok := c.Request().BasicAuth()
		       if !ok || username == "" || password == "" {
			       c.Response().Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
			       return c.JSON(http.StatusUnauthorized, map[string]string{"error": "missing or invalid basic auth"})
		       }

		       var storedToken string
		       err := db.QueryRow("SELECT token FROM user_tokens WHERE username = ?", username).Scan(&storedToken)
		       if err != nil {
			       if err == sql.ErrNoRows {
				       c.Response().Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
				       return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid credentials"})
			       }
			       return c.JSON(http.StatusInternalServerError, map[string]string{"error": "database error"})
		       }

		       if password != storedToken {
			       c.Response().Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
			       return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid credentials"})
		       }

		       c.Set("username", username)
		       return next(c)
	       }
       }
}

type Handler struct {
	db database.DatabaseInterface
}

type Token struct {
	Username  string    `json:"username" db:"username"`
	Token     string    `json:"token" db:"token"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type CreateTokenRequest struct {
	Username string `json:"username" validate:"required"`
	Token    string `json:"token" validate:"required"`
}

func NewHandler(db *sql.DB) *Handler {
	return &Handler{db: db}
}

// NewHandlerWithDB creates a new handler with a custom database interface (for testing)
func NewHandlerWithDB(db database.DatabaseInterface) *Handler {
	return &Handler{db: db}
}

// Readiness endpoint - checks if the service is ready to serve traffic
func (h *Handler) Readiness(c echo.Context) error {
	// Check database connection
	if err := h.db.Ping(); err != nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{
			"status": "not ready",
			"error":  "database connection failed",
		})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"status": "ready",
	})
}

// Liveness endpoint - checks if the service is alive
func (h *Handler) Liveness(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{
		"status": "alive",
	})
}

// CreateToken creates a new username/token pair
func (h *Handler) CreateToken(c echo.Context) error {
	var req CreateTokenRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "invalid request body",
		})
	}

	if req.Username == "" || req.Token == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "username and token are required",
		})
	}

	// Check if username already exists
	var existingToken string
	err := h.db.QueryRow("SELECT token FROM user_tokens WHERE username = ?", req.Username).Scan(&existingToken)
	if err == nil {
		return c.JSON(http.StatusConflict, map[string]string{
			"error": "username already exists",
		})
	} else if err != sql.ErrNoRows {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "database error",
		})
	}

	// Insert new token
	_, err = h.db.Exec("INSERT INTO user_tokens (username, token, created_at) VALUES (?, ?, ?)",
		req.Username, req.Token, time.Now())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "failed to create token",
		})
	}

	return c.JSON(http.StatusCreated, map[string]string{
		"message": "token created successfully",
	})
}

// GetTokens returns all username/token pairs
func (h *Handler) GetTokens(c echo.Context) error {
	rows, err := h.db.Query("SELECT username, token, created_at FROM user_tokens ORDER BY created_at DESC")
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "database error",
		})
	}
	defer rows.Close()

	var tokens []Token
	for rows.Next() {
		var token Token
		if err := rows.Scan(&token.Username, &token.Token, &token.CreatedAt); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": "failed to scan token",
			})
		}
		tokens = append(tokens, token)
	}

	return c.JSON(http.StatusOK, tokens)
}

// GetToken returns a specific token by username
func (h *Handler) GetToken(c echo.Context) error {
	username := c.Param("username")
	if username == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "username parameter is required",
		})
	}

	var token Token
	err := h.db.QueryRow("SELECT username, token, created_at FROM user_tokens WHERE username = ?", username).
		Scan(&token.Username, &token.Token, &token.CreatedAt)
	if err == sql.ErrNoRows {
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": "token not found",
		})
	} else if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "database error",
		})
	}

	return c.JSON(http.StatusOK, token)
}

// DeleteToken removes a token by username
func (h *Handler) DeleteToken(c echo.Context) error {
	username := c.Param("username")
	if username == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "username parameter is required",
		})
	}

	result, err := h.db.Exec("DELETE FROM user_tokens WHERE username = ?", username)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "database error",
		})
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "failed to check affected rows",
		})
	}

	if rowsAffected == 0 {
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": "token not found",
		})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "token deleted successfully",
	})
}