package trading

import (
	"errors"
	"fmt"
)

var ErrPriceUnavailable = errors.New("portfolio: price unavailable for instrument")

type Portfolio struct {
	cash      Money
	positions map[string]Position // keyed by Instrument.Key()
}

func NewPortfolio(cash Money) *Portfolio {
	positions := make(map[string]Position)
	return &Portfolio{
		cash,
		positions,
	}
}

// Position returns the current position for an instrument and whether one exists.
// Position is a value type, so the caller receives a copy — the internal map is
// never exposed and cannot be mutated through this accessor.
func (p *Portfolio) Position(instrument Instrument) (Position, bool) {
	pos, ok := p.positions[instrument.Key()]
	return pos, ok
}
func (p *Portfolio) updatePosition(instrument Instrument, position Position) {
	p.positions[instrument.Key()] = position
}
func (p *Portfolio) Equity(prices map[string]Money) (Money, error) {
	total := p.cash
	for key, pos := range p.positions {
		price, ok := prices[key]
		if !ok {
			return Money{}, fmt.Errorf("valuing %s: %w", key, ErrPriceUnavailable)
		}
		value := price.Mul(pos.Quantity) // signed quantity → shorts subtract
		sum, err := total.Add(value)
		if err != nil {
			return Money{}, fmt.Errorf("valuing %s: %w", key, err)
		}
		total = sum
	}
	return total, nil
}
func (p *Portfolio) ApplyFill(fill Fill) error {
	currentPosition, exists := p.Position(fill.Instrument)
	if !exists {
		currentPosition = Position{
			CostBasis:   Money{Currency: fill.Price.Currency},
			RealizedPnL: Money{Currency: fill.Price.Currency},
		}
	}

	cost := fill.Price.Mul(fill.Quantity)
	newCostBasis, err := currentPosition.CostBasis.Add(cost)
	if err != nil {
		return fmt.Errorf("applying fill for %s: %w", fill.Instrument.Key(), err)
	}
	newPosition := Position{
		Instrument: fill.Instrument,
		Quantity:   currentPosition.Quantity + fill.Quantity,
		CostBasis:  newCostBasis,
	}
	p.updatePosition(fill.Instrument, newPosition)
	return nil
}
