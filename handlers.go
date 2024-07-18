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
	logger := slog.Default().With("component", "bot handler")
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

	if update.Message.Text == "ping" {
		msg := bh.Message(b, chatID, "On how many?")
		msg.ReplyMarkup = &models.InlineKeyboardMarkup{
			InlineKeyboard: [][]models.InlineKeyboardButton{
				{
					{Text: "10", CallbackData: "ping on 10"},
					{Text: "20", CallbackData: "ping on 20"},
					{Text: "30", CallbackData: "ping on 30"},
				},
			},
		}
		b.SendMessage(ctx, msg)
		return
	}

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

func (bh *BotHandler) PingButtonHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.CallbackQuery == nil {
		return
	}

	b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
		CallbackQueryID: update.CallbackQuery.ID,
		ShowAlert:       false,
	})

	chatID := update.CallbackQuery.Message.Message.Chat.ID

	args := strings.Split(update.CallbackQuery.Data, " ")
	if len(args) != 3 || args[1] != "on" {
		return
	}

	number, err := strconv.Atoi(args[2])
	if err != nil {
		bh.logger.Error("can't parse number", "msg", err)
		return
	}

	bh.storage.SetCallback(func(c Counter) bool {
		if c.Count <= number {
			b.SendMessage(ctx, bh.Message(b, chatID, fmt.Sprintf("Hey, %s", c)))
			return true
		}
		return false
	})

	msg := fmt.Sprintf("Ok, I'll ping you once there are %d people on the wall.", number)
	b.SendMessage(ctx, bh.Message(b, chatID, msg))
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

	if update.CallbackQuery.Data == "gym_in" {
		msg := "Have a great climb!"
		err := bh.storage.GetGym().In()
		if err != nil {
			msg = err.Error()
		}
		b.SendMessage(ctx, bh.Message(b, chatID, msg))
	}

	if update.CallbackQuery.Data == "gym_out" {
		msg, err := bh.storage.GetGym().Out()
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
