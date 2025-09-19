package bluetooth

// BluetoothManagerInterface defines the interface for Bluetooth operations
type BluetoothManagerInterface interface {
	GetAdapters() ([]Adapter, error)
	GetAdapterPathByMAC(macAddress string) (string, error)
	GetDevices(adapterPath string) ([]Device, error)
	GetTrustedDevices(adapterPath string) ([]Device, error)
	GetConnectedDevices(adapterPath string) ([]Device, error)
	ConnectDevice(adapterPath, macAddress string) error
	TrustDevice(adapterPath, macAddress string) error
	PairDevice(adapterPath, macAddress string) error
	RemoveDevice(adapterPath, macAddress string) error
	SetDiscoverable(adapterPath string, enable bool) error
	SetDiscovering(adapterPath string, enable bool) error
	Close()
}

// Ensure BluetoothManager implements the interface
var _ BluetoothManagerInterface = (*BluetoothManager)(nil)