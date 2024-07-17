package store

import (
	_ "database/sql"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	"github.com/golang-migrate/migrate/v4"
	"github.com/jmoiron/sqlx"
)

type DB struct {
	Client *sqlx.DB

	url string
}

func New(url string) (db DB, err error) {
	db.Client, err = sqlx.Connect("postgres", url)
	if err != nil {
		return
	}

	db.url = url

	return
}

func (db *DB) Close() error {
	if db.Client != nil {
		db.Client.Close()
	}
	return nil
}

func Migrate(url string) error {
	if url != "" {
		mg, err := migrate.New("file://../../migrations/postgres", url)
		if err != nil {
			return err
		}

		if err = mg.Up(); err != nil && err != migrate.ErrNoChange {
			return err
		}
	}

	return nil
}
