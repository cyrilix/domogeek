# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this is

domogeek is a small Go HTTP service that exposes calendar/holiday information for home automation (domotic) use. It answers whether "today" is a working day, a French public holiday (`ferie`), and/or a personal holiday read from a CalDAV calendar. The same information is also exposed as MCP tools over a Streamable HTTP endpoint (`/mcp`), so LLM agents/MCP clients can query it directly.

## Commands

Build and test use Go's standard tooling with vendored dependencies (`-mod=vendor`). The module requires **Go 1.25+** (bumped from 1.18 to use the official MCP SDK). If the `go` binary on `PATH` resolves to an older toolchain (e.g. this box also has `golang-1.18` installed and it can shadow the newer one), invoke a newer toolchain explicitly, overriding `GOROOT` if it's pinned to the old one (e.g. by an IDE's `_INTELLIJ_FORCE_SET_GOROOT`):

```bash
GOROOT=/usr/lib/go-1.27 /usr/lib/go-1.27/bin/go build -mod=vendor ./...
```

Otherwise, plain invocations work the same as before:

```bash
go build -mod=vendor ./...                 # build
go vet -mod=vendor ./...                   # vet
go test -mod=vendor ./...                  # run all tests
go test -mod=vendor ./pkg/calendar/ -run TestCalendar_IsWorkingDay -v   # run a single test
```

Run locally (caldav flags are optional — omitting them disables CalDAV-based holiday checks):

```bash
go run -mod=vendor ./cmd/domogeek \
  -port 8080 \
  -caldav-url https://example.com \
  -caldav-path /calendars/user/holidays/ \
  -caldav-username user \
  -caldav-password pass \
  -caldav-summary-pattern Holidays
```

Running with no flags at all prints usage and exits (see the `len(os.Args) <= 1` check in `main`). Note: despite the "caldav flags are optional" framing, `main()` currently calls `calendar.NewCaldav` unconditionally, even with an empty `-caldav-url` — with no caldav server to validate against, this hits the 1000-retry exponential backoff described below and the process never gets to `http.ListenAndServe`. This is a pre-existing quirk, not something introduced by the MCP work; if you need a quick local instance (e.g. to exercise `/mcp` or `/calendar` without a real CalDAV server), it's easiest to call `newMCPHandler()`/`calendar.New(location)` directly from a small throwaway `main`/test rather than running the built binary.

Docker image (multi-stage, distroless):

```bash
docker build -t domogeek .
```

`build-docker.sh` is the multi-arch (amd64/arm64/arm/v7) release build using `buildah`, producing a manifest pushed to `docker.io/cyrilix/domogeek:<git describe>`. It is a release/CI script, not needed for local development.

## Architecture

- `cmd/domogeek/domogeek.go` — the binary entrypoint. Wires up flags, zap logging, the `calendar.Calendar`, Prometheus metrics/instrumented handlers, and a `health-go` health check, then serves four HTTP routes:
  - `GET /calendar` — returns a `CalendarDay` JSON payload (`working_day`, `ferie`, `holiday`, `weekday`) for "now" in `Europe/Paris`.
  - `POST /mcp` — MCP Streamable HTTP endpoint (see `cmd/domogeek/mcp.go` below).
  - `GET /metrics` — Prometheus metrics (request counter/summary/histogram wrapping the calendar handler, defined in `init()`).
  - `GET /status` — health-go handler with two checks: a static `calendar` check and a `caldav` check that calls `cal.IsHolidaysFromCaldav`.
  - The process blocks on `SIGTERM` (not `SIGINT`) via a signal channel; `http.ListenAndServe` runs in a goroutine and `zap.S().Fatal`s the process on error.

- `cmd/domogeek/mcp.go` — MCP server built with `github.com/modelcontextprotocol/go-sdk/mcp`, mounted at `/mcp` via `mcp.NewStreamableHTTPHandler`. Exposes three tools, all reading the same package-level `cal`/`location` globals as the `/calendar` handler:
  - `get_calendar_today` — same payload as `GET /calendar`, for now.
  - `get_calendar_for_date` — same payload for an arbitrary `date` (`YYYY-MM-DD`); an invalid date is returned as a tool error (`CallToolResult.IsError`), not a protocol-level error, so an MCP client/LLM can see and self-correct.
  - `get_holidays` — lists French public holidays (`Calendar.GetHolidays`) for a given `year`, defaulting to the current year if omitted.
  - `calendarDayFor` in this file is the single place that builds a `CalendarDay` from a `time.Time`; both the HTTP `/calendar` handler and the MCP tools call it, so there's one code path for the working-day/ferie/holiday/weekday logic.
  - `cmd/domogeek/mcp_test.go` unit-tests the handler functions directly; `cmd/domogeek/mcp_e2e_test.go` drives the real Streamable HTTP transport with an MCP client (`httptest.Server` + `mcp.NewClient(...).Connect`) to check the wiring end-to-end (tool listing, structured content, and the tool-error-vs-protocol-error distinction).

- `pkg/calendar/calendar.go` — all domain logic, built around the `Calendar` struct and functional options (`WithCaldav`, `WithCaldavPath`, `WithCaldavSummaryPattern`):
  - `GetEasterDay` implements the Gauss/Meeus algorithm to compute Easter Sunday for a given year (in `cal.Location`); `GetHolidays` derives the fixed-date French public holidays plus two computed from it: Easter Monday (Easter+1) and Ascension (Easter+39).
  - `IsHoliday` combines the static French holiday set with `IsHolidaysFromCaldav`.
  - `IsHolidaysFromCaldav` queries a CalDAV calendar for events on a given day and treats any event whose `Summary` contains `caldavSummaryPattern` (default `"Holidays"`) as a personal holiday. If no `Caldav` client is configured, it returns `false, nil` — CalDAV is optional.
  - The `Caldav` interface (`QueryEvents`) abstracts `github.com/dolanor/caldav-go`, letting tests substitute `MockCaldav` (see `calendar_test.go`) instead of hitting a real server.
  - `NewCaldav` validates the CalDAV connection at startup with `avast/retry-go`, retrying up to 1000 times with exponential backoff (max delay 24h) — this means a bad/unreachable CalDAV server does **not** fail fast; `main()` will effectively hang retrying rather than exiting quickly if the caldav server is down (it's only `Fatal`'d if `NewCaldav` itself returns a non-nil error, which currently only happens after all retries are exhausted).

- `pkg/metrics/metrics.go` — currently a single unused/unconfigured Prometheus counter (empty `Namespace`/`Subsystem`/`Name`/`Help`). Not wired into `cmd/domogeek`; treat as in-progress scaffolding rather than an established pattern to follow.

## Dependencies

Dependencies are vendored under `vendor/` and checked into the repo — always pass `-mod=vendor` (or rely on `GOFLAGS=-mod=vendor` if set) for build/test/vet so Go doesn't try to hit the network or the module cache. After changing `go.mod`/`go.sum`, run `go mod vendor` to keep `vendor/` in sync (use the newer toolchain per the note above, and `-mod=mod` instead of `-mod=vendor` for the `go get`/`go mod tidy` step itself).

`github.com/modelcontextprotocol/go-sdk` (the official Go MCP SDK) is the dependency added for the `/mcp` endpoint; it pulled in `google/jsonschema-go`, `segmentio/encoding`, `yosida95/uritemplate`, and a few `golang.org/x/*` packages as transitive deps, and required bumping the `go` directive in `go.mod` from `1.18` to `1.25.0`.
