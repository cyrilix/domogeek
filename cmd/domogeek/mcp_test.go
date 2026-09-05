package main

import (
	"context"
	"domogeek/pkg/calendar"
	"testing"
	"time"
)

func withTestCalendar(t *testing.T) {
	t.Helper()
	previous := cal
	cal = calendar.New(location)
	t.Cleanup(func() { cal = previous })
}

func TestCalendarDayFor(t *testing.T) {
	withTestCalendar(t)

	tests := []struct {
		name           string
		day            time.Time
		wantWorkingDay bool
		wantFerie      bool
		wantWeekday    bool
	}{
		{
			name:           "public holiday",
			day:            time.Date(2020, time.January, 1, 0, 0, 0, 0, location),
			wantWorkingDay: false,
			wantFerie:      true,
			wantWeekday:    true,
		},
		{
			name:           "regular working day",
			day:            time.Date(2020, time.January, 2, 0, 0, 0, 0, location),
			wantWorkingDay: true,
			wantFerie:      false,
			wantWeekday:    true,
		},
		{
			name:           "week-end",
			day:            time.Date(2020, time.January, 4, 0, 0, 0, 0, location),
			wantWorkingDay: false,
			wantFerie:      false,
			wantWeekday:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cd := calendarDayFor(tt.day)
			if cd.WorkingDay != tt.wantWorkingDay {
				t.Errorf("WorkingDay = %v, want %v", cd.WorkingDay, tt.wantWorkingDay)
			}
			if cd.Ferie != tt.wantFerie {
				t.Errorf("Ferie = %v, want %v", cd.Ferie, tt.wantFerie)
			}
			if cd.Weekday != tt.wantWeekday {
				t.Errorf("Weekday = %v, want %v", cd.Weekday, tt.wantWeekday)
			}
			if !cd.Day.Equal(tt.day) {
				t.Errorf("Day = %v, want %v", cd.Day, tt.day)
			}
		})
	}
}

func TestHandleCalendarToday(t *testing.T) {
	withTestCalendar(t)

	_, cd, err := handleCalendarToday(context.Background(), nil, struct{}{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	now := time.Now().In(location)
	wantDay := time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute(), 0, 0, location)
	gotDay := time.Date(cd.Day.Year(), cd.Day.Month(), cd.Day.Day(), cd.Day.Hour(), cd.Day.Minute(), 0, 0, location)
	if !gotDay.Equal(wantDay) {
		t.Errorf("Day = %v, want close to %v", cd.Day, now)
	}
}

func TestHandleCalendarForDate(t *testing.T) {
	withTestCalendar(t)

	t.Run("valid date", func(t *testing.T) {
		_, cd, err := handleCalendarForDate(context.Background(), nil, mcpDateParams{Date: "2020-01-01"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !cd.Ferie {
			t.Errorf("expected 2020-01-01 to be a ferie day")
		}
	})

	t.Run("invalid date", func(t *testing.T) {
		_, _, err := handleCalendarForDate(context.Background(), nil, mcpDateParams{Date: "not-a-date"})
		if err == nil {
			t.Fatal("expected an error for an invalid date, got nil")
		}
	})
}

func TestHandleGetHolidays(t *testing.T) {
	withTestCalendar(t)

	t.Run("explicit year", func(t *testing.T) {
		_, result, err := handleGetHolidays(context.Background(), nil, mcpHolidaysParams{Year: 2020})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.Year != 2020 {
			t.Errorf("Year = %d, want 2020", result.Year)
		}
		if len(result.Holidays) != 10 {
			t.Errorf("got %d holidays, want 10", len(result.Holidays))
		}
	})

	t.Run("defaults to current year", func(t *testing.T) {
		_, result, err := handleGetHolidays(context.Background(), nil, mcpHolidaysParams{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.Year != time.Now().In(location).Year() {
			t.Errorf("Year = %d, want current year %d", result.Year, time.Now().In(location).Year())
		}
	})
}
