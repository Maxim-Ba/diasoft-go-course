package app

import (
	"context"
	"time"

	"github.com/Maxim-Ba/hw12_13_14_15_calendar/internal/storage"
)

type App struct {
	logger  Logger
	storage Storage
}

type Logger interface {
	Info(msg string)
	Infof(format string, args ...interface{})
	Error(msg string)
	Errorf(format string, args ...interface{})
}

type Storage interface {
	CreateEvent(ctx context.Context, event storage.Event) (string, error)
	UpdateEvent(ctx context.Context, id string, event storage.Event) error
	DeleteEvent(ctx context.Context, id string) error
	GetEventByID(ctx context.Context, id string) (*storage.Event, error)
	GetEventsForDay(ctx context.Context, date time.Time) ([]storage.Event, error)
	GetEventsForWeek(ctx context.Context, startDate time.Time) ([]storage.Event, error)
	GetEventsForMonth(ctx context.Context, startDate time.Time) ([]storage.Event, error)
}

func New(logger Logger, storage Storage) *App {
	return &App{
		logger:  logger,
		storage: storage, // TODO потом буду использовать через слой сервис
	}
}

func (a *App) CreateEvent(ctx context.Context, event storage.Event) (string, error) {
	id, err := a.storage.CreateEvent(ctx, event)
	if err != nil {
		a.logger.Errorf("failed to create event: %v", err)
		return "", err
	}
	a.logger.Infof("event created: id=%s", id)
	return id, nil
}

func (a *App) UpdateEvent(ctx context.Context, id string, event storage.Event) error {
	if err := a.storage.UpdateEvent(ctx, id, event); err != nil {
		a.logger.Errorf("failed to update event id=%s: %v", id, err)
		return err
	}
	a.logger.Infof("event updated: id=%s", id)
	return nil
}

func (a *App) DeleteEvent(ctx context.Context, id string) error {
	if err := a.storage.DeleteEvent(ctx, id); err != nil {
		a.logger.Errorf("failed to delete event id=%s: %v", id, err)
		return err
	}
	a.logger.Infof("event deleted: id=%s", id)
	return nil
}

func (a *App) GetEventByID(ctx context.Context, id string) (*storage.Event, error) {
	event, err := a.storage.GetEventByID(ctx, id)
	if err != nil {
		a.logger.Errorf("failed to get event id=%s: %v", id, err)
		return nil, err
	}
	return event, nil
}

func (a *App) GetEventsForDay(ctx context.Context, date time.Time) ([]storage.Event, error) {
	events, err := a.storage.GetEventsForDay(ctx, date)
	if err != nil {
		a.logger.Errorf("failed to get events for day %s: %v", date.Format("2006-01-02"), err)
		return nil, err
	}
	return events, nil
}

func (a *App) GetEventsForWeek(ctx context.Context, startDate time.Time) ([]storage.Event, error) {
	events, err := a.storage.GetEventsForWeek(ctx, startDate)
	if err != nil {
		a.logger.Errorf("failed to get events for week %s: %v", startDate.Format("2006-01-02"), err)
		return nil, err
	}
	return events, nil
}

func (a *App) GetEventsForMonth(ctx context.Context, startDate time.Time) ([]storage.Event, error) {
	events, err := a.storage.GetEventsForMonth(ctx, startDate)
	if err != nil {
		a.logger.Errorf("failed to get events for month %s: %v", startDate.Format("2006-01-02"), err)
		return nil, err
	}
	return events, nil
}
