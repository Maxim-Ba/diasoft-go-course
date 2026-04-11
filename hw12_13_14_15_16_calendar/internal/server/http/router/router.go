package router

import (
	"encoding/json"
	"net/http"

	"github.com/Maxim-Ba/hw12_13_14_15_calendar/internal/server/http/generated"
	"github.com/Maxim-Ba/hw12_13_14_15_calendar/internal/server/http/handlers"
	"github.com/Maxim-Ba/hw12_13_14_15_calendar/internal/server/http/middlewares"
)

type Logger interface {
	Info(msg string)
	Infof(format string, args ...interface{})
}

func NewRouter(logger Logger, app handlers.Application) http.Handler {
	mux := http.NewServeMux()

	mux.Handle("/hello", handlers.NewHelloHandler())

	mux.HandleFunc("/openapi.json", func(w http.ResponseWriter, r *http.Request) {
		swagger, err := generated.GetSwagger()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(swagger)
	})

	eventsHandler := handlers.NewEventsHandler(app)
	generated.HandlerFromMux(eventsHandler, mux)

	return middlewares.LoggingMiddleware(logger)(mux)
}
