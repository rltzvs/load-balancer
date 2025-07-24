package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"load-balancer/internal/config"
	httpcontroller "load-balancer/internal/controller/http"
	"load-balancer/internal/healthchecker"
	"load-balancer/internal/loadbalancer"
	"load-balancer/internal/logger"
	"load-balancer/internal/ratelimiter"
)

func main() {
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
	go limiter.RefillAll(context.Background())

	healthChecker := healthchecker.New(balancer.Upstreams, cfg.Balancer.HealthCheckInterval, applog)
	go healthChecker.Start(context.Background())

	handler := httpcontroller.New(balancer, applog)

	http.Handle("/", handler)

	server := &http.Server{
		Addr:    ":" + cfg.Server.Port,
		Handler: limiter.Middleware(handler),
	}
	applog.Info("server started", "addr", server.Addr)
	if err := server.ListenAndServe(); err != nil {
		applog.Error("server listen error", "error", err)
	}
}
