package engine

import "github.com/AhanShankar/algo-trading/internal/trading"

type Strategy interface {
	OnMarketEvent(trading.MarketEvent) []trading.SignalEvent
	Instruments() []trading.Instrument
}
