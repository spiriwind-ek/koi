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
	if len(os.Args) > 1 && os.Args[1] == "shell" {
		runShellMode()
		return
	}
	runServerMode()
}

func runServerMode() {
	configPath := flag.String("config", "config/koi.toml", "config file path")
	listen := flag.String("listen", "", "override listen address")
	dbPath := flag.String("db", "", "override database path")
	webDir := flag.String("web", "web", "web directory")
	apiKey := flag.String("api-key", "", "API key for authentication (empty = no auth)")
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

	key := *apiKey
	if key == "" {
		key = os.Getenv("KOI_API_KEY")
	}

	db, err := storage.NewDB(cfg.Database.Path)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()
	log.Printf("Database: %s", cfg.Database.Path)

	srv := gateway.NewServer(db, cfg, *webDir, key)

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

	if key != "" {
		log.Printf("API key authentication enabled")
	} else {
		log.Printf("WARNING: No API key set, all endpoints are open")
	}

	fmt.Printf("Listening on %s\n", cfg.Server.Listen)
	if err := httpServer.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatalf("Server error: %v", err)
	}
}

func runShellMode() {
	configPath := "config/koi.toml"
	dbPath := ""

	args := os.Args[2:]
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "-config":
			if i+1 < len(args) {
				configPath = args[i+1]
				i++
			}
		case "-db":
			if i+1 < len(args) {
				dbPath = args[i+1]
				i++
			}
		}
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	if dbPath != "" {
		cfg.Database.Path = dbPath
	}

	db, err := storage.NewDB(cfg.Database.Path)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	runShell(db, cfg)
}
