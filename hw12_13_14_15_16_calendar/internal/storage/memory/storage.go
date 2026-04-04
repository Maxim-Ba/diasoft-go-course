package memorystorage

import (
	"context"
	"sync"
	"time"

	"github.com/Maxim-Ba/hw12_13_14_15_calendar/internal/storage"
	"github.com/google/uuid"
)

// Storage реализует хранилище событий в памяти с потокобезопасностью.
type Storage struct {
	mu     sync.RWMutex
	events map[string]storage.Event // хранилище событий по ID
}

// New создает новое in-memory хранилище.
func New() *Storage {
	return &Storage{
		events: make(map[string]storage.Event),
	}
}

// CreateEvent создает новое событие в хранилище и возвращает сгенерированный ID.
func (s *Storage) CreateEvent(ctx context.Context, event storage.Event) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if event.Title == "" {
		return "", storage.ErrInvalidEvent
	}

	event.ID = uuid.New().String()

	// проверяем, что время не занято другим событием того же пользователя
	if s.isTimeBusy(event) {
		return "", storage.ErrDateBusy
	}

	s.events[event.ID] = event
	return event.ID, nil
}

// UpdateEvent обновляет существующее событие по ID.
func (s *Storage) UpdateEvent(ctx context.Context, id string, event storage.Event) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.events[id]; !exists {
		return storage.ErrEventNotFound
	}

	if event.Title == "" {
		return storage.ErrInvalidEvent
	}

	if s.isTimeBusyExcept(event, id) {
		return storage.ErrDateBusy
	}

	event.ID = id
	s.events[id] = event
	return nil
}

// DeleteEvent удаляет событие по ID.
func (s *Storage) DeleteEvent(ctx context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.events[id]; !exists {
		return storage.ErrEventNotFound
	}

	delete(s.events, id)
	return nil
}

// GetEventByID возвращает событие по ID.
func (s *Storage) GetEventByID(ctx context.Context, id string) (*storage.Event, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	event, exists := s.events[id]
	if !exists {
		return nil, storage.ErrEventNotFound
	}

	return &event, nil
}

// GetEventsForDay возвращает список событий на указанный день.
func (s *Storage) GetEventsForDay(ctx context.Context, date time.Time) ([]storage.Event, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	start := startOfDay(date)
	end := start.Add(24 * time.Hour)

	return s.getEventsBetween(start, end), nil
}

// GetEventsForWeek возвращает список событий на неделю начиная с указанной даты.
func (s *Storage) GetEventsForWeek(ctx context.Context, startDate time.Time) ([]storage.Event, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	start := startOfDay(startDate)
	end := start.Add(7 * 24 * time.Hour)

	return s.getEventsBetween(start, end), nil
}

// GetEventsForMonth возвращает список событий на месяц начиная с указанной даты.
func (s *Storage) GetEventsForMonth(ctx context.Context, startDate time.Time) ([]storage.Event, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	start := startOfDay(startDate)
	end := start.AddDate(0, 1, 0)

	return s.getEventsBetween(start, end), nil
}

// getEventsBetween возвращает события в указанном временном промежутке.
// Метод НЕ потокобезопасен, вызывающий код должен держать блокировку.
func (s *Storage) getEventsBetween(start, end time.Time) []storage.Event {
	var result []storage.Event

	for _, event := range s.events {
		eventEnd := event.StartTime.Add(event.Duration)

		// Проверяем пересечение временных интервалов.
		if event.StartTime.Before(end) && eventEnd.After(start) {
			result = append(result, event)
		}
	}

	return result
}

// isTimeBusy проверяет, занято ли время события другим событием того же пользователя.
// Метод НЕ потокобезопасен, вызывающий код должен держать блокировку.
func (s *Storage) isTimeBusy(newEvent storage.Event) bool {
	newStart := newEvent.StartTime
	newEnd := newStart.Add(newEvent.Duration)

	for _, event := range s.events {
		// проверяем только события того же пользователя
		if event.UserID != newEvent.UserID {
			continue
		}

		eventStart := event.StartTime
		eventEnd := eventStart.Add(event.Duration)

		// проверяем пересечение временных интервалов
		if newStart.Before(eventEnd) && newEnd.After(eventStart) {
			return true
		}
	}

	return false
}

// isTimeBusyExcept проверяет, занято ли время события другим событием того же пользователя,
// исключая событие с указанным ID.
// Метод НЕ потокобезопасен, вызывающий код должен держать блокировку.
func (s *Storage) isTimeBusyExcept(newEvent storage.Event, exceptID string) bool {
	newStart := newEvent.StartTime
	newEnd := newStart.Add(newEvent.Duration)

	for id, event := range s.events {
		// пропускаем событие, которое обновляется
		if id == exceptID {
			continue
		}

		if event.UserID != newEvent.UserID {
			continue
		}

		eventStart := event.StartTime
		eventEnd := eventStart.Add(event.Duration)

		// проверяем пересечение временных интервалов
		if newStart.Before(eventEnd) && newEnd.After(eventStart) {
			return true
		}
	}

	return false
}

// GetEventsBetween возвращает события в указанном временном промежутке.
func (s *Storage) GetEventsBetween(start, end time.Time) []storage.Event {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.getEventsBetween(start, end)
}

// IsTimeBusy проверяет, занято ли время события другим событием того же пользователя.
func (s *Storage) IsTimeBusy(event storage.Event) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.isTimeBusy(event)
}

// IsTimeBusyExcept проверяет, занято ли время события другим событием того же пользователя,
// исключая событие с указанным ID.
func (s *Storage) IsTimeBusyExcept(event storage.Event, exceptID string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.isTimeBusyExcept(event, exceptID)
}

// startOfDay возвращает начало дня (00:00:00) для указанной даты.
func startOfDay(t time.Time) time.Time {
	year, month, day := t.Date()
	return time.Date(year, month, day, 0, 0, 0, 0, t.Location())
}
