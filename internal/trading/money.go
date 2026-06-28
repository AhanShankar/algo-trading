package trading

import (
	"errors"
	"fmt"
)

type Money struct {
	Amount   int64
	Currency Currency
}
type Currency string

const (
	INR Currency = "INR"
)

var ErrCurrencyMismatch = errors.New("money: currency mismatch")

func (m Money) Add(other Money) (Money, error) {
	if m.Currency != other.Currency {
		return Money{}, fmt.Errorf("add %s to %s: %w", other.Currency, m.Currency, ErrCurrencyMismatch)
	}
	return Money{Amount: m.Amount + other.Amount, Currency: m.Currency}, nil
}

func (m Money) Sub(other Money) (Money, error) {
	if m.Currency != other.Currency {
		return Money{}, fmt.Errorf("sub %s from %s: %w", other.Currency, m.Currency, ErrCurrencyMismatch)
	}
	return Money{Amount: m.Amount - other.Amount, Currency: m.Currency}, nil
}

func (m Money) Mul(factor int64) Money {
	return Money{Amount: m.Amount * factor, Currency: m.Currency}
}

func (m Money) Cmp(other Money) (int, error) {
	if m.Currency != other.Currency {
		return 0, fmt.Errorf("compare %s with %s: %w", m.Currency, other.Currency, ErrCurrencyMismatch)
	}
	switch {
	case m.Amount < other.Amount:
		return -1, nil
	case m.Amount > other.Amount:
		return 1, nil
	default:
		return 0, nil
	}
}

func (m Money) IsZero() bool { return m.Amount == 0 }
