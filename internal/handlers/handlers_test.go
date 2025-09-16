package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestHandler_Liveness(t *testing.T) {
	// Setup
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/livez", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	h := &Handler{}

	// Test
	err := h.Liveness(c)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	
	var response map[string]string
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "alive", response["status"])
}

func TestHandler_Readiness(t *testing.T) {
	tests := []struct {
		name           string
		setupMock      func(sqlmock.Sqlmock)
		expectedStatus int
		expectedBody   map[string]string
	}{
		{
			name: "success - database is healthy",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectPing()
			},
			expectedStatus: http.StatusOK,
			expectedBody:   map[string]string{"status": "ready"},
		},
		{
			name: "failure - database connection failed",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectPing().WillReturnError(errors.New("connection failed"))
			},
			expectedStatus: http.StatusServiceUnavailable,
			expectedBody:   map[string]string{"status": "not ready", "error": "database connection failed"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
			assert.NoError(t, err)
			defer db.Close()

			tt.setupMock(mock)

			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			h := NewHandlerWithDB(db)

			// Test
			err = h.Readiness(c)

			// Assert
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, rec.Code)
			
			var response map[string]string
			err = json.Unmarshal(rec.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedBody, response)

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestHandler_CreateToken(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    string
		setupMock      func(sqlmock.Sqlmock)
		expectedStatus int
		expectedBody   map[string]string
	}{
		{
			name:        "success - token created",
			requestBody: `{"username":"testuser","token":"testtoken"}`,
			setupMock: func(mock sqlmock.Sqlmock) {
				// First query to check if username exists
				mock.ExpectQuery("SELECT token FROM user_tokens WHERE username = ?").
					WithArgs("testuser").
					WillReturnError(sql.ErrNoRows)
				
				// Insert new token
				mock.ExpectExec("INSERT INTO user_tokens \\(username, token, created_at\\) VALUES \\(\\?, \\?, \\?\\)").
					WithArgs("testuser", "testtoken", sqlmock.AnyArg()).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			expectedStatus: http.StatusCreated,
			expectedBody:   map[string]string{"message": "token created successfully"},
		},
		{
			name:        "failure - username already exists",
			requestBody: `{"username":"testuser","token":"testtoken"}`,
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT token FROM user_tokens WHERE username = ?").
					WithArgs("testuser").
					WillReturnRows(sqlmock.NewRows([]string{"token"}).AddRow("existingtoken"))
			},
			expectedStatus: http.StatusConflict,
			expectedBody:   map[string]string{"error": "username already exists"},
		},
		{
			name:           "failure - invalid request body",
			requestBody:    `{"username":"","token":"testtoken"}`,
			setupMock:      func(mock sqlmock.Sqlmock) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   map[string]string{"error": "username and token are required"},
		},
		{
			name:        "failure - database error on insert",
			requestBody: `{"username":"testuser","token":"testtoken"}`,
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT token FROM user_tokens WHERE username = ?").
					WithArgs("testuser").
					WillReturnError(sql.ErrNoRows)
				
				mock.ExpectExec("INSERT INTO user_tokens \\(username, token, created_at\\) VALUES \\(\\?, \\?, \\?\\)").
					WithArgs("testuser", "testtoken", sqlmock.AnyArg()).
					WillReturnError(errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   map[string]string{"error": "failed to create token"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			db, mock, err := sqlmock.New()
			assert.NoError(t, err)
			defer db.Close()

			tt.setupMock(mock)

			e := echo.New()
			req := httptest.NewRequest(http.MethodPost, "/api/v1/tokens", strings.NewReader(tt.requestBody))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			h := NewHandlerWithDB(db)

			// Test
			err = h.CreateToken(c)

			// Assert
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, rec.Code)
			
			var response map[string]string
			err = json.Unmarshal(rec.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedBody, response)

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestHandler_GetTokens(t *testing.T) {
	tests := []struct {
		name           string
		setupMock      func(sqlmock.Sqlmock)
		expectedStatus int
		expectedTokens []Token
	}{
		{
			name: "success - returns tokens",
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"username", "token", "created_at"}).
					AddRow("user1", "token1", time.Now()).
					AddRow("user2", "token2", time.Now())
				
				mock.ExpectQuery("SELECT username, token, created_at FROM user_tokens ORDER BY created_at DESC").
					WillReturnRows(rows)
			},
			expectedStatus: http.StatusOK,
			expectedTokens: []Token{
				{Username: "user1", Token: "token1"},
				{Username: "user2", Token: "token2"},
			},
		},
		{
			name: "success - empty result",
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"username", "token", "created_at"})
				mock.ExpectQuery("SELECT username, token, created_at FROM user_tokens ORDER BY created_at DESC").
					WillReturnRows(rows)
			},
			expectedStatus: http.StatusOK,
			expectedTokens: []Token{},
		},
		{
			name: "failure - database error",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT username, token, created_at FROM user_tokens ORDER BY created_at DESC").
					WillReturnError(errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			db, mock, err := sqlmock.New()
			assert.NoError(t, err)
			defer db.Close()

			tt.setupMock(mock)

			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/api/v1/tokens", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			h := NewHandlerWithDB(db)

			// Test
			err = h.GetTokens(c)

			// Assert
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, rec.Code)
			
			if tt.expectedStatus == http.StatusOK {
				var response []Token
				err = json.Unmarshal(rec.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Len(t, response, len(tt.expectedTokens))
				
				for i, token := range response {
					assert.Equal(t, tt.expectedTokens[i].Username, token.Username)
					assert.Equal(t, tt.expectedTokens[i].Token, token.Token)
				}
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestHandler_GetToken(t *testing.T) {
	tests := []struct {
		name           string
		username       string
		setupMock      func(sqlmock.Sqlmock)
		expectedStatus int
		expectedToken  *Token
	}{
		{
			name:     "success - token found",
			username: "testuser",
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"username", "token", "created_at"}).
					AddRow("testuser", "testtoken", time.Now())
				
				mock.ExpectQuery("SELECT username, token, created_at FROM user_tokens WHERE username = ?").
					WithArgs("testuser").
					WillReturnRows(rows)
			},
			expectedStatus: http.StatusOK,
			expectedToken:  &Token{Username: "testuser", Token: "testtoken"},
		},
		{
			name:     "failure - token not found",
			username: "nonexistent",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT username, token, created_at FROM user_tokens WHERE username = ?").
					WithArgs("nonexistent").
					WillReturnError(sql.ErrNoRows)
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:     "failure - empty username",
			username: "",
			setupMock: func(mock sqlmock.Sqlmock) {},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			db, mock, err := sqlmock.New()
			assert.NoError(t, err)
			defer db.Close()

			tt.setupMock(mock)

			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/api/v1/tokens/"+tt.username, nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetParamNames("username")
			c.SetParamValues(tt.username)

			h := NewHandlerWithDB(db)

			// Test
			err = h.GetToken(c)

			// Assert
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, rec.Code)
			
			if tt.expectedToken != nil {
				var response Token
				err = json.Unmarshal(rec.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedToken.Username, response.Username)
				assert.Equal(t, tt.expectedToken.Token, response.Token)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestHandler_DeleteToken(t *testing.T) {
	tests := []struct {
		name           string
		username       string
		setupMock      func(sqlmock.Sqlmock)
		expectedStatus int
		expectedBody   map[string]string
	}{
		{
			name:     "success - token deleted",
			username: "testuser",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("DELETE FROM user_tokens WHERE username = ?").
					WithArgs("testuser").
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
			expectedStatus: http.StatusOK,
			expectedBody:   map[string]string{"message": "token deleted successfully"},
		},
		{
			name:     "failure - token not found",
			username: "nonexistent",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("DELETE FROM user_tokens WHERE username = ?").
					WithArgs("nonexistent").
					WillReturnResult(sqlmock.NewResult(0, 0))
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   map[string]string{"error": "token not found"},
		},
		{
			name:           "failure - empty username",
			username:       "",
			setupMock:      func(mock sqlmock.Sqlmock) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   map[string]string{"error": "username parameter is required"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			db, mock, err := sqlmock.New()
			assert.NoError(t, err)
			defer db.Close()

			tt.setupMock(mock)

			e := echo.New()
			req := httptest.NewRequest(http.MethodDelete, "/api/v1/tokens/"+tt.username, nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetParamNames("username")
			c.SetParamValues(tt.username)

			h := NewHandlerWithDB(db)

			// Test
			err = h.DeleteToken(c)

			// Assert
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, rec.Code)
			
			var response map[string]string
			err = json.Unmarshal(rec.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedBody, response)

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}