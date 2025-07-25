package loadbalancer

import (
	"errors"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync/atomic"

	"load-balancer/internal/controller/http/util"
	"load-balancer/internal/logger"
)

var (
	ErrNoUpstreams      = errors.New("no upstreams available")
	ErrNoAliveUpstreams = errors.New("no alive upstreams available")
)

type Balancer struct {
	Upstreams []*Upstream
	nextIndex atomic.Uint32
	Logger    logger.Logger
}

func New(upstreams []string, logger logger.Logger) (*Balancer, error) {
	var Upstreams []*Upstream

	if len(upstreams) == 0 {
		return nil, ErrNoUpstreams
	}

	for _, upstream := range upstreams {
		origin, err := url.Parse(upstream)

		if err != nil {
			return nil, err
		}
		proxy := httputil.NewSingleHostReverseProxy(origin)

		newUpstream := NewUpstream(origin, proxy)

		proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
			newUpstream.SetAlive(false)
			logger.Error("proxy error", "backend", newUpstream.URL.String(), "error", err)

			util.RespondError(w, http.StatusBadGateway, "Backend is unavailable. Please try again later.")
		}

		Upstreams = append(Upstreams, newUpstream)
	}

	return &Balancer{
		Upstreams: Upstreams,
		nextIndex: atomic.Uint32{},
		Logger:    logger,
	}, nil
}

func (b *Balancer) Next() (*Upstream, error) {
	total := len(b.Upstreams)
	if total == 0 {
		return nil, ErrNoUpstreams
	}

	for i := 0; i < total; i++ {
		idx := int(b.nextIndex.Add(1)) % total

		candidate := b.Upstreams[idx]

		if candidate.IsAlive() {
			return candidate, nil
		}
	}

	return nil, ErrNoAliveUpstreams
}
