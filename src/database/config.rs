use anyhow::{anyhow, Context, Result};
use serde::{Deserialize, Serialize};
use sqlx::SqlitePool;

#[derive(Debug, Serialize, Deserialize)]
pub struct Config {
    pub key: String,
    pub value: String,
}

pub async fn get_config(pool: &SqlitePool, key: &str) -> Result<Config> {
    let row: Option<(String, String)> =
        sqlx::query_as("SELECT config_key, config_value FROM config WHERE config_key = ?")
            .bind(key)
            .fetch_optional(pool)
            .await
            .context("failed to query config")?;

    row.map(|(k, v)| Config { key: k, value: v })
        .ok_or_else(|| anyhow!("config key '{}' not found", key))
}

pub async fn set_config(pool: &SqlitePool, key: &str, value: &str) -> Result<()> {
    sqlx::query(
        "INSERT OR REPLACE INTO config (config_key, config_value) VALUES (?, ?)",
    )
    .bind(key)
    .bind(value)
    .execute(pool)
    .await
    .context("failed to set config")?;
    Ok(())
}

pub async fn delete_config(pool: &SqlitePool, key: &str) -> Result<()> {
    let result = sqlx::query("DELETE FROM config WHERE config_key = ?")
        .bind(key)
        .execute(pool)
        .await
        .context("failed to delete config")?;
    if result.rows_affected() == 0 {
        return Err(anyhow!("config key '{}' not found", key));
    }
    Ok(())
}
