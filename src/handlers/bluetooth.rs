use axum::{
    extract::{Path, State},
    http::StatusCode,
    response::{IntoResponse, Response},
    Json,
};
use serde::Deserialize;

use super::err_json;
use crate::AppState;

pub async fn get_adapters(State(state): State<AppState>) -> Response {
    match state.bluetooth.get_adapters().await {
        Ok(adapters) => Json(serde_json::json!({ "adapters": adapters })).into_response(),
        Err(e) => err_json(
            StatusCode::INTERNAL_SERVER_ERROR,
            &format!("failed to get adapters: {}", e),
        ),
    }
}

pub async fn get_devices(
    State(state): State<AppState>,
    Path(adapter): Path<String>,
) -> Response {
    let path = match state.bluetooth.get_adapter_path_by_mac(&adapter).await {
        Ok(p) => p,
        Err(e) => return err_json(StatusCode::NOT_FOUND, &format!("adapter not found: {}", e)),
    };
    match state.bluetooth.get_devices(&path).await {
        Ok(devices) => Json(serde_json::json!({ "devices": devices })).into_response(),
        Err(e) => err_json(
            StatusCode::INTERNAL_SERVER_ERROR,
            &format!("failed to get devices: {}", e),
        ),
    }
}

pub async fn get_trusted_devices(
    State(state): State<AppState>,
    Path(adapter): Path<String>,
) -> Response {
    let path = match state.bluetooth.get_adapter_path_by_mac(&adapter).await {
        Ok(p) => p,
        Err(e) => return err_json(StatusCode::NOT_FOUND, &format!("adapter not found: {}", e)),
    };
    match state.bluetooth.get_trusted_devices(&path).await {
        Ok(devices) => Json(serde_json::json!({ "trusted_devices": devices })).into_response(),
        Err(e) => err_json(
            StatusCode::INTERNAL_SERVER_ERROR,
            &format!("failed to get trusted devices: {}", e),
        ),
    }
}

pub async fn get_connected_devices(
    State(state): State<AppState>,
    Path(adapter): Path<String>,
) -> Response {
    let path = match state.bluetooth.get_adapter_path_by_mac(&adapter).await {
        Ok(p) => p,
        Err(e) => return err_json(StatusCode::NOT_FOUND, &format!("adapter not found: {}", e)),
    };
    match state.bluetooth.get_connected_devices(&path).await {
        Ok(devices) => {
            Json(serde_json::json!({ "connected_devices": devices })).into_response()
        }
        Err(e) => err_json(
            StatusCode::INTERNAL_SERVER_ERROR,
            &format!("failed to get connected devices: {}", e),
        ),
    }
}

#[derive(Deserialize)]
pub struct EnableRequest {
    pub enable: bool,
}

pub async fn set_discoverable(
    State(state): State<AppState>,
    Path(adapter): Path<String>,
    Json(req): Json<EnableRequest>,
) -> Response {
    let path = match state.bluetooth.get_adapter_path_by_mac(&adapter).await {
        Ok(p) => p,
        Err(e) => return err_json(StatusCode::NOT_FOUND, &format!("adapter not found: {}", e)),
    };
    match state.bluetooth.set_discoverable(&path, req.enable).await {
        Ok(_) => Json(serde_json::json!({ "message": "discoverable updated" })).into_response(),
        Err(e) => err_json(
            StatusCode::INTERNAL_SERVER_ERROR,
            &format!("failed to set discoverable: {}", e),
        ),
    }
}

pub async fn set_discovering(
    State(state): State<AppState>,
    Path(adapter): Path<String>,
    Json(req): Json<EnableRequest>,
) -> Response {
    let path = match state.bluetooth.get_adapter_path_by_mac(&adapter).await {
        Ok(p) => p,
        Err(e) => return err_json(StatusCode::NOT_FOUND, &format!("adapter not found: {}", e)),
    };
    match state.bluetooth.set_discovering(&path, req.enable).await {
        Ok(_) => Json(serde_json::json!({ "message": "discovering updated" })).into_response(),
        Err(e) => err_json(
            StatusCode::INTERNAL_SERVER_ERROR,
            &format!("failed to set discovering: {}", e),
        ),
    }
}

pub async fn connect_device(
    State(state): State<AppState>,
    Path((adapter, mac)): Path<(String, String)>,
) -> Response {
    let path = match state.bluetooth.get_adapter_path_by_mac(&adapter).await {
        Ok(p) => p,
        Err(e) => return err_json(StatusCode::NOT_FOUND, &format!("adapter not found: {}", e)),
    };
    match state.bluetooth.connect_device(&path, &mac).await {
        Ok(_) => Json(serde_json::json!({
            "message": "device connection initiated successfully"
        }))
        .into_response(),
        Err(e) => err_json(
            StatusCode::INTERNAL_SERVER_ERROR,
            &format!("failed to connect device: {}", e),
        ),
    }
}

pub async fn trust_device(
    State(state): State<AppState>,
    Path((adapter, mac)): Path<(String, String)>,
) -> Response {
    let path = match state.bluetooth.get_adapter_path_by_mac(&adapter).await {
        Ok(p) => p,
        Err(e) => return err_json(StatusCode::NOT_FOUND, &format!("adapter not found: {}", e)),
    };
    match state.bluetooth.trust_device(&path, &mac).await {
        Ok(_) => {
            Json(serde_json::json!({ "message": "device trusted successfully" })).into_response()
        }
        Err(e) => err_json(
            StatusCode::INTERNAL_SERVER_ERROR,
            &format!("failed to trust device: {}", e),
        ),
    }
}

pub async fn pair_device(
    State(state): State<AppState>,
    Path((adapter, mac)): Path<(String, String)>,
) -> Response {
    let path = match state.bluetooth.get_adapter_path_by_mac(&adapter).await {
        Ok(p) => p,
        Err(e) => return err_json(StatusCode::NOT_FOUND, &format!("adapter not found: {}", e)),
    };
    match state.bluetooth.pair_device(&path, &mac).await {
        Ok(_) => Json(serde_json::json!({
            "message": "device pairing initiated successfully"
        }))
        .into_response(),
        Err(e) => err_json(
            StatusCode::INTERNAL_SERVER_ERROR,
            &format!("failed to pair device: {}", e),
        ),
    }
}

pub async fn remove_device(
    State(state): State<AppState>,
    Path((adapter, mac)): Path<(String, String)>,
) -> Response {
    let path = match state.bluetooth.get_adapter_path_by_mac(&adapter).await {
        Ok(p) => p,
        Err(e) => return err_json(StatusCode::NOT_FOUND, &format!("adapter not found: {}", e)),
    };
    match state.bluetooth.remove_device(&path, &mac).await {
        Ok(_) => {
            Json(serde_json::json!({ "message": "device removed successfully" })).into_response()
        }
        Err(e) => err_json(
            StatusCode::INTERNAL_SERVER_ERROR,
            &format!("failed to remove device: {}", e),
        ),
    }
}
