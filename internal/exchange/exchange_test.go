package exchange

import (
	"testing"
	"time"
)

// nseLike builds an exchange resembling NSE: IST timezone, a single regular
// session 09:15–15:30, weekends off, and one holiday.
func nseLike(t *testing.T) Exchange {
	t.Helper()
	ist, err := time.LoadLocation("Asia/Kolkata")
	if err != nil {
		t.Fatalf("loading IST: %v", err)
	}
	return Exchange{
		Venue:    "NSE",
		Timezone: ist,
		Sessions: []Session{
			{Name: "regular", StartTime: TimeOfDay{9, 15}, EndTime: TimeOfDay{15, 30}, IsTradeable: true},
		},
		Holidays:   []Holiday{{Date: "2026-01-26", Name: "Republic Day"}},
		WeeklyOffs: []time.Weekday{time.Saturday, time.Sunday},
	}
}

func TestExchangeIsTradeable(t *testing.T) {
	e := nseLike(t)
	ist := e.Timezone

	tests := []struct {
		name string
		t    time.Time
		want bool
	}{
		// 2026-06-29 is a Monday.
		{"during session", time.Date(2026, 6, 29, 10, 0, 0, 0, ist), true},
		{"at open is tradeable", time.Date(2026, 6, 29, 9, 15, 0, 0, ist), true},
		{"at close is not (half-open)", time.Date(2026, 6, 29, 15, 30, 0, 0, ist), false},
		{"one minute before close", time.Date(2026, 6, 29, 15, 29, 0, 0, ist), true},
		{"before open", time.Date(2026, 6, 29, 9, 14, 0, 0, ist), false},
		{"after close", time.Date(2026, 6, 29, 16, 0, 0, 0, ist), false},
		// 2026-07-04 is a Saturday, 2026-07-05 a Sunday.
		{"saturday", time.Date(2026, 7, 4, 10, 0, 0, 0, ist), false},
		{"sunday", time.Date(2026, 7, 5, 10, 0, 0, 0, ist), false},
		// 2026-01-26 (Republic Day) is a Monday, so this exercises isHoliday, not the weekend rule.
		{"holiday on a weekday", time.Date(2026, 1, 26, 10, 0, 0, 0, ist), false},
		// Timezone handling: 04:30 UTC == 10:00 IST, a tradeable moment.
		{"utc input converted to ist", time.Date(2026, 6, 29, 4, 30, 0, 0, time.UTC), true},
		// 10:00 UTC == 15:30 IST, which is at/after close.
		{"utc input at ist close", time.Date(2026, 6, 29, 10, 0, 0, 0, time.UTC), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := e.IsTradeable(tt.t); got != tt.want {
				t.Errorf("IsTradeable(%s) = %v, want %v", tt.t, got, tt.want)
			}
		})
	}
}

func TestExchangeMultipleSessions(t *testing.T) {
	ist, err := time.LoadLocation("Asia/Kolkata")
	if err != nil {
		t.Fatalf("loading IST: %v", err)
	}
	e := Exchange{
		Venue:    "TEST",
		Timezone: ist,
		Sessions: []Session{
			{Name: "morning", StartTime: TimeOfDay{9, 0}, EndTime: TimeOfDay{11, 0}, IsTradeable: true},
			{Name: "lunch", StartTime: TimeOfDay{11, 0}, EndTime: TimeOfDay{13, 0}, IsTradeable: false},
			{Name: "afternoon", StartTime: TimeOfDay{13, 0}, EndTime: TimeOfDay{15, 0}, IsTradeable: true},
		},
	}

	tests := []struct {
		name string
		hour int
		want bool
	}{
		{"morning session", 10, true},
		{"lunch break not tradeable", 12, false},
		{"afternoon session", 14, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			when := time.Date(2026, 6, 29, tt.hour, 0, 0, 0, ist)
			if got := e.IsTradeable(when); got != tt.want {
				t.Errorf("IsTradeable(%dh) = %v, want %v", tt.hour, got, tt.want)
			}
		})
	}
}

func TestSessionContains(t *testing.T) {
	s := Session{StartTime: TimeOfDay{9, 15}, EndTime: TimeOfDay{15, 30}}
	tests := []struct {
		name string
		tod  TimeOfDay
		want bool
	}{
		{"at start (inclusive)", TimeOfDay{9, 15}, true},
		{"just after start", TimeOfDay{9, 16}, true},
		{"middle", TimeOfDay{12, 0}, true},
		{"just before end", TimeOfDay{15, 29}, true},
		{"at end (exclusive)", TimeOfDay{15, 30}, false},
		{"before start", TimeOfDay{9, 14}, false},
		{"after end", TimeOfDay{16, 0}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := s.contains(tt.tod); got != tt.want {
				t.Errorf("contains(%v) = %v, want %v", tt.tod, got, tt.want)
			}
		})
	}
}

func TestTimeOfDayBefore(t *testing.T) {
	tests := []struct {
		name string
		a, b TimeOfDay
		want bool
	}{
		{"earlier hour", TimeOfDay{9, 30}, TimeOfDay{10, 0}, true},
		{"later hour", TimeOfDay{11, 0}, TimeOfDay{10, 0}, false},
		{"same hour earlier minute", TimeOfDay{10, 15}, TimeOfDay{10, 30}, true},
		{"same hour later minute", TimeOfDay{10, 45}, TimeOfDay{10, 30}, false},
		{"equal is not before", TimeOfDay{10, 0}, TimeOfDay{10, 0}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.a.before(tt.b); got != tt.want {
				t.Errorf("%v.before(%v) = %v, want %v", tt.a, tt.b, got, tt.want)
			}
		})
	}
}
