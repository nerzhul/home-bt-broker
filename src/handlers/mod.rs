pub mod aliases;
pub mod bluetooth;

use axum::{
    extract::{Path, Request, State},
    http::{header, StatusCode},
    middleware::Next,
    response::{IntoResponse, Response},
    Json,
};
use base64::Engine;
use chrono::NaiveDateTime;
use serde::{Deserialize, Serialize};

use crate::AppState;

// ── Shared helpers ─────────────────────────────────────────────────────────────

pub(crate) fn err_json(status: StatusCode, msg: &str) -> Response {
    (status, Json(serde_json::json!({ "error": msg }))).into_response()
}

// ── Auth middleware ────────────────────────────────────────────────────────────

pub async fn auth_middleware(
    State(state): State<AppState>,
    request: Request,
    next: Next,
) -> Response {
    let auth_header = request
        .headers()
        .get(header::AUTHORIZATION)
        .and_then(|v| v.to_str().ok());

    let (username, password) = match auth_header.and_then(parse_basic_auth) {
        Some(creds) => creds,
        None => {
            let mut resp =
                err_json(StatusCode::UNAUTHORIZED, "missing or invalid basic auth");
            resp.headers_mut().insert(
                header::WWW_AUTHENTICATE,
                r#"Basic realm="Restricted""#.parse().unwrap(),
            );
            return resp;
        }
    };

    let stored: Option<String> =
        sqlx::query_scalar("SELECT token FROM user_tokens WHERE username = ?")
            .bind(&username)
            .fetch_optional(&state.pool)
            .await
            .ok()
            .flatten();

    match stored {
        Some(token) if token == password => next.run(request).await,
        _ => {
            let mut resp = err_json(StatusCode::UNAUTHORIZED, "invalid credentials");
            resp.headers_mut().insert(
                header::WWW_AUTHENTICATE,
                r#"Basic realm="Restricted""#.parse().unwrap(),
            );
            resp
        }
    }
}

fn parse_basic_auth(header: &str) -> Option<(String, String)> {
    let encoded = header.strip_prefix("Basic ")?;
    let decoded = base64::engine::general_purpose::STANDARD
        .decode(encoded)
        .ok()?;
    let s = String::from_utf8(decoded).ok()?;
    let (user, pass) = s.split_once(':')?;
    Some((user.to_string(), pass.to_string()))
}

// ── Health ─────────────────────────────────────────────────────────────────────

pub async fn readiness(State(state): State<AppState>) -> Response {
    match sqlx::query("SELECT 1").execute(&state.pool).await {
        Ok(_) => Json(serde_json::json!({ "status": "ready" })).into_response(),
        Err(_) => (
            StatusCode::SERVICE_UNAVAILABLE,
            Json(serde_json::json!({
                "status": "not ready",
                "error": "database connection failed"
            })),
        )
            .into_response(),
    }
}

pub async fn liveness() -> impl IntoResponse {
    Json(serde_json::json!({ "status": "alive" }))
}

// ── Token types ────────────────────────────────────────────────────────────────

#[derive(Debug, Serialize)]
pub struct Token {
    pub username: String,
    pub token: String,
    pub created_at: NaiveDateTime,
}

#[derive(Debug, Deserialize)]
pub struct CreateTokenRequest {
    pub username: String,
    pub token: String,
}

// ── Token handlers ─────────────────────────────────────────────────────────────

pub async fn create_token(
    State(state): State<AppState>,
    Json(req): Json<CreateTokenRequest>,
) -> Response {
    if req.username.is_empty() || req.token.is_empty() {
        return err_json(StatusCode::BAD_REQUEST, "username and token are required");
    }

    let count: i64 =
        sqlx::query_scalar("SELECT COUNT(*) FROM user_tokens WHERE username = ?")
            .bind(&req.username)
            .fetch_one(&state.pool)
            .await
            .unwrap_or(0);

    if count > 0 {
        return err_json(StatusCode::CONFLICT, "username already exists");
    }

    match sqlx::query(
        "INSERT INTO user_tokens (username, token, created_at) VALUES (?, ?, datetime('now'))",
    )
    .bind(&req.username)
    .bind(&req.token)
    .execute(&state.pool)
    .await
    {
        Ok(_) => (
            StatusCode::CREATED,
            Json(serde_json::json!({ "message": "token created successfully" })),
        )
            .into_response(),
        Err(_) => err_json(StatusCode::INTERNAL_SERVER_ERROR, "failed to create token"),
    }
}

pub async fn get_tokens(State(state): State<AppState>) -> Response {
    let rows: Vec<(String, String, NaiveDateTime)> = match sqlx::query_as(
        "SELECT username, token, created_at FROM user_tokens ORDER BY created_at DESC",
    )
    .fetch_all(&state.pool)
    .await
    {
        Ok(r) => r,
        Err(_) => return err_json(StatusCode::INTERNAL_SERVER_ERROR, "database error"),
    };

    let tokens: Vec<Token> = rows
        .into_iter()
        .map(|(username, token, created_at)| Token {
            username,
            token,
            created_at,
        })
        .collect();

    Json(tokens).into_response()
}

pub async fn get_token(
    State(state): State<AppState>,
    Path(username): Path<String>,
) -> Response {
    let row: Option<(String, String, NaiveDateTime)> = match sqlx::query_as(
        "SELECT username, token, created_at FROM user_tokens WHERE username = ?",
    )
    .bind(&username)
    .fetch_optional(&state.pool)
    .await
    {
        Ok(r) => r,
        Err(_) => return err_json(StatusCode::INTERNAL_SERVER_ERROR, "database error"),
    };

    match row {
        Some((username, token, created_at)) => {
            Json(Token { username, token, created_at }).into_response()
        }
        None => err_json(StatusCode::NOT_FOUND, "token not found"),
    }
}

pub async fn delete_token(
    State(state): State<AppState>,
    Path(username): Path<String>,
) -> Response {
    match sqlx::query("DELETE FROM user_tokens WHERE username = ?")
        .bind(&username)
        .execute(&state.pool)
        .await
    {
        Ok(r) if r.rows_affected() > 0 => {
            Json(serde_json::json!({ "message": "token deleted successfully" }))
                .into_response()
        }
        Ok(_) => err_json(StatusCode::NOT_FOUND, "token not found"),
        Err(_) => err_json(StatusCode::INTERNAL_SERVER_ERROR, "database error"),
    }
}
