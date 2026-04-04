package router

import (
	"net/http"

	"github.com/Maxim-Ba/hw12_13_14_15_calendar/internal/server/http/handlers"
	"github.com/Maxim-Ba/hw12_13_14_15_calendar/internal/server/http/middlewares"
)

type Logger interface {
	Info(msg string)
	Infof(format string, args ...interface{})
}

func NewRouter(logger Logger) http.Handler {
	mux := http.NewServeMux()

	helloHandler := handlers.NewHelloHandler()
	mux.Handle("/hello", helloHandler)
	mux.Handle("/", helloHandler)

	handler := middlewares.LoggingMiddleware(logger)(mux)

	return handler
}
