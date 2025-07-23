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

	healthChecker := healthchecker.New(balancer.Upstreams, cfg.Balancer.HealthCheckInterval, applog)
	go healthChecker.Start(context.Background())

	handler := httpcontroller.New(balancer, applog)

	http.Handle("/", handler)

	server := &http.Server{
		Addr:    ":" + cfg.Server.Port,
		Handler: handler,
	}
	applog.Info("server started", "addr", server.Addr)
	if err := server.ListenAndServe(); err != nil {
		applog.Error("server listen error", "error", err)
	}
}
