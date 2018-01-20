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
	"net/http"
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/vterdunov/janna-api/config"
	"github.com/vterdunov/janna-api/version"
)

func main() {
	// Load ENV configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("Cannot read config")
	}

	logger := zerolog.New(os.Stderr).With().Timestamp().Logger()
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

	// router := handlers.Router(version.BuildTime, version.Commit, version.Release)

	// interrupt := make(chan os.Signal, 1)
	// signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	// srv := &http.Server{
	// 	Addr:    ":" + cfg.Port,
	// 	Handler: router,
	// }

	// go func() {
	// 	if err := srv.ListenAndServe(); err != nil {
	// 		log.Fatal().Err(err).Msg("Startup failed")
	// 	}
	// }()

	// log.Info().Msg("The service is ready to listen and serve.")

	// killSignal := <-interrupt
	// switch killSignal {
	// case os.Interrupt:
	// 	log.Info().Msg("Got SIGINT...")
	// case syscall.SIGTERM:
	// 	log.Info().Msg("Got SIGTERM...")
	// }

	// log.Info().Msg("The service is shutting down...")
	// srv.Shutdown(context.Background())
	// log.Info().Msg("Done")
	// logger := log.NewLogfmtLogger(os.Stdout)
	// var logger log.Logger
	// logger = log.NewJSONLogger(log.NewSyncWriter(os.Stdout))

	var svc Service
	svc = service{}
	svc = LoggingMiddleware(logger)(svc)

	var h http.Handler
	h = MakeHTTPHandler(svc)

	// vmInfoHandler := httptransport.NewServer(
	// 	makeVMInfoEndpoint(svc),
	// 	decodeVMInfoRequest,
	// 	encodeResponse,
	// )

	// http.Handle("/vm/info", vmInfoHandler)

	http.ListenAndServe(":8080", h)
}
