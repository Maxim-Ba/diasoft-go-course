package integration

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// ListingSuite тестирует листинг событий за день/неделю/месяц.
type ListingSuite struct {
	CalendarSuite
}

// baseDate — фиксированная дата в будущем для тестов, чтобы не зависеть от текущего времени.
func baseDate() time.Time {
	now := time.Now().UTC()
	// Первый день следующего месяца — чтобы тесты не пересекались с событиями текущего дня.
	return time.Date(now.Year(), now.Month()+1, 1, 10, 0, 0, 0, time.UTC)
}

// TestGetEventsForDay_ReturnsOnlyThatDay проверяет фильтрацию по дню:
// если есть два события в разные дни, GET /events/day?date=X возвращает только событие нужного дня.
func (s *ListingSuite) TestGetEventsForDay_ReturnsOnlyThatDay() {
	base := baseDate()
	day1 := base
	day2 := base.AddDate(0, 0, 1)

	s.createEvent("Event Day1", "user-listing", day1, 1800, nil)
	s.createEvent("Event Day2", "user-listing", day2, 1800, nil)

	dateStr := day1.Format("2006-01-02")
	status, body := s.get(fmt.Sprintf("/events/day?date=%s", dateStr))
	s.Require().Equal(http.StatusOK, status, string(body))

	var resp eventListResponse
	s.Require().NoError(json.Unmarshal(body, &resp))
	s.Require().Len(resp.Events, 1)
	s.Equal("Event Day1", resp.Events[0].Title)
}

// TestGetEventsForDay_EmptyResult проверяет, что запрос за день без событий
// возвращает 200 и пустой список (не null).
func (s *ListingSuite) TestGetEventsForDay_EmptyResult() {
	emptyDate := baseDate().AddDate(0, 0, 15).Format("2006-01-02")
	status, body := s.get(fmt.Sprintf("/events/day?date=%s", emptyDate))
	s.Require().Equal(http.StatusOK, status, string(body))

	var resp eventListResponse
	s.Require().NoError(json.Unmarshal(body, &resp))
	s.Empty(resp.Events)
}

// TestGetEventsForWeek_ReturnsEventsInRange проверяет фильтрацию по неделе:
// из 3 событий (2 в неделю + 1 за ней) GET /events/week?date=X возвращает только 2.
func (s *ListingSuite) TestGetEventsForWeek_ReturnsEventsInRange() {
	base := baseDate()
	// Три события: два в течение недели, одно — за её пределами.
	inWeek1 := base
	inWeek2 := base.AddDate(0, 0, 3)
	outOfWeek := base.AddDate(0, 0, 8)

	s.createEvent("Week Event 1", "user-week", inWeek1, 1800, nil)
	s.createEvent("Week Event 2", "user-week", inWeek2, 1800, nil)
	s.createEvent("Out of Week", "user-week", outOfWeek, 1800, nil)

	dateStr := base.Format("2006-01-02")
	status, body := s.get(fmt.Sprintf("/events/week?date=%s", dateStr))
	s.Require().Equal(http.StatusOK, status, string(body))

	var resp eventListResponse
	s.Require().NoError(json.Unmarshal(body, &resp))
	s.Require().Len(resp.Events, 2, "expected 2 events within the week")

	titles := make([]string, 0, len(resp.Events))
	for _, e := range resp.Events {
		titles = append(titles, e.Title)
	}
	s.Contains(titles, "Week Event 1")
	s.Contains(titles, "Week Event 2")
}

// TestGetEventsForMonth_ReturnsEventsInRange проверяет фильтрацию по месяцу:
// из 3 событий (2 в месяце + 1 в следующем) GET /events/month?date=X возвращает только 2.
func (s *ListingSuite) TestGetEventsForMonth_ReturnsEventsInRange() {
	base := baseDate()
	// Два события в рамках месяца, одно — за его пределами.
	inMonth1 := base
	inMonth2 := base.AddDate(0, 0, 15)
	outOfMonth := base.AddDate(0, 1, 1)

	s.createEvent("Month Event 1", "user-month", inMonth1, 1800, nil)
	s.createEvent("Month Event 2", "user-month", inMonth2, 1800, nil)
	s.createEvent("Out of Month", "user-month", outOfMonth, 1800, nil)

	dateStr := base.Format("2006-01-02")
	status, body := s.get(fmt.Sprintf("/events/month?date=%s", dateStr))
	s.Require().Equal(http.StatusOK, status, string(body))

	var resp eventListResponse
	s.Require().NoError(json.Unmarshal(body, &resp))
	s.Require().Len(resp.Events, 2, "expected 2 events within the month")

	titles := make([]string, 0, len(resp.Events))
	for _, e := range resp.Events {
		titles = append(titles, e.Title)
	}
	s.Contains(titles, "Month Event 1")
	s.Contains(titles, "Month Event 2")
}
