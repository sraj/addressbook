package main

import (
	"github.com/mobentum/kern"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func registerMetrics(server *kern.App) {
	reg := prometheus.NewRegistry()
	reg.MustRegister(
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
	)

	handler := promhttp.HandlerFor(reg, promhttp.HandlerOpts{})

	server.GET("/api/metrics", func(c *kern.Context) {
		handler.ServeHTTP(c.Response, c.Request)
	})
}
