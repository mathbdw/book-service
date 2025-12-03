package handlers

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/mathbdw/book/internal/domain/entities"
	"github.com/mathbdw/book/internal/infrastructure/persistence/postgres"
	"github.com/mathbdw/book/internal/interfaces/observability"
)

type KeyData struct {
	Action string `json:"action"`
	Page   string `json:"page"`
}

func (h *BotHandler) handleCallback(ctx context.Context, callback *tgbotapi.CallbackQuery) {
	start := time.Now()
	logger := h.observ.WithContext(ctx)
	ctx, span := h.observ.StartSpan(ctx, "v1.HandleCommand")
	defer span.End()

	statusCode := int(200)

	defer func() {
		duration := time.Since(start).Seconds()
		h.observ.RecordHanderRequest(ctx, "send", "v1/callback", statusCode, duration)
	}()

	var unData KeyData
	err := json.Unmarshal([]byte(callback.Data), &unData)
	if err != nil {
		logger.Error("botHandler.handleCallback: json unmarshel", map[string]any{"error": err})
		span.RecordError(err)
		span.SetAttributes([]observability.Attribute{{Key: "json.unmarshel.failed", Value: true}})

		statusCode = 500
		return
	}

	if unData.Action != "list" {
		logger.Error("botHandler.handleCallback: format data", map[string]any{"data": unData})
		span.RecordError(err)
		span.SetAttributes([]observability.Attribute{{Key: "format.data.failed", Value: true}})

		statusCode = 500

		return
	}
	cursor, err := postgres.DecodeCursor(unData.Page)
	if err != nil {
		logger.Error("botHandler.handleCallback: invalid cursor", map[string]any{"data": unData})
		span.RecordError(err)
		span.SetAttributes([]observability.Attribute{{Key: "invalid.cursor", Value: true}})

		return
	}

	params := entities.PaginationParams{
		Limit:     10,
		Cursor:    cursor,
		SortBy:    entities.CursorTypeBookID,
		SortOrder: entities.SortOrderTypeAsc,
	}
	respBooks, err := h.uc.List.Execute(ctx, params)
	if err != nil {
		logger.Info("botHandler.handleCallback: executing usecases", map[string]any{"error": err.Error()})
		span.RecordError(err)
		span.SetAttributes([]observability.Attribute{{Key: "usecases.failed", Value: true}})
		statusCode = 500

		return
	}

	var outMess []string

	for _, book := range respBooks.Data {
		outMess = append(outMess, book.String())
	}

	msg := tgbotapi.NewMessage(callback.Message.Chat.ID, strings.Join(outMess, "\n"))

	if respBooks.PageInfo.NextCursor != "" {
		var dataKey = KeyData{Action: "list", Page: respBooks.PageInfo.NextCursor}
		jsonData, err := json.Marshal(dataKey)
		if err != nil {
			statusCode = 500
			logger.Error("botHandler.handleCallback: json marshal", map[string]any{"error": err})
			span.RecordError(err)
			span.SetAttributes([]observability.Attribute{{Key: "json.marshal.failed", Value: true}})
			return
		}

		var key = tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("More", string(jsonData)),
			),
		)
		msg.ReplyMarkup = key
	}

	_, err = h.bot.Send(msg)
	if err != nil {
		logger.Error("botHandler.handleCallback: sending message", map[string]any{"error": err})
		span.RecordError(err)
		span.SetAttributes([]observability.Attribute{{Key: "sending.failed", Value: true}})

		statusCode = 500
	}
}
