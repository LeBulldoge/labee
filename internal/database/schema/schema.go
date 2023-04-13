package schema

import (
	"context"
	"strings"

	"github.com/jmoiron/sqlx"
)

const schema = `CREATE TABLE File (
  id    INTEGER NOT NULL
                UNIQUE,
  path  TEXT    NOT NULL
                UNIQUE,
  PRIMARY KEY (
      id AUTOINCREMENT
  )
);

CREATE TABLE FileInfo (
  fileId INTEGER NOT NULL
               REFERENCES File (id) ON DELETE CASCADE,
  labelId  INTEGER NOT NULL
               REFERENCES Label (id) ON DELETE CASCADE,
  UNIQUE(fileId, labelId)
  FOREIGN KEY (
      fileId
  )
  REFERENCES File (id),
  FOREIGN KEY (
      labelId
  )
  REFERENCES Label (id),
  PRIMARY KEY (
      fileId, labelId
  )
);

CREATE TABLE Label (
  id    INTEGER NOT NULL
                UNIQUE,
  name  TEXT    NOT NULL
                UNIQUE,
  color TEXT,
  PRIMARY KEY (
      id AUTOINCREMENT
  )
);`

func ApplySchema(ctx context.Context, tx *sqlx.Tx) error {
	_, err := tx.ExecContext(ctx, schema)
	return err
}

func CheckIfSchemaDiffers(ctx context.Context, tx *sqlx.Tx) (bool, error) {
	stmt := `
SELECT sql FROM sqlite_master WHERE
    type = 'table' AND
    name NOT LIKE 'sqlite_%'
ORDER BY name`

	rows, err := tx.QueryxContext(ctx, stmt)
	if err != nil {
		return true, err
	}

	var stmts = []string{}
	for rows.Next() {
		var s string
		err = rows.Scan(&s)
		if err != nil {
			return true, err
		}

		stmts = append(stmts, s+";")
	}

	curSchema := strings.Join(stmts, "\n\n")

	return schema != curSchema, nil
}
