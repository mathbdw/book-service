package entities

import (
	"fmt"
	"time"
)

type Book struct {
	ID          int64     `db:"id"`
	Title       string    `db:"title"`
	Description string    `db:"description"`
	Year        int       `db:"year"`
	Genre       string    `db:"genre"`
	Removed     bool      `db:"removed"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

func (b *Book) String() string {
	return fmt.Sprintf("%s, %s, %d, %s", b.Title, b.Description, b.Year, b.Genre)
}

type ResponseBooks struct {
	Data     []Book
	PageInfo PageInfo
}

func (b Book) GetFieldAsString(field string) (string, error) {
	switch field {
	case string(CursorTypeBookID):
		if b.ID < 1 {
			return "", nil
		}
		return fmt.Sprint(b.ID), nil
	case string(CursorTypeBookTitle):
		return b.Title, nil
	case string(CursorTypeBookYear):
		return fmt.Sprint(b.Year), nil
	default:
		return "", fmt.Errorf("bookEntity.GetFieldAsString: unknown field %s", field)
	}
}

func (b Book) GetCreatedAt() time.Time {
	return b.CreatedAt
}
