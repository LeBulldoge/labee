package schema

import (
	"context"
	"fmt"
	"strconv"

	"github.com/jmoiron/sqlx"
)

const TargetVersion = 2

type migration struct {
	up   func(context.Context, *sqlx.Tx) error
	down func(context.Context, *sqlx.Tx) error
}

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
