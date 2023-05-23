package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/jmoiron/sqlx"
)

type File struct {
	Id   int64  `db:"id"`
	Path string `db:"path"`
}

func (m *DB) DeleteFiles(paths []string) error {
	err := tx(context.TODO(), m.db, func(ctx context.Context, tx *sqlx.Tx) error {
		for _, v := range paths {
			err := deleteFile(ctx, tx, v)
			if err != nil {
				return err
			}
		}

		return nil
	})

	return err
}

var ErrCouldNotDeleteFile = errors.New("could not delete file")

func deleteFile(ctx context.Context, tx *sqlx.Tx, path string) error {
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

func InsertFile(ctx context.Context, tx *sqlx.Tx, path string) error {
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

func (m *DB) GetFilesFilteredWithLabels(labels []string, pattern string, pathPrefix string) ([]File, error) {
	stmt := `SELECT File.id, File.path
      FROM File, Label
      INNER JOIN FileInfo ON File.id  = FileInfo.fileId
                         AND Label.id = FileInfo.labelId
      WHERE %s
    GROUP BY File.path
    HAVING COUNT(File.path) = ` + strconv.Itoa(len(labels))

	filters := []string{"Label.name IN ('" + strings.Join(labels, "','") + "')"}

	if len(pattern) > 0 || len(pathPrefix) > 0 {
		filters = append(filters, buildFilenameFilters(pattern, pathPrefix))
	}

	stmt = fmt.Sprintf(stmt, strings.Join(filters, " AND "))

	files := []File{}
	err := m.db.Select(&files, stmt)
	if err != nil {
		return nil, err
	}

	return files, nil
}

func (m *DB) GetFilesFiltered(pattern string, pathPrefix string) ([]File, error) {
	stmt := `SELECT File.* FROM File`

	if len(pattern) > 0 || len(pathPrefix) > 0 {
		stmt += " WHERE " + buildFilenameFilters(pattern, pathPrefix)
	}

	files := []File{}
	err := m.db.Select(&files, stmt)
	if err != nil {
		return nil, err
	}

	return files, nil
}

func buildFilenameFilters(pattern string, pathPrefix string) string {
	var filterBuilder strings.Builder

	filterBuilder.Grow(len(pattern) + len(pathPrefix))
	filterBuilder.WriteString("path GLOB '")
	filterBuilder.WriteString(pathPrefix)
	filterBuilder.WriteString("*")
	filterBuilder.WriteString(pattern)
	filterBuilder.WriteString("'")

	return filterBuilder.String()
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
	err := tx(context.TODO(), m.db, func(ctx context.Context, tx *sqlx.Tx) error {
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
