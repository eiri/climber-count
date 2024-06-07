package main

import (
	"context"
	"fmt"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

type JobHandler struct {
	gym     string
	client  *Client
	storage Storer
}

func NewJobHandler(gym string, client *Client, storage Storer) *JobHandler {
	return &JobHandler{
		gym:     gym,
		client:  client,
		storage: storage,
	}
}

func (jh *JobHandler) Execute(ctx context.Context) error {
	counters, err := jh.client.Counters()
	if err != nil {
		return err
	}

	counter := counters.Counter(jh.gym)
	return jh.storage.Store(counter)
}

func (jh *JobHandler) Description() string {
	return fmt.Sprintf("Climber Count Job for %q", jh.gym)
}

type BotHandler struct {
	storage Storer
}

func NewBotHandler(storage Storer) *BotHandler {
	return &BotHandler{
		storage: storage,
	}
}

func (bh *BotHandler) Handler(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.Message == nil {
		return
	}

	if counter, ok := bh.storage.Last(); ok {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:    update.Message.Chat.ID,
			Text:      fmt.Sprintln(counter),
			ParseMode: models.ParseModeMarkdown,
		})
	}
}
