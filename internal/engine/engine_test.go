package engine_test

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/AhanShankar/algo-trading/internal/engine"
	"github.com/AhanShankar/algo-trading/internal/trading"
)

// --- fakes ---

type fakeMarketDataProvider struct {
	ch  chan trading.MarketEvent
	err error
}

func (f *fakeMarketDataProvider) Subscribe(_ context.Context, _ ...trading.Instrument) (<-chan trading.MarketEvent, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.ch, nil
}

type fakeBroker struct {
	mu             sync.Mutex
	placedOrders   [][]trading.Order
	placeOrderFunc func(orders []trading.Order) []trading.OrderResult
	fillCh         chan trading.Fill
}

func (f *fakeBroker) PlaceOrder(_ context.Context, orders []trading.Order) []trading.OrderResult {
	f.mu.Lock()
	f.placedOrders = append(f.placedOrders, orders)
	f.mu.Unlock()
	if f.placeOrderFunc != nil {
		return f.placeOrderFunc(orders)
	}
	return nil
}

func (f *fakeBroker) GetOrderStatus(_ context.Context, order trading.Order) (trading.Order, error) {
	return order, nil
}

func (f *fakeBroker) Fills() <-chan trading.Fill {
	return f.fillCh
}

func (f *fakeBroker) placedOrderCount() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.placedOrders)
}

type fakeRiskManager struct {
	evaluateFunc func(signals []trading.SignalEvent, portfolio engine.PortfolioView) ([]trading.Order, error)
}

func (f *fakeRiskManager) Evaluate(signals []trading.SignalEvent, portfolio engine.PortfolioView) ([]trading.Order, error) {
	if f.evaluateFunc != nil {
		return f.evaluateFunc(signals, portfolio)
	}
	return nil, nil
}

type fakeStrategy struct {
	instruments       []trading.Instrument
	onMarketEventFunc func(trading.MarketEvent) []trading.SignalEvent
}

func (f *fakeStrategy) OnMarketEvent(ev trading.MarketEvent) []trading.SignalEvent {
	if f.onMarketEventFunc != nil {
		return f.onMarketEventFunc(ev)
	}
	return nil
}

func (f *fakeStrategy) Instruments() []trading.Instrument {
	return f.instruments
}

type fakeStore struct{}

func (f *fakeStore) PersistOrders(_ []trading.Order) error   { return nil }
func (f *fakeStore) PersistFills(_ []trading.Fill) error     { return nil }
func (f *fakeStore) PersistCandles(_ []trading.Candle) error { return nil }

// --- helpers ---

func newTestLogger() (*slog.Logger, *bytes.Buffer) {
	var buf bytes.Buffer
	return slog.New(slog.NewTextHandler(&buf, nil)), &buf
}

// waitForRun waits for Run to return, failing the test if it takes too long.
func waitForRun(t *testing.T, errCh <-chan error) error {
	t.Helper()
	select {
	case err := <-errCh:
		return err
	case <-time.After(2 * time.Second):
		t.Fatal("engine.Run did not return in time")
		return nil
	}
}

func testInstrument() trading.Instrument {
	return trading.Instrument{Ticker: "INFY", Exchange: "NSE", Type: trading.Equity}
}

// --- tests ---

func TestEngine_Run_HappyPath(t *testing.T) {
	instrument := testInstrument()
	marketCh := make(chan trading.MarketEvent)
	fillCh := make(chan trading.Fill)

	order := trading.Order{
		ClientID:   "client-1",
		Instrument: instrument,
		Side:       trading.Buy,
		Quantity:   10,
		Price:      trading.Money{Amount: 100, Currency: trading.INR},
	}

	strategy := &fakeStrategy{
		instruments: []trading.Instrument{instrument},
		onMarketEventFunc: func(trading.MarketEvent) []trading.SignalEvent {
			return []trading.SignalEvent{{Instrument: instrument, Direction: trading.Long}}
		},
	}
	risk := &fakeRiskManager{
		evaluateFunc: func(signals []trading.SignalEvent, _ engine.PortfolioView) ([]trading.Order, error) {
			if len(signals) != 1 {
				t.Fatalf("expected 1 signal, got %d", len(signals))
			}
			return []trading.Order{order}, nil
		},
	}
	broker := &fakeBroker{
		fillCh: fillCh,
		placeOrderFunc: func(orders []trading.Order) []trading.OrderResult {
			return []trading.OrderResult{{ClientID: orders[0].ClientID, BrokerID: "broker-1"}}
		},
	}
	portfolio := trading.NewPortfolio(trading.Money{Currency: trading.INR})
	logger, _ := newTestLogger()

	eng := engine.New(broker, &fakeMarketDataProvider{ch: marketCh}, risk, strategy, &fakeStore{}, portfolio, logger)

	errCh := make(chan error, 1)
	go func() { errCh <- eng.Run(context.Background()) }()

	marketCh <- trading.Tick{Instrument: instrument}
	fillCh <- trading.Fill{
		OrderID:    "broker-1",
		Instrument: instrument,
		Side:       trading.Buy,
		Quantity:   10,
		Price:      trading.Money{Amount: 100, Currency: trading.INR},
	}
	close(marketCh)

	if err := waitForRun(t, errCh); err != nil {
		t.Fatalf("expected Run to return nil, got %v", err)
	}

	if broker.placedOrderCount() != 1 {
		t.Fatalf("expected 1 PlaceOrder call, got %d", broker.placedOrderCount())
	}

	pos, ok := portfolio.Position(instrument)
	if !ok {
		t.Fatal("expected a position to exist after fill")
	}
	if pos.Quantity != 10 {
		t.Errorf("expected quantity 10, got %d", pos.Quantity)
	}
	wantCostBasis := trading.Money{Amount: 1000, Currency: trading.INR}
	if pos.CostBasis != wantCostBasis {
		t.Errorf("expected cost basis %+v, got %+v", wantCostBasis, pos.CostBasis)
	}
}

func TestEngine_Run_RiskEvaluationError(t *testing.T) {
	instrument := testInstrument()
	marketCh := make(chan trading.MarketEvent)
	fillCh := make(chan trading.Fill)

	strategy := &fakeStrategy{
		instruments: []trading.Instrument{instrument},
		onMarketEventFunc: func(trading.MarketEvent) []trading.SignalEvent {
			return []trading.SignalEvent{{Instrument: instrument, Direction: trading.Long}}
		},
	}
	risk := &fakeRiskManager{
		evaluateFunc: func([]trading.SignalEvent, engine.PortfolioView) ([]trading.Order, error) {
			return nil, errors.New("risk check failed")
		},
	}
	broker := &fakeBroker{fillCh: fillCh}
	portfolio := trading.NewPortfolio(trading.Money{Currency: trading.INR})
	logger, buf := newTestLogger()

	eng := engine.New(broker, &fakeMarketDataProvider{ch: marketCh}, risk, strategy, &fakeStore{}, portfolio, logger)

	errCh := make(chan error, 1)
	go func() { errCh <- eng.Run(context.Background()) }()

	marketCh <- trading.Tick{Instrument: instrument}
	close(marketCh)

	if err := waitForRun(t, errCh); err != nil {
		t.Fatalf("expected Run to return nil, got %v", err)
	}

	if broker.placedOrderCount() != 0 {
		t.Fatalf("expected no orders placed after risk evaluation error, got %d", broker.placedOrderCount())
	}
	if !strings.Contains(buf.String(), "risk evaluation failed") {
		t.Errorf("expected log to mention risk evaluation failure, got: %s", buf.String())
	}
}

func TestEngine_Run_OrderPlacementError(t *testing.T) {
	instrument := testInstrument()
	marketCh := make(chan trading.MarketEvent)
	fillCh := make(chan trading.Fill)

	order := trading.Order{
		ClientID:   "client-1",
		Instrument: instrument,
		Side:       trading.Buy,
		Quantity:   5,
		Price:      trading.Money{Amount: 50, Currency: trading.INR},
	}
	strategy := &fakeStrategy{
		instruments: []trading.Instrument{instrument},
		onMarketEventFunc: func(trading.MarketEvent) []trading.SignalEvent {
			return []trading.SignalEvent{{Instrument: instrument, Direction: trading.Long}}
		},
	}
	risk := &fakeRiskManager{
		evaluateFunc: func([]trading.SignalEvent, engine.PortfolioView) ([]trading.Order, error) {
			return []trading.Order{order}, nil
		},
	}
	broker := &fakeBroker{
		fillCh: fillCh,
		placeOrderFunc: func(orders []trading.Order) []trading.OrderResult {
			return []trading.OrderResult{{ClientID: orders[0].ClientID, Err: errors.New("rejected: insufficient margin")}}
		},
	}
	portfolio := trading.NewPortfolio(trading.Money{Currency: trading.INR})
	logger, buf := newTestLogger()

	eng := engine.New(broker, &fakeMarketDataProvider{ch: marketCh}, risk, strategy, &fakeStore{}, portfolio, logger)

	errCh := make(chan error, 1)
	go func() { errCh <- eng.Run(context.Background()) }()

	marketCh <- trading.Tick{Instrument: instrument}
	close(marketCh)

	if err := waitForRun(t, errCh); err != nil {
		t.Fatalf("expected Run to return nil, got %v", err)
	}

	if !strings.Contains(buf.String(), "order not accepted") {
		t.Errorf("expected log to mention order rejection, got: %s", buf.String())
	}
	if _, exists := portfolio.Position(instrument); exists {
		t.Error("expected no position to be created for a rejected order")
	}
}

func TestEngine_Run_FillChannelClosed(t *testing.T) {
	instrument := testInstrument()
	marketCh := make(chan trading.MarketEvent)
	fillCh := make(chan trading.Fill)

	strategy := &fakeStrategy{instruments: []trading.Instrument{instrument}}
	risk := &fakeRiskManager{}
	broker := &fakeBroker{fillCh: fillCh}
	portfolio := trading.NewPortfolio(trading.Money{Currency: trading.INR})
	logger, buf := newTestLogger()

	eng := engine.New(broker, &fakeMarketDataProvider{ch: marketCh}, risk, strategy, &fakeStore{}, portfolio, logger)

	errCh := make(chan error, 1)
	go func() { errCh <- eng.Run(context.Background()) }()

	close(fillCh)
	// The engine should disable the fill case and keep processing market
	// events rather than spinning or crashing.
	marketCh <- trading.Tick{Instrument: instrument}
	close(marketCh)

	if err := waitForRun(t, errCh); err != nil {
		t.Fatalf("expected Run to return nil, got %v", err)
	}
	if !strings.Contains(buf.String(), "fill stream closed") {
		t.Errorf("expected log to mention fill stream closure, got: %s", buf.String())
	}
}

func TestEngine_Run_MarketChannelClosed(t *testing.T) {
	marketCh := make(chan trading.MarketEvent)
	fillCh := make(chan trading.Fill)

	strategy := &fakeStrategy{instruments: []trading.Instrument{testInstrument()}}
	risk := &fakeRiskManager{}
	broker := &fakeBroker{fillCh: fillCh}
	portfolio := trading.NewPortfolio(trading.Money{Currency: trading.INR})
	logger, _ := newTestLogger()

	eng := engine.New(broker, &fakeMarketDataProvider{ch: marketCh}, risk, strategy, &fakeStore{}, portfolio, logger)

	errCh := make(chan error, 1)
	go func() { errCh <- eng.Run(context.Background()) }()

	close(marketCh)

	if err := waitForRun(t, errCh); err != nil {
		t.Fatalf("expected Run to return nil when market data stream closes, got %v", err)
	}
}

func TestEngine_Run_ContextCancelled(t *testing.T) {
	marketCh := make(chan trading.MarketEvent)
	fillCh := make(chan trading.Fill)

	strategy := &fakeStrategy{instruments: []trading.Instrument{testInstrument()}}
	risk := &fakeRiskManager{}
	broker := &fakeBroker{fillCh: fillCh}
	portfolio := trading.NewPortfolio(trading.Money{Currency: trading.INR})
	logger, _ := newTestLogger()

	eng := engine.New(broker, &fakeMarketDataProvider{ch: marketCh}, risk, strategy, &fakeStore{}, portfolio, logger)

	ctx, cancel := context.WithCancel(context.Background())
	errCh := make(chan error, 1)
	go func() { errCh <- eng.Run(ctx) }()

	cancel()

	err := waitForRun(t, errCh)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
}

func TestEngine_Run_SubscribeError(t *testing.T) {
	wantErr := errors.New("subscribe failed")
	strategy := &fakeStrategy{instruments: []trading.Instrument{testInstrument()}}
	risk := &fakeRiskManager{}
	broker := &fakeBroker{fillCh: make(chan trading.Fill)}
	portfolio := trading.NewPortfolio(trading.Money{Currency: trading.INR})
	logger, _ := newTestLogger()

	eng := engine.New(broker, &fakeMarketDataProvider{err: wantErr}, risk, strategy, &fakeStore{}, portfolio, logger)

	err := eng.Run(context.Background())
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected error to wrap %v, got %v", wantErr, err)
	}
}

func TestEngine_Run_FillApplyError(t *testing.T) {
	instrument := testInstrument()
	marketCh := make(chan trading.MarketEvent)
	fillCh := make(chan trading.Fill)

	strategy := &fakeStrategy{instruments: []trading.Instrument{instrument}}
	risk := &fakeRiskManager{}
	broker := &fakeBroker{fillCh: fillCh}
	portfolio := trading.NewPortfolio(trading.Money{Currency: trading.INR})
	logger, buf := newTestLogger()

	eng := engine.New(broker, &fakeMarketDataProvider{ch: marketCh}, risk, strategy, &fakeStore{}, portfolio, logger)

	errCh := make(chan error, 1)
	go func() { errCh <- eng.Run(context.Background()) }()

	// First fill establishes a position priced in INR.
	fillCh <- trading.Fill{
		OrderID:    "broker-1",
		Instrument: instrument,
		Side:       trading.Buy,
		Quantity:   10,
		Price:      trading.Money{Amount: 100, Currency: trading.INR},
	}
	// Second fill uses a mismatched currency, which ApplyFill must reject
	// rather than silently corrupt the position.
	fillCh <- trading.Fill{
		OrderID:    "broker-2",
		Instrument: instrument,
		Side:       trading.Buy,
		Quantity:   5,
		Price:      trading.Money{Amount: 50, Currency: "USD"},
	}
	close(marketCh)

	if err := waitForRun(t, errCh); err != nil {
		t.Fatalf("expected Run to return nil, got %v", err)
	}

	if !strings.Contains(buf.String(), "failed to apply fill") {
		t.Errorf("expected log to mention fill application failure, got: %s", buf.String())
	}

	pos, ok := portfolio.Position(instrument)
	if !ok {
		t.Fatal("expected the position from the first fill to remain")
	}
	if pos.Quantity != 10 {
		t.Errorf("expected quantity to remain 10 after the failed second fill, got %d", pos.Quantity)
	}
}
