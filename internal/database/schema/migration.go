package schema

import (
	"context"
	"fmt"
	"strconv"

	"github.com/jmoiron/sqlx"
)

func CurrentVersion(ctx context.Context, tx *sqlx.Tx) (int, error) {
	curVersion := -1
	err := tx.GetContext(ctx, &curVersion, "PRAGMA user_version")
	if err != nil {
		return curVersion, err
	}

	return curVersion, err
}

func setVersion(ctx context.Context, tx *sqlx.Tx, version int) error {
	_, err := tx.ExecContext(ctx, "PRAGMA user_version = "+strconv.Itoa(version))

	return err
}

func ApplyMigrations(ctx context.Context, tx *sqlx.Tx, fromVer int, toVer int) error {
	if fromVer < toVer {
		err := migrateUp(ctx, tx, fromVer, toVer)
		if err != nil {
			return err
		}
	} else if fromVer > toVer {
		err := migrateDown(ctx, tx, fromVer, toVer)
		if err != nil {
			return err
		}
	} else {
		return nil
	}

	return setVersion(ctx, tx, toVer)
}

func migrateUp(ctx context.Context, tx *sqlx.Tx, curVersion int, targetVersion int) error {
	for v := curVersion + 1; v <= targetVersion; v++ {
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

func migrateDown(ctx context.Context, tx *sqlx.Tx, curVersion int, targetVersion int) error {
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
	2: version2,
}

// Add default value to color to not have to deal with sql NULLs
func version2() migration {
	up := func(ctx context.Context, tx *sqlx.Tx) error {
		stmt := `ALTER TABLE Label RENAME TO Label_b;

CREATE TABLE Label (
  id    INTEGER NOT NULL
                UNIQUE,
  name  TEXT    NOT NULL
                UNIQUE,
  color TEXT    NOT NULL
                DEFAULT 'NONE',
  PRIMARY KEY (
      id AUTOINCREMENT
  )
);

UPDATE Label_b SET color = 'NONE' WHERE color IS NULL;
INSERT INTO Label SELECT * FROM Label_b;

DROP TABLE Label_b;`

		_, err := tx.ExecContext(ctx, stmt)

		return err
	}

	return migration{up: up}
}

// The initial schema
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

		return nil
	}

	return migration{up: up}
}
