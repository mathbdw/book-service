package handlers

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/mathbdw/book/internal/domain/entities"
	"github.com/mathbdw/book/internal/interfaces/observability"
)

func (h *BotHandler) handleCommandList(ctx context.Context, mess *tgbotapi.Message) {
	start := time.Now()
	logger := h.observ.WithContext(ctx)
	ctx, span := h.observ.StartSpan(ctx, "v1.HandleCommand")
	defer span.End()

	statusCode := int(200)

	defer func() {
		duration := time.Since(start).Seconds()
		h.observ.RecordHanderRequest(ctx, "send", "v1/list", statusCode, duration)
	}()

	params := entities.PaginationParams{
		Limit:     10,
		SortBy:    entities.CursorTypeBookID,
		SortOrder: entities.SortOrderTypeAsc,
	}

	respBooks, err := h.uc.List.Execute(ctx, params)
	if err != nil {
		logger.Info("botHandler.handleCommandList: executing usecases", map[string]any{"error": err.Error()})
		span.RecordError(err)
		span.SetAttributes([]observability.Attribute{{Key: "usecases.failed", Value: true}})
		statusCode = 500

		msg := tgbotapi.NewMessage(mess.Chat.ID, "An error has occurred")
		_, err := h.bot.Send(msg)
		if err != nil {
			logger.Error("botHandler.handleCommandList: sending message", map[string]any{"error": err})
			span.RecordError(err)
			span.SetAttributes([]observability.Attribute{{Key: "sending.failed", Value: true}})

			return
		}
	}

	var outMess []string
	outMess = append(outMess, "Available book:")
	for _, book := range respBooks.Data {
		outMess = append(outMess, book.String())
	}
	msg := tgbotapi.NewMessage(mess.Chat.ID, strings.Join(outMess, "\n"))

	if respBooks.PageInfo.NextCursor != "" {
		var dataKey = KeyData{Action: "list", Page: respBooks.PageInfo.NextCursor}
		jsonData, err := json.Marshal(dataKey)
		if err != nil {
			statusCode = 500
			logger.Error("botHandler.handleCommandList: json marshal", map[string]any{"error": err})
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
		logger.Error("botHandler.handleCommandList: sending message", map[string]any{"error": err})
		span.RecordError(err)
		span.SetAttributes([]observability.Attribute{{Key: "sending.failed", Value: true}})

		statusCode = 500
	}

}
