package main

import (
	"context"
	"log"
	"os"
	"os/signal"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/reugn/go-quartz/quartz"
)

func main() {
	cfg, err := NewConfig()
	if err != nil {
		log.Fatal(err)
	}

	client := NewClient(cfg)
	storage := NewStorage(cfg.Storage)
	jh := NewJobHandler(cfg.Gym, client, storage)
	bh := NewBotHandler(storage)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()
	// pull and store the latest count
	err = jh.Execute(ctx)
	if err != nil {
		log.Fatal(err)
	}

	sched := quartz.NewStdScheduler()
	sched.Start(ctx)
	// pull the count every 5 minutes
	crontab := "0 */5 * * * *"
	cronTrigger, _ := quartz.NewCronTrigger(crontab)
	sched.ScheduleJob(quartz.NewJobDetail(jh, quartz.NewJobKey(cfg.Gym)), cronTrigger)
	defer func() {
		sched.Stop()
		sched.Wait(ctx)
	}()

	opts := []bot.Option{
		// just no-op
		bot.WithDefaultHandler(func(ctx context.Context, b *bot.Bot, update *models.Update) {}),
	}

	b, err := bot.New(cfg.BotToken, opts...)
	if nil != err {
		log.Fatal(err)
	}

	b.RegisterHandler(bot.HandlerTypeMessageText, "/count", bot.MatchTypeExact, bh.Handler)

	b.Start(ctx)
}
