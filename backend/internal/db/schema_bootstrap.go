package db

import (
	"context"
	"database/sql"
	_ "embed"
	"fmt"
)

//go:embed schema.sql
var baseSchemaSQL string

func ensureBaseSchema(ctx context.Context, tx *sql.Tx) error {
	if _, err := tx.ExecContext(ctx, baseSchemaSQL); err != nil {
		return fmt.Errorf("apply base schema: %w", err)
	}

	return nil
}
