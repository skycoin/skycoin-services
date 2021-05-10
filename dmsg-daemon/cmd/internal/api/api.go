package api

import (
	"net/http"
	"time"

	"github.com/go-chi/chi"
	"github.com/sirupsen/logrus"
	"github.com/skycoin/dmsg/httputil"
)

type API struct {
	http.Handler
	startedAt time.Time
}

// New returns a new API object, which can be started as a server
func New(log logrus.FieldLogger) *API {
	r := chi.NewRouter()
	api := &API{
		Handler:   r,
		startedAt: time.Now(),
	}
	r.Use(httputil.SetLoggerMiddleware(log))
	r.Get("/", api.health())

	return api
}

func (a *API) health() http.HandlerFunc {
	const expBase = "health"
	return httputil.MakeHealthHandler(expBase, nil)
}
