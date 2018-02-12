package microjson

import (
	"context"
	"net/http"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
)

var loggingContextKey = struct{}{}

func RequestLogEntry(req *http.Request) *logrus.Entry {
	return ContextLogEntry(req.Context())
}

func ContextLogEntry(ctx context.Context) *logrus.Entry {
	entry, ok := ctx.Value(loggingContextKey).(*logrus.Entry)
	if !ok || entry == nil {
		return logrus.WithField("context", "unavailable")
	}
	return entry
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (rec *statusRecorder) WriteHeader(status int) {
	rec.status = status
	rec.ResponseWriter.WriteHeader(status)
}

func LoggingMiddleware(rootEntry *logrus.Entry) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			start := time.Now()
			rwRecorder := &statusRecorder{
				status:         200,
				ResponseWriter: rw,
			}
			entry := rootEntry.WithFields(logrus.Fields{
				"url":      req.URL.String(),
				"status":   rwRecorder.status,
				"duration": start.Sub(time.Now()).Seconds(),
			})

			logContext := context.WithValue(req.Context(), loggingContextKey, entry)

			next.ServeHTTP(rwRecorder, req.WithContext(logContext))

			if route := mux.CurrentRoute(req); route != nil {
				entry = entry.WithFields(logrus.Fields{
					"route": route.GetName(),
				})
			}

			entry.Info("Request")

		})
	}
}
