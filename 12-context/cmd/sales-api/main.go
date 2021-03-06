package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kelseyhightower/envconfig"
	"github.com/pkg/errors"

	"github.com/ardanlabs/service-training/12-context/cmd/sales-api/internal/handlers"
	"github.com/ardanlabs/service-training/12-context/internal/platform/database"
	"github.com/ardanlabs/service-training/12-context/internal/platform/log"
)

// This is for parsing the environment.
const envKey = "sales"

type config struct {
	DB   database.Config
	HTTP struct {
		Address string `default:":8000"`
	}
}

func main() {
	if err := run(); err != nil {
		log.Log("shutting down", "error", err)
		os.Exit(1)
	}
}

func run() error {

	// Process command line flags.
	var flags struct {
		configOnly bool
	}
	flag.BoolVar(&flags.configOnly, "config-only", false, "only show parsed configuration then exit")
	flag.Usage = func() {
		fmt.Print("This program is a service for managing inventory and sales at a Garage Sale.\n\nUsage of sales-api:\n\nsales-api [flags]\n\n")
		flag.CommandLine.SetOutput(os.Stdout)
		flag.PrintDefaults()
		fmt.Print("\nConfiguration:\n\n")
		envconfig.Usage(envKey, &config{})
	}
	flag.Parse()

	// Get configuration from environment.
	var cfg config
	if err := envconfig.Process(envKey, &cfg); err != nil {
		return errors.Wrap(err, "parsing config")
	}

	// Print config and exit if requested.
	if flags.configOnly {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "	")
		if err := enc.Encode(cfg); err != nil {
			return errors.Wrap(err, "encoding config as json")
		}
		return nil
	}

	// Initialize dependencies.
	db, err := database.Open(cfg.DB)
	if err != nil {
		return errors.Wrap(err, "connecting to db")
	}
	defer db.Close()

	server := http.Server{
		Addr:    cfg.HTTP.Address,
		Handler: handlers.NewProducts(db),
	}

	serverErrors := make(chan error, 1)
	go func() {
		serverErrors <- server.ListenAndServe()
	}()

	osSignals := make(chan os.Signal, 1)
	signal.Notify(osSignals, os.Interrupt, syscall.SIGTERM)

	log.Log("startup complete")

	select {
	case err := <-serverErrors:
		return errors.Wrap(err, "listening and serving")

	case <-osSignals:
		log.Log("caught signal, shutting down")

		// Give outstanding requests 15 seconds to complete.
		const timeout = 15 * time.Second
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			log.Log("gracefully shutting down server", "error", err)
			if err := server.Close(); err != nil {
				log.Log("closing server", "error", err)
			}
		}
	}

	log.Log("done")

	return nil
}
