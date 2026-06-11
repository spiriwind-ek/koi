package storage

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type Node struct {
	Key      string
	ParentKey *string
	Name     string
	ObjType  string
	Value    *string
	BlobData []byte
	BlobMeta *string
	CreatedAt int64
	UpdatedAt int64
}

type DB struct {
	conn *sql.DB
}

func NewDB(dbPath string) (*DB, error) {
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("create db dir: %w", err)
	}

	conn, err := sql.Open("sqlite3", dbPath+"?_journal_mode=WAL&_synchronous=NORMAL&_foreign_keys=ON")
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}

	db := &DB{conn: conn}

	schemaPath := filepath.Join(filepath.Dir(dbPath), "..", "sql", "schema.sql")
	if _, err := os.Stat(schemaPath); err == nil {
		schema, err := os.ReadFile(schemaPath)
		if err == nil {
			for _, stmt := range splitSQL(string(schema)) {
				stmt = strings.TrimSpace(stmt)
				if stmt != "" && !strings.HasPrefix(stmt, "--") {
					conn.Exec(stmt)
				}
			}
		}
	}

	return db, nil
}

func (db *DB) Close() error {
	return db.conn.Close()
}

func splitSQL(sql string) []string {
	var stmts []string
	var current strings.Builder
	for _, line := range strings.Split(sql, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "PRAGMA") || strings.HasPrefix(trimmed, "--") {
			continue
		}
		current.WriteString(line)
		current.WriteString("\n")
		if strings.HasSuffix(trimmed, ";") {
			stmts = append(stmts, current.String())
			current.Reset()
		}
	}
	if s := strings.TrimSpace(current.String()); s != "" {
		stmts = append(stmts, s)
	}
	return stmts
}

func (db *DB) GetNode(key string) (*Node, error) {
	var n Node
	err := db.conn.QueryRow(
		`SELECT key, parent_key, name, obj_type, value, blob_data, blob_meta, created_at, updated_at
		 FROM nodes WHERE key = ?`, key,
	).Scan(&n.Key, &n.ParentKey, &n.Name, &n.ObjType, &n.Value, &n.BlobData, &n.BlobMeta, &n.CreatedAt, &n.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &n, nil
}

func (db *DB) CreateNode(n *Node) error {
	now := time.Now().Unix()
	if n.CreatedAt == 0 {
		n.CreatedAt = now
	}
	if n.UpdatedAt == 0 {
		n.UpdatedAt = now
	}
	_, err := db.conn.Exec(
		`INSERT INTO nodes (key, parent_key, name, obj_type, value, blob_data, blob_meta, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		n.Key, n.ParentKey, n.Name, n.ObjType, n.Value, n.BlobData, n.BlobMeta, n.CreatedAt, n.UpdatedAt,
	)
	return err
}

func (db *DB) UpdateNode(n *Node) error {
	n.UpdatedAt = time.Now().Unix()
	_, err := db.conn.Exec(
		`UPDATE nodes SET value = ?, blob_data = ?, blob_meta = ?, updated_at = ? WHERE key = ?`,
		n.Value, n.BlobData, n.BlobMeta, n.UpdatedAt, n.Key,
	)
	return err
}

func (db *DB) DeleteNode(key string) error {
	_, err := db.conn.Exec(`DELETE FROM nodes WHERE key = ?`, key)
	return err
}

func (db *DB) ListChildren(parentKey string) ([]Node, error) {
	rows, err := db.conn.Query(
		`SELECT key, parent_key, name, obj_type, value, created_at, updated_at
		 FROM nodes WHERE parent_key = ? ORDER BY name`, parentKey,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var nodes []Node
	for rows.Next() {
		var n Node
		if err := rows.Scan(&n.Key, &n.ParentKey, &n.Name, &n.ObjType, &n.Value, &n.CreatedAt, &n.UpdatedAt); err != nil {
			return nil, err
		}
		nodes = append(nodes, n)
	}
	return nodes, nil
}

func (db *DB) AutoMkdir(path string) error {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	current := ""

	tx, err := db.conn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, part := range parts {
		if current == "" {
			current = "/" + part
		} else {
			current = current + "/" + part
		}

		var exists bool
		err := tx.QueryRow("SELECT EXISTS(SELECT 1 FROM nodes WHERE key = ?)", current).Scan(&exists)
		if err != nil {
			return err
		}

		if !exists {
			parent := filepath.Dir(current)
			if parent == "." {
				parent = "/"
			}
			now := time.Now().Unix()
			_, err := tx.Exec(
				`INSERT INTO nodes (key, parent_key, name, obj_type, created_at, updated_at) VALUES (?, ?, ?, 'dir', ?, ?)`,
				current, parent, filepath.Base(current), now, now,
			)
			if err != nil {
				return err
			}
		}
	}

	return tx.Commit()
}

func (db *DB) NodeExists(key string) bool {
	var exists bool
	db.conn.QueryRow("SELECT EXISTS(SELECT 1 FROM nodes WHERE key = ?)", key).Scan(&exists)
	return exists
}
