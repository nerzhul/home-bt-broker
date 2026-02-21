use std::sync::Arc;
use std::time::Duration;

use super::BluetoothManager;

/// Spawns a background task that periodically reconnects paired+trusted devices
/// that are not currently connected.
pub fn start_reconnect_loop(manager: Arc<BluetoothManager>, interval: Duration) {
    tokio::spawn(async move {
        tracing::info!("Reconnect: loop started (interval={:?})", interval);
        let mut ticker = tokio::time::interval(interval);

        loop {
            ticker.tick().await;

            let adapters = match manager.get_adapters().await {
                Ok(a) => a,
                Err(e) => {
                    tracing::warn!("Reconnect: get adapters error: {}", e);
                    continue;
                }
            };

            for adapter in &adapters {
                let devices = match manager.get_devices(&adapter.path).await {
                    Ok(d) => d,
                    Err(e) => {
                        tracing::warn!(
                            "Reconnect: get devices error for {}: {}",
                            adapter.address,
                            e
                        );
                        continue;
                    }
                };

                for device in &devices {
                    if device.paired && device.trusted && !device.connected {
                        match manager.connect_device(&adapter.path, &device.address).await {
                            Ok(_) => tracing::info!(
                                "Reconnect: connect attempt for {} initiated",
                                device.address
                            ),
                            Err(e) => tracing::warn!(
                                "Reconnect: connect {} failed: {}",
                                device.address,
                                e
                            ),
                        }
                    }
                }
            }
        }
    });
}

/// Returns the reconnect interval from `RECONNECT_INTERVAL_SECONDS` env var
/// or the provided default (in seconds).
pub fn get_reconnect_interval(default_seconds: u64) -> Duration {
    std::env::var("RECONNECT_INTERVAL_SECONDS")
        .ok()
        .and_then(|v| v.parse::<u64>().ok())
        .filter(|&n| n > 0)
        .map(Duration::from_secs)
        .unwrap_or(Duration::from_secs(default_seconds))
}
