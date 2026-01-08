package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/ankulikov/rapidmin/backend/pkg/config"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"

	"github.com/ankulikov/rapidmin/backend/app"
)

func main() {
	addr := flag.String("addr", ":8080", "HTTP listen address")
	configPath := flag.String("config", "config.yaml", "path to config YAML")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("load config error: %v", err)
	}

	srv, err := app.NewServer(cfg)
	if err != nil {
		log.Fatalf("init server error: %v", err)
	}

	handler := srv.Handler()

	log.Printf("listening on %s", *addr)
	if err := http.ListenAndServe(*addr, handler); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
