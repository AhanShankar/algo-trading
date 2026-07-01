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

func (p Portfolio) Equity(prices map[string]Money) (Money, error) {
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
