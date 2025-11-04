-- Create table for bluetooth device aliases
CREATE TABLE IF NOT EXISTS bluetooth_aliases (
    mac TEXT PRIMARY KEY,
    alias TEXT NOT NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Index for faster lookup by alias (optional)
CREATE INDEX IF NOT EXISTS idx_bluetooth_aliases_alias ON bluetooth_aliases(alias);
