package schema

import (
	"context"
	"testing"

	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"
)

func testNewDatabase(t *testing.T) *sqlx.DB {
	t.Helper()

	db, err := sqlx.Open("sqlite", ":memory:")
	if err != nil {
		t.Errorf("failed to open db: %v", err)
	}

	return db
}

func TestMigrateDatabase(t *testing.T) {
	db := testNewDatabase(t)
	t.Cleanup(func() {
		db.Close()
	})

	tx, err := db.Beginx()
	if err != nil {
		t.Errorf("couldn't start a transaction: %v", err)
	}

	err = ApplyMigrations(context.TODO(), tx, 0, TargetVersion)
	if err != nil {
		t.Errorf("failed applying migrations: %v", err)
	}

	version, err := CurrentVersion(context.TODO(), tx)
	if err != nil {
		t.Errorf("failed checking db version: %v", err)
	}

	if version != TargetVersion {
		t.Errorf("version: %d does't equal %d: %v", version, TargetVersion, err)
	}
}
