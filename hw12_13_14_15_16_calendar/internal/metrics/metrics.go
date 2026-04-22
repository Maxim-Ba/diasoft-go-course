package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	HTTPRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests.",
		},
		[]string{"method", "path", "status"},
	)

	HTTPRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Duration of HTTP requests in seconds.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)

	EventsCreatedTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "calendar_events_created_total",
			Help: "Total number of calendar events created.",
		},
	)

	EventsUpdatedTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "calendar_events_updated_total",
			Help: "Total number of calendar events updated.",
		},
	)

	EventsDeletedTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "calendar_events_deleted_total",
			Help: "Total number of calendar events deleted.",
		},
	)

	EventsFetchedTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "calendar_events_fetched_total",
			Help: "Total number of calendar event list requests by scope (day/week/month).",
		},
		[]string{"scope"},
	)

	SchedulerNotificationsSentTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "calendar_scheduler_notifications_sent_total",
			Help: "Total number of event notifications sent by the scheduler.",
		},
	)

	SchedulerNotificationsErrorsTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "calendar_scheduler_notifications_errors_total",
			Help: "Total number of errors when sending event notifications.",
		},
	)

	SchedulerCleanupsTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "calendar_scheduler_cleanups_total",
			Help: "Total number of old events deleted by the scheduler cleanup.",
		},
	)

	SchedulerLastRunTimestamp = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "calendar_scheduler_last_run_timestamp_seconds",
			Help: "Unix timestamp of the last scheduler ScanAndNotify run.",
		},
	)
)
