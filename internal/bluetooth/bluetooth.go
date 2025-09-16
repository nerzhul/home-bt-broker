package bluetooth

import (
	"fmt"
	"strings"

	"github.com/godbus/dbus/v5"
)

const (
	BluezService        = "org.bluez"
	BluezObjectPath     = "/org/bluez"
	AdapterInterface    = "org.bluez.Adapter1"
	DeviceInterface     = "org.bluez.Device1"
	AgentManagerIface   = "org.bluez.AgentManager1"
	AgentInterface      = "org.bluez.Agent1"
	ObjectManagerIface  = "org.freedesktop.DBus.ObjectManager"
)

type BluetoothManager struct {
	conn *dbus.Conn
}

type Adapter struct {
	Path         string `json:"path"`
	Name         string `json:"name"`
	Address      string `json:"address"`
	Powered      bool   `json:"powered"`
	Discoverable bool   `json:"discoverable"`
	Discovering  bool   `json:"discovering"`
}

type Device struct {
	Path      string `json:"path"`
	Name      string `json:"name"`
	Address   string `json:"address"`
	Paired    bool   `json:"paired"`
	Trusted   bool   `json:"trusted"`
	Connected bool   `json:"connected"`
	Adapter   string `json:"adapter"`
}

// NewBluetoothManager creates a new Bluetooth manager instance
func NewBluetoothManager() (*BluetoothManager, error) {
	conn, err := dbus.SystemBus()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to D-Bus: %w", err)
	}

	return &BluetoothManager{conn: conn}, nil
}

// Close closes the D-Bus connection
func (bm *BluetoothManager) Close() {
	if bm.conn != nil {
		bm.conn.Close()
	}
}

// GetAdapters returns a list of all Bluetooth adapters
func (bm *BluetoothManager) GetAdapters() ([]Adapter, error) {
	obj := bm.conn.Object(BluezService, BluezObjectPath)
	call := obj.Call(ObjectManagerIface+".GetManagedObjects", 0)
	if call.Err != nil {
		return nil, fmt.Errorf("failed to get managed objects: %w", call.Err)
	}

	var objects map[dbus.ObjectPath]map[string]map[string]dbus.Variant
	err := call.Store(&objects)
	if err != nil {
		return nil, fmt.Errorf("failed to parse managed objects: %w", err)
	}

	var adapters []Adapter
	for path, interfaces := range objects {
		if adapterProps, exists := interfaces[AdapterInterface]; exists {
			adapter := Adapter{
				Path: string(path),
			}
			
			if name, ok := adapterProps["Name"]; ok {
				adapter.Name = name.Value().(string)
			}
			if address, ok := adapterProps["Address"]; ok {
				adapter.Address = address.Value().(string)
			}
			if powered, ok := adapterProps["Powered"]; ok {
				adapter.Powered = powered.Value().(bool)
			}
			if discoverable, ok := adapterProps["Discoverable"]; ok {
				adapter.Discoverable = discoverable.Value().(bool)
			}
			if discovering, ok := adapterProps["Discovering"]; ok {
				adapter.Discovering = discovering.Value().(bool)
			}
			
			adapters = append(adapters, adapter)
		}
	}

	return adapters, nil
}

// GetDevices returns all devices for a specific adapter
func (bm *BluetoothManager) GetDevices(adapterPath string) ([]Device, error) {
	obj := bm.conn.Object(BluezService, BluezObjectPath)
	call := obj.Call(ObjectManagerIface+".GetManagedObjects", 0)
	if call.Err != nil {
		return nil, fmt.Errorf("failed to get managed objects: %w", call.Err)
	}

	var objects map[dbus.ObjectPath]map[string]map[string]dbus.Variant
	err := call.Store(&objects)
	if err != nil {
		return nil, fmt.Errorf("failed to parse managed objects: %w", err)
	}

	var devices []Device
	for path, interfaces := range objects {
		if deviceProps, exists := interfaces[DeviceInterface]; exists {
			pathStr := string(path)
			// Check if device belongs to the specified adapter
			if !strings.HasPrefix(pathStr, adapterPath+"/") {
				continue
			}

			device := Device{
				Path:    pathStr,
				Adapter: adapterPath,
			}
			
			if name, ok := deviceProps["Name"]; ok {
				device.Name = name.Value().(string)
			}
			if address, ok := deviceProps["Address"]; ok {
				device.Address = address.Value().(string)
			}
			if paired, ok := deviceProps["Paired"]; ok {
				device.Paired = paired.Value().(bool)
			}
			if trusted, ok := deviceProps["Trusted"]; ok {
				device.Trusted = trusted.Value().(bool)
			}
			if connected, ok := deviceProps["Connected"]; ok {
				device.Connected = connected.Value().(bool)
			}
			
			devices = append(devices, device)
		}
	}

	return devices, nil
}

// GetTrustedDevices returns only trusted devices for a specific adapter
func (bm *BluetoothManager) GetTrustedDevices(adapterPath string) ([]Device, error) {
	devices, err := bm.GetDevices(adapterPath)
	if err != nil {
		return nil, err
	}

	var trustedDevices []Device
	for _, device := range devices {
		if device.Trusted {
			trustedDevices = append(trustedDevices, device)
		}
	}

	return trustedDevices, nil
}

// GetConnectedDevices returns only connected devices for a specific adapter
func (bm *BluetoothManager) GetConnectedDevices(adapterPath string) ([]Device, error) {
	devices, err := bm.GetDevices(adapterPath)
	if err != nil {
		return nil, err
	}

	var connectedDevices []Device
	for _, device := range devices {
		if device.Connected {
			connectedDevices = append(connectedDevices, device)
		}
	}

	return connectedDevices, nil
}

// ConnectDevice connects to a device by MAC address
func (bm *BluetoothManager) ConnectDevice(adapterPath, macAddress string) error {
	devicePath := fmt.Sprintf("%s/dev_%s", adapterPath, strings.ReplaceAll(macAddress, ":", "_"))
	
	obj := bm.conn.Object(BluezService, dbus.ObjectPath(devicePath))
	call := obj.Call(DeviceInterface+".Connect", 0)
	if call.Err != nil {
		return fmt.Errorf("failed to connect to device %s: %w", macAddress, call.Err)
	}

	return nil
}

// TrustDevice sets a device as trusted by MAC address
func (bm *BluetoothManager) TrustDevice(adapterPath, macAddress string) error {
	devicePath := fmt.Sprintf("%s/dev_%s", adapterPath, strings.ReplaceAll(macAddress, ":", "_"))
	
	obj := bm.conn.Object(BluezService, dbus.ObjectPath(devicePath))
	call := obj.Call("org.freedesktop.DBus.Properties.Set", 0, DeviceInterface, "Trusted", dbus.MakeVariant(true))
	if call.Err != nil {
		return fmt.Errorf("failed to trust device %s: %w", macAddress, call.Err)
	}

	return nil
}

// GetAdapterPathByMAC resolves an adapter MAC address to its D-Bus path
func (bm *BluetoothManager) GetAdapterPathByMAC(macAddress string) (string, error) {
	adapters, err := bm.GetAdapters()
	if err != nil {
		return "", err
	}

	for _, adapter := range adapters {
		if adapter.Address == macAddress {
			return adapter.Path, nil
		}
	}

	return "", fmt.Errorf("adapter with MAC address %s not found", macAddress)
}

// PairDevice pairs with a device by MAC address and auto-accepts PIN/passkey
func (bm *BluetoothManager) PairDevice(adapterPath, macAddress string) error {
	devicePath := fmt.Sprintf("%s/dev_%s", adapterPath, strings.ReplaceAll(macAddress, ":", "_"))
	
	obj := bm.conn.Object(BluezService, dbus.ObjectPath(devicePath))
	call := obj.Call(DeviceInterface+".Pair", 0)
	if call.Err != nil {
		return fmt.Errorf("failed to pair with device %s: %w", macAddress, call.Err)
	}

	return nil
}

// RemoveDevice removes a device by MAC address
func (bm *BluetoothManager) RemoveDevice(adapterPath, macAddress string) error {
	devicePath := fmt.Sprintf("%s/dev_%s", adapterPath, strings.ReplaceAll(macAddress, ":", "_"))
	
	obj := bm.conn.Object(BluezService, dbus.ObjectPath(adapterPath))
	call := obj.Call(AdapterInterface+".RemoveDevice", 0, dbus.ObjectPath(devicePath))
	if call.Err != nil {
		return fmt.Errorf("failed to remove device %s: %w", macAddress, call.Err)
	}

	return nil
}