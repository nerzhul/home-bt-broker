package database

import "database/sql"

// DatabaseInterface defines the interface for database operations
type DatabaseInterface interface {
	Ping() error
	QueryRow(query string, args ...interface{}) *sql.Row
	Query(query string, args ...interface{}) (*sql.Rows, error)
	Exec(query string, args ...interface{}) (sql.Result, error)
}

// Ensure *sql.DB implements the interface
var _ DatabaseInterface = (*sql.DB)(nil)