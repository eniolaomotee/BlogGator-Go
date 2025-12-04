package service

import (
	"context"
	"log"
	"time"
)

type HealthCheck func(ctx context.Context) error

type HealthChecker struct {
	service     *Service
	healthCheck HealthCheck
	interval    time.Duration
	timeout     time.Duration
	failures    int
	maxFailures int
	ctx         context.Context
	cancel      context.CancelFunc
}

func NewHealthChecker(service *Service, check HealthCheck, interval, timeout time.Duration, maxFailures int) *HealthChecker {
	ctx, cancel := context.WithCancel(context.Background())
	return &HealthChecker{
		service:     service,
		healthCheck: check,
		interval:    interval,
		timeout:     timeout,
		maxFailures: maxFailures,
		ctx:         ctx,
		cancel:      cancel,
	}
}

func (h *HealthChecker) Start() {
	go h.run()
}

func (h *HealthChecker) Stop() {
	h.cancel()
}

func (h *HealthChecker) run() {
	ticker := time.NewTicker(h.interval)
	defer ticker.Stop()

	for {
		select {
		case <-h.ctx.Done():
			return
		case <-ticker.C:
			h.check()
		}
	}
}

func (h *HealthChecker) check() {
	ctx, cancel := context.WithTimeout(h.ctx, h.timeout)
	defer cancel()

	err := h.healthCheck(ctx)

	if err != nil {
		h.failures++
		log.Printf("[%s] Health check failed (%d/%d): %v",
			h.service.config.Name, h.failures, h.maxFailures, err)

		if h.failures >= h.maxFailures {
			log.Printf("[%s] Max health check failures reached, restarting service",
				h.service.config.Name)
			h.failures = 0

			// Restart the service
			if err := h.service.stop(); err != nil {
				log.Printf("[%s] Error stopping service: %v", h.service.config.Name, err)
			}
			time.Sleep(2 * time.Second)
			if err := h.service.start(); err != nil {
				log.Printf("[%s] Error restarting service: %v", h.service.config.Name, err)
			}
		}
	} else {
		// Reset failure counter on success
		if h.failures > 0 {
			log.Printf("[%s] Health check recovered", h.service.config.Name)
		}
		h.failures = 0
	}
}
