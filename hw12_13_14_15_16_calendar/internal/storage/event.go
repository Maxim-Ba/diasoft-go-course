package storage

import "time"

// Event представляет собой календарное событие.
type Event struct {
	ID               string        // ID события .
	Title            string        // заголовок - короткий текст.
	StartTime        time.Time     // дата и время начала события.
	Duration         time.Duration // дллительность события.
	Description      string        // описание события - длинный текст, опционально.
	UserID           string        // ID пользователя, владельца события.
	NotificationTime time.Duration // за сколько времени высылать уведомление, опционально.
}
