package db

import (
	"database/sql"
	"embed"
	"fmt"

	_ "github.com/lib/pq"
)

//go:embed migrations/*.sql
var EmbedMigrations embed.FS

func Connect(host, port, user, pass, name string) (*sql.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, pass, name,
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	return db, db.Ping()
}
