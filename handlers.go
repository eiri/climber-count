package main

import (
	"context"
	"fmt"
	"log/slog"
	"sort"
	"strings"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

type JobHandler struct {
	storageDir string
	client     *Client
	storers    map[string]Storer
}

func NewJobHandler(storageDir string, client *Client, storers map[string]Storer) *JobHandler {
	return &JobHandler{
		storageDir: storageDir,
		client:     client,
		storers:    storers,
	}
}

func (jh *JobHandler) Execute(ctx context.Context) error {
	logger := slog.Default().With("component", "cron handler")

	counters, err := jh.client.Counters()
	if err != nil {
		logger.Error("can't get counters from client", "msg", err)
		return err
	}

	var firstErr error
	for gym, storer := range jh.storers {
		counter := counters.Counter(gym)
		logger.Info("got counter from client", "gym", gym, "counter", counter)
		if err := storer.Store(counter); err != nil {
			logger.Error("failed to store counter", "gym", gym, "msg", err)
			if firstErr == nil {
				firstErr = err
			}
		}
	}
	return firstErr
}

func (jh *JobHandler) Description() string {
	return fmt.Sprintf("Climber Count Job for %d gym(s)", len(jh.storers))
}

type BotHandler struct {
	storers    map[string]Storer
	defaultGym string
	logger     *slog.Logger
}

func NewBotHandler(defaultGym string, storers map[string]Storer) *BotHandler {
	logger := slog.Default().With("component", "bot handler")
	return &BotHandler{
		storers:    storers,
		defaultGym: defaultGym,
		logger:     logger,
	}
}

func (bh *BotHandler) CountHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	bh.logger.Info("CountHandler", "text", update.Message)
	if update.Message == nil {
		return
	}
	chatID := update.Message.Chat.ID

	gymKey := bh.defaultGym
	if fields := strings.Fields(update.Message.Text); len(fields) >= 2 {
		gymKey = strings.ToUpper(fields[1])
	}

	storer, ok := bh.storers[gymKey]
	if !ok {
		b.SendMessage(ctx, bh.Message(b, chatID,
			fmt.Sprintf("Unknown gym %q. Known gyms: %s", gymKey, bh.gymKeys())))
		return
	}

	if counter, ok := storer.Last(); ok {
		b.SendMessage(ctx, bh.Message(b, chatID, counter.String()))
	}
}

func (bh *BotHandler) GymHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.Message == nil {
		return
	}
	chatID := update.Message.Chat.ID
	msg := bh.Message(b, chatID, "Going into the gym?")
	msg.ReplyMarkup = &models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{
				{Text: "Yeah", CallbackData: "gym_in"},
				{Text: "Done", CallbackData: "gym_out"},
			},
		},
	}
	b.SendMessage(ctx, msg)
}

func (bh *BotHandler) GymButtonHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.CallbackQuery == nil {
		return
	}

	b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
		CallbackQueryID: update.CallbackQuery.ID,
		ShowAlert:       false,
	})

	chatID := update.CallbackQuery.Message.Message.Chat.ID

	storer, ok := bh.storers[bh.defaultGym]
	if !ok {
		b.SendMessage(ctx, bh.Message(b, chatID,
			fmt.Sprintf("Unknown gym %q. Known gyms: %s", bh.defaultGym, bh.gymKeys())))
		return
	}

	if update.CallbackQuery.Data == "gym_in" {
		msg := "Have a great climb!"
		err := storer.GetGym().In()
		if err != nil {
			msg = err.Error()
		}
		b.SendMessage(ctx, bh.Message(b, chatID, msg))
	}

	if update.CallbackQuery.Data == "gym_out" {
		msg, err := storer.GetGym().Out()
		if err != nil {
			b.SendMessage(ctx, bh.Message(b, chatID, err.Error()))
			return
		}
		b.SendMessage(ctx, bh.Message(b, chatID, fmt.Sprintf("You went to gym %s. Good job!", msg)))
	}
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

func (bh *BotHandler) gymKeys() string {
	keys := make([]string, 0, len(bh.storers))
	for k := range bh.storers {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return strings.Join(keys, ", ")
}
