package trading

type Position struct {
	Instrument  Instrument
	Quantity    int64 // signed: >0 long, <0 short
	CostBasis   Money // total cash to open the current quantity
	RealizedPnL Money
}

func (p Position) AvgPrice() Money {
	if p.Quantity == 0 {
		return Money{Currency: p.CostBasis.Currency}
	}
	qty := p.Quantity
	if qty < 0 {
		qty = -qty
	}
	return Money{Amount: p.CostBasis.Amount / qty, Currency: p.CostBasis.Currency}
}
