package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"go.uber.org/zap"
)

type mcpDateParams struct {
	Date string `json:"date" jsonschema:"Date to check, in YYYY-MM-DD format"`
}

type mcpHolidaysParams struct {
	Year int `json:"year,omitempty" jsonschema:"Year to list French public holidays for (defaults to the current year)"`
}

type mcpHolidaysResult struct {
	Year     int         `json:"year"`
	Holidays []time.Time `json:"holidays"`
}

func calendarDayFor(day time.Time) CalendarDay {
	calDavHolidays, err := cal.IsHolidaysFromCaldav(day)
	if err != nil {
		zap.S().Warnf("unable to read holiday status from caldav: %v", err)
		calDavHolidays = false
	}
	return CalendarDay{
		Day:        day,
		WorkingDay: cal.IsWorkingDay(day),
		Ferie:      cal.IsHoliday(day),
		Holiday:    calDavHolidays,
		Weekday:    cal.IsWeekDay(day),
	}
}

func handleCurrentDateTime(_ context.Context, _ *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, CurrentDateTime, error) {
	return nil, CurrentDateTime{DateTime: time.Now().In(location)}, nil
}

func handleCalendarToday(_ context.Context, _ *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, CalendarDay, error) {
	return nil, calendarDayFor(time.Now().In(location)), nil
}

func handleCalendarForDate(_ context.Context, _ *mcp.CallToolRequest, params mcpDateParams) (*mcp.CallToolResult, CalendarDay, error) {
	day, err := time.ParseInLocation("2006-01-02", params.Date, location)
	if err != nil {
		return nil, CalendarDay{}, fmt.Errorf("invalid date %q, expected format YYYY-MM-DD: %w", params.Date, err)
	}
	return nil, calendarDayFor(day), nil
}

func handleGetHolidays(_ context.Context, _ *mcp.CallToolRequest, params mcpHolidaysParams) (*mcp.CallToolResult, mcpHolidaysResult, error) {
	year := params.Year
	if year == 0 {
		year = time.Now().In(location).Year()
	}
	return nil, mcpHolidaysResult{Year: year, Holidays: *cal.GetHolidays(year)}, nil
}

func newMCPHandler() http.Handler {
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "domogeek",
		Version: "1.0.0",
	}, nil)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_current_datetime",
		Description: "Get the current date and time",
	}, handleCurrentDateTime)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_calendar_today",
		Description: "Get calendar information for today: working day, French public holiday (ferie), personal holiday (from CalDAV) and weekday",
	}, handleCalendarToday)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_calendar_for_date",
		Description: "Get calendar information for a given date: working day, French public holiday (ferie), personal holiday (from CalDAV) and weekday",
	}, handleCalendarForDate)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_holidays",
		Description: "List French public holidays for a given year (defaults to the current year)",
	}, handleGetHolidays)

	return mcp.NewStreamableHTTPHandler(func(*http.Request) *mcp.Server {
		return server
	}, nil)
}
