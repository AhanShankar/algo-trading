package trading

import (
	"strings"
	"testing"
)

func TestGenerateClientID_NotEmpty(t *testing.T) {
	o := Order{
		Instrument:  Instrument{Ticker: "RELIANCE", Exchange: "NSE", Type: Equity},
		Side:        Buy,
		Quantity:    10,
		Price:       Money{Amount: 250000, Currency: INR},
		Type:        Market,
		ProductType: Intraday,
		Validity:    Day,
		State:       Pending,
	}

	id := o.GenerateClientID()
	if id == "" {
		t.Error("expected non-empty ID, got empty string")
	}
}

func TestGenerateClientID_ContainsTicker(t *testing.T) {
	o := Order{
		Instrument:  Instrument{Ticker: "INFY", Exchange: "NSE", Type: Equity},
		Side:        Buy,
		Quantity:    5,
		Price:       Money{Amount: 150000, Currency: INR},
		Type:        Limit,
		ProductType: Delivery,
		Validity:    Day,
		State:       Pending,
	}

	id := o.GenerateClientID()
	if !strings.Contains(id, "INFY") {
		t.Errorf("expected ID to contain ticker INFY, got: %s", id)
	}
}

func TestGenerateClientID_ContainsSide(t *testing.T) {
	tests := []struct {
		side     Side
		expected string
	}{
		{Buy, "BUY"},
		{Sell, "SELL"},
	}

	for _, tt := range tests {
		o := Order{
			Instrument:  Instrument{Ticker: "RELIANCE", Exchange: "NSE", Type: Equity},
			Side:        tt.side,
			Quantity:    10,
			Price:       Money{Amount: 250000, Currency: INR},
			Type:        Market,
			ProductType: Intraday,
			Validity:    Day,
			State:       Pending,
		}

		id := o.GenerateClientID()
		if !strings.Contains(id, tt.expected) {
			t.Errorf("expected ID to contain side %s, got: %s", tt.expected, id)
		}
	}
}

func TestGenerateClientID_Unique(t *testing.T) {
	o := Order{
		Instrument:  Instrument{Ticker: "RELIANCE", Exchange: "NSE", Type: Equity},
		Side:        Buy,
		Quantity:    10,
		Price:       Money{Amount: 250000, Currency: INR},
		Type:        Market,
		ProductType: Intraday,
		Validity:    Day,
		State:       Pending,
	}

	ids := make(map[string]bool)
	for i := 0; i < 100; i++ {
		id := o.GenerateClientID()
		if ids[id] {
			t.Errorf("duplicate ID generated: %s", id)
		}
		ids[id] = true
	}
}
