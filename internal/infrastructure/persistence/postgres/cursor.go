package postgres

import (
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/mathbdw/book/internal/domain/entities"
)

// EncodeCursor - encodes cursor to base64
func EncodeCursor(value string, createdAt *time.Time) (string, error) {
	if value == "" {
		return "", fmt.Errorf("value can not be empty")
	}

	strCursor := value
	if createdAt != nil {
		strCursor = strCursor + ":" + fmt.Sprint(createdAt.UnixNano())
	}

	return base64.StdEncoding.EncodeToString([]byte(strCursor)), nil
}

// DecodeCursor - decodes cursor from base64
func DecodeCursor(strCursor string) (*entities.Cursor, error) {
	decoded, err := base64.StdEncoding.DecodeString(strCursor)
	if err != nil {
		return nil, fmt.Errorf("invalid cursor decoding: %w", err)
	}

	parts := strings.Split(string(decoded), ":")
	if len(parts[0]) == 0 {
		return nil, fmt.Errorf("empty value in cursor")
	}

	if len(parts) > 2 {
		return nil, fmt.Errorf("invalid parts in cursor: %d", len(parts))
	}

	cursor := entities.Cursor{Value: parts[0], CreatedAt: nil}
	if len(parts) > 1 {
		timestamp, err := strconv.ParseInt(parts[1], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid timestamp in cursor: %w", err)
		}
		createdAt := time.Unix(0, timestamp)
		cursor.CreatedAt = &createdAt
	}

	return &cursor, nil
}
