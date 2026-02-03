package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Define the PostgreSQL schema in snake_case
var postgresSchemas = map[string]string{
	"users": `
		DO $$ 
		BEGIN 
			IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'user_role') THEN
				CREATE TYPE user_role AS ENUM ('customer', 'seller');
			END IF;
		END $$;

		CREATE TABLE IF NOT EXISTS users (
			id UUID PRIMARY KEY,
			username TEXT NOT NULL,
			email TEXT NOT NULL UNIQUE,
			password TEXT NOT NULL,
			role user_role NOT NULL DEFAULT 'customer',
			is_verified BOOLEAN NOT NULL DEFAULT 'false',
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);
	`,
}

func CreatePostgresTables(ctx context.Context, postgresPool *pgxpool.Pool) error {
	for tableName, schema := range postgresSchemas {
		if err := ValidateTableName(tableName); err != nil {
			return fmt.Errorf("invalid table name: %s", tableName)
		}

		if _, err := postgresPool.Exec(ctx, schema); err != nil {
			return fmt.Errorf("failed to create table %s: %w", tableName, err)
		}
	}
	return nil
}
