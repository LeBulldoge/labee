package database

import (
	"context"
	"fmt"
)

type dbContextKeyType string

const dbContextKey dbContextKeyType = "db"

func WithDatabase(ctx context.Context, db *DB) context.Context {
	return context.WithValue(ctx, dbContextKey, db)
}

func FromContext(ctx context.Context) (*DB, error) {
	db, ok := ctx.Value(dbContextKey).(*DB)
	if !ok {
		return nil, fmt.Errorf("couldn't get database from context")
	}

	return db, nil
}
