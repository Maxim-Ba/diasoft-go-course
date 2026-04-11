package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/Maxim-Ba/hw12_13_14_15_calendar/internal/server/http/generated"
	"github.com/Maxim-Ba/hw12_13_14_15_calendar/internal/storage"
	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

type Application interface {
	storage.EventCreator
	storage.EventUpdater
	storage.EventDeleter
	storage.EventGetter
}

// EventsHandler реализует generated.ServerInterface.
type EventsHandler struct {
	app Application
}

func NewEventsHandler(app Application) *EventsHandler {
	return &EventsHandler{app: app}
}

// CreateEvent POST /events.
func (h *EventsHandler) CreateEvent(w http.ResponseWriter, r *http.Request) {
	var input generated.EventInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}

	event := inputToEvent(input)
	id, err := h.app.CreateEvent(r.Context(), event)
	if err != nil {
		writeAppError(w, err)
		return
	}

	event.ID = id
	writeJSON(w, http.StatusCreated, eventToResponse(event))
}

// UpdateEvent PUT  /events/{id}.
func (h *EventsHandler) UpdateEvent(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	var input generated.EventInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}

	event := inputToEvent(input)
	if err := h.app.UpdateEvent(r.Context(), id.String(), event); err != nil {
		writeAppError(w, err)
		return
	}

	event.ID = id.String()
	writeJSON(w, http.StatusOK, eventToResponse(event))
}

// DeleteEvent DELETE  /events/{id}.
func (h *EventsHandler) DeleteEvent(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	if err := h.app.DeleteEvent(r.Context(), id.String()); err != nil {
		writeAppError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// GetEventsForDay GET /events/day?date=YYYY-MM-DD.
func (h *EventsHandler) GetEventsForDay(w http.ResponseWriter, r *http.Request, params generated.GetEventsForDayParams) {
	events, err := h.app.GetEventsForDay(r.Context(), params.Date.Time)
	if err != nil {
		writeAppError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, generated.EventListResponse{Events: eventsToResponse(events)})
}

// GetEventsForWeek GET /events/week?date=YYYY-MM-DD.
func (h *EventsHandler) GetEventsForWeek(w http.ResponseWriter, r *http.Request, params generated.GetEventsForWeekParams) {
	events, err := h.app.GetEventsForWeek(r.Context(), params.Date.Time)
	if err != nil {
		writeAppError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, generated.EventListResponse{Events: eventsToResponse(events)})
}

// GetEventsForMonth GET /events/month?date=YYYY-MM-DD.
func (h *EventsHandler) GetEventsForMonth(w http.ResponseWriter, r *http.Request, params generated.GetEventsForMonthParams) {
	events, err := h.app.GetEventsForMonth(r.Context(), params.Date.Time)
	if err != nil {
		writeAppError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, generated.EventListResponse{Events: eventsToResponse(events)})
}

func inputToEvent(input generated.EventInput) storage.Event {
	event := storage.Event{
		Title:     input.Title,
		StartTime: input.StartTime,
		Duration:  time.Duration(input.Duration) * time.Second,
		UserID:    input.UserId,
	}
	if input.Description != nil {
		event.Description = *input.Description
	}
	if input.NotificationTime != nil {
		event.NotificationTime = time.Duration(*input.NotificationTime) * time.Second
	}
	return event
}

func eventToResponse(e storage.Event) generated.EventResponse {
	resp := generated.EventResponse{
		Title:     e.Title,
		StartTime: e.StartTime,
		Duration:  int64(e.Duration / time.Second),
		UserId:    e.UserID,
	}

	if parsed, err := uuid.Parse(e.ID); err == nil {
		resp.Id = openapi_types.UUID(parsed)
	}

	if e.Description != "" {
		resp.Description = &e.Description
	}
	if e.NotificationTime > 0 {
		nt := int64(e.NotificationTime / time.Second)
		resp.NotificationTime = &nt
	}
	return resp
}

func eventsToResponse(events []storage.Event) []generated.EventResponse {
	if events == nil {
		return []generated.EventResponse{}
	}
	result := make([]generated.EventResponse, 0, len(events))
	for _, e := range events {
		result = append(result, eventToResponse(e))
	}
	return result
}

// writeJSON сериализует v в JSON и пишет в ResponseWriter.
func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// writeError пишет JSON-ответ с ошибкой.
func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, generated.ErrorResponse{Error: msg})
}

// writeAppError отображает ошибки бизнес-логики на HTTP-статусы.
func writeAppError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, storage.ErrEventNotFound):
		writeError(w, http.StatusNotFound, err.Error())
	case errors.Is(err, storage.ErrDateBusy):
		writeError(w, http.StatusConflict, err.Error())
	case errors.Is(err, storage.ErrInvalidEvent):
		writeError(w, http.StatusBadRequest, err.Error())
	default:
		writeError(w, http.StatusInternalServerError, "internal server error")
	}
}
