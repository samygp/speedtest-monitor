package handlers

import (
	"context"
	"net/http"

	spt "github.com/speedtest-monitor/app/speedtest"
	log "github.com/sirupsen/logrus"
)

// Handler is the main struct to reference when calling
// the handling methods
type Handler struct {
	latestResult *spt.LatestResult
	targets      *spt.Servers
}

// SetLatestResultPtr sets the pointer to the latest measurement result
func (h *Handler) SetLatestResultPtr(lr *spt.LatestResult) {
	h.latestResult = lr
}

// SetTargetServers sets the pointer to the servers to be queried
func (h *Handler) SetTargetServers(targets *spt.Servers) {
	h.targets = targets
}

// HandlerFunction is the callback function type to be called in each of a
// router's endpoints
type HandlerFunction func(context.Context, http.ResponseWriter, *http.Request) error

// GetLatestResult returns the latest stored speedtest results
func (h *Handler) GetLatestResult(ctx context.Context, writer http.ResponseWriter, request *http.Request) error {
	log.Debug("Retrieving last speedtest results")
	h.targets.ShowResult(h.latestResult)
	Respond(ctx, writer, *h.latestResult, http.StatusOK)
	return nil
}

// TestSpeedNow performs a speed test and returns the latest stored speedtest results
func (h *Handler) TestSpeedNow(ctx context.Context, writer http.ResponseWriter, request *http.Request) error {
	log.Debug("Requested to test connection speed")
	h.targets.TestNow(h.latestResult)
	Respond(ctx, writer, *h.latestResult, http.StatusOK)
	return nil
}

// Index just prints a message, indicating the service is still alive
func (h *Handler) Index(ctx context.Context, writer http.ResponseWriter, request *http.Request) error {
	Respond(ctx, writer, "Internet SpeedTest", http.StatusOK)
	return nil
}
