package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/reugn/go-quartz/logger"
	"github.com/reugn/go-quartz/quartz"
)

func main() {
	SetLogger()

	cfg, err := NewConfig()
	if err != nil {
		log.Fatal(err)
	}

	// Fetch counters once to discover all available gym keys.
	client := NewClient(cfg)
	counters, err := client.Counters()
	if err != nil {
		log.Fatal(err)
	}

	// Build one Storage (and Gym) per gym found in the scraped counters.
	storers := make(map[string]Storer)
	for gymKey := range *counters {
		st, err := NewStorage(cfg.Storage, gymKey)
		if err != nil {
			log.Fatalf("create storage for gym %q: %v", gymKey, err)
		}
		if err := st.NewGym(); err != nil {
			log.Fatalf("init gym for %q: %v", gymKey, err)
		}
		storers[gymKey] = st
	}

	if len(storers) == 0 {
		log.Fatal("no gyms found in scraped data")
	}

	// The configured GYM drives the Telegram bot responses.
	botStorage, ok := storers[cfg.Gym]
	if !ok {
		// Fallback: use the first storer (shouldn't happen in normal operation).
		slog.Warn("configured GYM not found in scraped data, falling back",
			"gym", cfg.Gym,
			"available", gymKeys(storers),
		)
		for _, s := range storers {
			botStorage = s
			break
		}
	}

	jh := NewJobHandler(cfg.Storage, client, storers)
	bh := NewBotHandler(botStorage)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	// Store the freshly-fetched counters for all gyms immediately.
	if err := jh.Execute(ctx); err != nil {
		log.Fatal(err)
	}

	loc := time.Now().Location()
	sched, err := quartz.NewStdScheduler(quartz.WithLogger(logger.NoOpLogger{}))
	if err != nil {
		log.Fatal(err)
	}
	sched.Start(ctx)

	for key, crontab := range cfg.Schedule {
		slog.Info("schedule job", "job_key", key, "crontab", crontab, "loc", loc)
		cronTrigger, _ := quartz.NewCronTriggerWithLoc(crontab, loc)
		err := sched.ScheduleJob(quartz.NewJobDetail(jh, quartz.NewJobKey(key)), cronTrigger)
		if err != nil {
			log.Fatal(err)
		}
	}

	defer func() {
		sched.Stop()
		sched.Wait(ctx)
	}()

	opts := []bot.Option{
		bot.WithDefaultHandler(func(ctx context.Context, b *bot.Bot, update *models.Update) {}),
		bot.WithDebugHandler(func(format string, args ...any) {
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
	b.RegisterHandler(bot.HandlerTypeMessageText, "/gym", bot.MatchTypeExact, bh.GymHandler)
	b.RegisterHandler(bot.HandlerTypeCallbackQueryData, "gym", bot.MatchTypePrefix, bh.GymButtonHandler)

	b.Start(ctx)
}

func gymKeys(m map[string]Storer) string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return strings.Join(keys, ", ")
}
