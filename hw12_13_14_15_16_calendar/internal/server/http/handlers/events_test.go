package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Maxim-Ba/hw12_13_14_15_calendar/internal/server/http/generated"
	"github.com/Maxim-Ba/hw12_13_14_15_calendar/internal/server/http/handlers"
	"github.com/Maxim-Ba/hw12_13_14_15_calendar/internal/storage"
)

// mockApp — тестовая реализация handlers.Application.
type mockApp struct {
	createFn       func(ctx context.Context, event storage.Event) (string, error)
	updateFn       func(ctx context.Context, id string, event storage.Event) error
	deleteFn       func(ctx context.Context, id string) error
	getByIDFn      func(ctx context.Context, id string) (*storage.Event, error)
	getForDayFn    func(ctx context.Context, date time.Time) ([]storage.Event, error)
	getForWeekFn   func(ctx context.Context, startDate time.Time) ([]storage.Event, error)
	getForMonthFn  func(ctx context.Context, startDate time.Time) ([]storage.Event, error)
}

func (m *mockApp) CreateEvent(ctx context.Context, event storage.Event) (string, error) {
	return m.createFn(ctx, event)
}
func (m *mockApp) UpdateEvent(ctx context.Context, id string, event storage.Event) error {
	return m.updateFn(ctx, id, event)
}
func (m *mockApp) DeleteEvent(ctx context.Context, id string) error {
	return m.deleteFn(ctx, id)
}
func (m *mockApp) GetEventByID(ctx context.Context, id string) (*storage.Event, error) {
	return m.getByIDFn(ctx, id)
}
func (m *mockApp) GetEventsForDay(ctx context.Context, date time.Time) ([]storage.Event, error) {
	return m.getForDayFn(ctx, date)
}
func (m *mockApp) GetEventsForWeek(ctx context.Context, startDate time.Time) ([]storage.Event, error) {
	return m.getForWeekFn(ctx, startDate)
}
func (m *mockApp) GetEventsForMonth(ctx context.Context, startDate time.Time) ([]storage.Event, error) {
	return m.getForMonthFn(ctx, startDate)
}

// newTestHandler создаёт http.Handler с зарегистрированными маршрутами.
func newTestHandler(app handlers.Application) http.Handler {
	h := handlers.NewEventsHandler(app)
	return generated.Handler(h)
}

// jsonBody сериализует v в io.Reader для использования в запросах.
func jsonBody(t *testing.T, v interface{}) *bytes.Buffer {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("jsonBody: %v", err)
	}
	return bytes.NewBuffer(b)
}

var testTime = time.Date(2025, 6, 1, 10, 0, 0, 0, time.UTC)

var testInput = generated.EventInput{
	Title:     "Meeting",
	StartTime: testTime,
	Duration:  3600,
	UserId:    "user-1",
}

var testEvent = storage.Event{
	ID:        "00000000-0000-0000-0000-000000000001",
	Title:     "Meeting",
	StartTime: testTime,
	Duration:  3600 * time.Second,
	UserID:    "user-1",
}

// --- CreateEvent ---

func TestCreateEvent_Success(t *testing.T) {
	app := &mockApp{
		createFn: func(_ context.Context, _ storage.Event) (string, error) {
			return testEvent.ID, nil
		},
	}
	handler := newTestHandler(app)

	req := httptest.NewRequest(http.MethodPost, "/events", jsonBody(t, testInput))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d: %s", w.Code, w.Body.String())
	}

	var resp generated.EventResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Title != testInput.Title {
		t.Errorf("expected title %q, got %q", testInput.Title, resp.Title)
	}
	if resp.UserId != testInput.UserId {
		t.Errorf("expected userId %q, got %q", testInput.UserId, resp.UserId)
	}
}

func TestCreateEvent_InvalidJSON(t *testing.T) {
	app := &mockApp{}
	handler := newTestHandler(app)

	req := httptest.NewRequest(http.MethodPost, "/events", bytes.NewBufferString("{bad json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestCreateEvent_DateBusy(t *testing.T) {
	app := &mockApp{
		createFn: func(_ context.Context, _ storage.Event) (string, error) {
			return "", storage.ErrDateBusy
		},
	}
	handler := newTestHandler(app)

	req := httptest.NewRequest(http.MethodPost, "/events", jsonBody(t, testInput))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusConflict {
		t.Errorf("expected 409, got %d", w.Code)
	}
}

func TestCreateEvent_InvalidEvent(t *testing.T) {
	app := &mockApp{
		createFn: func(_ context.Context, _ storage.Event) (string, error) {
			return "", storage.ErrInvalidEvent
		},
	}
	handler := newTestHandler(app)

	req := httptest.NewRequest(http.MethodPost, "/events", jsonBody(t, testInput))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

// --- UpdateEvent ---

func TestUpdateEvent_Success(t *testing.T) {
	app := &mockApp{
		updateFn: func(_ context.Context, _ string, _ storage.Event) error { return nil },
	}
	handler := newTestHandler(app)

	id := "00000000-0000-0000-0000-000000000001"
	req := httptest.NewRequest(http.MethodPut, "/events/"+id, jsonBody(t, testInput))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp generated.EventResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Title != testInput.Title {
		t.Errorf("expected title %q, got %q", testInput.Title, resp.Title)
	}
}

func TestUpdateEvent_NotFound(t *testing.T) {
	app := &mockApp{
		updateFn: func(_ context.Context, _ string, _ storage.Event) error {
			return storage.ErrEventNotFound
		},
	}
	handler := newTestHandler(app)

	id := "00000000-0000-0000-0000-000000000099"
	req := httptest.NewRequest(http.MethodPut, "/events/"+id, jsonBody(t, testInput))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestUpdateEvent_InvalidJSON(t *testing.T) {
	app := &mockApp{}
	handler := newTestHandler(app)

	id := "00000000-0000-0000-0000-000000000001"
	req := httptest.NewRequest(http.MethodPut, "/events/"+id, bytes.NewBufferString("bad"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

// --- DeleteEvent ---

func TestDeleteEvent_Success(t *testing.T) {
	app := &mockApp{
		deleteFn: func(_ context.Context, _ string) error { return nil },
	}
	handler := newTestHandler(app)

	id := "00000000-0000-0000-0000-000000000001"
	req := httptest.NewRequest(http.MethodDelete, "/events/"+id, nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", w.Code)
	}
}

func TestDeleteEvent_NotFound(t *testing.T) {
	app := &mockApp{
		deleteFn: func(_ context.Context, _ string) error { return storage.ErrEventNotFound },
	}
	handler := newTestHandler(app)

	id := "00000000-0000-0000-0000-000000000099"
	req := httptest.NewRequest(http.MethodDelete, "/events/"+id, nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

// --- GetEventsForDay ---

func TestGetEventsForDay_Success(t *testing.T) {
	app := &mockApp{
		getForDayFn: func(_ context.Context, _ time.Time) ([]storage.Event, error) {
			return []storage.Event{testEvent}, nil
		},
	}
	handler := newTestHandler(app)

	req := httptest.NewRequest(http.MethodGet, "/events/day?date=2025-06-01", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp generated.EventListResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(resp.Events) != 1 {
		t.Errorf("expected 1 event, got %d", len(resp.Events))
	}
	if resp.Events[0].Title != testEvent.Title {
		t.Errorf("expected title %q, got %q", testEvent.Title, resp.Events[0].Title)
	}
}

func TestGetEventsForDay_MissingDate(t *testing.T) {
	app := &mockApp{}
	handler := newTestHandler(app)

	req := httptest.NewRequest(http.MethodGet, "/events/day", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestGetEventsForDay_Empty(t *testing.T) {
	app := &mockApp{
		getForDayFn: func(_ context.Context, _ time.Time) ([]storage.Event, error) {
			return nil, nil
		},
	}
	handler := newTestHandler(app)

	req := httptest.NewRequest(http.MethodGet, "/events/day?date=2025-06-01", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	var resp generated.EventListResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(resp.Events) != 0 {
		t.Errorf("expected 0 events, got %d", len(resp.Events))
	}
}

// --- GetEventsForWeek ---

func TestGetEventsForWeek_Success(t *testing.T) {
	app := &mockApp{
		getForWeekFn: func(_ context.Context, _ time.Time) ([]storage.Event, error) {
			return []storage.Event{testEvent}, nil
		},
	}
	handler := newTestHandler(app)

	req := httptest.NewRequest(http.MethodGet, "/events/week?date=2025-06-01", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp generated.EventListResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(resp.Events) != 1 {
		t.Errorf("expected 1 event, got %d", len(resp.Events))
	}
}

// --- GetEventsForMonth ---

func TestGetEventsForMonth_Success(t *testing.T) {
	app := &mockApp{
		getForMonthFn: func(_ context.Context, _ time.Time) ([]storage.Event, error) {
			return []storage.Event{testEvent}, nil
		},
	}
	handler := newTestHandler(app)

	req := httptest.NewRequest(http.MethodGet, "/events/month?date=2025-06-01", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp generated.EventListResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(resp.Events) != 1 {
		t.Errorf("expected 1 event, got %d", len(resp.Events))
	}
}
