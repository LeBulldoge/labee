package database

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"

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
		err := os.CreatePath(dbPath)
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
		//err := schema.ApplySchema(ctx, tx)

		needSchemaUpdate, err := schema.CheckIfSchemaDiffers(ctx, tx)
		if err != nil {
			return err
		}

		if needSchemaUpdate {
			err := schema.Migrate(ctx, tx)
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

func (m *DB) GetFileLabels(path string) ([]Label, error) {
	stmt :=
		`SELECT Label.* FROM Label, File
    JOIN FileInfo ON
    Label.id = FileInfo.labelId AND
    FileInfo.fileId = File.id
    WHERE File.path = $1`

	labels := []Label{}
	err := m.db.Select(&labels, stmt, path)
	if err != nil {
		return nil, err
	}

	return labels, nil
}

func (m *DB) LabelExists(name string) bool {
	row, err := m.db.Query(`SELECT id FROM Label WHERE name = ?`, name)
	if err != nil {
		return false
	}
	exists := row.Next()
	row.Close()
	return exists
}

func (m *DB) GetSimilarLabel(name string) *Label {
	letters := strings.Join(strings.Split(name, ""), "%")

	var label Label
	err := m.db.Get(&label,
		`SELECT name, color FROM Label
        WHERE name LIKE '%`+letters+"%'")

	if err != nil {
		return nil
	}

	return &label
}

func (m *DB) GetFilesByLabels(labels []string) ([]File, error) {
	stmt := fmt.Sprintf(
		`SELECT File.id, File.path
      FROM File, Label
      INNER JOIN FileInfo ON File.id  = FileInfo.fileId
                         AND Label.id = FileInfo.labelId
      WHERE Label.name IN ('%s')
    GROUP BY File.path
    HAVING COUNT(File.path) = %d`,
		strings.Join(labels, "','"), len(labels),
	)

	files := []File{}
	err := m.db.Select(&files, stmt)
	if err != nil {
		return nil, err
	}

	return files, nil
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
