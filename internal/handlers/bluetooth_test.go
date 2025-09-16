package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/nerzhul/home-bt-broker/internal/bluetooth"
	"github.com/stretchr/testify/assert"
)

func TestBluetoothHandler_GetAdapters(t *testing.T) {
	tests := []struct {
		name           string
		setupMock      func(*bluetooth.MockBluetoothManager)
		expectedStatus int
		expectedCount  int
	}{
		{
			name: "success - returns adapters",
			setupMock: func(mock *bluetooth.MockBluetoothManager) {
				adapters := []bluetooth.Adapter{
					{
						Path:         "/org/bluez/hci0",
						Name:         "Intel Bluetooth",
						Address:      "AA:BB:CC:DD:EE:00",
						Powered:      true,
						Discoverable: false,
						Discovering:  false,
					},
					{
						Path:         "/org/bluez/hci1",
						Name:         "USB Bluetooth",
						Address:      "11:22:33:44:55:66",
						Powered:      true,
						Discoverable: true,
						Discovering:  true,
					},
				}
				mock.On("GetAdapters").Return(adapters, nil)
			},
			expectedStatus: http.StatusOK,
			expectedCount:  2,
		},
		{
			name: "failure - bluetooth manager error",
			setupMock: func(mock *bluetooth.MockBluetoothManager) {
				mock.On("GetAdapters").Return([]bluetooth.Adapter{}, errors.New("D-Bus connection failed"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedCount:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mock := bluetooth.NewMockBluetoothManager(t)
			tt.setupMock(mock)

			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/api/v1/bluetooth/adapters", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			h := NewBluetoothHandlerWithManager(mock)

			// Test
			err := h.GetAdapters(c)

			// Assert
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, rec.Code)

			if tt.expectedStatus == http.StatusOK {
				var response map[string][]bluetooth.Adapter
				err = json.Unmarshal(rec.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Len(t, response["adapters"], tt.expectedCount)
			}
		})
	}
}

func TestBluetoothHandler_GetDevices(t *testing.T) {
	tests := []struct {
		name           string
		adapterMAC     string
		setupMock      func(*bluetooth.MockBluetoothManager)
		expectedStatus int
		expectedCount  int
	}{
		{
			name:       "success - returns devices",
			adapterMAC: "AA:BB:CC:DD:EE:00",
			setupMock: func(mock *bluetooth.MockBluetoothManager) {
				mock.On("GetAdapterPathByMAC", "AA:BB:CC:DD:EE:00").Return("/org/bluez/hci0", nil)
				
				devices := []bluetooth.Device{
					{
						Path:      "/org/bluez/hci0/dev_11_22_33_44_55_66",
						Name:      "Test Device",
						Address:   "11:22:33:44:55:66",
						Paired:    true,
						Trusted:   true,
						Connected: true,
						Adapter:   "/org/bluez/hci0",
					},
				}
				mock.On("GetDevices", "/org/bluez/hci0").Return(devices, nil)
			},
			expectedStatus: http.StatusOK,
			expectedCount:  1,
		},
		{
			name:       "failure - adapter not found",
			adapterMAC: "FF:FF:FF:FF:FF:FF",
			setupMock: func(mock *bluetooth.MockBluetoothManager) {
				mock.On("GetAdapterPathByMAC", "FF:FF:FF:FF:FF:FF").Return("", errors.New("adapter with MAC address FF:FF:FF:FF:FF:FF not found"))
			},
			expectedStatus: http.StatusNotFound,
			expectedCount:  0,
		},
		{
			name:           "failure - empty adapter MAC",
			adapterMAC:     "",
			setupMock:      func(mock *bluetooth.MockBluetoothManager) {},
			expectedStatus: http.StatusBadRequest,
			expectedCount:  0,
		},
		{
			name:       "failure - get devices error",
			adapterMAC: "AA:BB:CC:DD:EE:00",
			setupMock: func(mock *bluetooth.MockBluetoothManager) {
				mock.On("GetAdapterPathByMAC", "AA:BB:CC:DD:EE:00").Return("/org/bluez/hci0", nil)
				mock.On("GetDevices", "/org/bluez/hci0").Return([]bluetooth.Device{}, errors.New("D-Bus error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedCount:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mock := bluetooth.NewMockBluetoothManager(t)
			tt.setupMock(mock)

			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/api/v1/bluetooth/adapters/"+tt.adapterMAC+"/devices", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetParamNames("adapter")
			c.SetParamValues(tt.adapterMAC)

			h := NewBluetoothHandlerWithManager(mock)

			// Test
			err := h.GetDevices(c)

			// Assert
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, rec.Code)

			if tt.expectedStatus == http.StatusOK {
				var response map[string][]bluetooth.Device
				err = json.Unmarshal(rec.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Len(t, response["devices"], tt.expectedCount)
			}
		})
	}
}

func TestBluetoothHandler_GetTrustedDevices(t *testing.T) {
	tests := []struct {
		name           string
		adapterMAC     string
		setupMock      func(*bluetooth.MockBluetoothManager)
		expectedStatus int
		expectedCount  int
	}{
		{
			name:       "success - returns trusted devices",
			adapterMAC: "AA:BB:CC:DD:EE:00",
			setupMock: func(mock *bluetooth.MockBluetoothManager) {
				mock.On("GetAdapterPathByMAC", "AA:BB:CC:DD:EE:00").Return("/org/bluez/hci0", nil)
				
				devices := []bluetooth.Device{
					{
						Path:      "/org/bluez/hci0/dev_11_22_33_44_55_66",
						Name:      "Trusted Device",
						Address:   "11:22:33:44:55:66",
						Paired:    true,
						Trusted:   true,
						Connected: false,
						Adapter:   "/org/bluez/hci0",
					},
				}
				mock.On("GetTrustedDevices", "/org/bluez/hci0").Return(devices, nil)
			},
			expectedStatus: http.StatusOK,
			expectedCount:  1,
		},
		{
			name:       "failure - adapter not found",
			adapterMAC: "FF:FF:FF:FF:FF:FF",
			setupMock: func(mock *bluetooth.MockBluetoothManager) {
				mock.On("GetAdapterPathByMAC", "FF:FF:FF:FF:FF:FF").Return("", errors.New("adapter not found"))
			},
			expectedStatus: http.StatusNotFound,
			expectedCount:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mock := bluetooth.NewMockBluetoothManager(t)
			tt.setupMock(mock)

			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/api/v1/bluetooth/adapters/"+tt.adapterMAC+"/devices/trusted", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetParamNames("adapter")
			c.SetParamValues(tt.adapterMAC)

			h := NewBluetoothHandlerWithManager(mock)

			// Test
			err := h.GetTrustedDevices(c)

			// Assert
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, rec.Code)

			if tt.expectedStatus == http.StatusOK {
				var response map[string][]bluetooth.Device
				err = json.Unmarshal(rec.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Len(t, response["trusted_devices"], tt.expectedCount)
			}
		})
	}
}

func TestBluetoothHandler_ConnectDevice(t *testing.T) {
	tests := []struct {
		name           string
		adapterMAC     string
		deviceMAC      string
		setupMock      func(*bluetooth.MockBluetoothManager)
		expectedStatus int
		expectedBody   map[string]string
	}{
		{
			name:       "success - device connected",
			adapterMAC: "AA:BB:CC:DD:EE:00",
			deviceMAC:  "11:22:33:44:55:66",
			setupMock: func(mock *bluetooth.MockBluetoothManager) {
				mock.On("GetAdapterPathByMAC", "AA:BB:CC:DD:EE:00").Return("/org/bluez/hci0", nil)
				mock.On("ConnectDevice", "/org/bluez/hci0", "11:22:33:44:55:66").Return(nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   map[string]string{"message": "device connection initiated successfully"},
		},
		{
			name:           "failure - empty adapter MAC",
			adapterMAC:     "",
			deviceMAC:      "11:22:33:44:55:66",
			setupMock:      func(mock *bluetooth.MockBluetoothManager) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   map[string]string{"error": "adapter MAC address parameter is required"},
		},
		{
			name:           "failure - empty device MAC",
			adapterMAC:     "AA:BB:CC:DD:EE:00",
			deviceMAC:      "",
			setupMock:      func(mock *bluetooth.MockBluetoothManager) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   map[string]string{"error": "device MAC address parameter is required"},
		},
		{
			name:       "failure - adapter not found",
			adapterMAC: "FF:FF:FF:FF:FF:FF",
			deviceMAC:  "11:22:33:44:55:66",
			setupMock: func(mock *bluetooth.MockBluetoothManager) {
				mock.On("GetAdapterPathByMAC", "FF:FF:FF:FF:FF:FF").Return("", errors.New("adapter not found"))
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   map[string]string{"error": "adapter not found: adapter not found"},
		},
		{
			name:       "failure - connect device error",
			adapterMAC: "AA:BB:CC:DD:EE:00",
			deviceMAC:  "11:22:33:44:55:66",
			setupMock: func(mock *bluetooth.MockBluetoothManager) {
				mock.On("GetAdapterPathByMAC", "AA:BB:CC:DD:EE:00").Return("/org/bluez/hci0", nil)
				mock.On("ConnectDevice", "/org/bluez/hci0", "11:22:33:44:55:66").Return(errors.New("connection failed"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   map[string]string{"error": "failed to connect device: connection failed"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mock := bluetooth.NewMockBluetoothManager(t)
			tt.setupMock(mock)

			e := echo.New()
			req := httptest.NewRequest(http.MethodPost, "/api/v1/bluetooth/adapters/"+tt.adapterMAC+"/devices/"+tt.deviceMAC+"/connect", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetParamNames("adapter", "mac")
			c.SetParamValues(tt.adapterMAC, tt.deviceMAC)

			h := NewBluetoothHandlerWithManager(mock)

			// Test
			err := h.ConnectDevice(c)

			// Assert
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, rec.Code)
			
			var response map[string]string
			err = json.Unmarshal(rec.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedBody, response)
		})
	}
}

func TestBluetoothHandler_PairDevice(t *testing.T) {
	tests := []struct {
		name           string
		adapterMAC     string
		deviceMAC      string
		setupMock      func(*bluetooth.MockBluetoothManager)
		expectedStatus int
		expectedBody   map[string]string
	}{
		{
			name:       "success - device paired",
			adapterMAC: "AA:BB:CC:DD:EE:00",
			deviceMAC:  "11:22:33:44:55:66",
			setupMock: func(mock *bluetooth.MockBluetoothManager) {
				mock.On("GetAdapterPathByMAC", "AA:BB:CC:DD:EE:00").Return("/org/bluez/hci0", nil)
				mock.On("PairDevice", "/org/bluez/hci0", "11:22:33:44:55:66").Return(nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   map[string]string{"message": "device pairing initiated successfully"},
		},
		{
			name:       "failure - pair device error",
			adapterMAC: "AA:BB:CC:DD:EE:00",
			deviceMAC:  "11:22:33:44:55:66",
			setupMock: func(mock *bluetooth.MockBluetoothManager) {
				mock.On("GetAdapterPathByMAC", "AA:BB:CC:DD:EE:00").Return("/org/bluez/hci0", nil)
				mock.On("PairDevice", "/org/bluez/hci0", "11:22:33:44:55:66").Return(errors.New("pairing failed"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   map[string]string{"error": "failed to pair device: pairing failed"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mock := bluetooth.NewMockBluetoothManager(t)
			tt.setupMock(mock)

			e := echo.New()
			req := httptest.NewRequest(http.MethodPost, "/api/v1/bluetooth/adapters/"+tt.adapterMAC+"/devices/"+tt.deviceMAC+"/pair", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetParamNames("adapter", "mac")
			c.SetParamValues(tt.adapterMAC, tt.deviceMAC)

			h := NewBluetoothHandlerWithManager(mock)

			// Test
			err := h.PairDevice(c)

			// Assert
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, rec.Code)
			
			var response map[string]string
			err = json.Unmarshal(rec.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedBody, response)
		})
	}
}

func TestBluetoothHandler_TrustDevice(t *testing.T) {
	tests := []struct {
		name           string
		adapterMAC     string
		deviceMAC      string
		setupMock      func(*bluetooth.MockBluetoothManager)
		expectedStatus int
		expectedBody   map[string]string
	}{
		{
			name:       "success - device trusted",
			adapterMAC: "AA:BB:CC:DD:EE:00",
			deviceMAC:  "11:22:33:44:55:66",
			setupMock: func(mock *bluetooth.MockBluetoothManager) {
				mock.On("GetAdapterPathByMAC", "AA:BB:CC:DD:EE:00").Return("/org/bluez/hci0", nil)
				mock.On("TrustDevice", "/org/bluez/hci0", "11:22:33:44:55:66").Return(nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   map[string]string{"message": "device trusted successfully"},
		},
		{
			name:       "failure - trust device error",
			adapterMAC: "AA:BB:CC:DD:EE:00",
			deviceMAC:  "11:22:33:44:55:66",
			setupMock: func(mock *bluetooth.MockBluetoothManager) {
				mock.On("GetAdapterPathByMAC", "AA:BB:CC:DD:EE:00").Return("/org/bluez/hci0", nil)
				mock.On("TrustDevice", "/org/bluez/hci0", "11:22:33:44:55:66").Return(errors.New("trust failed"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   map[string]string{"error": "failed to trust device: trust failed"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mock := bluetooth.NewMockBluetoothManager(t)
			tt.setupMock(mock)

			e := echo.New()
			req := httptest.NewRequest(http.MethodPost, "/api/v1/bluetooth/adapters/"+tt.adapterMAC+"/devices/"+tt.deviceMAC+"/trust", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetParamNames("adapter", "mac")
			c.SetParamValues(tt.adapterMAC, tt.deviceMAC)

			h := NewBluetoothHandlerWithManager(mock)

			// Test
			err := h.TrustDevice(c)

			// Assert
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, rec.Code)
			
			var response map[string]string
			err = json.Unmarshal(rec.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedBody, response)
		})
	}
}

func TestBluetoothHandler_RemoveDevice(t *testing.T) {
	tests := []struct {
		name           string
		adapterMAC     string
		deviceMAC      string
		setupMock      func(*bluetooth.MockBluetoothManager)
		expectedStatus int
		expectedBody   map[string]string
	}{
		{
			name:       "success - device removed",
			adapterMAC: "AA:BB:CC:DD:EE:00",
			deviceMAC:  "11:22:33:44:55:66",
			setupMock: func(mock *bluetooth.MockBluetoothManager) {
				mock.On("GetAdapterPathByMAC", "AA:BB:CC:DD:EE:00").Return("/org/bluez/hci0", nil)
				mock.On("RemoveDevice", "/org/bluez/hci0", "11:22:33:44:55:66").Return(nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   map[string]string{"message": "device removed successfully"},
		},
		{
			name:       "failure - remove device error",
			adapterMAC: "AA:BB:CC:DD:EE:00",
			deviceMAC:  "11:22:33:44:55:66",
			setupMock: func(mock *bluetooth.MockBluetoothManager) {
				mock.On("GetAdapterPathByMAC", "AA:BB:CC:DD:EE:00").Return("/org/bluez/hci0", nil)
				mock.On("RemoveDevice", "/org/bluez/hci0", "11:22:33:44:55:66").Return(errors.New("remove failed"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   map[string]string{"error": "failed to remove device: remove failed"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mock := bluetooth.NewMockBluetoothManager(t)
			tt.setupMock(mock)

			e := echo.New()
			req := httptest.NewRequest(http.MethodDelete, "/api/v1/bluetooth/adapters/"+tt.adapterMAC+"/devices/"+tt.deviceMAC, nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetParamNames("adapter", "mac")
			c.SetParamValues(tt.adapterMAC, tt.deviceMAC)

			h := NewBluetoothHandlerWithManager(mock)

			// Test
			err := h.RemoveDevice(c)

			// Assert
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, rec.Code)
			
			var response map[string]string
			err = json.Unmarshal(rec.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedBody, response)
		})
	}
}