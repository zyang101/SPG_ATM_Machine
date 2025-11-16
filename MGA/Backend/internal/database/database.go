package database

import (
    "context"
    "database/sql"
    "embed"
    "errors"
    "fmt"
    _ "modernc.org/sqlite"
)

var (
    dbHandle *sql.DB
)

//go:embed schema.sql
var schemaFS embed.FS

// Init opens a SQLite database at dbPath and configures pragmas.
func Init(dbPath string) (*sql.DB, error) {
    if dbHandle != nil {
        return dbHandle, nil
    }
    dsn := fmt.Sprintf("file:%s?_busy_timeout=5000&_journal_mode=WAL&_fk=1", dbPath)
    handle, err := sql.Open("sqlite", dsn)
    if err != nil {
        return nil, err
    }
    if err = handle.Ping(); err != nil {
        _ = handle.Close()
        return nil, err
    }

    dbHandle = handle
    return dbHandle, nil
}

func GetDB() (*sql.DB, error) {
    if dbHandle == nil {
        return nil, errors.New("database not initialized")
    }
    return dbHandle, nil
}

func Close() error {
    if dbHandle == nil {
        return nil
    }
    err := dbHandle.Close()
    dbHandle = nil
    return err
}

// Migrate executes the embedded schema.
func Migrate(ctx context.Context) error {
    if dbHandle == nil {
        return errors.New("database not initialized")
    }
    contents, err := schemaFS.ReadFile("schema.sql")
    if err != nil {
        return err
    }
    _, err = dbHandle.ExecContext(ctx, string(contents))
    return err
}


