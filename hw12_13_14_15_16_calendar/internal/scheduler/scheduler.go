package scheduler

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Maxim-Ba/hw12_13_14_15_calendar/internal/metrics"
	"github.com/Maxim-Ba/hw12_13_14_15_calendar/internal/queue"
	"github.com/Maxim-Ba/hw12_13_14_15_calendar/internal/storage"
)

type EventStorage interface {
	ListEvents(ctx context.Context, from, to time.Time) ([]storage.Event, error)
	DeleteEvent(ctx context.Context, id string) error
	MarkEventNotified(ctx context.Context, id string) error
}

type Logger interface {
	Info(msg string)
	Error(msg string)
}

type Scheduler struct {
	storage  EventStorage
	producer queue.Producer
	topic    string
	logger   Logger
}

func New(storage EventStorage, producer queue.Producer, topic string, logger Logger) *Scheduler {
	return &Scheduler{
		storage:  storage,
		producer: producer,
		topic:    topic,
		logger:   logger,
	}
}

func (s *Scheduler) ScanAndNotify(ctx context.Context) error {
	now := time.Now()
	endTime := now.Add(24 * time.Hour)

	events, err := s.storage.ListEvents(ctx, now, endTime)
	if err != nil {
		return fmt.Errorf("failed to list events: %w", err)
	}

	for _, event := range events {
		if event.NotificationTime == 0 {
			continue
		}

		if event.NotifiedAt != nil {
			continue
		}

		notifyAt := event.StartTime.Add(-event.NotificationTime)
		if notifyAt.After(now) && notifyAt.Before(now.Add(time.Hour)) {
			if err := s.sendNotification(ctx, event); err != nil {
				metrics.SchedulerNotificationsErrorsTotal.Inc()
				s.logger.Error(fmt.Sprintf("failed to send notification for event %s: %v", event.ID, err))
				continue
			}
			if err := s.storage.MarkEventNotified(ctx, event.ID); err != nil {
				s.logger.Error(fmt.Sprintf("failed to mark event %s as notified: %v", event.ID, err))
			}
			metrics.SchedulerNotificationsSentTotal.Inc()
			s.logger.Info(fmt.Sprintf("notification sent for event %s", event.ID))
		}
	}

	metrics.SchedulerLastRunTimestamp.SetToCurrentTime()
	return nil
}

func (s *Scheduler) CleanupOldEvents(ctx context.Context) error {
	oneYearAgo := time.Now().AddDate(-1, 0, 0)
	events, err := s.storage.ListEvents(ctx, time.Time{}, oneYearAgo)
	if err != nil {
		return fmt.Errorf("failed to list old events: %w", err)
	}

	for _, event := range events {
		if err := s.storage.DeleteEvent(ctx, event.ID); err != nil {
			s.logger.Error(fmt.Sprintf("failed to delete event %s: %v", event.ID, err))
			continue
		}
		metrics.SchedulerCleanupsTotal.Inc()
		s.logger.Info(fmt.Sprintf("deleted old event %s", event.ID))
	}

	return nil
}

func (s *Scheduler) sendNotification(ctx context.Context, event storage.Event) error {
	notification := storage.Notification{
		EventID:   event.ID,
		Title:     event.Title,
		EventDate: event.StartTime,
		UserID:    event.UserID,
	}

	data, err := json.Marshal(notification)
	if err != nil {
		return fmt.Errorf("failed to marshal notification: %w", err)
	}

	if err := s.producer.SendMessage(ctx, s.topic, event.ID, data); err != nil {
		return fmt.Errorf("failed to send message to kafka: %w", err)
	}

	return nil
}
