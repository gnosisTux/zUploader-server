package internal

import (
	"database/sql"
	"time"

	_ "modernc.org/sqlite"
)

var DB *sql.DB

func InitDB(path string) {
	var err error
	DB, err = sql.Open("sqlite", path)
	if err != nil {
		Error.Panicf("[FATAL] failed to open database: %v", err)
	}

	var exists int
	err = DB.QueryRow(`SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='audit'`).Scan(&exists)
	if err != nil {
		Error.Panicf("[FATAL] failed to check database: %v", err)
	}

	if exists == 0 {
		Error.Printf("[DB] audit.db not found, creating...")
	} else {
		Error.Printf("[DB] audit.db found at %s", path)
	}

	_, err = DB.Exec(`
		CREATE TABLE IF NOT EXISTS audit (
			id         INTEGER PRIMARY KEY AUTOINCREMENT,
			timestamp  TEXT NOT NULL,
			action     TEXT NOT NULL,
			ip         TEXT NOT NULL,
			file       TEXT,
			size_bytes INTEGER,
			ua         TEXT,
			mode       TEXT,
			reason     TEXT
		);
		CREATE INDEX IF NOT EXISTS idx_timestamp ON audit(timestamp);
		CREATE INDEX IF NOT EXISTS idx_ip        ON audit(ip);
		CREATE INDEX IF NOT EXISTS idx_file      ON audit(file);
	`)
	if err != nil {
		Error.Panicf("[FATAL] failed to create audit table: %v", err)
	}

	if exists == 0 {
		Error.Printf("[DB] audit database created successfully at %s", path)
	}
}

func AuditLog(action, ip, file string, sizeBytes int64, ua, mode, reason string) {
	ts := time.Now().UTC().Format(time.RFC3339)
	_, err := DB.Exec(`
		INSERT INTO audit (timestamp, action, ip, file, size_bytes, ua, mode, reason)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		ts, action, ip, file, sizeBytes, ua, mode, reason,
	)
	if err != nil {
		Error.Printf("[ERROR] failed to write audit log: %v", err)
		return
	}
	Error.Printf("[DB] audit record saved: action=%s ip=%s file=%s size=%d ua=\"%s\" mode=%s reason=%s",
		action, ip, file, sizeBytes, ua, mode, reason)
}

func PurgeOldRecords() {
	res, err := DB.Exec(`
		DELETE FROM audit WHERE timestamp < datetime('now', '-12 months')
	`)
	if err != nil {
		Error.Printf("[ERROR] failed to purge old audit records: %v", err)
		return
	}
	n, _ := res.RowsAffected()
	if n > 0 {
		Error.Printf("[PURGE] deleted %d audit records older than 12 months", n)
	}
}
