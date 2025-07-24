package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"load-balancer/internal/config"
	httpcontroller "load-balancer/internal/controller/http"
	"load-balancer/internal/healthchecker"
	"load-balancer/internal/loadbalancer"
	"load-balancer/internal/logger"
	"load-balancer/internal/ratelimiter"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	applog := logger.New(cfg.Logging.Level)

	balancer, err := loadbalancer.New(cfg.Balancer.Upstreams, applog)
	if err != nil {
		applog.Error("cannot create balancer", "error", err)
		os.Exit(1)
	}

	limiter := ratelimiter.New(applog, uint32(cfg.RateLimits.DefaultCapacity), uint32(cfg.RateLimits.DefaultRate))
	go limiter.RefillAll(ctx)

	healthChecker := healthchecker.New(balancer.Upstreams, cfg.Balancer.HealthCheckInterval, applog)
	go healthChecker.Start(ctx)

	handler := httpcontroller.New(balancer, applog)

	http.Handle("/", handler)

	server := &http.Server{
		Addr:    ":" + cfg.Server.Port,
		Handler: limiter.Middleware(handler),
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			applog.Error("failed to start server", "error", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	applog.Info("shutdown signal received")

	shutdownCtx, cancel := context.WithTimeout(ctx, time.Duration(cfg.Server.ShutdownTimeout)*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		applog.Error("failed to shutdown server", "error", err)
	} else {
		applog.Info("server shutdown gracefully")
	}
}
