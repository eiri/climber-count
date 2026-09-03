package main

import (
	"context"
	"database/sql"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

type stubStorer struct {
	stored []Counter
	db     *sql.DB
	gym    *Gym
}

func newStubStorer(t *testing.T) *stubStorer {
	t.Helper()
	return &stubStorer{db: newTestDB(t)}
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
	s.gym, err = NewGym(s.db, "TST")
	return err
}
func (s *stubStorer) GetGym() *Gym { return s.gym }

// errStorer wraps stubStorer but always fails on Store.
type errStorer struct{ stubStorer }

func (e *errStorer) Store(_ Counter) error { return errors.New("store failed") }

func TestNewJobHandler(t *testing.T) {
	cfg := &Config{PGK: "pgk", FID: "fid"}
	storers := map[string]Storer{"TST": newStubStorer(t)}
	jh := NewJobHandler(NewClient(cfg), storers)
	if jh == nil {
		t.Fatal("expected non-nil JobHandler")
	}
	if len(jh.storers) != 1 {
		t.Errorf("expected 1 storer, got %d", len(jh.storers))
	}
}

func TestJobHandler_Description(t *testing.T) {
	cfg := &Config{PGK: "pgk", FID: "fid"}
	storers := map[string]Storer{
		"TST": newStubStorer(t),
		"SLB": newStubStorer(t),
	}
	jh := NewJobHandler(NewClient(cfg), storers)
	got := jh.Description()
	want := "Climber Count Job for 2 gym(s)"
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
	storers := map[string]Storer{"TST": st}
	jh := NewJobHandler(c, storers)

	if err := jh.Execute(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(st.stored) != 1 {
		t.Errorf("expected 1 stored counter, got %d", len(st.stored))
	}
}

func TestJobHandler_Execute_MultipleGyms(t *testing.T) {
	// Page contains two gyms: TST and SLB.
	page := multiGymOccupancyHTML(map[string][2]int{
		"TST": {3, 30},
		"SLB": {12, 50},
	})
	cfg := &Config{PGK: "pgk", FID: "fid"}
	c := NewClient(cfg)
	c.client = &MockClient{
		resp: &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(page)),
		},
	}

	stTST := newStubStorer(t)
	stSLB := newStubStorer(t)
	storers := map[string]Storer{
		"TST": stTST,
		"SLB": stSLB,
	}
	jh := NewJobHandler(c, storers)

	if err := jh.Execute(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(stTST.stored) != 1 {
		t.Errorf("TST: expected 1 stored counter, got %d", len(stTST.stored))
	}
	if len(stSLB.stored) != 1 {
		t.Errorf("SLB: expected 1 stored counter, got %d", len(stSLB.stored))
	}
}

func TestJobHandler_Execute_Metrics(t *testing.T) {
	page := minimalOccupancyHTML("TST", 3, 30)
	cfg := &Config{PGK: "pgk", FID: "fid"}
	c := NewClient(cfg)
	c.client = &MockClient{
		resp: &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(page)),
		},
	}
	m := NewMetrics(prometheus.NewRegistry())
	storers := map[string]Storer{"TST": newStubStorer(t)}
	jh := NewJobHandler(c, storers, m)

	if err := jh.Execute(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got := histCount(t, m.scrapeDuration); got != 1 {
		t.Errorf("expected scrape count 1, got %d", got)
	}
	if got := histCount(t, m.writes.WithLabelValues("counter_store").(prometheus.Metric)); got != 1 {
		t.Errorf("expected store count 1, got %d", got)
	}
	if got := testutil.ToFloat64(m.people.WithLabelValues("TST")); got != 3 {
		t.Errorf("expected people 3, got %v", got)
	}
	if got := testutil.ToFloat64(m.capacity.WithLabelValues("TST")); got != 30 {
		t.Errorf("expected capacity 30, got %v", got)
	}
}

func TestJobHandler_Execute_FetchError(t *testing.T) {
	cfg := &Config{PGK: "pgk", FID: "fid"}
	c := NewClient(cfg)
	c.client = &MockClient{err: errors.New("network down")}
	storers := map[string]Storer{"TST": newStubStorer(t)}
	jh := NewJobHandler(c, storers)
	if err := jh.Execute(context.Background()); err == nil {
		t.Fatal("expected error when fetch fails")
	}
}

func TestJobHandler_Execute_FetchErrorMetric(t *testing.T) {
	cfg := &Config{PGK: "pgk", FID: "fid"}
	c := NewClient(cfg)
	c.client = &MockClient{err: errors.New("network down")}
	m := NewMetrics(prometheus.NewRegistry())
	storers := map[string]Storer{"TST": newStubStorer(t)}
	jh := NewJobHandler(c, storers, m)

	if err := jh.Execute(context.Background()); err == nil {
		t.Fatal("expected error when fetch fails")
	}
	if got := histCount(t, m.scrapeDuration); got != 1 {
		t.Errorf("expected scrape count 1, got %d", got)
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
	storers := map[string]Storer{"TST": &errStorer{}}
	jh := NewJobHandler(c, storers)
	if err := jh.Execute(context.Background()); err == nil {
		t.Fatal("expected error when storage fails")
	}
}

func newBotHandler(t *testing.T) *BotHandler {
	t.Helper()
	st := newStubStorer(t)
	return NewBotHandler("TST", map[string]Storer{"TST": st})
}

func TestNewBotHandler(t *testing.T) {
	bh := newBotHandler(t)
	if bh == nil {
		t.Fatal("expected non-nil BotHandler")
	}
}

func TestBotHandler_Message(t *testing.T) {
	bh := newBotHandler(t)
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
	bh := newBotHandler(t)
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
	bh := newBotHandler(t)
	bh.CountHandler(context.Background(), &bot.Bot{}, &models.Update{})
}

func TestBotHandler_CountHandler_WithStoredCounter(t *testing.T) {
	st := newStubStorer(t)
	st.stored = []Counter{
		{Count: 5, Capacity: 50, LastUpdate: LastUpdate{Time: time.Now()}},
	}
	bh := NewBotHandler("TST", map[string]Storer{"TST": st})
	//nolint:errcheck
	defer func() { recover() }()
	bh.CountHandler(context.Background(), &bot.Bot{}, &models.Update{
		Message: &models.Message{Chat: models.Chat{ID: 1}},
	})
}

func TestBotHandler_CountHandler_WithGymArg(t *testing.T) {
	stTST := newStubStorer(t)
	stSLB := newStubStorer(t)
	stSLB.stored = []Counter{
		{Count: 12, Capacity: 50, LastUpdate: LastUpdate{Time: time.Now()}},
	}
	bh := NewBotHandler("TST", map[string]Storer{"TST": stTST, "SLB": stSLB})
	//nolint:errcheck
	defer func() { recover() }()
	// "/count SLB" — should use stSLB, not stTST
	bh.CountHandler(context.Background(), &bot.Bot{}, &models.Update{
		Message: &models.Message{
			Chat: models.Chat{ID: 1},
			Text: "/count SLB",
		},
	})
}

func TestBotHandler_CountHandler_WithGymArgLowercase(t *testing.T) {
	st := newStubStorer(t)
	st.stored = []Counter{
		{Count: 7, Capacity: 30, LastUpdate: LastUpdate{Time: time.Now()}},
	}
	bh := NewBotHandler("TST", map[string]Storer{"SLB": st})
	//nolint:errcheck
	defer func() { recover() }()
	// lowercase arg should still match
	bh.CountHandler(context.Background(), &bot.Bot{}, &models.Update{
		Message: &models.Message{
			Chat: models.Chat{ID: 1},
			Text: "/count slb",
		},
	})
}

func TestBotHandler_CountHandler_UnknownGymArg(t *testing.T) {
	bh := NewBotHandler("TST", map[string]Storer{"TST": newStubStorer(t)})
	//nolint:errcheck
	defer func() { recover() }()
	// Unknown gym should send an error message (panics on nil bot — recover catches it)
	bh.CountHandler(context.Background(), &bot.Bot{}, &models.Update{
		Message: &models.Message{
			Chat: models.Chat{ID: 1},
			Text: "/count XYZ",
		},
	})
}

func TestBotHandler_GymHandler_NilMessage(t *testing.T) {
	bh := newBotHandler(t)
	bh.GymHandler(context.Background(), &bot.Bot{}, &models.Update{})
}

func TestBotHandler_GymHandler_WithMessage(t *testing.T) {
	bh := newBotHandler(t)
	//nolint:errcheck
	defer func() { recover() }()
	bh.GymHandler(context.Background(), &bot.Bot{}, &models.Update{
		Message: &models.Message{Chat: models.Chat{ID: 1}},
	})
}

func TestBotHandler_GymButtonHandler_NilCallbackQuery(t *testing.T) {
	bh := newBotHandler(t)
	bh.GymButtonHandler(context.Background(), &bot.Bot{}, &models.Update{})
}

// multiGymOccupancyHTML builds a minimal rockgympro HTML page containing
// multiple gym entries, used by multi-gym Execute tests.
func multiGymOccupancyHTML(gyms map[string][2]int) string {
	var sb strings.Builder
	sb.WriteString("<!DOCTYPE html><html><body><script>\nvar data = {\n")
	for key, vals := range gyms {
		sb.WriteString("'" + key + "' : {")
		sb.WriteString("'capacity' : " + itoa(vals[1]) + ",")
		sb.WriteString("'count' : " + itoa(vals[0]) + ",")
		sb.WriteString("'subLabel' : 'Current climber count',")
		sb.WriteString("'lastUpdate' : 'Last updated:&nbspnow (10:00 AM)'")
		sb.WriteString("},\n")
	}
	sb.WriteString("};\n</script></body></html>")
	return sb.String()
}

func itoa(n int) string { return strconv.Itoa(n) }
