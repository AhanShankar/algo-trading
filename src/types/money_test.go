package types

import (
	"errors"
	"testing"
)

func TestMoneyAdd(t *testing.T) {
	tests := []struct {
		name    string
		a, b    Money
		want    int64
		wantErr error
	}{
		{"same currency", Money{100, INR}, Money{250, INR}, 350, nil},
		{"with negative", Money{500, INR}, Money{-200, INR}, 300, nil},
		{"currency mismatch", Money{100, INR}, Money{100, "USD"}, 0, ErrCurrencyMismatch},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.a.Add(tt.b)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("err = %v, want %v", err, tt.wantErr)
			}
			if tt.wantErr == nil && got.Amount != tt.want {
				t.Errorf("Amount = %d, want %d", got.Amount, tt.want)
			}
		})
	}
}

func TestMoneySub(t *testing.T) {
	tests := []struct {
		name    string
		a, b    Money
		want    int64
		wantErr error
	}{
		{"same currency", Money{500, INR}, Money{200, INR}, 300, nil},
		{"goes negative", Money{100, INR}, Money{250, INR}, -150, nil},
		{"currency mismatch", Money{100, INR}, Money{100, "USD"}, 0, ErrCurrencyMismatch},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.a.Sub(tt.b)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("err = %v, want %v", err, tt.wantErr)
			}
			if tt.wantErr == nil && got.Amount != tt.want {
				t.Errorf("Amount = %d, want %d", got.Amount, tt.want)
			}
		})
	}
}

func TestMoneyMul(t *testing.T) {
	tests := []struct {
		name   string
		m      Money
		factor int64
		want   int64
	}{
		{"long quantity", Money{10000, INR}, 5, 50000},
		{"short quantity is negative", Money{10000, INR}, -3, -30000},
		{"zero quantity", Money{10000, INR}, 0, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.m.Mul(tt.factor); got.Amount != tt.want {
				t.Errorf("Amount = %d, want %d", got.Amount, tt.want)
			}
		})
	}
}

func TestMoneyCmp(t *testing.T) {
	tests := []struct {
		name    string
		a, b    Money
		want    int
		wantErr error
	}{
		{"less", Money{100, INR}, Money{200, INR}, -1, nil},
		{"equal", Money{200, INR}, Money{200, INR}, 0, nil},
		{"greater", Money{300, INR}, Money{200, INR}, 1, nil},
		{"currency mismatch", Money{100, INR}, Money{100, "USD"}, 0, ErrCurrencyMismatch},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.a.Cmp(tt.b)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("err = %v, want %v", err, tt.wantErr)
			}
			if tt.wantErr == nil && got != tt.want {
				t.Errorf("Cmp = %d, want %d", got, tt.want)
			}
		})
	}
}
