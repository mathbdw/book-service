package handlers

import (
	"context"
	"runtime/debug"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/mathbdw/book/internal/interfaces/observability"
	"github.com/mathbdw/book/internal/usecases/book"
)

type BotHandler struct {
	bot *tgbotapi.BotAPI
	uc  *book.BookUsecases

	observ observability.HandlerObservability
}

func New(bot *tgbotapi.BotAPI, uc *book.BookUsecases, observ observability.HandlerObservability) *BotHandler {
	return &BotHandler{
		bot:    bot,
		uc:     uc,
		observ: observ,
	}
}

func (h *BotHandler) Start(ctx context.Context, update tgbotapi.Update) {
	defer func() {
		if err := recover(); err != nil {
			h.observ.Error(
				"handler.Start: recovered from panic",
				map[string]any{"error": err, "stack": string(debug.Stack())},
			)
		}
	}()

	if update.CallbackQuery != nil {
		h.handleCallback(ctx, update.CallbackQuery)
		return
	}

	if update.Message == nil {
		return
	}

	switch update.Message.Command() {
	case "add":
		h.handleCommandAdd(ctx, update.Message)
	case "get":
		h.handleCommandGet(ctx, update.Message)
	case "list":
		h.handleCommandList(ctx, update.Message)
	case "remove":
		h.handleCommandRemove(ctx, update.Message)
	case "help":
		h.handleCommandHelp(ctx, update.Message)
	default:
		h.handleDefaultMessage(ctx, update.Message)
	}
}
