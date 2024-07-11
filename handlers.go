package main

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"strings"

	"github.com/enescakir/emoji"
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
	logger  *slog.Logger
}

func NewBotHandler(storage Storer) *BotHandler {
	logger := slog.Default().With("component", "bot count handler")
	return &BotHandler{
		storage: storage,
		logger:  logger,
	}
}

func (bh *BotHandler) CountHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.Message == nil {
		return
	}

	if counter, ok := bh.storage.Last(); ok {
		b.SendMessage(ctx, bh.Message(b, update.Message.Chat.ID, counter.String()))
	}
}

func (bh *BotHandler) PingHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.Message == nil {
		return
	}

	chatID := update.Message.Chat.ID
	msgID := update.Message.ID

	if update.Message.Text == "ping off" {
		bh.storage.RemoveCallback()
		b.SetMessageReaction(ctx, bh.Reaction(b, chatID, msgID, emoji.ThumbsUp.String()))
		return
	}

	args := strings.Split(update.Message.Text, " ")
	if len(args) != 3 || args[1] != "on" {
		return
	}

	number, err := strconv.Atoi(args[2])
	if err != nil {
		bh.logger.Error("can't parse number", "msg", err)
		b.SendMessage(ctx, bh.Message(b, chatID, "Please provide a valid number."))
		return
	}

	bh.storage.SetCallback(func(c Counter) bool {
		if c.Count <= number {
			b.SendMessage(ctx, bh.Message(b, chatID, fmt.Sprintf("Hey, %s", c)))
			return true
		}
		return false
	})
	b.SetMessageReaction(ctx, bh.Reaction(b, chatID, msgID, emoji.Handshake.String()))
}

func (bh *BotHandler) Message(b *bot.Bot, chatID int64, msg string) *bot.SendMessageParams {
	bh.logger.Info("sending reply", "chat_id", chatID, "text", msg)
	return &bot.SendMessageParams{ChatID: chatID, Text: msg}
}

func (bh *BotHandler) Reaction(b *bot.Bot, chatID int64, msgId int, emoji string) *bot.SetMessageReactionParams {
	bh.logger.Info("sending reply", "chat_id", chatID, "reply", emoji)
	return &bot.SetMessageReactionParams{
		ChatID:    chatID,
		MessageID: msgId,
		Reaction: []models.ReactionType{{
			Type: models.ReactionTypeTypeEmoji,
			ReactionTypeEmoji: &models.ReactionTypeEmoji{
				Type:  models.ReactionTypeTypeEmoji,
				Emoji: emoji,
			}},
		},
	}
}
