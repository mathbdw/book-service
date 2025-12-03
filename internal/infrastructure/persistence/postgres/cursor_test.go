package postgres

import (
	"encoding/base64"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestPagination_EncodeCursor_Error(t *testing.T) {
	_, err := EncodeCursor("", nil)

	assert.Error(t, err)
}

func TestPagination_EncodeCursor_Success(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name      string
		value     string
		createdAt *time.Time
	}{
		{"CreatedAtNil", "test", nil},
		{"CreatedAt", "test", &now},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			strCursor, err := EncodeCursor(tt.value, tt.createdAt)

			assert.NoError(t, err)
			assert.Equal(t, true, len(strCursor) > 0)
		})
	}
}

func TestPagination_DecodeCursor_ErrorList(t *testing.T) {
	tests := []struct {
		name             string
		value            string
		wantErrorContent string
	}{
		{"InvalidDecode", "df df", "invalid cursor decoding"},
		{"EmptyValue",
			base64.StdEncoding.EncodeToString([]byte(":sfd")),
			"empty value in cursor",
		},
		{"InvalidParts",
			base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("test:%d:", time.Now().UnixNano()))),
			"invalid parts in cursor",
		},
		{"InvalidTimestamp",
			base64.StdEncoding.EncodeToString([]byte("test:63456346345634563563")),
			"invalid timestamp in cursor",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := DecodeCursor(tt.value)

			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErrorContent)
		})
	}
}

func TestPagination_DecodeCursor_Success(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name      string
		strCursor string
		value     any
		createdAt *time.Time
	}{
		{"Value",
			base64.StdEncoding.EncodeToString([]byte("test")),
			"test",
			nil,
		},
		{"ValueAndCreatedAt",
			base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("10:%d", time.Now().UnixNano()))),
			"10",
			&now,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cursor, err := DecodeCursor(tt.strCursor)

			assert.NoError(t, err)
			assert.Equal(t, tt.value, cursor.Value)
			if tt.createdAt != nil {
				assert.Equal(t, tt.createdAt.Unix(), cursor.CreatedAt.Unix())
			}
		})
	}
}
