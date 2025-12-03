package handlers

import (
	"context"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/mathbdw/book/internal/interfaces/observability"
)

func (h *BotHandler) handleCommandHelp(ctx context.Context, mess *tgbotapi.Message) {
	start := time.Now()
	logger := h.observ.WithContext(ctx)
	ctx, span := h.observ.StartSpan(ctx, "v1.HandleCommand")
	defer span.End()

	statusCode := int(200)

	defer func() {
		duration := time.Since(start).Seconds()
		h.observ.RecordHanderRequest(ctx, "send", "v1/help", statusCode, duration)
	}()

	var outMess []string
	outMess = append(outMess, "Available commands")
	outMess = append(outMess, "/add - The command adds a product. Example:\n/add\nTitle - New title\nDescription - New description\nYear - year write\nGenre - history")
	outMess = append(outMess, "/delete {ID} - The command deletes a product by ID . Example:\n/delete 1")
	outMess = append(outMess, "/help - The command help.")
	outMess = append(outMess, "/get {ID} - The command gets a product by ID . Example:\n/get 1")
	outMess = append(outMess, "/list - The command view list of products")

	msg := tgbotapi.NewMessage(mess.Chat.ID, strings.Join(outMess[:], "\n\n"))
	_, err := h.bot.Send(msg)
	if err != nil {
		logger.Error("botHandler.handleCommandRemove: sending message", map[string]any{"error": err})
		span.RecordError(err)
		span.SetAttributes([]observability.Attribute{{Key: "sending.failed", Value: true}})
	}
}
