package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/sahib/config"
	"github.com/sahib/wedlist/cache"
	"github.com/sahib/wedlist/db"
	"github.com/sahib/wedlist/defaults"
	"github.com/sahib/wedlist/server"
)

func main() {
	configPath := "./config.cfg"
	if len(os.Args) > 1 {
		configPath = os.Args[1]
	}

	cfg, err := defaults.OpenMigratedConfig(configPath)
	if err != nil {
		log.Printf("failed to open config: %v", err)

		if _, err := os.Stat(configPath); err != nil && !os.IsNotExist(err) {
			os.Exit(1)
		}

		log.Printf("creating empty config at %s", configPath)
		cfg, err = config.Open(nil, defaults.Defaults, config.StrictnessPanic)
		if err != nil {
			log.Fatalf("failed to load defaults: %v", err)
		}

		fd, err := os.OpenFile(configPath, os.O_CREATE|os.O_WRONLY, 0600)
		if err != nil {
			log.Fatalf("failed to open config location: %v", err)
		}

		if err := cfg.Save(config.NewYamlEncoder(fd)); err != nil {
			log.Fatalf("failed to save default config: %v", err)
		}
	}

	cache, err := cache.NewSessionCache(cfg.String("database.session_cache"))
	if err != nil {
		log.Fatalf("failed to open cache: %v", err)
	}

	defer cache.Close()

	db, err := db.NewDatabase(cfg.String("database.sqlite_path"))
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}

	defer db.Close()

	srv := server.NewServer(cfg, db, cache)
	defer srv.Close()

	go func() {
		if err := srv.Serve(); err != nil {
			if err != http.ErrServerClosed {
				log.Fatalf("failed to serve: %v", err)
			}
		}
	}()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	sig := <-sigs

	log.Printf("Received %v signal; quitting", sig)
	srv.Terminate()
}
