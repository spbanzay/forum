package db

import (
	"database/sql"
	"fmt"
	"os"
)

func InitDatabase(db *sql.DB) error {
	sqlBytes, err := os.ReadFile("internal/db/schema.sql")
	if err != nil {
		return fmt.Errorf("ошибка чтения schema.sql: %w", err)
	}

	_, err = db.Exec(string(sqlBytes))
	if err != nil {
		return fmt.Errorf("ошибка выполнения SQL схемы: %w", err)
	}

	return nil
}
