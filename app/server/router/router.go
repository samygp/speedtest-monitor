package router

import (
	"github.com/gorilla/mux"
	"github.com/speedtest-monitor/app/server/handlers"
	spt "github.com/speedtest-monitor/app/speedtest"
)

// Endpoint defines an http endpoint to be queried from an external source
type Endpoint struct {
	Name        string
	Method      string
	Pattern     string
	HandlerFunc handlers.HandlerFunc
}

// NewRouter creates the routes to listen
func NewRouter(targets *spt.Servers, maesurementsPtr *spt.LatestResult) *mux.Router {

	handler := &handlers.Handler{}
	handler.SetLatestResultPtr(maesurementsPtr)
	handler.SetTargetServers(targets)

	var endpoints = []Endpoint{
		{
			"Index",
			"GET",
			"/",
			handler.Index,
		},
		{
			"GetLatestResult",
			"GET",
			"/getLatestResult",
			handler.GetLatestResult,
		},
		{
			"TestSpeedNow",
			"GET",
			"/testSpeedNow",
			handler.TestSpeedNow,
		},
	}

	router := mux.NewRouter().StrictSlash(true)
	for _, endpoint := range endpoints {
		router.
			Methods(endpoint.Method).
			Path(endpoint.Pattern).
			Name(endpoint.Name).
			Handler(endpoint.HandlerFunc)
	}

	return router

}
