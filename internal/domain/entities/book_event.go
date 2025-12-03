package entities

import "time"

type (
	EventType   uint16
	EventStatus uint16
)

const (
	Created EventType = iota + 1
	Updated
	Deleted
)
const (
	EventStatusNew EventStatus = iota + 1
	EventStatusLock
	EventStatusUnlock
)

type BookEvent struct {
	ID        int64       `db:"id"`
	BookId    int64       `db:"book_id"`
	Type      EventType   `db:"type"`
	Status    EventStatus `db:"status"`
	Payload   []byte      `db:"payload"`
	UpdatedAt time.Time   `db:"updated_at"`
}
