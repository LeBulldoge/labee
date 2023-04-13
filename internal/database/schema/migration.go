package schema

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
)

const targetVersion = 1

func Migrate(ctx context.Context, tx *sqlx.Tx) error {
	curVersion := -1
	err := tx.GetContext(ctx, &curVersion, "PRAGMA user_version")
	if err != nil {
		return err
	}

	if curVersion < targetVersion {
		err = migrateUp(ctx, tx, curVersion)
		if err != nil {
			return err
		}
	} else if curVersion > targetVersion {
		err = migrateDown(ctx, tx, curVersion)
		if err != nil {
			return err
		}
	}

	return nil
}

func migrateUp(ctx context.Context, tx *sqlx.Tx, curVersion int) error {
	for v := curVersion + 1; v <= len(versionMap); v++ {
		migration := versionMap[v]()
		if migration.up == nil {
			return fmt.Errorf("cannot migrate database further up than v%d, should never happen", v)
		}

		err := migration.up(ctx, tx)
		if err != nil {
			return fmt.Errorf("error migrating database from v%d to v%d: %w", curVersion, v, err)
		}
	}

	return nil
}

func migrateDown(ctx context.Context, tx *sqlx.Tx, curVersion int) error {
	for v := curVersion; v > targetVersion; v-- {
		migration := versionMap[v]()
		if migration.down == nil {
			return fmt.Errorf("cannot migrate database further down than v%d", v)
		}

		err := migration.down(ctx, tx)
		if err != nil {
			return fmt.Errorf("error migrating database from v%d to v%d: %w", curVersion, v, err)
		}
	}

	return nil
}

type migration struct {
	up   func(context.Context, *sqlx.Tx) error
	down func(context.Context, *sqlx.Tx) error
}

var versionMap = map[int](func() migration){
	1: version1,
}

func version1() migration {
	up := func(ctx context.Context, tx *sqlx.Tx) error {

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

		return nil
	}

	return migration{up: up}
}
