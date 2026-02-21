use zbus::interface;
use zbus::zvariant::ObjectPath;

/// BlueZ auto-pair agent. Accepts all authentication requests without user interaction.
///
/// How it works:
///  1. Registered at `/org/bluez/AutoPairAgent` with capability `NoInputNoOutput`.
///  2. BlueZ calls these methods during pairing.
///  3. All requests are auto-accepted (PIN "0000", passkey 0, confirmations OK).
pub struct BluetoothAgent;

#[interface(name = "org.bluez.Agent1")]
impl BluetoothAgent {
    async fn release(&self) {
        tracing::info!("Bluetooth Agent: Release called - agent being released");
    }

    async fn request_pin_code(&self, device: ObjectPath<'_>) -> zbus::fdo::Result<String> {
        tracing::info!("Bluetooth Agent: RequestPinCode for {} - providing 0000", device);
        Ok("0000".to_string())
    }

    async fn display_pin_code(
        &self,
        device: ObjectPath<'_>,
        pincode: &str,
    ) -> zbus::fdo::Result<()> {
        tracing::info!("Bluetooth Agent: DisplayPinCode for {} - PIN: {}", device, pincode);
        Ok(())
    }

    async fn request_passkey(&self, device: ObjectPath<'_>) -> zbus::fdo::Result<u32> {
        tracing::info!("Bluetooth Agent: RequestPasskey for {} - providing 0", device);
        Ok(0)
    }

    async fn display_passkey(
        &self,
        device: ObjectPath<'_>,
        passkey: u32,
        entered: u16,
    ) -> zbus::fdo::Result<()> {
        tracing::info!(
            "Bluetooth Agent: DisplayPasskey for {} - passkey: {}, entered: {}",
            device,
            passkey,
            entered
        );
        Ok(())
    }

    async fn request_confirmation(
        &self,
        device: ObjectPath<'_>,
        passkey: u32,
    ) -> zbus::fdo::Result<()> {
        tracing::info!(
            "Bluetooth Agent: RequestConfirmation for {} - passkey: {} - auto-confirming",
            device,
            passkey
        );
        Ok(())
    }

    async fn request_authorization(&self, device: ObjectPath<'_>) -> zbus::fdo::Result<()> {
        tracing::info!("Bluetooth Agent: RequestAuthorization for {} - auto-authorizing", device);
        Ok(())
    }

    async fn authorize_service(
        &self,
        device: ObjectPath<'_>,
        uuid: &str,
    ) -> zbus::fdo::Result<()> {
        tracing::info!(
            "Bluetooth Agent: AuthorizeService for {}, UUID: {} - auto-authorizing",
            device,
            uuid
        );
        Ok(())
    }

    async fn cancel(&self) -> zbus::fdo::Result<()> {
        tracing::info!("Bluetooth Agent: Cancel called - pairing process cancelled");
        Ok(())
    }
}
