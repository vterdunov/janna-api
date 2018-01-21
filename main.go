// Janna API docs
//
// Janna is a little REST API interface for VMware.
//
//     Schemes: http
//     Host: localhost
//     BasePath: /v1/
//     Version: 0.0.1
// 		 License: MIT http://opensource.org/licenses/MIT
//
//     Consumes:
//     	- application/json
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

	"github.com/go-kit/kit/log"

	"github.com/vterdunov/janna-api/config"
	"github.com/vterdunov/janna-api/version"
)

func main() {
	// Load ENV configuration
	cfg, err := config.Load()
	if err != nil {
		// log.Fatal().Err(err).Msg("Cannot read config")
	}
	var logger log.Logger
	if cfg.Debug {
		logger = log.NewLogfmtLogger(log.NewSyncWriter(os.Stderr))
	} else {
		logger = log.NewJSONLogger(os.Stderr)
	}
	logger = log.With(logger, "ts", log.DefaultTimestampUTC)
	logger = log.With(logger, "caller", log.DefaultCaller)

	ctx := context.Background()
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, syscall.SIGINT, syscall.SIGTERM)

	var svc Service
	svc = service{
		logger: logger,
	}
	svc = NewLoggingMiddleware(logger)(svc)

	h := MakeHTTPHandler(svc, log.With(logger, "component", "http"))

	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: h,
	}

	go func() {
		logger.Log(
			"commit", version.Commit,
			"build_time", version.BuildTime,
			"msg", "Starting the service",
		)
		if err := srv.ListenAndServe(); err != nil {
			logger.Log("msg", "Startup failed")
		}
	}()

	switch <-interrupt {
	case syscall.SIGINT:
		logger.Log("msg", "Got SIGINT")
	case syscall.SIGTERM:
		logger.Log("msg", "Got SIGTERM")
	}

	logger.Log("msg", "The service is shutting down")
	srv.Shutdown(ctx)
	logger.Log("msg", "Done")
}
