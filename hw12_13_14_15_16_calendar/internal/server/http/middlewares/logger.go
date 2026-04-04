package middlewares

import (
	"net/http"
	"time"
)

type Logger interface {
	Info(msg string)
	Infof(format string, args ...interface{})
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
	written    int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	n, err := rw.ResponseWriter.Write(b)
	rw.written += n
	return n, err
}

func LoggingMiddleware(logger Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			rw := &responseWriter{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
			}

			next.ServeHTTP(rw, r)

			duration := time.Since(start)

			ip := r.RemoteAddr
			if forwardedFor := r.Header.Get("X-Forwarded-For"); forwardedFor != "" {
				ip = forwardedFor
			}

			timestamp := start.Format("02/Jan/2006:15:04:05 -0700")
			method := r.Method
			path := r.URL.RequestURI()
			proto := r.Proto
			status := rw.statusCode
			latency := duration.Milliseconds()
			userAgent := r.UserAgent()

			logger.Infof("%s [%s] %s %s %s %d %d %q",
				ip, timestamp, method, path, proto, status, latency, userAgent)
		})
	}
}
