package handlers

import (
	"net/http"

	"github.com/statping-ng/statping-ng/types/core"
	"github.com/statping-ng/statping-ng/types/services"
)

// PollingSettingsResponse is the API response for polling settings
type PollingSettingsResponse struct {
	Workers            int `json:"polling_workers"`
	QueueSize          int `json:"polling_queue_size"`
	RateLimitPerDomain int `json:"polling_rate_limit"`
}

// apiPollingSettingsHandler returns polling/worker pool settings (GET /api/polling)
func apiPollingSettingsHandler(w http.ResponseWriter, r *http.Request) {
	workers := core.App.PollingWorkers
	if workers == 0 {
		workers = services.DefaultWorkers
	}
	queueSize := core.App.PollingQueueSize
	if queueSize == 0 {
		queueSize = services.DefaultQueueSize
	}
	rateLimit := core.App.PollingRateLimitPerDomain
	if rateLimit == 0 {
		rateLimit = services.DefaultRateLimitPerDomain
	}

	resp := PollingSettingsResponse{
		Workers:            workers,
		QueueSize:          queueSize,
		RateLimitPerDomain: rateLimit,
	}
	returnJson(resp, w, r)
}

// PollingSaveRequest is the request body for saving polling settings
type PollingSaveRequest struct {
	Workers            int `json:"polling_workers"`
	QueueSize          int `json:"polling_queue_size"`
	RateLimitPerDomain int `json:"polling_rate_limit"`
}

// apiPollingSaveHandler saves polling settings (POST /api/polling)
func apiPollingSaveHandler(w http.ResponseWriter, r *http.Request) {
	var req PollingSaveRequest
	if err := DecodeJSON(r, &req); err != nil {
		sendErrorJson(err, w, r)
		return
	}

	// Validate and clamp values
	if req.Workers < services.MinWorkers {
		req.Workers = services.MinWorkers
	} else if req.Workers > services.MaxWorkers {
		req.Workers = services.MaxWorkers
	}
	if req.QueueSize < services.MinQueueSize {
		req.QueueSize = services.MinQueueSize
	} else if req.QueueSize > services.MaxQueueSize {
		req.QueueSize = services.MaxQueueSize
	}
	if req.RateLimitPerDomain < 0 {
		req.RateLimitPerDomain = 0
	}

	// Update core settings
	core.App.PollingWorkers = req.Workers
	core.App.PollingQueueSize = req.QueueSize
	core.App.PollingRateLimitPerDomain = req.RateLimitPerDomain

	if err := core.App.Update(); err != nil {
		sendErrorJson(err, w, r)
		return
	}

	// Update worker pool config (rate limit can change immediately)
	wp := services.GetWorkerPool()
	wp.UpdateConfig(req.Workers, req.QueueSize, req.RateLimitPerDomain)

	log.Infof("Polling settings updated: workers=%d, queue_size=%d, rate_limit=%d",
		req.Workers, req.QueueSize, req.RateLimitPerDomain)

	returnJson(PollingSettingsResponse(req), w, r)
}

// apiPollingStatsHandler returns worker pool statistics (GET /api/polling/stats)
func apiPollingStatsHandler(w http.ResponseWriter, r *http.Request) {
	wp := services.GetWorkerPool()
	stats := wp.Stats()
	returnJson(stats, w, r)
}
