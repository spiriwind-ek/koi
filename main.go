package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"koi/config"
	"koi/gateway"
	"koi/storage"
)

func main() {
	configPath := flag.String("config", "config/koi.toml", "config file path")
	listen := flag.String("listen", "", "override listen address")
	dbPath := flag.String("db", "", "override database path")
	webDir := flag.String("web", "", "web directory")
	flag.Parse()

	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Printf("Koi 0.1.0-mvp starting...")

	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	log.Printf("Config: %s", *configPath)

	if *listen != "" {
		cfg.Server.Listen = *listen
	}
	if *dbPath != "" {
		cfg.Database.Path = *dbPath
	}

	db, err := storage.NewDB(cfg.Database.Path)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()
	log.Printf("Database: %s", cfg.Database.Path)

	dir := *webDir
	if dir == "" {
		dir = gateway.FindWebDir()
	}

	srv := gateway.NewServer(db, cfg, dir)

	httpServer := &http.Server{
		Addr:         cfg.Server.Listen,
		Handler:      srv.Handler(),
		ReadTimeout:  30e9,
		WriteTimeout: 30e9,
	}

	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan
		log.Println("Shutting down...")
		httpServer.Close()
	}()

	fmt.Printf("Listening on %s\n", cfg.Server.Listen)
	if err := httpServer.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatalf("Server error: %v", err)
	}
}
