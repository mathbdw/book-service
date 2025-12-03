package tbot

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Bot struct {
	Api *tgbotapi.BotAPI

	readTimeout int
	token       string
	debug       bool
}

// New -.
func New(opt ...Option) (*Bot, error) {
	var err error
	b := &Bot{
		Api:   nil,
		token: "",
	}

	for _, opt := range opt {
		opt(b)
	}

	b.Api, err = tgbotapi.NewBotAPI(b.token)
	if err != nil {
		return nil, err
	}

	b.Api.Debug = b.debug


	return b, nil
}

// Start - gets the channel updates
func (b *Bot) Start() tgbotapi.UpdatesChannel {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = b.readTimeout

	return b.Api.GetUpdatesChan(u)
}

// Shutdown -.
func (b *Bot) Shutdown() {
	b.Api.StopReceivingUpdates()
}