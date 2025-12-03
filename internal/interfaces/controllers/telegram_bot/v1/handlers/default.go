package handlers

import (
	"context"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/mathbdw/book/internal/interfaces/observability"
)

func (h *BotHandler) handleDefaultMessage(ctx context.Context, mess *tgbotapi.Message) {
	start := time.Now()
	logger := h.observ.WithContext(ctx)
	ctx, span := h.observ.StartSpan(ctx, "v1.HandleCommand")
	defer span.End()

	statusCode := int(200)

	defer func() {
		duration := time.Since(start).Seconds()
		h.observ.RecordHanderRequest(ctx, "send", "v1/default", statusCode, duration)
	}()

	msg := tgbotapi.NewMessage(mess.Chat.ID, "The command was not recognized")
	_, err := h.bot.Send(msg)
	if err != nil {
		logger.Error("botHandler.handleDefaultMessage: sending message", map[string]any{"error": err})
		span.RecordError(err)
		span.SetAttributes([]observability.Attribute{{Key: "sending.failed", Value: true}})

		statusCode = 500
	}
}
