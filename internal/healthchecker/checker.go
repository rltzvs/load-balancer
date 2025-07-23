package healthchecker

import (
	"context"
	"load-balancer/internal/loadbalancer"
	"load-balancer/internal/logger"
	"net/http"
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
	ticker := time.NewTicker(hc.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			for _, upstream := range hc.Upstreams {
				hc.checkUpstream(upstream)
			}
		case <-ctx.Done():
			return
		}
	}
}

func (hc *HealthChecker) checkUpstream(upstream *loadbalancer.Upstream) {
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
