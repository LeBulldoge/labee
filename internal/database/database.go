package database

import (
	"context"
	"errors"
	"path/filepath"

	"github.com/LeBulldoge/labee/internal/database/schema"
	"github.com/LeBulldoge/labee/internal/os"
	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"
)

type DB struct {
	db *sqlx.DB
}

func New() (*DB, error) {
	config := os.ConfigPath()
	dbPath := filepath.Join(config, "storage.db")

	if !os.FileExists(dbPath) {
		err := os.CreateFile(dbPath)
		if err != nil {
			return nil, err
		}
	}

	db, err := sqlx.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(1)

	err = tx(db, context.TODO(), func(ctx context.Context, tx *sqlx.Tx) error {
		curVersion, err := schema.CurrentVersion(ctx, tx)
		if err != nil {
			return err
		}

		needSchemaUpdate := curVersion != schema.TargetVersion

		if needSchemaUpdate {
			err := schema.ApplyMigrations(ctx, tx, curVersion, schema.TargetVersion)
			if err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	res := &DB{db: db}

	return res, nil
}

func (m *DB) Close() error {
	_, err := m.db.Exec("PRAGMA optimize")
	if err != nil {
		return err
	}

	return m.db.Close()
}

func tx(db *sqlx.DB, ctx context.Context, f func(context.Context, *sqlx.Tx) error) error {
	tx, err := db.BeginTxx(ctx, nil)

	if err != nil {
		return err
	}

	err = f(ctx, tx)
	if err != nil {
		e := tx.Rollback()
		return errors.Join(err, e)
	}

	return tx.Commit()
}

func insertFileInfo(tx *sqlx.Tx, fileId int64, labelIds []int64) error {
	for _, labelId := range labelIds {
		_, err := tx.Exec(`INSERT OR IGNORE INTO FileInfo (fileId, labelId) VALUES ($1, $2)`, fileId, labelId)
		if err != nil {
			return err
		}
	}

	return nil
}
