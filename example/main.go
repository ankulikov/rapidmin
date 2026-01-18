package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/ankulikov/rapidmin/config"
	_ "github.com/mattn/go-sqlite3"

	"github.com/ankulikov/rapidmin"
)

const (
	dbFile     = "rapidmin.db"
	schemaPath = "db/schema.sql"
	dataPath   = "db/data.sql"
	configPath = "config.yaml"
	serverAddr = ":8080"
)

func main() {
	if err := migrate(); err != nil {
		log.Fatalf("migration failed: %v", err)
	}
	if err := startServer(); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}

func migrate() error {
	if err := os.MkdirAll("db", 0o755); err != nil {
		return fmt.Errorf("create db dir: %w", err)
	}

	if _, err := os.Stat(dbFile); err == nil {
		return nil
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("stat db: %w", err)
	}

	db, err := sql.Open("sqlite3", dbFile)
	if err != nil {
		return fmt.Errorf("open db: %w", err)
	}
	defer db.Close()

	for _, path := range []string{schemaPath, dataPath} {
		sqlBytes, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read %s: %w", path, err)
		}
		if len(sqlBytes) == 0 {
			return fmt.Errorf("%s is empty", path)
		}
		if _, err := db.Exec(string(sqlBytes)); err != nil {
			return fmt.Errorf("exec %s: %w", path, err)
		}
	}

	return ensureConfig()
}

func ensureConfig() error {
	_, err := os.Stat(configPath)
	if err == nil {
		return nil
	}
	if os.IsNotExist(err) {
		return fmt.Errorf("missing %s", configPath)
	}
	return fmt.Errorf("stat config: %w", err)
}

func startServer() error {
	cfg, err := config.Load(configPath)
	if err != nil {
		log.Fatalf("load config error: %v", err)
	}

	srv, err := rapidmin.NewServer(cfg)
	if err != nil {
		return err
	}

	log.Printf("listening on %s", serverAddr)
	return http.ListenAndServe(serverAddr, srv.Handler())
}
