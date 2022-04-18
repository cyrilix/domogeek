package calendar

import (
	"testing"
	"time"
)

func TestCalendar_GetEasterDay(t *testing.T) {
	loc, err := time.LoadLocation("Europe/Paris")
	if err != nil {
		t.Errorf("unable to load time location: %v", err)
		t.Fail()
	}

	easterDays := []time.Time{
		time.Date(2019, time.April, 21, 0, 0, 0, 0, loc),
		time.Date(2020, time.April, 12, 0, 0, 0, 0, loc),
		time.Date(2021, time.April, 4, 0, 0, 0, 0, loc),
	}

	c := New(loc)

	for _, d := range easterDays {
		easter := c.GetEasterDay(d.Year())
		if easter != d {
			t.Errorf("bad date for year %d, expected:%v ; actual:%v", d.Year(), d, easter)
		}
	}
}

func TestCalendar_GetHolidays(t *testing.T) {
	loc, err := time.LoadLocation("Europe/Paris")
	if err != nil {
		t.Errorf("unable to load time location: %v", err)
		t.Fail()
	}

	expectedHolidays := map[time.Time]bool{
		time.Date(2020, time.January, 1, 0, 0, 0, 0, loc):   true,
		time.Date(2020, time.April, 13, 0, 0, 0, 0, loc):    true,
		time.Date(2020, time.May, 1, 0, 0, 0, 0, loc):       true,
		time.Date(2020, time.May, 8, 0, 0, 0, 0, loc):       true,
		time.Date(2020, time.May, 21, 0, 0, 0, 0, loc):      true,
		time.Date(2020, time.July, 14, 0, 0, 0, 0, loc):     true,
		time.Date(2020, time.August, 15, 0, 0, 0, 0, loc):   true,
		time.Date(2020, time.November, 1, 0, 0, 0, 0, loc):  true,
		time.Date(2020, time.November, 11, 0, 0, 0, 0, loc): true,
		time.Date(2020, time.December, 25, 0, 0, 0, 0, loc): true,
	}

	c := Calendar{loc}
	holidays := c.GetHolidays(2020)
	if len(*holidays) != len(expectedHolidays) {
		t.Errorf("bad number of holidays, %d but %d are expected", len(*holidays), len(expectedHolidays))
	}
	for _, h := range *holidays {
		if !expectedHolidays[h] {
			t.Errorf("%v is not a holiday", h)
		}
	}
}

func TestCalendar_GetHolidaysSet(t *testing.T) {
	loc, err := time.LoadLocation("Europe/Paris")
	if err != nil {
		t.Errorf("unable to load time location: %v", err)
		t.Fail()
	}

	expectedHolidays := []time.Time{
		time.Date(2020, time.January, 1, 0, 0, 0, 0, loc),
		time.Date(2020, time.April, 13, 0, 0, 0, 0, loc),
		time.Date(2020, time.May, 1, 0, 0, 0, 0, loc),
		time.Date(2020, time.May, 8, 0, 0, 0, 0, loc),
		time.Date(2020, time.May, 21, 0, 0, 0, 0, loc),
		time.Date(2020, time.July, 14, 0, 0, 0, 0, loc),
		time.Date(2020, time.August, 15, 0, 0, 0, 0, loc),
		time.Date(2020, time.November, 1, 0, 0, 0, 0, loc),
		time.Date(2020, time.November, 11, 0, 0, 0, 0, loc),
		time.Date(2020, time.December, 25, 0, 0, 0, 0, loc),
	}

	c := Calendar{loc}
	holidays := c.GetHolidaysSet(2020)
	if len(*holidays) != len(expectedHolidays) {
		t.Errorf("bad number of holidays, %d but %d are expected", len(*holidays), len(expectedHolidays))
	}
	for _, h := range expectedHolidays {
		if !(*holidays)[h] {
			t.Errorf("%v is not a holiday", h)
		}
	}
}

func TestCalendar_IsHolidays(t *testing.T) {
	loc, err := time.LoadLocation("Europe/Paris")
	if err != nil {
		t.Errorf("unable to load time location: %v", err)
		t.Fail()
	}

	expectedHolidays := []time.Time{
		time.Date(2020, time.January, 1, 0, 0, 0, 0, loc),
		time.Date(2020, time.April, 13, 0, 0, 0, 0, loc),
		time.Date(2020, time.May, 1, 0, 0, 0, 0, loc),
		time.Date(2020, time.May, 8, 0, 0, 0, 0, loc),
		time.Date(2020, time.May, 21, 0, 0, 0, 0, loc),
		time.Date(2020, time.July, 14, 0, 0, 0, 0, loc),
		time.Date(2020, time.August, 15, 0, 0, 0, 0, loc),
		time.Date(2020, time.November, 1, 0, 0, 0, 0, loc),
		time.Date(2020, time.November, 11, 0, 0, 0, 0, loc),
		time.Date(2020, time.December, 25, 0, 0, 0, 0, loc),
	}

	c := Calendar{loc}
	holidays := c.GetHolidaysSet(2020)
	if len(*holidays) != len(expectedHolidays) {
		t.Errorf("bad number of holidays, %d but %d are expected", len(*holidays), len(expectedHolidays))
	}
	for _, h := range expectedHolidays {
		if !c.IsHoliday(h) {
			t.Errorf("%v is a holiday", h)
		}
	}
	if c.IsHoliday(time.Date(2019, time.January, 02, 0, 0, 0, 0, loc)) {
		t.Error("02 january should not be a holiday")
	}
}

func TestCalendar_IsWorkingDay(t *testing.T) {
	loc, err := time.LoadLocation("Europe/Paris")
	if err != nil {
		t.Errorf("unable to load time location: %v", err)
		t.Fail()
	}
	c := Calendar{loc}

	if c.IsWorkingDay(time.Date(2019, time.January, 01, 0, 0, 0, 0, loc)) {
		t.Error("1st january is not a working day")
	}
	if !c.IsWorkingDay(time.Date(2019, time.January, 02, 0, 0, 0, 0, loc)) {
		t.Error("02 january should be a working day")
	}

	if !c.IsWorkingDay(time.Date(2019, time.January, 7, 0, 0, 0, 0, loc)) {
		t.Error("Monday should be a working day")
	}
	if !c.IsWorkingDay(time.Date(2019, time.January, 8, 0, 0, 0, 0, loc)) {
		t.Error("Tuesday should be a working day")
	}
	if !c.IsWorkingDay(time.Date(2019, time.January, 9, 0, 0, 0, 0, loc)) {
		t.Error("Wednesday should be a working day")
	}
	if !c.IsWorkingDay(time.Date(2019, time.January, 10, 0, 0, 0, 0, loc)) {
		t.Error("Thursday should be a working day")
	}
	if !c.IsWorkingDay(time.Date(2019, time.January, 11, 0, 0, 0, 0, loc)) {
		t.Error("Friday should be a working day")
	}
	if c.IsWorkingDay(time.Date(2019, time.January, 12, 0, 0, 0, 0, loc)) {
		t.Error("Saturday should not be a working day")
	}
	if c.IsWorkingDay(time.Date(2019, time.January, 13, 0, 0, 0, 0, loc)) {
		t.Error("Sunday should not be a working day")
	}
}
