use anyhow::{anyhow, Context, Result};
use sqlx::SqlitePool;
use std::collections::HashMap;

use crate::utils::{is_valid_mac, normalize_mac};

/// Creates or updates an alias for a device MAC address.
pub async fn set_alias(pool: &SqlitePool, mac: &str, alias: &str) -> Result<()> {
    if !is_valid_mac(mac) {
        return Err(anyhow!("invalid MAC address format"));
    }
    let mac = normalize_mac(mac);
    if alias.is_empty() {
        return Err(anyhow!("alias cannot be empty"));
    }
    sqlx::query(
        "INSERT INTO bluetooth_aliases (mac, alias, updated_at)
         VALUES (?, ?, CURRENT_TIMESTAMP)
         ON CONFLICT(mac) DO UPDATE SET alias = excluded.alias, updated_at = CURRENT_TIMESTAMP",
    )
    .bind(&mac)
    .bind(alias)
    .execute(pool)
    .await
    .context("failed to set alias")?;
    Ok(())
}

/// Returns the alias for a MAC address, or `None` if not found.
pub async fn get_alias(pool: &SqlitePool, mac: &str) -> Result<Option<String>> {
    if !is_valid_mac(mac) {
        return Err(anyhow!("invalid MAC address format"));
    }
    let mac = normalize_mac(mac);
    let row: Option<(String,)> =
        sqlx::query_as("SELECT alias FROM bluetooth_aliases WHERE mac = ?")
            .bind(&mac)
            .fetch_optional(pool)
            .await
            .context("failed to get alias")?;
    Ok(row.map(|(a,)| a))
}

/// Deletes the alias for a MAC address. Returns `true` if a row was deleted.
pub async fn delete_alias(pool: &SqlitePool, mac: &str) -> Result<bool> {
    if !is_valid_mac(mac) {
        return Err(anyhow!("invalid MAC address format"));
    }
    let mac = normalize_mac(mac);
    let result = sqlx::query("DELETE FROM bluetooth_aliases WHERE mac = ?")
        .bind(&mac)
        .execute(pool)
        .await
        .context("failed to delete alias")?;
    Ok(result.rows_affected() > 0)
}

/// Returns all aliases as a `mac → alias` map.
pub async fn get_all_aliases(pool: &SqlitePool) -> Result<HashMap<String, String>> {
    let rows: Vec<(String, String)> =
        sqlx::query_as("SELECT mac, alias FROM bluetooth_aliases")
            .fetch_all(pool)
            .await
            .context("failed to list aliases")?;
    Ok(rows.into_iter().collect())
}
