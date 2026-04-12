package memorystorage

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/Maxim-Ba/hw12_13_14_15_calendar/internal/storage"
)

// Тест создания события.
func TestCreateEvent(t *testing.T) {
	s := New()
	ctx := context.Background()

	event := storage.Event{
		Title:     "Meeting",
		StartTime: time.Now(),
		Duration:  time.Hour,
		UserID:    "user-1",
	}

	id, err := s.CreateEvent(ctx, event)
	if err != nil {
		t.Fatalf("Failed to create event: %v", err)
	}

	if id == "" {
		t.Error("Expected non-empty ID")
	}

	retrieved, err := s.GetEventByID(ctx, id)
	if err != nil {
		t.Fatalf("Failed to get event: %v", err)
	}

	if retrieved.Title != event.Title {
		t.Errorf("Expected title %s, got %s", event.Title, retrieved.Title)
	}

	if retrieved.ID != id {
		t.Errorf("Expected ID %s, got %s", id, retrieved.ID)
	}
}

// Тест создания невалидного события.
func TestCreateInvalidEvent(t *testing.T) {
	s := New()
	ctx := context.Background()

	// Событие без заголовка.
	event := storage.Event{
		Title:     "",
		StartTime: time.Now(),
		Duration:  time.Hour,
		UserID:    "user-1",
	}

	_, err := s.CreateEvent(ctx, event)
	if err != storage.ErrInvalidEvent {
		t.Errorf("Expected ErrInvalidEvent, got %v", err)
	}
}

// Тест проверки занятости времени.
func TestDateBusy(t *testing.T) {
	s := New()
	ctx := context.Background()

	startTime := time.Now()

	event1 := storage.Event{
		Title:     "Meeting 1",
		StartTime: startTime,
		Duration:  time.Hour,
		UserID:    "user-1",
	}

	_, err := s.CreateEvent(ctx, event1)
	if err != nil {
		t.Fatalf("Failed to create event: %v", err)
	}

	// пытаемся создать событие в то же время для того же пользователя
	event2 := storage.Event{
		Title:     "Meeting 2",
		StartTime: startTime.Add(30 * time.Minute), // пересекается с event1
		Duration:  time.Hour,
		UserID:    "user-1",
	}

	_, err = s.CreateEvent(ctx, event2)
	if err != storage.ErrDateBusy {
		t.Errorf("Expected ErrDateBusy, got %v", err)
	}

	// событие для другого пользователя в то же время должно быть создано.
	event3 := storage.Event{
		Title:     "Meeting 3",
		StartTime: startTime,
		Duration:  time.Hour,
		UserID:    "user-2",
	}

	_, err = s.CreateEvent(ctx, event3)
	if err != nil {
		t.Errorf("Expected success for different user, got %v", err)
	}
}

// Тест обновления события.
func TestUpdateEvent(t *testing.T) {
	s := New()
	ctx := context.Background()

	event := storage.Event{
		Title:     "Meeting",
		StartTime: time.Now(),
		Duration:  time.Hour,
		UserID:    "user-1",
	}

	id, err := s.CreateEvent(ctx, event)
	if err != nil {
		t.Fatalf("Failed to create event: %v", err)
	}

	updatedEvent := event
	updatedEvent.Title = "Updated Meeting"
	updatedEvent.Duration = 2 * time.Hour

	err = s.UpdateEvent(ctx, id, updatedEvent)
	if err != nil {
		t.Fatalf("Failed to update event: %v", err)
	}

	retrieved, err := s.GetEventByID(ctx, id)
	if err != nil {
		t.Fatalf("Failed to get event: %v", err)
	}

	if retrieved.Title != "Updated Meeting" {
		t.Errorf("Expected title 'Updated Meeting', got %s", retrieved.Title)
	}

	if retrieved.Duration != 2*time.Hour {
		t.Errorf("Expected duration 2h, got %v", retrieved.Duration)
	}
}

// Тест обновления несуществующего события.
func TestUpdateNonExistentEvent(t *testing.T) {
	s := New()
	ctx := context.Background()

	event := storage.Event{
		Title:     "Meeting",
		StartTime: time.Now(),
		Duration:  time.Hour,
		UserID:    "user-1",
	}

	err := s.UpdateEvent(ctx, "non-existent", event)
	if err != storage.ErrEventNotFound {
		t.Errorf("Expected ErrEventNotFound, got %v", err)
	}
}

// Тест удаления события.
func TestDeleteEvent(t *testing.T) {
	s := New()
	ctx := context.Background()

	event := storage.Event{
		Title:     "Meeting",
		StartTime: time.Now(),
		Duration:  time.Hour,
		UserID:    "user-1",
	}

	id, err := s.CreateEvent(ctx, event)
	if err != nil {
		t.Fatalf("Failed to create event: %v", err)
	}

	err = s.DeleteEvent(ctx, id)
	if err != nil {
		t.Fatalf("Failed to delete event: %v", err)
	}

	_, err = s.GetEventByID(ctx, id)
	if err != storage.ErrEventNotFound {
		t.Errorf("Expected ErrEventNotFound after delete, got %v", err)
	}
}

// Тест получения событий на день.
func TestGetEventsForDay(t *testing.T) {
	s := New()
	ctx := context.Background()

	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 10, 0, 0, 0, now.Location())
	tomorrow := today.Add(24 * time.Hour)

	event1 := storage.Event{
		Title:     "Meeting 1",
		StartTime: today,
		Duration:  time.Hour,
		UserID:    "user-1",
	}

	event2 := storage.Event{
		Title:     "Meeting 2",
		StartTime: today.Add(3 * time.Hour),
		Duration:  time.Hour,
		UserID:    "user-1",
	}

	event3 := storage.Event{
		Title:     "Meeting 3",
		StartTime: tomorrow,
		Duration:  time.Hour,
		UserID:    "user-1",
	}

	_, err := s.CreateEvent(ctx, event1)
	if err != nil {
		t.Fatalf("Failed to create event 1: %v", err)
	}
	_, err = s.CreateEvent(ctx, event2)
	if err != nil {
		t.Fatalf("Failed to create event 2: %v", err)
	}
	_, err = s.CreateEvent(ctx, event3)
	if err != nil {
		t.Fatalf("Failed to create event 3: %v", err)
	}
	events, err := s.GetEventsForDay(ctx, today)
	if err != nil {
		t.Fatalf("Failed to get events for day: %v", err)
	}

	if len(events) != 2 {
		t.Errorf("Expected 2 events for today, got %d", len(events))
	}
}

// Тест получения событий на неделю.
func TestGetEventsForWeek(t *testing.T) {
	s := New()
	ctx := context.Background()

	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 10, 0, 0, 0, now.Location())

	for i := 0; i < 7; i++ {
		event := storage.Event{
			Title:     "Meeting",
			StartTime: today.Add(time.Duration(i) * 24 * time.Hour),
			Duration:  time.Hour,
			UserID:    "user-1",
		}
		_, err := s.CreateEvent(ctx, event)
		if err != nil {
			t.Fatalf("Failed to get events for day: %v", err)
		}
	}

	eventOutside := storage.Event{
		Title:     "Meeting Outside",
		StartTime: today.Add(8 * 24 * time.Hour),
		Duration:  time.Hour,
		UserID:    "user-1",
	}
	_, err := s.CreateEvent(ctx, eventOutside)
	if err != nil {
		t.Fatalf("Failed to get events for day: %v", err)
	}

	events, err := s.GetEventsForWeek(ctx, today)
	if err != nil {
		t.Fatalf("Failed to get events for week: %v", err)
	}

	if len(events) != 7 {
		t.Errorf("Expected 7 events for week, got %d", len(events))
	}
}

// Тест получения событий на месяц.
func TestGetEventsForMonth(t *testing.T) {
	s := New()
	ctx := context.Background()

	now := time.Now()
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 10, 0, 0, 0, now.Location())

	// создаем события в течение месяца
	for i := 0; i < 30; i++ {
		event := storage.Event{
			Title:     "Meeting",
			StartTime: startOfMonth.Add(time.Duration(i) * 24 * time.Hour),
			Duration:  time.Hour,
			UserID:    "user-1",
		}
		_, err := s.CreateEvent(ctx, event)
		if err != nil {
			t.Fatalf("Failed to get events for day: %v", err)
		}
	}

	// событие в следующем месяце
	nextMonth := startOfMonth.AddDate(0, 1, 5)
	eventOutside := storage.Event{
		Title:     "Meeting Next Month",
		StartTime: nextMonth,
		Duration:  time.Hour,
		UserID:    "user-1",
	}
	_, err := s.CreateEvent(ctx, eventOutside)
	if err != nil {
		t.Fatalf("Failed to get events for day: %v", err)
	}

	events, err := s.GetEventsForMonth(ctx, startOfMonth)
	if err != nil {
		t.Fatalf("Failed to get events for month: %v", err)
	}

	if len(events) != 30 {
		t.Errorf("Expected 30 events for month, got %d", len(events))
	}
}

// Тест потокобезопасности при конкурентной записи/чтении.
func TestConcurrentAccess(t *testing.T) {
	s := New()
	ctx := context.Background()

	var wg sync.WaitGroup
	numGoroutines := 100

	// конкурентное создание событий.
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			event := storage.Event{
				Title:     "Concurrent Meeting",
				StartTime: time.Now().Add(time.Duration(id) * time.Hour),
				Duration:  30 * time.Minute,
				UserID:    "user-1",
			}

			_, err := s.CreateEvent(ctx, event)
			if err != nil {
				t.Errorf("Failed to create event concurrently: %v", err)
			}
		}(i)
	}

	// конкурентное чтение событий
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			_, _ = s.GetEventsForDay(ctx, time.Now())
		}()
	}

	wg.Wait()
}

// Тест потокобезопасности при конкурентном обновлении.
func TestConcurrentUpdate(t *testing.T) {
	s := New()
	ctx := context.Background()

	event := storage.Event{
		Title:     "Meeting",
		StartTime: time.Now(),
		Duration:  time.Hour,
		UserID:    "user-1",
	}

	id, err := s.CreateEvent(ctx, event)
	if err != nil {
		t.Fatalf("Failed to create event: %v", err)
	}

	var wg sync.WaitGroup
	numGoroutines := 50

	// конкурентное обновление одного события
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(num int) {
			defer wg.Done()

			updatedEvent := event
			updatedEvent.Title = "Updated Meeting"
			updatedEvent.StartTime = time.Now().Add(time.Duration(num*2) * time.Hour)

			_ = s.UpdateEvent(ctx, id, updatedEvent)
		}(i)
	}

	wg.Wait()

	// проверяем, что событие все еще существует и не повреждено
	retrieved, err := s.GetEventByID(ctx, id)
	if err != nil {
		t.Fatalf("Failed to get event after concurrent updates: %v", err)
	}

	if retrieved.ID != id {
		t.Errorf("Event ID was corrupted during concurrent updates")
	}
}

// TestGetEventsBetween тест публичного потокобезопасного метода.
func TestGetEventsBetween(t *testing.T) {
	s := New()
	ctx := context.Background()

	now := time.Now()
	start := time.Date(now.Year(), now.Month(), now.Day(), 10, 0, 0, 0, now.Location())

	event1 := storage.Event{
		Title:     "Event 1",
		StartTime: start,
		Duration:  time.Hour,
		UserID:    "user-1",
	}

	event2 := storage.Event{
		Title:     "Event 2",
		StartTime: start.Add(2 * time.Hour),
		Duration:  time.Hour,
		UserID:    "user-1",
	}

	event3 := storage.Event{
		Title:     "Event 3",
		StartTime: start.Add(5 * time.Hour),
		Duration:  time.Hour,
		UserID:    "user-1",
	}

	_, err := s.CreateEvent(ctx, event1)
	if err != nil {
		t.Fatalf("Failed to create event 1: %v", err)
	}
	_, err = s.CreateEvent(ctx, event2)
	if err != nil {
		t.Fatalf("Failed to create event 2: %v", err)
	}
	_, err = s.CreateEvent(ctx, event3)
	if err != nil {
		t.Fatalf("Failed to create event 3: %v", err)
	}

	// получаем события с 10:00 до 14:00 (должны попасть event1 и event2)
	events := s.GetEventsBetween(start, start.Add(4*time.Hour))

	if len(events) != 2 {
		t.Errorf("Expected 2 events, got %d", len(events))
	}
}

// TestGetEventsBetweenConcurrent тест конкурентного доступа к GetEventsBetween.
func TestGetEventsBetweenConcurrent(t *testing.T) {
	s := New()
	ctx := context.Background()

	now := time.Now()
	start := time.Date(now.Year(), now.Month(), now.Day(), 10, 0, 0, 0, now.Location())

	// создаем несколько событий
	for i := 0; i < 10; i++ {
		event := storage.Event{
			Title:     "Event",
			StartTime: start.Add(time.Duration(i) * time.Hour),
			Duration:  30 * time.Minute,
			UserID:    "user-1",
		}
		_, err := s.CreateEvent(ctx, event)
		if err != nil {
			t.Fatalf("Failed to create event: %v", err)
		}
	}

	var wg sync.WaitGroup
	numGoroutines := 100

	// конкурентное чтение через GetEventsBetween
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = s.GetEventsBetween(start, start.Add(24*time.Hour))
		}()
	}

	wg.Wait()
}

// TestIsTimeBusy тест публичного потокобезопасного метода.
func TestIsTimeBusy(t *testing.T) {
	s := New()
	ctx := context.Background()

	now := time.Now()
	start := time.Date(now.Year(), now.Month(), now.Day(), 10, 0, 0, 0, now.Location())

	existingEvent := storage.Event{
		Title:     "Existing Event",
		StartTime: start,
		Duration:  time.Hour,
		UserID:    "user-1",
	}

	_, err := s.CreateEvent(ctx, existingEvent)
	if err != nil {
		t.Fatalf("Failed to create event: %v", err)
	}

	// событие, которое пересекается
	conflictingEvent := storage.Event{
		Title:     "Conflicting Event",
		StartTime: start.Add(30 * time.Minute),
		Duration:  time.Hour,
		UserID:    "user-1",
	}

	if !s.IsTimeBusy(conflictingEvent) {
		t.Error("Expected time to be busy")
	}

	// событие, которое НЕ пересекается
	nonConflictingEvent := storage.Event{
		Title:     "Non-conflicting Event",
		StartTime: start.Add(2 * time.Hour),
		Duration:  time.Hour,
		UserID:    "user-1",
	}

	if s.IsTimeBusy(nonConflictingEvent) {
		t.Error("Expected time to be free")
	}

	// событие другого пользователя в то же время
	differentUserEvent := storage.Event{
		Title:     "Different User Event",
		StartTime: start,
		Duration:  time.Hour,
		UserID:    "user-2",
	}

	if s.IsTimeBusy(differentUserEvent) {
		t.Error("Expected time to be free for different user")
	}
}

// TestIsTimeBusyConcurrent тест конкурентного доступа к IsTimeBusy.
func TestIsTimeBusyConcurrent(t *testing.T) {
	s := New()
	ctx := context.Background()

	now := time.Now()
	start := time.Date(now.Year(), now.Month(), now.Day(), 10, 0, 0, 0, now.Location())

	event := storage.Event{
		Title:     "Test Event",
		StartTime: start,
		Duration:  time.Hour,
		UserID:    "user-1",
	}

	_, err := s.CreateEvent(ctx, event)
	if err != nil {
		t.Fatalf("Failed to create event: %v", err)
	}

	var wg sync.WaitGroup
	numGoroutines := 100

	// конкурентная проверка занятости времени
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(offset int) {
			defer wg.Done()

			testEvent := storage.Event{
				Title:     "Check Event",
				StartTime: start.Add(time.Duration(offset) * time.Minute),
				Duration:  time.Hour,
				UserID:    "user-1",
			}

			_ = s.IsTimeBusy(testEvent)
		}(i)
	}

	wg.Wait()
}

// TestIsTimeBusyExcept тест публичного потокобезопасного метода.
func TestIsTimeBusyExcept(t *testing.T) {
	s := New()
	ctx := context.Background()

	now := time.Now()
	start := time.Date(now.Year(), now.Month(), now.Day(), 10, 0, 0, 0, now.Location())

	event1 := storage.Event{
		Title:     "Event 1",
		StartTime: start,
		Duration:  time.Hour,
		UserID:    "user-1",
	}

	event2 := storage.Event{
		Title:     "Event 2",
		StartTime: start.Add(2 * time.Hour),
		Duration:  time.Hour,
		UserID:    "user-1",
	}

	id1, err := s.CreateEvent(ctx, event1)
	if err != nil {
		t.Fatalf("Failed to create event: %v", err)
	}
	id2, err := s.CreateEvent(ctx, event2)
	if err != nil {
		t.Fatalf("Failed to create event: %v", err)
	}

	// обновляем event1 на то же время - не должно быть конфликта с самим собой
	updatedEvent1 := storage.Event{
		Title:     "Updated Event 1",
		StartTime: start,
		Duration:  time.Hour,
		UserID:    "user-1",
	}

	if s.IsTimeBusyExcept(updatedEvent1, id1) {
		t.Error("Expected time to be free when excluding self")
	}

	// пытаемся переместить event1 на время event2 - должен быть конфликт
	conflictingUpdate := storage.Event{
		Title:     "Conflicting Update",
		StartTime: start.Add(2 * time.Hour),
		Duration:  time.Hour,
		UserID:    "user-1",
	}

	if !s.IsTimeBusyExcept(conflictingUpdate, id1) {
		t.Error("Expected time to be busy with another event")
	}

	// проверяем, что исключение работает для event2
	if s.IsTimeBusyExcept(event2, id2) {
		t.Error("Expected time to be free when excluding event2 itself")
	}
}

// TestIsTimeBusyExceptConcurrent тест конкурентного доступа к IsTimeBusyExcept.
func TestIsTimeBusyExceptConcurrent(t *testing.T) {
	s := New()
	ctx := context.Background()

	now := time.Now()
	start := time.Date(now.Year(), now.Month(), now.Day(), 10, 0, 0, 0, now.Location())

	// создаем несколько событий
	ids := make([]string, 10)
	for i := 0; i < 10; i++ {
		event := storage.Event{
			Title:     "Event",
			StartTime: start.Add(time.Duration(i*2) * time.Hour),
			Duration:  time.Hour,
			UserID:    "user-1",
		}
		id, err := s.CreateEvent(ctx, event)
		if err != nil {
			t.Fatalf("Failed to create event: %v", err)
		}
		ids[i] = id
	}

	var wg sync.WaitGroup
	numGoroutines := 100

	// конкурентная проверка с исключением
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()

			testEvent := storage.Event{
				Title:     "Check Event",
				StartTime: start.Add(time.Duration(idx%10*2) * time.Hour),
				Duration:  time.Hour,
				UserID:    "user-1",
			}

			exceptID := ids[idx%len(ids)]
			_ = s.IsTimeBusyExcept(testEvent, exceptID)
		}(i)
	}

	wg.Wait()
}
