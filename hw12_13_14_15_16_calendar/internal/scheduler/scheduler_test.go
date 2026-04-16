package scheduler

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Maxim-Ba/hw12_13_14_15_calendar/internal/storage"
)

// moks

type mockStorage struct {
	events            []storage.Event
	listEventsErr     error
	deleteEventErr    error
	markNotifiedErr   error
	deletedIDs        []string
	markedNotifiedIDs []string
}

func (m *mockStorage) ListEvents(_ context.Context, _, _ time.Time) ([]storage.Event, error) {
	return m.events, m.listEventsErr
}

func (m *mockStorage) DeleteEvent(_ context.Context, id string) error {
	m.deletedIDs = append(m.deletedIDs, id)
	return m.deleteEventErr
}

func (m *mockStorage) MarkEventNotified(_ context.Context, id string) error {
	m.markedNotifiedIDs = append(m.markedNotifiedIDs, id)
	return m.markNotifiedErr
}

type mockProducer struct {
	sentMessages int
	sendErr      error
}

func (m *mockProducer) SendMessage(_ context.Context, _ string, _ string, _ []byte) error {
	m.sentMessages++
	return m.sendErr
}

func (m *mockProducer) Close() error { return nil }

type mockLogger struct{}

func (l *mockLogger) Info(_ string)  {}
func (l *mockLogger) Error(_ string) {}

// Вспомогательная

func newTestScheduler(stor *mockStorage, prod *mockProducer) *Scheduler {
	return New(stor, prod, "test-topic", &mockLogger{})
}

// eventInWindow возвращает событие, notifyAt которого попадает в окно [now+30s, now+30m].
func eventInWindow(id string) storage.Event {
	notifyOffset := 30 * time.Minute
	startTime := time.Now().Add(notifyOffset + 30*time.Second)
	return storage.Event{
		ID:               id,
		Title:            "Test Event",
		UserID:           "user1",
		StartTime:        startTime,
		NotificationTime: notifyOffset,
	}
}

// Тесты ScanAndNotify

func TestScanAndNotify_NoEvents(t *testing.T) {
	stor := &mockStorage{events: []storage.Event{}}
	prod := &mockProducer{}
	sched := newTestScheduler(stor, prod)

	if err := sched.ScanAndNotify(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if prod.sentMessages != 0 {
		t.Errorf("expected 0 messages sent, got %d", prod.sentMessages)
	}
}

func TestScanAndNotify_NoNotificationTime(t *testing.T) {
	event := eventInWindow("evt-1")
	event.NotificationTime = 0
	stor := &mockStorage{events: []storage.Event{event}}
	prod := &mockProducer{}
	sched := newTestScheduler(stor, prod)

	if err := sched.ScanAndNotify(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if prod.sentMessages != 0 {
		t.Errorf("expected 0 messages sent, got %d", prod.sentMessages)
	}
}

func TestScanAndNotify_NotifyWindowHit(t *testing.T) {
	event := eventInWindow("evt-2")
	stor := &mockStorage{events: []storage.Event{event}}
	prod := &mockProducer{}
	sched := newTestScheduler(stor, prod)

	if err := sched.ScanAndNotify(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if prod.sentMessages != 1 {
		t.Errorf("expected 1 message sent, got %d", prod.sentMessages)
	}
	if len(stor.markedNotifiedIDs) != 1 || stor.markedNotifiedIDs[0] != "evt-2" {
		t.Errorf("expected MarkEventNotified called for evt-2, got %v", stor.markedNotifiedIDs)
	}
}

func TestScanAndNotify_AlreadyNotified(t *testing.T) {
	event := eventInWindow("evt-3")
	now := time.Now()
	event.NotifiedAt = &now
	stor := &mockStorage{events: []storage.Event{event}}
	prod := &mockProducer{}
	sched := newTestScheduler(stor, prod)

	if err := sched.ScanAndNotify(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if prod.sentMessages != 0 {
		t.Errorf("expected 0 messages sent for already-notified event, got %d", prod.sentMessages)
	}
	if len(stor.markedNotifiedIDs) != 0 {
		t.Errorf("expected MarkEventNotified not called, got %v", stor.markedNotifiedIDs)
	}
}

func TestScanAndNotify_NotifyWindowMiss(t *testing.T) {
	// notifyAt будет через 3 часа — вне окна [now, now+1h]
	startTime := time.Now().Add(4 * time.Hour)
	event := storage.Event{
		ID:               "evt-4",
		Title:            "Future Event",
		UserID:           "user1",
		StartTime:        startTime,
		NotificationTime: 1 * time.Hour,
	}
	stor := &mockStorage{events: []storage.Event{event}}
	prod := &mockProducer{}
	sched := newTestScheduler(stor, prod)

	if err := sched.ScanAndNotify(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if prod.sentMessages != 0 {
		t.Errorf("expected 0 messages sent (window miss), got %d", prod.sentMessages)
	}
}

func TestScanAndNotify_ListEventsError(t *testing.T) {
	stor := &mockStorage{listEventsErr: errors.New("db error")}
	prod := &mockProducer{}
	sched := newTestScheduler(stor, prod)

	err := sched.ScanAndNotify(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestScanAndNotify_MultipleEvents_OnlyUnnotified(t *testing.T) {
	notified := eventInWindow("evt-notified")
	now := time.Now()
	notified.NotifiedAt = &now

	unnotified := eventInWindow("evt-unnotified")

	stor := &mockStorage{events: []storage.Event{notified, unnotified}}
	prod := &mockProducer{}
	sched := newTestScheduler(stor, prod)

	if err := sched.ScanAndNotify(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if prod.sentMessages != 1 {
		t.Errorf("expected 1 message sent, got %d", prod.sentMessages)
	}
	if len(stor.markedNotifiedIDs) != 1 || stor.markedNotifiedIDs[0] != "evt-unnotified" {
		t.Errorf("expected only evt-unnotified marked, got %v", stor.markedNotifiedIDs)
	}
}

//  Тесты CleanupOldEvents

func TestCleanupOldEvents_DeletesOldEvents(t *testing.T) {
	oldEvents := []storage.Event{
		{ID: "old-1", Title: "Old 1"},
		{ID: "old-2", Title: "Old 2"},
	}
	stor := &mockStorage{events: oldEvents}
	prod := &mockProducer{}
	sched := newTestScheduler(stor, prod)

	if err := sched.CleanupOldEvents(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(stor.deletedIDs) != 2 {
		t.Errorf("expected 2 deletions, got %d", len(stor.deletedIDs))
	}
}

func TestCleanupOldEvents_NoOldEvents(t *testing.T) {
	stor := &mockStorage{events: []storage.Event{}}
	prod := &mockProducer{}
	sched := newTestScheduler(stor, prod)

	if err := sched.CleanupOldEvents(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(stor.deletedIDs) != 0 {
		t.Errorf("expected 0 deletions, got %d", len(stor.deletedIDs))
	}
}

func TestCleanupOldEvents_ListError(t *testing.T) {
	stor := &mockStorage{listEventsErr: errors.New("db error")}
	prod := &mockProducer{}
	sched := newTestScheduler(stor, prod)

	err := sched.CleanupOldEvents(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
