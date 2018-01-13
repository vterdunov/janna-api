// Janna API docs
//
// Janna is a little REST API interface for VMware.
// Janna can deploy your VM from OVA file or Template.
// Also Janna can destroy VMs, show information about VMs or change their power state.
//
//     Schemes: http
//     Host: localhost
//     BasePath: /v2/
//     Version: 0.0.1
// 		 License: MIT http://opensource.org/licenses/MIT
//
//     Consumes:
//     	- text/plain; charset=utf-8
//
//     Produces:
//     	- application/json
//
// swagger:meta
package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/vterdunov/janna-api/config"

	"github.com/vterdunov/janna-api/handlers"
	"github.com/vterdunov/janna-api/version"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	// Load ENV configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("Cannot read config")
	}

	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	zerolog.MessageFieldName = "msg"
	if cfg.Debug {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	log.Info().
		Str("commit", version.Commit).
		Str("build time", version.BuildTime).
		Str("release", version.Release).
		Msg("Starting the service...")

	router := handlers.Router(version.BuildTime, version.Commit, version.Release)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: router,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Fatal().Err(err).Msg("Startup failed")
		}
	}()

	log.Info().Msg("The service is ready to listen and serve.")

	killSignal := <-interrupt
	switch killSignal {
	case os.Interrupt:
		log.Info().Msg("Got SIGINT...")
	case syscall.SIGTERM:
		log.Info().Msg("Got SIGTERM...")
	}

	log.Info().Msg("The service is shutting down...")
	srv.Shutdown(context.Background())
	log.Info().Msg("Done")
}
