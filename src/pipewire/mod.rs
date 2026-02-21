use anyhow::{anyhow, Context, Result};
use serde::Deserialize;
use std::process::Command;

#[derive(Debug, Deserialize)]
struct PipeWireNode {
    id: u32,
    #[serde(rename = "type")]
    node_type: String,
    info: Option<serde_json::Value>,
}

pub struct PipeWireOutput {
    pub id: u32,
    pub node_name: String,
}

/// Runs `pw-dump` and looks for a node named `combined_output`.
pub fn check_combined_output() -> Result<PipeWireOutput> {
    tracing::info!("PipeWire: Checking for combined_output node");

    let out = Command::new("pw-dump")
        .output()
        .context("failed to execute pw-dump")?;

    let nodes: Vec<PipeWireNode> =
        serde_json::from_slice(&out.stdout).context("failed to parse pw-dump output")?;

    for node in nodes {
        if node.node_type != "PipeWire:Interface:Node" {
            continue;
        }
        let info = match &node.info {
            Some(v) => v,
            None => continue,
        };
        let props = match info.get("props").and_then(|p| p.as_object()) {
            Some(p) => p,
            None => continue,
        };
        let node_name = match props.get("node.name").and_then(|n| n.as_str()) {
            Some(n) => n,
            None => continue,
        };
        if node_name == "combined_output" {
            tracing::info!("PipeWire: Found combined_output node with ID {}", node.id);
            return Ok(PipeWireOutput {
                id: node.id,
                node_name: node_name.to_string(),
            });
        }
    }

    Err(anyhow!("combined_output node not found"))
}
