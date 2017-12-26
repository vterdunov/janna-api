package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/vterdunov/janna-api/config"

	"github.com/vterdunov/janna-api/handlers"
	"github.com/vterdunov/janna-api/version"
)

func main() {
	log.Println("Starting the service...")
	log.Printf("Commit: %s, build time: %s, release: %s",
		version.Commit, version.BuildTime, version.Release,
	)

	// Load ENV configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Cannot read config: %v", err)
	}

	router := handlers.Router(version.BuildTime, version.Commit, version.Release)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: router,
	}

	go func() {
		log.Fatal(srv.ListenAndServe())
	}()
	log.Print("The service is ready to listen and serve.")

	killSignal := <-interrupt
	switch killSignal {
	case os.Interrupt:
		log.Print("Got SIGINT...")
	case syscall.SIGTERM:
		log.Print("Got SIGTERM...")
	}

	log.Print("The service is shutting down...")
	srv.Shutdown(context.Background())
	log.Print("Done")
}
