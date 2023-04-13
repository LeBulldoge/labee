package database

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"
)

type Label struct {
	Id    int64          `db:"id"`
	Name  string         `db:"name"`
	Color sql.NullString `db:"color"`
}

func (m *DB) UpdateLabel(oldName string, newName string, newColor string) error {
	labelId, err := getLabelId(m.db, oldName)
	if err != nil {
		return err
	}

	_, err = m.db.Exec("UPDATE Label SET (name, color) VALUES ($1, $2) WHERE labelId = $3",
		newName, newColor, labelId)

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

func getLabelId(db *sqlx.DB, name string) (int64, error) {
	var id int64
	err := db.Get(&id, "SELECT id FROM Label WHERE name = $1", name)
	return id, err
}

func (m *DB) AddLabel(name string, color string) (*Label, error) {
	var result *Label
	err := tx(m.db, context.TODO(), func(ctx context.Context, tx *sqlx.Tx) error {
		label, err := getOrInsertLabel(ctx, tx, name)
		if err != nil {
			return err
		}

		if label.Color.String == color {
			return nil
		}

		err = UpsertLabel(tx, ctx, name, color)
		if err != nil {
			return err
		}

		label.Color.String = color
		result = label

		return nil
	})

	if err != nil {
		return nil, err
	}

	return result, nil
}

func (m *DB) DeleteLabel(name string) error {
	err := tx(m.db, context.TODO(), func(ctx context.Context, tx *sqlx.Tx) error {
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
	if err == nil || errors.Is(err, sql.ErrNoRows) {
		return &label, nil
	} else if err != nil {
		return nil, err
	}

	id, err := insertLabel(ctx, tx, name)
	if err != nil {
		return nil, err
	}

	label = Label{Id: id, Name: name}

	return &label, err
}

func UpsertLabel(tx *sqlx.Tx, ctx context.Context, name string, color string) error {
	stmt :=
		`INSERT INTO Label (name, color) VALUES ($1, $2)
        ON CONFLICT(name) DO UPDATE SET color=excluded.color`

	_, err := tx.ExecContext(ctx, stmt, name, color)

	return err
}

func RenameLabel(tx *sqlx.Tx, ctx context.Context, oldName string, newName string) error {
	stmt :=
		`UPDATE Label SET
    name = $1
    WHERE name = $2`

	_, err := tx.ExecContext(ctx, stmt, newName, oldName)

	return err
}
