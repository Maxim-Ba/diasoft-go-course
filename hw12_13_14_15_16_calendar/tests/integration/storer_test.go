package integration

import (
	"time"
)

// StorerSuite тестирует сохранение уведомлений в БД через цепочку:
// Rest API -> scheduler -> Kafka -> storer -> таблица notifications.
type StorerSuite struct {
	CalendarSuite
}

// Строка таблицы notifications.
type dbNotification struct {
	ID        int       `db:"id"`
	EventID   string    `db:"event_id"`
	Title     string    `db:"title"`
	EventDate time.Time `db:"event_date"`
	UserID    string    `db:"user_id"`
	CreatedAt time.Time `db:"created_at"`
}

// TestStorerSavesNotificationToDB проверяет сквозное сохранение уведомлений через всю цепочку:
// создаём событие (startTime=time.Now()+30m, notificationTime=29m) -> scheduler обнаруживает его
// за примерно 5с -> отправляет в Kafka -> storer сохраняет в таблицу notifications.
// Поллингом проверяет появление записи с правильными полями в течение 30 секунд.
func (s *StorerSuite) TestStorerSavesNotificationToDB() {
	// Создаём событие: startTime = time.Now() + 30 минут, notificationTime = 29 минут.
	// notifyAt = time.Now() + 30min - 29min = time.Now() + 1min — попадает в окно [time.Now(), time.Now()+1h] планировщика.
	startTime := time.Now().Add(30 * time.Minute).UTC()
	notificationSec := int64(29 * 60) // 29 минут

	eventID := s.createEvent(
		"Storer Test Event",
		"user-storer",
		startTime,
		1800,
		&notificationSec,
	)
	s.Require().NotEmpty(eventID)

	// Polling таблицы notifications до 30 секунд (scheduler с SCAN_INTERVAL=5s.
	var notification dbNotification
	found := waitCondition(30*time.Second, func() bool {
		var rows []dbNotification
		err := s.db.Select(&rows,
			"SELECT id, event_id, title, event_date, user_id, created_at FROM notifications WHERE event_id = $1",
			eventID,
		)
		if err != nil || len(rows) == 0 {
			return false
		}
		notification = rows[0]
		return true
	})

	s.Require().True(found, "notification was not saved to DB within timeout")
	s.Equal(eventID, notification.EventID)
	s.Equal("Storer Test Event", notification.Title)
	s.Equal("user-storer", notification.UserID)

	// проверяем что дата события сохранена корректно (с точностью до секунды)
	s.WithinDuration(startTime, notification.EventDate, time.Second)
}
