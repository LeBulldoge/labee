package schema

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
)

func Migrate(ctx context.Context, tx *sqlx.Tx) error {
	curVersion := -1
	err := tx.GetContext(ctx, &curVersion, "PRAGMA user_version")
	if err != nil {
		return err
	}

	for v := curVersion + 1; v <= len(versionMap); v++ {
		migration := versionMap[v]
		err = migration(ctx, tx)
		if err != nil {
			return fmt.Errorf("error migrating database from v%d to v%d: %w", curVersion, v, err)
		}
	}

	return nil
}

var versionMap = map[int](func(context.Context, *sqlx.Tx) error){
	1: version1,
}

func version1(ctx context.Context, tx *sqlx.Tx) error {
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

	_, err := tx.ExecContext(ctx, schema)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, "PRAGMA user_version = 1")
	if err != nil {
		return err
	}

	return err
}
