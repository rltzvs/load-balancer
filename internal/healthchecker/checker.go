package healthchecker

import (
	"context"
	"load-balancer/internal/loadbalancer"
	"load-balancer/internal/logger"
	"net/http"
	"sync"
	"time"
)

type HealthChecker struct {
	Upstreams []*loadbalancer.Upstream
	Interval  time.Duration
	Logger    logger.Logger
	Client    *http.Client
}

func New(upstreams []*loadbalancer.Upstream, interval time.Duration, logger logger.Logger) *HealthChecker {
	return &HealthChecker{
		Upstreams: upstreams,
		Interval:  interval,
		Logger:    logger,
		Client: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

func (hc *HealthChecker) Start(ctx context.Context) {
	hc.Logger.Info("starting health checker")
	ticker := time.NewTicker(hc.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			var wg sync.WaitGroup
			for _, upstream := range hc.Upstreams {
				wg.Add(1)
				go func(u *loadbalancer.Upstream) {
					defer wg.Done()
					hc.checkUpstream(u)
				}(upstream)
			}
			wg.Wait()
		case <-ctx.Done():
			hc.Logger.Info("stopping health checker")
			return
		}
	}
}

func (hc *HealthChecker) checkUpstream(upstream *loadbalancer.Upstream) {
	hc.Logger.Debug("checking upstream", "url", upstream.URL.String())
	resp, err := hc.Client.Get(upstream.URL.String())
	prevAlive := upstream.IsAlive()

	if err != nil {
		if prevAlive {
			hc.Logger.Warn("upstream became unavailable", "url", upstream.URL.String(), "error", err)
			upstream.SetAlive(false)
		}
		return
	}
	defer resp.Body.Close()

	isHealthy := resp.StatusCode >= 200 && resp.StatusCode < 300
	if isHealthy && !prevAlive {
		hc.Logger.Info("upstream recovered", "url", upstream.URL.String())
		upstream.SetAlive(true)
	} else if !isHealthy && prevAlive {
		hc.Logger.Warn("upstream returned unhealthy status", "url", upstream.URL.String(), "status", resp.StatusCode)
		upstream.SetAlive(false)
	}
}
