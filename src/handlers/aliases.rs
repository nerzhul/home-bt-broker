use axum::{
    extract::{Path, State},
    http::StatusCode,
    response::{IntoResponse, Response},
    Json,
};
use serde::Deserialize;

use super::err_json;
use crate::{database::aliases, utils, AppState};

pub async fn get_all_aliases(State(state): State<AppState>) -> Response {
    match aliases::get_all_aliases(&state.pool).await {
        Ok(map) => Json(serde_json::json!({ "aliases": map })).into_response(),
        Err(e) => err_json(StatusCode::INTERNAL_SERVER_ERROR, &e.to_string()),
    }
}

pub async fn get_alias(
    State(state): State<AppState>,
    Path(mac): Path<String>,
) -> Response {
    if !utils::is_valid_mac(&mac) {
        return err_json(StatusCode::BAD_REQUEST, "invalid MAC address format");
    }
    let mac = utils::normalize_mac(&mac);
    match aliases::get_alias(&state.pool, &mac).await {
        Ok(Some(alias)) => {
            Json(serde_json::json!({ "mac": mac, "alias": alias })).into_response()
        }
        Ok(None) => err_json(StatusCode::NOT_FOUND, "alias not found"),
        Err(e) => err_json(StatusCode::INTERNAL_SERVER_ERROR, &e.to_string()),
    }
}

#[derive(Deserialize)]
pub struct SetAliasRequest {
    pub alias: String,
}

pub async fn set_alias(
    State(state): State<AppState>,
    Path(mac): Path<String>,
    Json(req): Json<SetAliasRequest>,
) -> Response {
    if !utils::is_valid_mac(&mac) {
        return err_json(StatusCode::BAD_REQUEST, "invalid MAC address format");
    }
    let mac = utils::normalize_mac(&mac);
    if req.alias.is_empty() {
        return err_json(StatusCode::BAD_REQUEST, "alias is required");
    }
    match aliases::set_alias(&state.pool, &mac, &req.alias).await {
        Ok(_) => Json(serde_json::json!({
            "message": "alias set",
            "mac": mac,
            "alias": req.alias,
        }))
        .into_response(),
        Err(e) => err_json(StatusCode::INTERNAL_SERVER_ERROR, &e.to_string()),
    }
}

pub async fn delete_alias(
    State(state): State<AppState>,
    Path(mac): Path<String>,
) -> Response {
    if !utils::is_valid_mac(&mac) {
        return err_json(StatusCode::BAD_REQUEST, "invalid MAC address format");
    }
    let mac = utils::normalize_mac(&mac);
    match aliases::delete_alias(&state.pool, &mac).await {
        Ok(true) => {
            Json(serde_json::json!({ "message": "alias deleted", "mac": mac })).into_response()
        }
        Ok(false) => err_json(StatusCode::NOT_FOUND, "alias not found"),
        Err(e) => err_json(StatusCode::INTERNAL_SERVER_ERROR, &e.to_string()),
    }
}
