package schema

import (
	"context"

	"github.com/jmoiron/sqlx"
)

var versionMap = map[int](func() migration){
	2: version2,
	1: version1,
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
