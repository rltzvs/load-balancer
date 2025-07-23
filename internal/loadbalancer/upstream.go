package loadbalancer

import (
	"net/http/httputil"
	"net/url"
	"sync/atomic"
)

type Upstream struct {
	URL   *url.URL
	Proxy *httputil.ReverseProxy
	Alive *atomic.Bool
}

func NewUpstream(u *url.URL, proxy *httputil.ReverseProxy) *Upstream {
	alive := &atomic.Bool{}
	alive.Store(true)
	return &Upstream{
		URL:   u,
		Proxy: proxy,
		Alive: alive,
	}
}

func (u *Upstream) IsAlive() bool {
	return u.Alive.Load()
}

func (u *Upstream) SetAlive(alive bool) {
	u.Alive.Store(alive)
}
