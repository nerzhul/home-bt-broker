mod bluetooth;
mod database;
mod handlers;
mod pipewire;
mod utils;

use std::sync::Arc;

use axum::{
    middleware,
    routing::{delete, get, patch, post},
    Router,
};
use sqlx::SqlitePool;
use tower_http::{cors::CorsLayer, services::ServeFile, trace::TraceLayer};
use tracing::info;

use bluetooth::{reconnector, BluetoothManager};
use handlers::{aliases as alias_handlers, bluetooth as bt_handlers};

/// Shared application state injected into every handler.
#[derive(Clone)]
pub struct AppState {
    pub pool: SqlitePool,
    pub bluetooth: Arc<BluetoothManager>,
}

#[tokio::main]
async fn main() -> anyhow::Result<()> {
    // Logging
    tracing_subscriber::fmt()
        .with_env_filter(
            tracing_subscriber::EnvFilter::try_from_default_env()
                .unwrap_or_else(|_| "info".into()),
        )
        .init();

    // Database
    let pool = database::init_db().await?;
    database::run_migrations(&pool).await?;

    // Bluetooth
    let bt = Arc::new(BluetoothManager::new().await?);

    match bt.get_adapters().await {
        Ok(adapters) if adapters.is_empty() => info!("No Bluetooth adapters found."),
        Ok(adapters) => {
            info!("Bluetooth adapters detected:");
            for a in &adapters {
                info!(
                    "- Name: {}, Address: {}, Powered: {}, Discoverable: {}, Discovering: {}",
                    a.name, a.address, a.powered, a.discoverable, a.discovering
                );
            }
        }
        Err(e) => info!("Could not list Bluetooth adapters: {}", e),
    }

    // PipeWire – fatal if combined_output not found
    match pipewire::check_combined_output() {
        Ok(node) => info!(
            "Found combined_output node: ID={}, Name={}",
            node.id, node.node_name
        ),
        Err(e) => {
            tracing::error!("PipeWire: {}", e);
            std::process::exit(1);
        }
    }

    // Reconnect loop
    reconnector::start_reconnect_loop(
        Arc::clone(&bt),
        reconnector::get_reconnect_interval(10),
    );

    let state = AppState { pool, bluetooth: bt };

    // ── Router ──────────────────────────────────────────────────────────────

    let token_routes = Router::new()
        .route(
            "/",
            post(handlers::create_token).get(handlers::get_tokens),
        )
        .route(
            "/:username",
            get(handlers::get_token).delete(handlers::delete_token),
        );

    let bt_routes = Router::new()
        .route("/adapters", get(bt_handlers::get_adapters))
        .route(
            "/adapters/:adapter/discoverable",
            patch(bt_handlers::set_discoverable),
        )
        .route(
            "/adapters/:adapter/discovering",
            patch(bt_handlers::set_discovering),
        )
        .route("/adapters/:adapter/devices", get(bt_handlers::get_devices))
        .route(
            "/adapters/:adapter/devices/trusted",
            get(bt_handlers::get_trusted_devices),
        )
        .route(
            "/adapters/:adapter/devices/connected",
            get(bt_handlers::get_connected_devices),
        )
        .route(
            "/adapters/:adapter/devices/:mac/pair",
            post(bt_handlers::pair_device),
        )
        .route(
            "/adapters/:adapter/devices/:mac/connect",
            post(bt_handlers::connect_device),
        )
        .route(
            "/adapters/:adapter/devices/:mac/trust",
            post(bt_handlers::trust_device),
        )
        .route(
            "/adapters/:adapter/devices/:mac",
            delete(bt_handlers::remove_device),
        )
        .route("/aliases", get(alias_handlers::get_all_aliases))
        .route(
            "/aliases/:mac",
            get(alias_handlers::get_alias)
                .put(alias_handlers::set_alias)
                .delete(alias_handlers::delete_alias),
        );

    let api = Router::new()
        .nest("/tokens", token_routes)
        .nest("/bluetooth", bt_routes)
        .route_layer(middleware::from_fn_with_state(
            state.clone(),
            handlers::auth_middleware,
        ));

    let app = Router::new()
        .route_service("/", ServeFile::new("static/index.html"))
        .route("/readyz", get(handlers::readiness))
        .route("/livez", get(handlers::liveness))
        .nest("/api/v1", api)
        .layer(CorsLayer::permissive())
        .layer(TraceLayer::new_for_http())
        .with_state(state);

    let port = std::env::var("PORT").unwrap_or_else(|_| "8080".to_string());
    let addr = format!("0.0.0.0:{}", port);
    info!("Starting server on port {}", port);

    let listener = tokio::net::TcpListener::bind(&addr).await?;
    axum::serve(listener, app).await?;

    Ok(())
}
