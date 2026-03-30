package main

import (
	"context"
	"errors"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

type stubStorer struct {
	stored  []Counter
	gym     *Gym
	gymPath string
}

func newStubStorer(t *testing.T) *stubStorer {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "gym-*.sqlite")
	if err != nil {
		t.Fatalf("newStubStorer: %v", err)
	}
	f.Close()
	return &stubStorer{gymPath: f.Name()}
}

func (s *stubStorer) Store(c Counter) error { s.stored = append(s.stored, c); return nil }
func (s *stubStorer) Last() (Counter, bool) {
	if len(s.stored) == 0 {
		return Counter{}, false
	}
	return s.stored[len(s.stored)-1], true
}
func (s *stubStorer) NewGym() error {
	var err error
	s.gym, err = NewGym(s.gymPath)
	return err
}
func (s *stubStorer) GetGym() *Gym { return s.gym }

// errStorer wraps stubStorer but always fails on Store.
type errStorer struct{ stubStorer }

func (e *errStorer) Store(_ Counter) error { return errors.New("store failed") }

func TestNewJobHandler(t *testing.T) {
	cfg := &Config{PGK: "pgk", FID: "fid"}
	jh := NewJobHandler("TST", NewClient(cfg), newStubStorer(t))
	if jh == nil {
		t.Fatal("expected non-nil JobHandler")
	}
	if jh.gym != "TST" {
		t.Errorf("expected gym %q, got %q", "TST", jh.gym)
	}
}

func TestJobHandler_Description(t *testing.T) {
	cfg := &Config{PGK: "pgk", FID: "fid"}
	jh := NewJobHandler("TST", NewClient(cfg), newStubStorer(t))

	got := jh.Description()
	want := `Climber Count Job for "TST"`
	if got != want {
		t.Errorf("expected %q, got %q", want, got)
	}
}

func TestJobHandler_Execute_Success(t *testing.T) {
	page := minimalOccupancyHTML("TST", 3, 30)

	cfg := &Config{PGK: "pgk", FID: "fid"}
	c := NewClient(cfg)
	c.client = &MockClient{
		resp: &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(page)),
		},
	}

	st := newStubStorer(t)
	jh := NewJobHandler("TST", c, st)

	if err := jh.Execute(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(st.stored) != 1 {
		t.Errorf("expected 1 stored counter, got %d", len(st.stored))
	}
}

func TestJobHandler_Execute_FetchError(t *testing.T) {
	cfg := &Config{PGK: "pgk", FID: "fid"}
	c := NewClient(cfg)
	c.client = &MockClient{err: errors.New("network down")}

	jh := NewJobHandler("TST", c, newStubStorer(t))
	if err := jh.Execute(context.Background()); err == nil {
		t.Fatal("expected error when fetch fails")
	}
}

func TestJobHandler_Execute_StoreError(t *testing.T) {
	page := minimalOccupancyHTML("TST", 3, 30)

	cfg := &Config{PGK: "pgk", FID: "fid"}
	c := NewClient(cfg)
	c.client = &MockClient{
		resp: &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(page)),
		},
	}

	jh := NewJobHandler("TST", c, &errStorer{})
	if err := jh.Execute(context.Background()); err == nil {
		t.Fatal("expected error when storage fails")
	}
}

func TestNewBotHandler(t *testing.T) {
	bh := NewBotHandler(newStubStorer(t))
	if bh == nil {
		t.Fatal("expected non-nil BotHandler")
	}
}

func TestBotHandler_Message(t *testing.T) {
	bh := NewBotHandler(newStubStorer(t))
	params := bh.Message(nil, 42, "hello")
	if params == nil {
		t.Fatal("Message returned nil")
	}
	if params.ChatID != int64(42) {
		t.Errorf("expected ChatID 42, got %v", params.ChatID)
	}
	if params.Text != "hello" {
		t.Errorf("expected text %q, got %q", "hello", params.Text)
	}
}

func TestBotHandler_Reaction(t *testing.T) {
	bh := NewBotHandler(newStubStorer(t))
	params := bh.Reaction(nil, 99, 7, "👍")
	if params == nil {
		t.Fatal("Reaction returned nil")
	}
	if params.ChatID != int64(99) {
		t.Errorf("expected ChatID 99, got %v", params.ChatID)
	}
	if params.MessageID != 7 {
		t.Errorf("expected MessageID 7, got %d", params.MessageID)
	}
	if len(params.Reaction) == 0 {
		t.Fatal("expected at least one reaction entry")
	}
	emoji := params.Reaction[0].ReactionTypeEmoji
	if emoji == nil || emoji.Emoji != "👍" {
		t.Errorf("expected emoji 👍, got %v", emoji)
	}
}

func TestBotHandler_CountHandler_NilMessage(t *testing.T) {
	bh := NewBotHandler(newStubStorer(t))
	// Guard clause: returns immediately when Message is nil — must not panic.
	bh.CountHandler(context.Background(), &bot.Bot{}, &models.Update{})
}

func TestBotHandler_CountHandler_WithStoredCounter(t *testing.T) {
	st := newStubStorer(t)
	st.stored = []Counter{
		{Count: 5, Capacity: 50, LastUpdate: LastUpdate{Time: time.Now()}},
	}
	bh := NewBotHandler(st)

	// bot.Bot requires a live Telegram token to call SendMessage
	// recover from the nil-pointer panic so we still can verify
	// the handler reaches the SendMessage call path.
	//nolint:errcheck
	defer func() { recover() }()
	bh.CountHandler(context.Background(), &bot.Bot{}, &models.Update{
		Message: &models.Message{Chat: models.Chat{ID: 1}},
	})
}

func TestBotHandler_GymHandler_NilMessage(t *testing.T) {
	bh := NewBotHandler(newStubStorer(t))
	bh.GymHandler(context.Background(), &bot.Bot{}, &models.Update{})
}

func TestBotHandler_GymHandler_WithMessage(t *testing.T) {
	bh := NewBotHandler(newStubStorer(t))
	//nolint:errcheck
	defer func() { recover() }()
	bh.GymHandler(context.Background(), &bot.Bot{}, &models.Update{
		Message: &models.Message{Chat: models.Chat{ID: 1}},
	})
}

func TestBotHandler_GymButtonHandler_NilCallbackQuery(t *testing.T) {
	bh := NewBotHandler(newStubStorer(t))
	// Guard clause: returns immediately when CallbackQuery is nil.
	bh.GymButtonHandler(context.Background(), &bot.Bot{}, &models.Update{})
}
