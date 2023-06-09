package database

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/jmoiron/sqlx"
)

type Label struct {
	Id    int64  `db:"id"`
	Name  string `db:"name"`
	Color string `db:"color"`
}

func (m *DB) UpdateLabel(ctx context.Context, name string, newName string, newColor string) error {
	err := tx(ctx, m.db, func(ctx context.Context, tx *sqlx.Tx) error {
		if len(newName) > 0 {
			_, err := tx.ExecContext(ctx, `UPDATE Label SET name = $1 WHERE name = $3`, newName, newColor, name)
			if err != nil {
				return err
			}

			name = newName
		}

		if len(newColor) > 0 {
			err := UpsertLabel(ctx, tx, name, newColor)
			return err
		}

		return nil
	})

	return err
}

func (m *DB) GetAllLabels() ([]Label, error) {
	labels := []Label{}
	err := m.db.Select(&labels, "SELECT name, color FROM Label")
	if err != nil {
		return nil, err
	}

	return labels, nil
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

func getLabelId(db *sqlx.DB, name string) (int64, error) {
	var id int64
	err := db.Get(&id, "SELECT id FROM Label WHERE name = $1", name)
	return id, err
}

func (m *DB) LabelExists(name string) bool {
	_, err := getLabelId(m.db, name)
	return err == nil
}

func (m *DB) AddLabel(ctx context.Context, name string, color string) (*Label, error) {
	var result *Label
	err := tx(ctx, m.db, func(ctx context.Context, tx *sqlx.Tx) error {
		label, err := getOrInsertLabel(ctx, tx, name)
		if err != nil {
			return err
		}

		if label.Color == color {
			return nil
		}

		err = UpsertLabel(ctx, tx, name, color)
		if err != nil {
			return err
		}

		label.Color = color
		result = label

		return nil
	})

	if err != nil {
		return nil, err
	}

	return result, nil
}

func (m *DB) DeleteLabel(ctx context.Context, name string) error {
	err := tx(ctx, m.db, func(ctx context.Context, tx *sqlx.Tx) error {
		_, err := tx.ExecContext(ctx, `DELETE FROM Label WHERE name = ?`, name)
		return err
	})

	return err
}

func insertLabel(ctx context.Context, tx *sqlx.Tx, name string) (int64, error) {
	res, err := tx.ExecContext(ctx, `INSERT INTO Label (name) VALUES (?)`, name)
	if err != nil {
		return -1, err
	}

	return res.LastInsertId()
}

func getOrInsertLabel(ctx context.Context, tx *sqlx.Tx, name string) (*Label, error) {
	var label Label
	err := tx.GetContext(ctx, &label, `SELECT * FROM Label WHERE name = ?`, name)
	if err == nil {
		return &label, nil
	} else if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	id, err := insertLabel(ctx, tx, name)
	if err != nil {
		return nil, err
	}

	label = Label{Id: id, Name: name}

	return &label, err
}

func UpsertLabel(ctx context.Context, tx *sqlx.Tx, name string, color string) error {
	stmt :=
		`INSERT INTO Label (name, color) VALUES ($1, $2)
        ON CONFLICT(name) DO UPDATE SET color=excluded.color`

	_, err := tx.ExecContext(ctx, stmt, name, color)

	return err
}

func RenameLabel(ctx context.Context, tx *sqlx.Tx, oldName string, newName string) error {
	stmt :=
		`UPDATE Label SET
    name = $1
    WHERE name = $2`

	_, err := tx.ExecContext(ctx, stmt, newName, oldName)

	return err
}
