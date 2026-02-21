pub mod agent;
pub mod reconnector;

use std::collections::HashMap;
use std::ops::Deref;
use std::time::Duration;

use anyhow::{anyhow, Context, Result};
use serde::{Deserialize, Serialize};
use zbus::proxy;
use zbus::zvariant::{OwnedObjectPath, OwnedValue, Value};
use zbus::Connection;

use agent::BluetoothAgent;

// ── D-Bus interface constants ──────────────────────────────────────────────────

const ADAPTER_INTERFACE: &str = "org.bluez.Adapter1";
const DEVICE_INTERFACE: &str = "org.bluez.Device1";

// ── Types ──────────────────────────────────────────────────────────────────────

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Adapter {
    pub path: String,
    pub name: String,
    pub address: String,
    pub powered: bool,
    pub discoverable: bool,
    pub discovering: bool,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Device {
    pub path: String,
    pub name: String,
    pub address: String,
    pub paired: bool,
    pub trusted: bool,
    pub connected: bool,
    pub adapter: String,
}

// ── D-Bus proxies ──────────────────────────────────────────────────────────────

#[proxy(
    interface = "org.freedesktop.DBus.ObjectManager",
    default_service = "org.bluez",
    default_path = "/"
)]
trait ObjectManager {
    fn get_managed_objects(
        &self,
    ) -> zbus::Result<HashMap<OwnedObjectPath, HashMap<String, HashMap<String, OwnedValue>>>>;
}

#[proxy(
    interface = "org.bluez.AgentManager1",
    default_service = "org.bluez",
    default_path = "/org/bluez"
)]
trait AgentManager1 {
    fn register_agent(
        &self,
        agent: zbus::zvariant::ObjectPath<'_>,
        capability: &str,
    ) -> zbus::Result<()>;
    fn unregister_agent(&self, agent: zbus::zvariant::ObjectPath<'_>) -> zbus::Result<()>;
    fn request_default_agent(
        &self,
        agent: zbus::zvariant::ObjectPath<'_>,
    ) -> zbus::Result<()>;
}

#[proxy(
    interface = "org.bluez.Adapter1",
    default_service = "org.bluez"
)]
trait Adapter1 {
    fn start_discovery(&self) -> zbus::Result<()>;
    fn stop_discovery(&self) -> zbus::Result<()>;
    fn remove_device(&self, device: zbus::zvariant::ObjectPath<'_>) -> zbus::Result<()>;

    #[zbus(property)]
    fn discoverable(&self) -> zbus::Result<bool>;
    #[zbus(property)]
    fn set_discoverable(&self, value: bool) -> zbus::Result<()>;
}

#[proxy(
    interface = "org.bluez.Device1",
    default_service = "org.bluez"
)]
trait Device1 {
    fn connect(&self) -> zbus::Result<()>;
    fn pair(&self) -> zbus::Result<()>;

    #[zbus(property)]
    fn trusted(&self) -> zbus::Result<bool>;
    #[zbus(property)]
    fn set_trusted(&self, value: bool) -> zbus::Result<()>;
}

// ── BluetoothManager ──────────────────────────────────────────────────────────

/// Singleton-like manager that owns the D-Bus connection and the BlueZ agent.
pub struct BluetoothManager {
    connection: Connection,
}

impl BluetoothManager {
    /// Creates a new manager: connects to the system bus, registers the auto-pair agent.
    pub async fn new() -> Result<Self> {
        let connection = Connection::system()
            .await
            .context("failed to connect to D-Bus system bus")?;

        let agent_path = "/org/bluez/AutoPairAgent";
        tracing::info!("Bluetooth Agent: Registering agent at {}", agent_path);

        connection
            .object_server()
            .at(agent_path, BluetoothAgent)
            .await
            .context("failed to export Bluetooth agent object")?;

        let agent_manager = AgentManager1Proxy::new(&connection)
            .await
            .context("failed to get AgentManager1 proxy")?;

        let path = zbus::zvariant::ObjectPath::try_from(agent_path)
            .context("invalid agent D-Bus path")?;

        agent_manager
            .register_agent(path.as_ref(), "NoInputNoOutput")
            .await
            .context("failed to register Bluetooth agent")?;

        agent_manager
            .request_default_agent(path.as_ref())
            .await
            .context("failed to set agent as default")?;

        tracing::info!("Bluetooth Agent: Successfully registered and set as default agent");

        Ok(Self { connection })
    }

    /// Unregisters the agent and closes the D-Bus connection.
    pub async fn close(&self) {
        let agent_path = "/org/bluez/AutoPairAgent";
        if let Ok(proxy) = AgentManager1Proxy::new(&self.connection).await {
            if let Ok(path) = zbus::zvariant::ObjectPath::try_from(agent_path) {
                let _ = proxy.unregister_agent(path.as_ref()).await;
            }
        }
        tracing::info!("Bluetooth Agent: Unregistered");
    }

    // ── Internal helpers ───────────────────────────────────────────────────

    async fn get_managed_objects(
        &self,
    ) -> Result<HashMap<OwnedObjectPath, HashMap<String, HashMap<String, OwnedValue>>>> {
        let proxy = ObjectManagerProxy::new(&self.connection)
            .await
            .context("failed to create ObjectManager proxy")?;
        proxy
            .get_managed_objects()
            .await
            .context("failed to get managed objects from BlueZ")
    }

    // ── Public API ─────────────────────────────────────────────────────────

    pub async fn get_adapters(&self) -> Result<Vec<Adapter>> {
        let objects = self.get_managed_objects().await?;
        let adapters = objects
            .iter()
            .filter_map(|(path, ifaces)| {
                let props = ifaces.get(ADAPTER_INTERFACE)?;
                Some(Adapter {
                    path: path.as_str().to_string(),
                    name: prop_str(props, "Name"),
                    address: prop_str(props, "Address"),
                    powered: prop_bool(props, "Powered"),
                    discoverable: prop_bool(props, "Discoverable"),
                    discovering: prop_bool(props, "Discovering"),
                })
            })
            .collect();
        Ok(adapters)
    }

    pub async fn get_devices(&self, adapter_path: &str) -> Result<Vec<Device>> {
        let objects = self.get_managed_objects().await?;
        let prefix = format!("{}/", adapter_path);
        let devices = objects
            .iter()
            .filter_map(|(path, ifaces)| {
                let path_str = path.as_str();
                if !path_str.starts_with(&prefix) {
                    return None;
                }
                let props = ifaces.get(DEVICE_INTERFACE)?;
                Some(Device {
                    path: path_str.to_string(),
                    name: prop_str(props, "Name"),
                    address: prop_str(props, "Address"),
                    paired: prop_bool(props, "Paired"),
                    trusted: prop_bool(props, "Trusted"),
                    connected: prop_bool(props, "Connected"),
                    adapter: adapter_path.to_string(),
                })
            })
            .collect();
        Ok(devices)
    }

    pub async fn get_trusted_devices(&self, adapter_path: &str) -> Result<Vec<Device>> {
        Ok(self
            .get_devices(adapter_path)
            .await?
            .into_iter()
            .filter(|d| d.trusted)
            .collect())
    }

    pub async fn get_connected_devices(&self, adapter_path: &str) -> Result<Vec<Device>> {
        Ok(self
            .get_devices(adapter_path)
            .await?
            .into_iter()
            .filter(|d| d.connected)
            .collect())
    }

    pub async fn connect_device(&self, adapter_path: &str, mac: &str) -> Result<()> {
        let device_path = mac_to_device_path(adapter_path, mac);
        let proxy = Device1Proxy::builder(&self.connection)
            .path(device_path.as_str())
            .context("invalid device path")?
            .build()
            .await
            .context("failed to build Device1 proxy")?;

        tokio::time::timeout(Duration::from_millis(12_500), proxy.connect())
            .await
            .context("connect timed out")?
            .with_context(|| format!("failed to connect to device {}", mac))
    }

    pub async fn trust_device(&self, adapter_path: &str, mac: &str) -> Result<()> {
        let device_path = mac_to_device_path(adapter_path, mac);
        let proxy = Device1Proxy::builder(&self.connection)
            .path(device_path.as_str())
            .context("invalid device path")?
            .build()
            .await
            .context("failed to build Device1 proxy")?;

        proxy
            .set_trusted(true)
            .await
            .with_context(|| format!("failed to trust device {}", mac))
    }

    pub async fn pair_device(&self, adapter_path: &str, mac: &str) -> Result<()> {
        let device_path = mac_to_device_path(adapter_path, mac);
        let proxy = Device1Proxy::builder(&self.connection)
            .path(device_path.as_str())
            .context("invalid device path")?
            .build()
            .await
            .context("failed to build Device1 proxy")?;

        proxy
            .pair()
            .await
            .with_context(|| format!("failed to pair with device {}", mac))
    }

    pub async fn remove_device(&self, adapter_path: &str, mac: &str) -> Result<()> {
        let device_path = mac_to_device_path(adapter_path, mac);
        let proxy = Adapter1Proxy::builder(&self.connection)
            .path(adapter_path)
            .context("invalid adapter path")?
            .build()
            .await
            .context("failed to build Adapter1 proxy")?;

        let obj_path = zbus::zvariant::ObjectPath::try_from(device_path.as_str())
            .context("invalid device D-Bus path")?;

        proxy
            .remove_device(obj_path)
            .await
            .with_context(|| format!("failed to remove device {}", mac))
    }

    pub async fn set_discoverable(&self, adapter_path: &str, enable: bool) -> Result<()> {
        let proxy = Adapter1Proxy::builder(&self.connection)
            .path(adapter_path)
            .context("invalid adapter path")?
            .build()
            .await
            .context("failed to build Adapter1 proxy")?;

        proxy
            .set_discoverable(enable)
            .await
            .context("failed to set discoverable")
    }

    pub async fn set_discovering(&self, adapter_path: &str, enable: bool) -> Result<()> {
        let proxy = Adapter1Proxy::builder(&self.connection)
            .path(adapter_path)
            .context("invalid adapter path")?
            .build()
            .await
            .context("failed to build Adapter1 proxy")?;

        if enable {
            proxy.start_discovery().await.context("failed to start discovery")
        } else {
            proxy.stop_discovery().await.context("failed to stop discovery")
        }
    }

    /// Resolves an adapter MAC address to its D-Bus object path.
    pub async fn get_adapter_path_by_mac(&self, mac: &str) -> Result<String> {
        self.get_adapters()
            .await?
            .into_iter()
            .find(|a| a.address == mac)
            .map(|a| a.path)
            .ok_or_else(|| anyhow!("adapter with MAC address {} not found", mac))
    }
}

// ── Helpers ────────────────────────────────────────────────────────────────────

fn mac_to_device_path(adapter_path: &str, mac: &str) -> String {
    format!("{}/dev_{}", adapter_path, mac.replace(':', "_"))
}

fn prop_str(props: &HashMap<String, OwnedValue>, key: &str) -> String {
    props
        .get(key)
        .and_then(|v| match v.deref() {
            Value::Str(s) => Some(s.as_str().to_string()),
            _ => None,
        })
        .unwrap_or_default()
}

fn prop_bool(props: &HashMap<String, OwnedValue>, key: &str) -> bool {
    props
        .get(key)
        .and_then(|v| match v.deref() {
            Value::Bool(b) => Some(*b),
            _ => None,
        })
        .unwrap_or_default()
}
