package loadbalancer

import (
	"errors"
	"load-balancer/internal/logger"
	"net/http/httputil"
	"net/url"
	"sync/atomic"
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

		Upstreams = append(Upstreams, NewUpstream(origin, proxy))
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
