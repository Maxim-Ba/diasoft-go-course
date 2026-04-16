package storer

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Maxim-Ba/hw12_13_14_15_calendar/internal/queue"
	"github.com/Maxim-Ba/hw12_13_14_15_calendar/internal/storage"
)

type NotificationStorage interface {
	SaveNotification(ctx context.Context, notification storage.Notification) error
}

type Logger interface {
	Info(msg string)
	Error(msg string)
}

type Storer struct {
	storage  NotificationStorage
	consumer queue.Consumer
	logger   Logger
}

func New(storage NotificationStorage, consumer queue.Consumer, logger Logger) *Storer {
	return &Storer{
		storage:  storage,
		consumer: consumer,
		logger:   logger,
	}
}

func (s *Storer) Start(ctx context.Context, topic string) error {
	handler := func(ctx context.Context, message []byte) error {
		var notification storage.Notification
		if err := json.Unmarshal(message, &notification); err != nil {
			s.logger.Error(fmt.Sprintf("failed to unmarshal notification: %v", err))
			return nil
		}

		if err := s.storage.SaveNotification(ctx, notification); err != nil {
			s.logger.Error(fmt.Sprintf("failed to save notification: %v", err))
			return err
		}

		s.logger.Info(fmt.Sprintf("notification saved for event %s", notification.EventID))
		return nil
	}

	return s.consumer.ConsumeMessages(ctx, topic, handler)
}
