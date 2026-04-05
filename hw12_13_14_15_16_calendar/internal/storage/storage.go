package storage

import (
	"context"
	"errors"
	"time"
)

// Ошибки БЛ хранилища.
var (
	ErrEventNotFound = errors.New("event not found")
	ErrDateBusy      = errors.New("date/time is busy")
	ErrInvalidEvent  = errors.New("invalid event data")
)

type EventCreator interface {
	// CreateEvent создает новое событие в хранилище и возвращает сгенерированный ID.
	CreateEvent(ctx context.Context, event Event) (string, error)
}

type EventUpdater interface {
	// UpdateEvent обновляет существующее событие по ID.
	UpdateEvent(ctx context.Context, id string, event Event) error
}

type EventDeleter interface {
	// DeleteEvent удаляет событие по ID.
	DeleteEvent(ctx context.Context, id string) error
}

type EventGetter interface {
	// GetEventByID возвращает событие по ID.
	GetEventByID(ctx context.Context, id string) (*Event, error)

	// GetEventsForDay возвращает список событий на указанный день.
	GetEventsForDay(ctx context.Context, date time.Time) ([]Event, error)

	// GetEventsForWeek возвращает список событий на неделю начиная с указанной даты.
	GetEventsForWeek(ctx context.Context, startDate time.Time) ([]Event, error)

	// GetEventsForMonth возвращает список событий на месяц начиная с указанной даты.
	GetEventsForMonth(ctx context.Context, startDate time.Time) ([]Event, error)
}

type Storage interface {
	EventCreator
	EventUpdater
	EventDeleter
	EventGetter
}
