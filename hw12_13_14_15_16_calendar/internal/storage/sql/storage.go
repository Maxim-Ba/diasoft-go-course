package sqlstorage

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"github.com/Maxim-Ba/hw12_13_14_15_calendar/internal/storage"
)

type Storage struct {
	db *sqlx.DB
}

type dbEvent struct {
	ID               string         `db:"id"`
	Title            string         `db:"title"`
	StartTime        time.Time      `db:"start_time"`
	Duration         int64          `db:"duration"`
	Description      sql.NullString `db:"description"`
	UserID           string         `db:"user_id"`
	NotificationTime sql.NullInt64  `db:"notification_time"`
	CreatedAt        time.Time      `db:"created_at"`
	UpdatedAt        time.Time      `db:"updated_at"`
}

func New(dsn string) (*Storage, error) {
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	if err := RunMigrations(db.DB); err != nil {
		err := db.Close()
		if err != nil {
			return nil, fmt.Errorf("failed to close database: %w", err)
		}
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return &Storage{db: db}, nil
}

func (s *Storage) Close() error {
	return s.db.Close()
}

func toDBEvent(e storage.Event) dbEvent {
	dbe := dbEvent{
		ID:        e.ID,
		Title:     e.Title,
		StartTime: e.StartTime,
		Duration:  int64(e.Duration),
		UserID:    e.UserID,
	}

	if e.Description != "" {
		dbe.Description = sql.NullString{String: e.Description, Valid: true}
	}

	if e.NotificationTime > 0 {
		dbe.NotificationTime = sql.NullInt64{Int64: int64(e.NotificationTime), Valid: true}
	}

	return dbe
}

func fromDBEvent(dbe dbEvent) storage.Event {
	e := storage.Event{
		ID:        dbe.ID,
		Title:     dbe.Title,
		StartTime: dbe.StartTime,
		Duration:  time.Duration(dbe.Duration),
		UserID:    dbe.UserID,
	}

	if dbe.Description.Valid {
		e.Description = dbe.Description.String
	}

	if dbe.NotificationTime.Valid {
		e.NotificationTime = time.Duration(dbe.NotificationTime.Int64)
	}

	return e
}

func (s *Storage) CreateEvent(ctx context.Context, event storage.Event) (string, error) {
	if event.Title == "" {
		return "", storage.ErrInvalidEvent
	}

	event.ID = uuid.New().String()

	if err := s.checkTimeBusy(ctx, event); err != nil {
		return "", err
	}

	dbe := toDBEvent(event)

	query := `
		INSERT INTO events (id, title, start_time, duration, description, user_id, notification_time)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err := s.db.ExecContext(ctx, query,
		dbe.ID, dbe.Title, dbe.StartTime, dbe.Duration,
		dbe.Description, dbe.UserID, dbe.NotificationTime,
	)
	if err != nil {
		return "", fmt.Errorf("failed to create event: %w", err)
	}

	return event.ID, nil
}

func (s *Storage) UpdateEvent(ctx context.Context, id string, event storage.Event) error {
	var exists bool
	err := s.db.GetContext(ctx, &exists, "SELECT EXISTS(SELECT 1 FROM events WHERE id = $1)", id)
	if err != nil {
		return fmt.Errorf("failed to check event existence: %w", err)
	}
	if !exists {
		return storage.ErrEventNotFound
	}

	event.ID = id

	if err := s.checkTimeBusy(ctx, event); err != nil {
		return err
	}

	dbe := toDBEvent(event)

	query := `
		UPDATE events 
		SET title = $2, start_time = $3, duration = $4, 
		    description = $5, user_id = $6, notification_time = $7, 
		    updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
	`

	_, err = s.db.ExecContext(ctx, query,
		dbe.ID, dbe.Title, dbe.StartTime, dbe.Duration,
		dbe.Description, dbe.UserID, dbe.NotificationTime,
	)
	if err != nil {
		return fmt.Errorf("failed to update event: %w", err)
	}

	return nil
}

func (s *Storage) DeleteEvent(ctx context.Context, id string) error {
	result, err := s.db.ExecContext(ctx, "DELETE FROM events WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("failed to delete event: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return storage.ErrEventNotFound
	}

	return nil
}

func (s *Storage) GetEventByID(ctx context.Context, id string) (*storage.Event, error) {
	var dbe dbEvent

	query := "SELECT * FROM events WHERE id = $1"
	err := s.db.GetContext(ctx, &dbe, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, storage.ErrEventNotFound
		}
		return nil, fmt.Errorf("failed to get event: %w", err)
	}

	event := fromDBEvent(dbe)
	return &event, nil
}

func (s *Storage) GetEventsForDay(ctx context.Context, day time.Time) ([]storage.Event, error) {
	startOfDay := time.Date(day.Year(), day.Month(), day.Day(), 0, 0, 0, 0, day.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	return s.getEventsBetween(ctx, startOfDay, endOfDay)
}

func (s *Storage) GetEventsForWeek(ctx context.Context, week time.Time) ([]storage.Event, error) {
	startOfWeek := time.Date(week.Year(), week.Month(), week.Day(), 0, 0, 0, 0, week.Location())
	endOfWeek := startOfWeek.Add(7 * 24 * time.Hour)

	return s.getEventsBetween(ctx, startOfWeek, endOfWeek)
}

func (s *Storage) GetEventsForMonth(ctx context.Context, month time.Time) ([]storage.Event, error) {
	startOfMonth := time.Date(month.Year(), month.Month(), 1, 0, 0, 0, 0, month.Location())
	endOfMonth := startOfMonth.AddDate(0, 1, 0)

	return s.getEventsBetween(ctx, startOfMonth, endOfMonth)
}

func (s *Storage) getEventsBetween(ctx context.Context, start, end time.Time) ([]storage.Event, error) {
	var dbEvents []dbEvent

	query := `
		SELECT * FROM events 
		WHERE start_time >= $1 AND start_time < $2
		ORDER BY start_time
	`

	err := s.db.SelectContext(ctx, &dbEvents, query, start, end)
	if err != nil {
		return nil, fmt.Errorf("failed to get events: %w", err)
	}

	events := make([]storage.Event, 0, len(dbEvents))
	for _, dbe := range dbEvents {
		events = append(events, fromDBEvent(dbe))
	}

	return events, nil
}

func (s *Storage) checkTimeBusy(ctx context.Context, event storage.Event) error {
	eventEnd := event.StartTime.Add(event.Duration)

	query := `
		SELECT COUNT(*) FROM events 
		WHERE user_id = $1 
		  AND id != $2
		  AND start_time < $3 
		  AND (start_time + duration * INTERVAL '1 nanosecond') > $4
	`

	var count int
	err := s.db.GetContext(ctx, &count, query,
		event.UserID, event.ID, eventEnd, event.StartTime,
	)
	if err != nil {
		return fmt.Errorf("failed to check time busy: %w", err)
	}

	if count > 0 {
		return storage.ErrDateBusy
	}

	return nil
}

func (s *Storage) ListEvents(ctx context.Context, from, to time.Time) ([]storage.Event, error) {
	var dbEvents []dbEvent

	query := `
		SELECT * FROM events 
		WHERE start_time >= $1 AND start_time <= $2
		ORDER BY start_time
	`

	err := s.db.SelectContext(ctx, &dbEvents, query, from, to)
	if err != nil {
		return nil, fmt.Errorf("failed to list events: %w", err)
	}

	events := make([]storage.Event, 0, len(dbEvents))
	for _, dbe := range dbEvents {
		events = append(events, fromDBEvent(dbe))
	}

	return events, nil
}

func (s *Storage) SaveNotification(ctx context.Context, notification storage.Notification) error {
	query := `
		INSERT INTO notifications (event_id, title, event_date, user_id, created_at)
		VALUES ($1, $2, $3, $4, CURRENT_TIMESTAMP)
	`

	_, err := s.db.ExecContext(ctx, query,
		notification.EventID, notification.Title, notification.EventDate, notification.UserID,
	)
	if err != nil {
		return fmt.Errorf("failed to save notification: %w", err)
	}

	return nil
}
