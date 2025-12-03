package handlers

import (
	"context"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/mathbdw/book/internal/interfaces/controllers/telegram_bot/v1/validate"
	"github.com/mathbdw/book/internal/interfaces/observability"
)

func (h *BotHandler) handleCommandRemove(ctx context.Context, mess *tgbotapi.Message) {
	start := time.Now()
	logger := h.observ.WithContext(ctx)
	ctx, span := h.observ.StartSpan(ctx, "v1.HandleCommand")
	defer span.End()

	statusCode := int(200)

	defer func() {
		duration := time.Since(start).Seconds()
		h.observ.RecordHanderRequest(ctx, "send", "v1/remove", statusCode, duration)
	}()

	bookId, err := validate.GetBook(mess)
	if err != nil {
		logger.Error("botHandler.handleCommandRemove: validate", map[string]any{"error": err})
		span.RecordError(err)
		span.SetAttributes([]observability.Attribute{{Key: "validation.failed", Value: true}})

		statusCode = 422
		msg := tgbotapi.NewMessage(mess.Chat.ID, err.Error())
		_, err = h.bot.Send(msg)
		if err != nil {
			logger.Error("botHandler.handleCommandRemove: sending message", map[string]any{"error": err})
			span.RecordError(err)
			span.SetAttributes([]observability.Attribute{{Key: "sending.failed", Value: true}})
			statusCode = 500

			return
		}

		return
	}

	err = h.uc.Remove.Execute(ctx, []int64{bookId})
	if err != nil {
		logger.Info("botHandler.handleCommandRemove: executing usecases", map[string]any{
			"error": err.Error(),
			"id":    bookId,
		})
		statusCode = 500

		span.RecordError(err)
		span.SetAttributes([]observability.Attribute{{Key: "usecases.failed", Value: true}})

		msg := tgbotapi.NewMessage(mess.Chat.ID, "An error has occurred")
		_, err := h.bot.Send(msg)
		if err != nil {
			logger.Error("botHandler.handleCommandRemove: sending message", map[string]any{"error": err})
			span.RecordError(err)
			span.SetAttributes([]observability.Attribute{{Key: "sending.failed", Value: true}})
		}

		return
	}

	msg := tgbotapi.NewMessage(mess.Chat.ID, "Book remove successfully")
	_, err = h.bot.Send(msg)
	if err != nil {
		logger.Error("botHandler.handleCommandRemove: sending message", map[string]any{"error": err})
		span.RecordError(err)
		span.SetAttributes([]observability.Attribute{{Key: "sending.failed", Value: true}})
	}
}
