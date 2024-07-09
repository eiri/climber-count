package main

import (
	"context"
	"fmt"
	"log/slog"

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
	logger := slog.Default().With("component", "cron handler")
	counters, err := jh.client.Counters()
	if err != nil {
		logger.Error("can't get counters from client", "msg", err)
		return err
	}

	counter := counters.Counter(jh.gym)
	logger.Info("got counter from client", "counter", counter)
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

	logger := slog.Default().With("component", "bot handler")
	if counter, ok := bh.storage.Last(); ok {
		logger.Info("sending reply", "chat_id", update.Message.Chat.ID, "text", counter)
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:    update.Message.Chat.ID,
			Text:      fmt.Sprintln(counter),
			ParseMode: models.ParseModeMarkdown,
		})
	}
}
