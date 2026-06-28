package trading

import (
	"errors"
	"testing"
)

func TestPositionAvgPrice(t *testing.T) {
	tests := []struct {
		name      string
		qty       int64
		costBasis int64
		want      int64
	}{
		{"long exact", 10, 1000000, 100000},          // 10 @ ₹1000.00
		{"short uses magnitude", -5, 500000, 100000}, // -5 @ ₹1000.00
		{"rounds down", 3, 300001, 100000},           // truncated, not 100000.33
		{"flat is zero", 0, 0, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := Position{Quantity: tt.qty, CostBasis: Money{tt.costBasis, INR}}
			if got := p.AvgPrice(); got.Amount != tt.want {
				t.Errorf("AvgPrice = %d, want %d", got.Amount, tt.want)
			}
		})
	}
}

func TestPortfolioEquity(t *testing.T) {
	reliance := Instrument{Ticker: "RELIANCE", Exchange: "NSE", Type: Equity}
	infy := Instrument{Ticker: "INFY", Exchange: "NSE", Type: Equity}

	p := Portfolio{
		Cash: Money{100000, INR}, // ₹1000.00
		Positions: map[string]Position{
			reliance.Key(): {Instrument: reliance, Quantity: 10, CostBasis: Money{2500000, INR}},
			infy.Key():     {Instrument: infy, Quantity: -5, CostBasis: Money{750000, INR}}, // short
		},
	}

	// cash 100000 + 10*300000 (long) + (-5)*160000 (short) = 100000 + 3000000 - 800000
	prices := map[string]Money{
		reliance.Key(): {300000, INR},
		infy.Key():     {160000, INR},
	}
	got, err := p.Equity(prices)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if want := int64(2300000); got.Amount != want {
		t.Errorf("Equity = %d, want %d", got.Amount, want)
	}
}

func TestPortfolioEquityMissingPrice(t *testing.T) {
	reliance := Instrument{Ticker: "RELIANCE", Exchange: "NSE", Type: Equity}
	p := Portfolio{
		Cash: Money{100000, INR},
		Positions: map[string]Position{
			reliance.Key(): {Instrument: reliance, Quantity: 10, CostBasis: Money{2500000, INR}},
		},
	}
	if _, err := p.Equity(map[string]Money{}); !errors.Is(err, ErrPriceUnavailable) {
		t.Errorf("err = %v, want ErrPriceUnavailable", err)
	}
}

func TestPortfolioEquityCurrencyMismatch(t *testing.T) {
	reliance := Instrument{Ticker: "RELIANCE", Exchange: "NSE", Type: Equity}
	p := Portfolio{
		Cash: Money{100000, INR},
		Positions: map[string]Position{
			reliance.Key(): {Instrument: reliance, Quantity: 10, CostBasis: Money{2500000, INR}},
		},
	}
	prices := map[string]Money{reliance.Key(): {300000, "USD"}}
	if _, err := p.Equity(prices); !errors.Is(err, ErrCurrencyMismatch) {
		t.Errorf("err = %v, want ErrCurrencyMismatch", err)
	}
}
