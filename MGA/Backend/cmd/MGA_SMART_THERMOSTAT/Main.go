package main

import (
    "context"
    "log"
    "net/http"
    "os"
    "time"

    "mga_smart_thermostat/internal/api"
    "mga_smart_thermostat/internal/database"
)

func main() {
    dbPath := "smart_thermostat.db"
    if v := os.Getenv("MGA_DB_PATH"); v != "" { dbPath = v }

    db, err := database.Init(dbPath)
    if err != nil { log.Fatalf("db init failed: %v", err) }
    defer func() { _ = database.Close() }()

    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
    if err := database.Migrate(ctx); err != nil { log.Fatalf("migration failed: %v", err) }

    srv := api.NewServer(db)
    // Optional seed for quick start
    if os.Getenv("MGA_SEED") == "1" {
        if err := srv.SeedQuickStart(ctx); err != nil { log.Printf("seed error: %v", err) }
    }

    addr := ":8080"
    if v := os.Getenv("MGA_HTTP_ADDR"); v != "" { addr = v }
    server := &http.Server{ Addr: addr, Handler: srv.Router(), ReadHeaderTimeout: 10 * time.Second }
    log.Printf("Smart Thermostat server listening on %s", addr)
    if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
        log.Fatalf("server error: %v", err)
    }
}


