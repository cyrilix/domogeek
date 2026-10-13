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

## Continuous builds

Pushing a Git tag matching `v[0-9]*` (e.g. `v0.4.1`) to the Gitea remote triggers the Gitea Actions
workflow `.gitea/workflows/build.yaml` (no Docker, no privileges), which builds and pushes a
multi-arch (amd64 + arm64) image plus a CycloneDX SBOM to DependencyTrack:

```
git.cyrilix.bzh/cyrilix/domogeek:<tag>
```

```sh
git tag v0.4.1
git push --tags
```

Watch the run: Gitea → `cyrilix/domogeek` → Actions.
