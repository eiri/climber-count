package main

import (
	"context"
	"fmt"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

type Bot struct {
	client  *Client
	storage Storer
	gym     string
}

func NewBot(cfg *Config) *Bot {
	storage := NewStorage(cfg.Storage)
	return &Bot{
		client:  NewClient(cfg),
		storage: storage,
		gym:     cfg.Gym,
	}
}

func (bt *Bot) Handler(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.Message == nil {
		return
	}

	counters, err := bt.client.Counters()
	// no error handeling for now
	if err != nil {
		return
	}

	counter := counters.Counter(bt.gym)

	err = bt.storage.Store(counter)
	if err != nil {
		return
	}

	if c, ok := bt.storage.Last(); ok {
		counter = c
	}

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:    update.Message.Chat.ID,
		Text:      fmt.Sprintln(counter),
		ParseMode: models.ParseModeMarkdown,
	})
}
