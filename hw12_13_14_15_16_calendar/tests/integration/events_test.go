package integration

import (
	"encoding/json"
	"net/http"
	"time"
)

// EventsSuite тестирует создание/удаление событий и бизнес-ошибки.
type EventsSuite struct {
	CalendarSuite
}

// TestCreateEvent_Success проверяет успешное создание эвента:
// POST /events с корректными данными -> 201 Created.
func (s *EventsSuite) TestCreateEvent_Success() {
	startTime := time.Now().Add(2 * time.Hour).UTC()

	body := eventInput{
		Title:     "Meeting",
		StartTime: startTime.Format(time.RFC3339),
		Duration:  3600,
		UserID:    "user-1",
	}

	status, respBody := s.post("/events", body)
	s.Require().Equal(http.StatusCreated, status, string(respBody))

	var resp eventResponse
	s.Require().NoError(json.Unmarshal(respBody, &resp))
	s.NotEmpty(resp.ID)
	s.Equal("Meeting", resp.Title)
	s.Equal("user-1", resp.UserID)
	s.Equal(int64(3600), resp.Duration)
}

// TestCreateEvent_EmptyTitle_Returns400 проверяет валидацию обязательного поля:
// POST /events с пустым title -> 400 Bad Request с ошибкой.
func (s *EventsSuite) TestCreateEvent_EmptyTitle_Returns400() {
	body := eventInput{
		Title:     "",
		StartTime: time.Now().Add(time.Hour).UTC().Format(time.RFC3339),
		Duration:  3600,
		UserID:    "user-1",
	}

	status, respBody := s.post("/events", body)
	s.Equal(http.StatusBadRequest, status, string(respBody))

	var resp errorResponse
	s.Require().NoError(json.Unmarshal(respBody, &resp))
	s.NotEmpty(resp.Error)
}

// TestCreateEvent_TimeBusy_Returns409 проверяет бизнес-правило занятости временного слота:
// два события одного пользователя с пересекающимся временем -> 409 Conflict.
func (s *EventsSuite) TestCreateEvent_TimeBusy_Returns409() {
	startTime := time.Now().Add(3 * time.Hour).UTC()

	first := eventInput{
		Title:     "First Event",
		StartTime: startTime.Format(time.RFC3339),
		Duration:  7200,
		UserID:    "user-conflict",
	}
	status, body := s.post("/events", first)
	s.Require().Equal(http.StatusCreated, status, string(body))

	// второе событие пересекается по времени с первым для того же user
	second := eventInput{
		Title:     "Conflicting Event",
		StartTime: startTime.Add(30 * time.Minute).Format(time.RFC3339),
		Duration:  3600,
		UserID:    "user-conflict",
	}

	status, respBody := s.post("/events", second)
	s.Equal(http.StatusConflict, status, string(respBody))

	var resp errorResponse
	s.Require().NoError(json.Unmarshal(respBody, &resp))
	s.NotEmpty(resp.Error)
}

// TestDeleteEvent_NotFound_Returns404 проверяет обработку отсутствующего события:
// DELETE /events/{nonexistent-id} -> 404 Not Found.
func (s *EventsSuite) TestDeleteEvent_NotFound_Returns404() {
	fakeID := "00000000-0000-0000-0000-000000000000"
	status, respBody := s.delete("/events/" + fakeID)
	s.Equal(http.StatusNotFound, status, string(respBody))

	var resp errorResponse
	s.Require().NoError(json.Unmarshal(respBody, &resp))
	s.NotEmpty(resp.Error)
}

// TestDeleteEvent_Success проверяет успешное удаление события:
// создаём событие, затем DELETE /events/{id} -> 204 (No Content).
func (s *EventsSuite) TestDeleteEvent_Success() {
	id := s.createEvent("Delete Me", "user-del", time.Now().Add(5*time.Hour).UTC(), 1800, nil)
	status, _ := s.delete("/events/" + id)
	s.Equal(http.StatusNoContent, status)
}

// TestUpdateEvent_NotFound_Returns404 проверяет обновление несуществующего события:
// PUT /events/{nonexistent-id} -> 404.
func (s *EventsSuite) TestUpdateEvent_NotFound_Returns404() {
	fakeID := "00000000-0000-0000-0000-000000000001"
	body := eventInput{
		Title:     "Updated",
		StartTime: time.Now().Add(time.Hour).UTC().Format(time.RFC3339),
		Duration:  1800,
		UserID:    "user-1",
	}
	status, respBody := s.put("/events/"+fakeID, body)
	s.Equal(http.StatusNotFound, status, string(respBody))
}
