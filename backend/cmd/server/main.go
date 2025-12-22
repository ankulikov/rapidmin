package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"

	"github.com/ankulikov/rapidmin/internal/config"
	"github.com/ankulikov/rapidmin/internal/providers"
	sqlprovider "github.com/ankulikov/rapidmin/internal/providers/sql"
	"github.com/ankulikov/rapidmin/internal/server"
)

func main() {
	addr := flag.String("addr", ":8080", "HTTP listen address")
	configPath := flag.String("config", "config.yaml", "path to config YAML")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	providerRegistry := providers.Registry{
		"db": sqlprovider.New(openDB()),
	}

	app, err := server.New(cfg, providerRegistry)
	if err != nil {
		log.Fatalf("init server: %v", err)
	}

	log.Printf("listening on %s", *addr)
	if err := http.ListenAndServe(*addr, app.Handler()); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

func openDB() *sqlx.DB {
	//driver := strings.TrimSpace(os.Getenv("DB_DRIVER"))
	//dsn := strings.TrimSpace(os.Getenv("DB_DSN"))
	//if driver == "" || dsn == "" {
	//	return nil
	//}

	db, err := sqlx.Open("sqlite3",
		"file:/Users/ankulikov/GolandProjects/rapidmin/backend/data.db?_fk=1")
	if err != nil {
		log.Printf("db open failed: %v", err)
		return nil
	}

	return db
}
