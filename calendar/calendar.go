package calendar

import (
	"math"
	"time"
)

type Calendar struct {
	Location *time.Location
}

func (cal *Calendar) GetEasterDay(year int) time.Time {
	g := float64(year % 19.0)
	c := math.Floor(float64(year) / 100.0)
	c4 := math.Floor(c / 4.0)
	h := float64(int(19.0*g+c-c4-math.Floor((8.0*c+13)/25)+15) % 30.0)
	k := math.Floor(h / 28.0)
	i := (k*math.Floor(29./(h+1.))*math.Floor((21.-g)/11.)-1.)*k + h

	// jour de Pâques (0=dimanche, 1=lundi....)
	dayWeek := int(math.Floor(float64(year)/4.)+float64(year)+i+2+c4-c) % 7

	// Jour de Pâques en jours enpartant de 1 = 1er mars
	presJour := int(28 + int(i) - dayWeek)

	// mois (0 = janvier, ... 2 = mars, 3 = avril)
	month := 2
	if presJour > 31 {
		month = 3
	}

	// Mois dans l'année
	month += 1

	// jour du mois
	day := presJour - 31
	if month == 2 {
		day = presJour
	}

	return time.Date(year, 3, 31, 0, 0, 0, 0, cal.Location).AddDate(0, 0, day)
}

func (cal *Calendar) GetHolidays(year int) *[]time.Time {

	// Calcul du jour de pâques
	paques := cal.GetEasterDay(year)

	joursFeries := []time.Time{
		// Jour de l'an
		time.Date(year, time.January, 1, 0, 0, 0, 0, cal.Location),
		// Easter
		paques,
		// 1 mai
		time.Date(year, time.May, 1, 0, 0, 0, 0, cal.Location),
		// 8 mai
		time.Date(year, time.May, 8, 0, 0, 0, 0, cal.Location),
		// Ascension
		paques.AddDate(0, 0, 39),
		// 14 juillet
		time.Date(year, time.July, 14, 0, 0, 0, 0, cal.Location),
		// 15 aout
		time.Date(year, time.August, 15, 0, 0, 0, 0, cal.Location),
		// Toussaint
		time.Date(year, time.November, 1, 0, 0, 0, 0, cal.Location),
		// 11 novembre
		time.Date(year, time.November, 11, 0, 0, 0, 0, cal.Location),
		// noël
		time.Date(year, time.December, 25, 0, 0, 0, 0, cal.Location),
	}

	return &joursFeries
}

func (cal *Calendar) GetHolidaysSet(year int) *map[time.Time]bool {
	holidays := cal.GetHolidays(year)
	result := make(map[time.Time]bool, len(*holidays))
	for _, h := range *holidays {
		result[h] = true
	}
	return &result
}

func(cal *Calendar) IsHoliday(date time.Time) bool{
	h := cal.GetHolidaysSet(date.Year())
	d := date.In(cal.Location)
	day := time.Date(d.Year(), d.Month(), d.Day(), 0, 0, 0, 0, cal.Location)
	return (*h)[day]
}

func (cal *Calendar) IsWorkingDay(date time.Time) bool {
	return !cal.IsHoliday(date) && date.Weekday() >= time.Monday && date.Weekday() <= time.Friday
}

func (cal *Calendar) IsWorkingDayToday() bool {
	return cal.IsWorkingDay(time.Now())
}

func (cal *Calendar) IsWeekDay(day time.Time) bool{
	return day.Weekday() >= time.Monday && day.Weekday() <= time.Friday
}
