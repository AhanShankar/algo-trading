package engine

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/AhanShankar/algo-trading/internal/trading"
)

type Engine struct {
	broker             Broker
	marketDataProvider MarketDataProvider
	riskManager        RiskManager
	strategy           Strategy
	store              Store
	portfolio          *trading.Portfolio
	log                *slog.Logger
}

func New(broker Broker, marketDataProvider MarketDataProvider, riskManager RiskManager, strategy Strategy, store Store, portfolio *trading.Portfolio, log *slog.Logger) *Engine {
	return &Engine{
		broker:             broker,
		marketDataProvider: marketDataProvider,
		riskManager:        riskManager,
		strategy:           strategy,
		store:              store,
		portfolio:          portfolio,
		log:                log,
	}
}
func (e *Engine) Run(ctx context.Context) error {
	marketCh, err := e.marketDataProvider.Subscribe(ctx, e.strategy.Instruments()...)
	if err != nil {
		// Setup failure: there is no session without market data. Fatal.
		return fmt.Errorf("run: subscribe market data: %w", err)
	}
	fillCh := e.broker.Fills()
	for {
		select {
		case ev, ok := <-marketCh:
			if !ok {
				e.log.Info("market data stream closed; stopping engine")
				return nil
			}
			signals := e.strategy.OnMarketEvent(ev)
			orders, err := e.riskManager.Evaluate(signals, e.portfolio)
			if err != nil {
				e.log.Error("risk evaluation failed; skipping event",
					"err", err,
					"signalCount", len(signals),
					"instruments", signalInstruments(signals))
				continue
			}
			orderResults := e.broker.PlaceOrder(ctx, orders)
			ordersByClientID := make(map[string]trading.Order, len(orders))
			for _, o := range orders {
				ordersByClientID[o.ClientID] = o
			}
			for _, orderResult := range orderResults {
				if orderResult.Err != nil {
					o := ordersByClientID[orderResult.ClientID]
					e.log.Warn("order not accepted",
						"clientID", orderResult.ClientID,
						"ticker", o.Instrument.Ticker,
						"exchange", o.Instrument.Exchange,
						"side", o.Side,
						"quantity", o.Quantity,
						"err", orderResult.Err)
				}
			}
		case fill, ok := <-fillCh:
			if !ok {
				// The broker closed its fill stream while we're still running.
				// We can't track executions anymore, but market data may still
				// flow. Disable this case (nil channel is never selected) so the
				// loop doesn't spin, and keep going. Whether this should instead
				// be fatal is a real safety decision — revisit in Phase 6.
				e.log.Warn("fill stream closed; execution tracking disabled")
				fillCh = nil
				continue
			}
			if err := e.portfolio.ApplyFill(fill); err != nil {
				e.log.Error("failed to apply fill",
					"orderID", fill.OrderID,
					"ticker", fill.Instrument.Ticker,
					"exchange", fill.Instrument.Exchange,
					"side", fill.Side,
					"quantity", fill.Quantity,
					"err", err)
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

// signalInstruments returns the instrument keys of a batch of signals, for
// logging which instruments were involved when an event fails to process.
func signalInstruments(signals []trading.SignalEvent) []string {
	keys := make([]string, len(signals))
	for i, s := range signals {
		keys[i] = s.Instrument.Key()
	}
	return keys
}
