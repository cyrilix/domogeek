# domogeek

Tools for domotic

## Calendar

Return calendar informations about today

## Date and time

`GET /datetime` returns the current date and time.

## MCP

An MCP server is exposed over Streamable HTTP at `/mcp`, providing the same calendar
and date/time information as MCP tools (`get_current_datetime`, `get_calendar_today`,
`get_calendar_for_date`, `get_holidays`) for use by LLM agents/MCP clients.

