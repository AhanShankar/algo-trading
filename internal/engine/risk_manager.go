package engine

import "github.com/AhanShankar/algo-trading/internal/trading"

// PortfolioView is the read-only slice of the portfolio the risk manager is
// permitted to see. It is defined here (the consumer) so risk cannot mutate
// positions or reach methods it has no business calling. trading.Portfolio
// satisfies it implicitly.
type PortfolioView interface {
	Position(trading.Instrument) (trading.Position, bool)
	Equity(map[string]trading.Money) (trading.Money, error)
}

type RiskManager interface {
	Evaluate(signals []trading.SignalEvent, portfolio PortfolioView) ([]trading.Order, error)
}
