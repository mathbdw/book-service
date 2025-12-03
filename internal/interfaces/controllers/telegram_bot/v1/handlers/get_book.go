package handlers

import (
	"context"
	"errors"
	"fmt"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	errs "github.com/mathbdw/book/internal/errors"
	"github.com/mathbdw/book/internal/interfaces/controllers/telegram_bot/v1/validate"
	"github.com/mathbdw/book/internal/interfaces/observability"
)

func (h *BotHandler) handleCommandGet(ctx context.Context, mess *tgbotapi.Message) {
	start := time.Now()
	logger := h.observ.WithContext(ctx)
	ctx, span := h.observ.StartSpan(ctx, "v1.HandleCommand")
	defer span.End()

	statusCode := int(200)

	defer func() {
		duration := time.Since(start).Seconds()
		h.observ.RecordHanderRequest(ctx, "send", "v1/get", statusCode, duration)
	}()

	bookId, err := validate.GetBook(mess)
	if err != nil {
		logger.Error("botHandler.handleCommandGet: validate", map[string]any{"error": err})
		span.RecordError(err)
		span.SetAttributes([]observability.Attribute{{Key: "validation.failed", Value: true}})

		statusCode = 422
		msg := tgbotapi.NewMessage(mess.Chat.ID, err.Error())
		_, err = h.bot.Send(msg)
		if err != nil {
			logger.Error("botHandler.handleCommandGet: sending message", map[string]any{"error": err})
			span.RecordError(err)
			span.SetAttributes([]observability.Attribute{{Key: "sending.failed", Value: true}})
			statusCode = 500

			return
		}

		return
	}

	books, err := h.uc.Get.GetByIDs(ctx, []int64{bookId})
	if err != nil {
		msgStr := "Book not found"

		if errors.Is(err, errs.ErrNotFound) {
			logger.Info("botHandler.handleCommandGet: usecase not found ID", map[string]any{
				"error": err.Error(),
				"id":    bookId,
			})
		} else {
			logger.Info("botHandler.handleCommandGet: executing usecases", map[string]any{
				"error": err.Error(),
				"ids":   bookId,
			})
			statusCode = 500
			msgStr = "An error has occurred"
		}

		span.RecordError(err)
		span.SetAttributes([]observability.Attribute{{Key: "usecases.failed", Value: true}})

		msg := tgbotapi.NewMessage(mess.Chat.ID, msgStr)
		_, err := h.bot.Send(msg)
		if err != nil {
			logger.Error("botHandler.handleCommandGet: sending message", map[string]any{"error": err})
			span.RecordError(err)
			span.SetAttributes([]observability.Attribute{{Key: "sending.failed", Value: true}})
		}

		return
	}

	msg := tgbotapi.NewMessage(mess.Chat.ID, fmt.Sprintf("%v", books))
	_, err = h.bot.Send(msg)
	if err != nil {
		logger.Error("botHandler.handleCommandGet: sending message", map[string]any{"error": err})
		span.RecordError(err)
		span.SetAttributes([]observability.Attribute{{Key: "sending.failed", Value: true}})
	}
}
