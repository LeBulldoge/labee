package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/jmoiron/sqlx"
)

type File struct {
	Id   int64  `db:"id"`
	Path string `db:"path"`
}

func (m *DB) DeleteFiles(paths []string) error {
	err := tx(m.db, context.TODO(), func(ctx context.Context, tx *sqlx.Tx) error {
		for _, v := range paths {
			err := deleteFile(tx, ctx, v)
			if err != nil {
				return err
			}
		}

		return nil
	})

	return err
}

var ErrCouldNotDeleteFile = errors.New("could not delete file")

func deleteFile(tx *sqlx.Tx, ctx context.Context, path string) error {
	stmt := `DELETE FROM File WHERE path = ?`

	res, err := tx.ExecContext(ctx, stmt, path)
	if err != nil {
		return fmt.Errorf("%w %s: %w", ErrCouldNotDeleteFile, path, err)
	}

	if cnt, err := res.RowsAffected(); err != nil {
		return fmt.Errorf("%w %s: %w", ErrCouldNotDeleteFile, path, err)
	} else if cnt == 0 {
		return fmt.Errorf("%w %s. file doesn't exist in storage", ErrCouldNotDeleteFile, path)
	}

	return err
}

func InsertFile(tx *sqlx.Tx, ctx context.Context, path string) error {
	stmt := "INSERT INTO File (path) VALUES (?)"
	_, err := tx.ExecContext(ctx, stmt, path)

	return err
}

func sortAndPickFirstFile(files []File, q string) File {
	q = strings.ToLower(q)

	// Sort so that the provided path is matched starting from the end
	sort.SliceStable(files, func(i int, j int) bool {
		return strings.LastIndex(strings.ToLower(files[i].Path), q) >
			strings.LastIndex(strings.ToLower(files[j].Path), q)
	})

	return files[0]
}

func (m *DB) GetFiles(keywords []string) ([]File, error) {
	stmt := `SELECT File.id, File.path FROM File`

	if len(keywords) > 0 {
		stmt += " WHERE path LIKE '%"
		stmt += strings.Join(keywords, "%")
		stmt += "%'"
	}

	files := []File{}
	err := m.db.Select(&files, stmt)
	if err != nil {
		return nil, err
	}

	if len(files) == 0 {
		if keywords != nil {
			return nil, fmt.Errorf("%w with %v", ErrFilesNotFound, keywords)
		}
		return nil, ErrFilesNotFound
	}

	return files, nil
}

var (
	ErrFileAlreadyExists = errors.New("file already exists in storage")
	ErrFilesNotFound     = errors.New("couldn't find files")
)

func getOrInsertFile(tx *sqlx.Tx, path string) (int64, error) {
	var id int64
	err := tx.Get(&id, `SELECT id FROM File WHERE path = $1`, path)

	if err == nil {
		return id, nil
	} else if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return -1, err
	}

	res, err := tx.Exec(
		`INSERT INTO File (path) VALUES ($1)`,
		path,
	)
	if err != nil {
		return 0, err
	}
	id, _ = res.LastInsertId()

	return id, err
}

func (m *DB) AddFilesAndLinks(filepaths []string, labelNames []string) error {
	err := tx(m.db, context.TODO(), func(ctx context.Context, tx *sqlx.Tx) error {
		var labelIds []int64
		for _, name := range labelNames {
			label, err := getOrInsertLabel(ctx, tx, name)
			if err != nil {
				return err
			}

			labelIds = append(labelIds, label.Id)
		}

		for _, file := range filepaths {
			fileId, err := getOrInsertFile(tx, file)
			if err != nil {
				return err
			}

			err = insertFileInfo(tx, fileId, labelIds)
			if err != nil {
				return nil
			}
		}

		return nil
	})

	return err
}
