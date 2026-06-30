package engine

import (
	"context"

	"github.com/AhanShankar/algo-trading/internal/trading"
)

type MarketDataProvider interface {
	Subscribe(ctx context.Context, instrument ...trading.Instrument) (<-chan trading.MarketEvent, error)
}
