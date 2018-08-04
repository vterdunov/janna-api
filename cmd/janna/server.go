package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/metrics/prometheus"
	"github.com/gocraft/work"
	"github.com/gomodule/redigo/redis"
	"github.com/pkg/errors"
	stdprometheus "github.com/prometheus/client_golang/prometheus"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/session"
	"github.com/vmware/govmomi/vim25"
	"github.com/vmware/govmomi/vim25/soap"

	"github.com/vterdunov/janna-api/pkg/config"
	"github.com/vterdunov/janna-api/pkg/endpoint"
	"github.com/vterdunov/janna-api/pkg/service"
	"github.com/vterdunov/janna-api/pkg/transport"
	"github.com/vterdunov/janna-api/pkg/version"
)

type JobContext struct{}

func (c *JobContext) DeployOVA(job *work.Job) error {
	// c.logger.Log()

	// Extract arguments:
	addr := job.ArgString("address")

	subject := job.ArgString("subject")
	if err := job.ArgError(); err != nil {
		fmt.Println("args error")
		return err
	}
	fmt.Println("+++++++++++++++++")
	fmt.Println(addr)
	fmt.Println(subject)
	fmt.Println(job.ID)
	fmt.Println("+++++++++++++++++")

	// Go ahead and send the email...
	// sendEmailTo(addr, subject)

	return nil
}

func (c *JobContext) Log(job *work.Job, next work.NextMiddlewareFunc) error {
	fmt.Println("Starting job: ", job.Name)
	job.Checkin("test")
	return next()
}

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

	client, err := newGovmomiClient(ctx, cfg.VMWare.URL, cfg.VMWare.Insecure)
	if err != nil {
		logger.Log("err", errors.Wrap(err, "Could not create Govmomi client"))
		os.Exit(1)
	}

	// ------------------------------------------------------
	// ------------------------------------------------------
	// ------------------------------------------------------
	// WORKERS
	redisPool := &redis.Pool{
		MaxActive: 5,
		MaxIdle:   5,
		Wait:      true,
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", "janna-redis:6379")
		},
	}
	enqueuer := work.NewEnqueuer("janna", redisPool)

	pool := work.NewWorkerPool(JobContext{}, 10, "janna", redisPool)
	pool.Middleware((*JobContext).Log)
	pool.Start()

	pool.Job("send_email", (*JobContext).DeployOVA)

	_, err = enqueuer.Enqueue("send_email", work.Q{"address": "test@example.com", "subject": "This is subject"})
	if err != nil {
		fmt.Println(err)
	}

	// ------------------------------------------------------
	// ------------------------------------------------------
	// ------------------------------------------------------

	duration := prometheus.NewSummaryFrom(stdprometheus.SummaryOpts{
		Namespace: "http",
		Subsystem: "request",
		Name:      "request_latency_seconds",
		Help:      "Total duration of requests in seconds.",
	}, []string{"method", "success"})

	// Build the layers of the service "onion" from the inside out.
	svc := service.New(logger, cfg, client.Client, duration)

	endpoints := endpoint.New(svc, logger)
	httpHandler := transport.NewHTTPHandler(endpoints, logger)
	jsonrpcHandler := transport.NewJSONRPCHandler(endpoints, logger)

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
