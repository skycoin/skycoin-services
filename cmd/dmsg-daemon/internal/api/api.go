package api

import (
	"net/http"

	"github.com/go-chi/chi"
	"github.com/sirupsen/logrus"
	"github.com/skycoin/dmsg/httputil"
)

// NewApi returns a new *chi.Mux object, which can be started as a server
func NewApi(log logrus.FieldLogger) *chi.Mux {
	r := chi.NewRouter()
	r.Use(httputil.SetLoggerMiddleware(log))
	r.Get("/", health())

	return r
}

func health() http.HandlerFunc {
	const expBase = "health"
	return httputil.MakeHealthHandler(expBase, nil)
}
