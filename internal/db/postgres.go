package db

import (
    "database/sql"
    "github.com/lib/pq"
)

func Connect(url string) (*sql.DB, error) {
    return sql.Open("postgres", url)
}
