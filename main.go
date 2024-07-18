package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"time"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/reugn/go-quartz/quartz"
)

func main() {
	SetLogger()

	cfg, err := NewConfig()
	if err != nil {
		log.Fatal(err)
	}

	client := NewClient(cfg)
	storage, err := NewStorage(cfg.Storage)
	if err != nil {
		log.Fatal(err)
	}
	err = storage.NewGym()
	if err != nil {
		log.Fatal(err)
	}

	jh := NewJobHandler(cfg.Gym, client, storage)
	bh := NewBotHandler(storage)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()
	// pull and store the latest count
	err = jh.Execute(ctx)
	if err != nil {
		log.Fatal(err)
	}

	loc := time.Now().Location()
	sched := quartz.NewStdScheduler()
	sched.Start(ctx)
	for key, crontab := range cfg.Schedule {
		slog.Info("schedule job", "job_key", key, "crontab", crontab, "loc", loc)
		cronTrigger, _ := quartz.NewCronTriggerWithLoc(crontab, loc)
		sched.ScheduleJob(quartz.NewJobDetail(jh, quartz.NewJobKey(key)), cronTrigger)
	}
	// shutdown sched on exit
	defer func() {
		sched.Stop()
		sched.Wait(ctx)
	}()

	opts := []bot.Option{
		// just no-op
		bot.WithDefaultHandler(func(ctx context.Context, b *bot.Bot, update *models.Update) {}),
		bot.WithDebugHandler(func(format string, args ...interface{}) {
			slog.Debug(fmt.Sprintf(format, args), "component", "telegram bot")
		}),
		bot.WithErrorsHandler(func(err error) {
			slog.Error("telegram error", "msg", err, "component", "telegram bot")
		}),
	}

	b, err := bot.New(cfg.BotToken, opts...)
	if nil != err {
		log.Fatal(err)
	}

	b.RegisterHandler(bot.HandlerTypeMessageText, "/count", bot.MatchTypeExact, bh.CountHandler)
	b.RegisterHandler(bot.HandlerTypeMessageText, "ping", bot.MatchTypePrefix, bh.PingHandler)
	b.RegisterHandler(bot.HandlerTypeCallbackQueryData, "ping", bot.MatchTypePrefix, bh.PingButtonHandler)
	b.RegisterHandler(bot.HandlerTypeMessageText, "/gym", bot.MatchTypeExact, bh.GymHandler)
	b.RegisterHandler(bot.HandlerTypeCallbackQueryData, "gym", bot.MatchTypePrefix, bh.GymButtonHandler)

	b.Start(ctx)
}
