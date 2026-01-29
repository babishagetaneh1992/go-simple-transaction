package database

import (
	"database/sql"
	"strings"

	_ "github.com/lib/pq"
)

func NewPostgres(dsn string) (*sql.DB, error) {
	// Auto-append sslmode=disable if not present to avoid "SSL is not enabled on the server" errors
	if !strings.Contains(dsn, "sslmode") {
		if strings.Contains(dsn, "://") {
			if strings.Contains(dsn, "?") {
				dsn += "&sslmode=disable"
			} else {
				dsn += "?sslmode=disable"
			}
		} else {
			dsn += " sslmode=disable"
		}
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	db.SetMaxIdleConns(10)

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}
