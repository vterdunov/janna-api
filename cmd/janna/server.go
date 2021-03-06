package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/vterdunov/janna-api/internal/status"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/metrics/prometheus"
	"github.com/pkg/errors"
	stdprometheus "github.com/prometheus/client_golang/prometheus"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/session"
	"github.com/vmware/govmomi/vim25"
	"github.com/vmware/govmomi/vim25/soap"

	"github.com/vterdunov/janna-api/internal/config"
	"github.com/vterdunov/janna-api/internal/endpoint"
	"github.com/vterdunov/janna-api/internal/service"
	"github.com/vterdunov/janna-api/internal/transport"
	"github.com/vterdunov/janna-api/internal/version"
)

func main() {
	// Load ENV configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("Could not read config. Err: %s\n", err)
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

	client, err := newGovmomiClient(ctx, cfg.VMWare.URL, cfg.VMWare.Insecure)
	if err != nil {
		logger.Log("err", errors.Wrap(err, "Could not create Govmomi client"))
		os.Exit(1)
	}

	duration := prometheus.NewSummaryFrom(stdprometheus.SummaryOpts{
		Namespace: "http",
		Subsystem: "request",
		Name:      "request_latency_seconds",
		Help:      "Total duration of requests in seconds.",
	}, []string{"method", "success"})

	inMemoryStorage := status.NewStorage()
	statusStorage := service.Statuser(inMemoryStorage)

	svc := service.New(logger, cfg, client.Client, duration, statusStorage)

	endpoints := endpoint.New(svc, logger)
	httpHandler := transport.NewHTTPHandler(endpoints, logger, cfg.DebugHTTP)

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
		if err := httpServer.Shutdown(ctx); err != nil {
			logger.Log("msg", "Somethig went wrong while HTTP server stopping", "err", err)
		}
	}
	logger.Log("msg", "Stopped")
}

func newGovmomiClient(ctx context.Context, url string, insecure bool) (*govmomi.Client, error) {
	u, err := soap.ParseURL(url)
	if err != nil {
		return nil, err
	}

	soapClient := soap.NewClient(u, insecure)
	vimClient, err := vim25.NewClient(ctx, soapClient)
	if err != nil {
		return nil, err
	}

	vimClient.RoundTripper = session.KeepAlive(vimClient.RoundTripper, 1*time.Minute)
	client := &govmomi.Client{
		Client:         vimClient,
		SessionManager: session.NewManager(vimClient),
	}

	err = client.SessionManager.Login(ctx, u.User)
	if err != nil {
		return nil, err
	}

	return client, nil
}
