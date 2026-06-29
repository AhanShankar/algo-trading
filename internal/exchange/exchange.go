package exchange

import (
	"slices"
	"time"
)

type Exchange struct {
	Venue      string
	Timezone   *time.Location
	Sessions   []Session
	Holidays   []Holiday
	WeeklyOffs []time.Weekday
}
type Session struct {
	Name        string
	StartTime   TimeOfDay
	EndTime     TimeOfDay
	IsTradeable bool
}
type TimeOfDay struct {
	Hours  int
	Minute int
}
type Holiday struct {
	Date string
	Name string
}

const dateLayout = "2006-01-02"

func (e Exchange) IsTradeable(t time.Time) bool {
	t = t.In(e.Timezone)

	if e.isWeeklyOff(t) || e.isHoliday(t) {
		return false
	}

	tod := TimeOfDay{Hours: t.Hour(), Minute: t.Minute()}
	for _, s := range e.Sessions {
		if s.IsTradeable && s.contains(tod) {
			return true
		}
	}
	return false
}

func (e Exchange) isWeeklyOff(t time.Time) bool {
	return slices.Contains(e.WeeklyOffs, t.Weekday())
}

func (e Exchange) isHoliday(t time.Time) bool {
	date := t.Format(dateLayout)
	for _, h := range e.Holidays {
		if h.Date == date {
			return true
		}
	}
	return false
}

func (s Session) contains(tod TimeOfDay) bool {
	return !tod.before(s.StartTime) && tod.before(s.EndTime)
}

func (t TimeOfDay) before(other TimeOfDay) bool {
	if t.Hours != other.Hours {
		return t.Hours < other.Hours
	}
	return t.Minute < other.Minute
}
