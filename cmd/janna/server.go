// Janna provides HTTP endpoints for some operations on VMware.
// Deploy OVF/OVA, get info about VM, etc.
//
//     Schemes: http
//     Host: localhost
//     BasePath: /
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
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-kit/kit/log"
	"github.com/pkg/errors"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/vim25/soap"

	"github.com/vterdunov/janna-api/pkg/config"
	"github.com/vterdunov/janna-api/pkg/jannaendpoint"
	"github.com/vterdunov/janna-api/pkg/jannaservice"
	"github.com/vterdunov/janna-api/pkg/jannatransport"
	"github.com/vterdunov/janna-api/pkg/version"
)

func main() {
	// Load ENV configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("Cannot read config. Err: %s\n", err)
		os.Exit(1)
	}

	// Create logger, which we'll use and give to other components.
	var logger log.Logger
	if cfg.Debug {
		logger = log.NewLogfmtLogger(log.NewSyncWriter(os.Stdout))
	} else {
		logger = log.NewJSONLogger(os.Stdout)
	}

	logger = log.With(logger, "ts", log.DefaultTimestampUTC)
	logger = log.With(logger, "caller", log.DefaultCaller)

	ctx := context.Background()

	// TODO: add retries with backoff
	client, err := newGovmomiClient(ctx, cfg.VMWare.URL, cfg.VMWare.Insecure)
	if err != nil {
		logger.Log("err", errors.Wrap(err, "Could not create Govmomi client"))
		os.Exit(1)
	}
	vimClient := client.Client

	// Build the layers of the service "onion" from the inside out.
	svc := jannaservice.New(logger, cfg, vimClient)
	endpoints := jannaendpoint.New(svc, logger)
	httpHandler := jannatransport.NewHTTPHandler(endpoints, logger)
	jsonrpcHandler := jannatransport.NewJSONRPCHandler(endpoints, logger)

	logger.Log(
		"commit", version.Commit,
		"build_time", version.BuildTime,
		"msg", "Starting application",
	)

	// HTTP server
	var httpServer *http.Server
	if cfg.Protocols.HTTP.Port != "" {
		httpServer = &http.Server{
			Addr:    ":" + cfg.Protocols.HTTP.Port,
			Handler: httpHandler,
		}

		go func() {
			logger.Log(
				"msg", "Starting HTTP server",
				"address", httpServer.Addr,
			)
			if err := httpServer.ListenAndServe(); err != nil {
				logger.Log("msg", "Startup failed", "err", err)
				os.Exit(1)
			}
		}()
	}

	// JSON RPC server
	var jsonrpcServer *http.Server
	if cfg.Protocols.JSONRPC.Port != "" {
		jsonrpcServer = &http.Server{
			Addr:    ":" + cfg.Protocols.JSONRPC.Port,
			Handler: jsonrpcHandler,
		}

		go func() {
			logger.Log(
				"msg", "Starting JSON RPC over HTTP server",
				"address", jsonrpcServer.Addr,
			)
			if err := jsonrpcServer.ListenAndServe(); err != nil {
				logger.Log("msg", "Startup failed", "err", err)
				os.Exit(1)
			}
		}()
	}

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, syscall.SIGINT, syscall.SIGTERM)

	switch <-interrupt {
	case syscall.SIGINT:
		logger.Log("msg", "Got SIGINT")
	case syscall.SIGTERM:
		logger.Log("msg", "Got SIGTERM")
	}

	logger.Log("msg", "The service is going shutting down")
	client.Logout(ctx)

	if cfg.Protocols.HTTP.Port != "" {
		httpServer.Shutdown(ctx)
	}
	if cfg.Protocols.JSONRPC.Port != "" {
		jsonrpcServer.Shutdown(ctx)
	}
	logger.Log("msg", "Stopped")
}

func newGovmomiClient(ctx context.Context, URL string, insecure bool) (*govmomi.Client, error) {
	u, err := soap.ParseURL(URL)
	if err != nil {
		return nil, err
	}

	c, err := govmomi.NewClient(ctx, u, insecure)
	if err != nil {
		return nil, err
	}
	return c, nil
}
