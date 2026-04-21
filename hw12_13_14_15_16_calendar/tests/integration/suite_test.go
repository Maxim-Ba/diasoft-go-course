package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/suite"
)

func calendarURL() string {
	if u := os.Getenv("CALENDAR_URL"); u != "" {
		return u
	}
	return "http://localhost:8082"
}

func databaseDSN() string {
	if d := os.Getenv("DATABASE_DSN"); d != "" {
		return d
	}
	return "postgres://calendar:calendar@localhost:5432/calendar?sslmode=disable"
}

// CalendarSuite — базовый suite для всех интеграционных тестов.
type CalendarSuite struct {
	suite.Suite
	baseURL string
	db      *sqlx.DB
	client  *http.Client
}

func (s *CalendarSuite) SetupSuite() {
	s.baseURL = calendarURL()
	s.client = &http.Client{Timeout: 10 * time.Second}

	db, err := sqlx.Connect("postgres", databaseDSN())
	s.Require().NoError(err, "failed to connect to database")
	s.db = db
}

func (s *CalendarSuite) TearDownSuite() {
	if s.db != nil {
		s.db.Close()
	}
}

func (s *CalendarSuite) SetupTest() {
	_, err := s.db.Exec("DELETE FROM notifications")
	s.Require().NoError(err)
	_, err = s.db.Exec("DELETE FROM events")
	s.Require().NoError(err)
}

// post выполняет POST-запрос и возвращает статус и тело ответа.
func (s *CalendarSuite) post(path string, body interface{}) (int, []byte) {
	data, err := json.Marshal(body)
	s.Require().NoError(err)

	resp, err := s.client.Post(s.baseURL+path, "application/json", bytes.NewReader(data))
	s.Require().NoError(err)
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	s.Require().NoError(err)
	return resp.StatusCode, respBody
}

// put выполняет PUT-запрос.
func (s *CalendarSuite) put(path string, body interface{}) (int, []byte) {
	data, err := json.Marshal(body)
	s.Require().NoError(err)

	req, err := http.NewRequest(http.MethodPut, s.baseURL+path, bytes.NewReader(data))
	s.Require().NoError(err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	s.Require().NoError(err)
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	s.Require().NoError(err)
	return resp.StatusCode, respBody
}

// delete выполняет DELETE-запрос.
func (s *CalendarSuite) delete(path string) (int, []byte) {
	req, err := http.NewRequest(http.MethodDelete, s.baseURL+path, nil)
	s.Require().NoError(err)

	resp, err := s.client.Do(req)
	s.Require().NoError(err)
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	s.Require().NoError(err)
	return resp.StatusCode, respBody
}

// get выполняет GET-запрос.
func (s *CalendarSuite) get(path string) (int, []byte) {
	resp, err := s.client.Get(s.baseURL + path)
	s.Require().NoError(err)
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	s.Require().NoError(err)
	return resp.StatusCode, respBody
}

// createEvent — хелпер для создания события через API, возвращает ID.
func (s *CalendarSuite) createEvent(title, userID string, startTime time.Time, durationSec int64, notificationSec *int64) string {
	body := map[string]interface{}{
		"title":     title,
		"startTime": startTime.Format(time.RFC3339),
		"duration":  durationSec,
		"userId":    userID,
	}
	if notificationSec != nil {
		body["notificationTime"] = *notificationSec
	}

	status, respBody := s.post("/events", body)
	s.Require().Equal(http.StatusCreated, status, "createEvent failed: %s", string(respBody))

	var result map[string]interface{}
	s.Require().NoError(json.Unmarshal(respBody, &result))
	id, ok := result["id"].(string)
	s.Require().True(ok, "response missing 'id' field")
	return id
}

// waitCondition делает polling условия каждую секунду до timeout.
func waitCondition(timeout time.Duration, condition func() bool) bool {
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		if condition() {
			return true
		}
		time.Sleep(time.Second)
	}
	return false
}

// TestMain — точка входа для запуска всех suite.
func TestMain(m *testing.M) {
	// ждём готовности API (до 30 секунд).
	url := calendarURL() + "/events/day?date=" + time.Now().Format("2006-01-02")
	client := &http.Client{Timeout: 2 * time.Second}
	deadline := time.Now().Add(30 * time.Second)
	for time.Now().Before(deadline) {
		resp, err := client.Get(url)
		if err == nil {
			resp.Body.Close()
			break
		}
		time.Sleep(time.Second)
	}

	os.Exit(m.Run())
}

// eventInput — вспомогательная структура для сериализации.
type eventInput struct {
	Title            string  `json:"title"`
	StartTime        string  `json:"startTime"`
	Duration         int64   `json:"duration"`
	UserID           string  `json:"userId"`
	Description      *string `json:"description,omitempty"`
	NotificationTime *int64  `json:"notificationTime,omitempty"`
}

// eventResponse — вспомогательная структура для десериализации.
type eventResponse struct {
	ID               string  `json:"id"`
	Title            string  `json:"title"`
	StartTime        string  `json:"startTime"`
	Duration         int64   `json:"duration"`
	UserID           string  `json:"userId"`
	Description      *string `json:"description,omitempty"`
	NotificationTime *int64  `json:"notificationTime,omitempty"`
}

type eventListResponse struct {
	Events []eventResponse `json:"events"`
}

type errorResponse struct {
	Error string `json:"error"`
}

func TestCalendar(t *testing.T) {
	fmt.Println("=== Integration Tests: Calendar API ===")
	suite.Run(t, new(EventsSuite))
	suite.Run(t, new(ListingSuite))
	suite.Run(t, new(StorerSuite))
}
