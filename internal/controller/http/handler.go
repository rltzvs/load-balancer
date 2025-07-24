package http

import (
	"load-balancer/internal/loadbalancer"
	"load-balancer/internal/logger"
	"net/http"
)

type Handler struct {
	Balancer *loadbalancer.Balancer
	Logger   logger.Logger
}

func New(balancer *loadbalancer.Balancer, logger logger.Logger) *Handler {
	return &Handler{
		Balancer: balancer,
		Logger:   logger,
	}
}

// TODO: Балансировщик должен корректно обрабатывать ситуацию, когда один или несколько бэкендов недоступны
// (выводить понятное сообщение об ошибке или перенаправлять запросы на работающие серверы).

// Реализовать обработку ошибок при обращении к бэкендам.
// Выводить понятные сообщения ошибок в лог (например, при недоступности сервера).
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	upstream, err := h.Balancer.Next()
	if err != nil {
		h.Logger.Error("failed to get next upstream", "error", err)
		http.Error(w, "Service not available", http.StatusServiceUnavailable)
		return
	}
	h.Logger.Info("forwarding request", "url", r.URL.String(), "to", upstream.URL.String())
	upstream.Proxy.ServeHTTP(w, r)
}
