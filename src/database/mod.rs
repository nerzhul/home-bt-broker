pub mod aliases;
pub mod config;

use anyhow::{Context, Result};
use sqlx::SqlitePool;

/// Opens (or creates) the SQLite database pointed to by `DATABASE_PATH` env var.
pub async fn init_db() -> Result<SqlitePool> {
    let db_path =
        std::env::var("DATABASE_PATH").unwrap_or_else(|_| "./data.db".to_string());

    // Create parent directories if they don't exist.
    if let Some(parent) = std::path::Path::new(&db_path).parent() {
        if parent != std::path::Path::new("") && parent != std::path::Path::new(".") {
            std::fs::create_dir_all(parent)
                .context("failed to create database directory")?;
        }
    }

    let url = format!("sqlite://{}?mode=rwc", db_path);
    let pool = SqlitePool::connect(&url)
        .await
        .context("failed to open database")?;

    Ok(pool)
}

/// Runs embedded SQL migrations from the `migrations/` directory.
pub async fn run_migrations(pool: &SqlitePool) -> Result<()> {
    sqlx::migrate!("./migrations")
        .run(pool)
        .await
        .context("failed to run database migrations")
}
