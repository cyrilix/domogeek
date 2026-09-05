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

Pushing a Git tag matching `v[0-9]*` (e.g. `v0.4.0`) to the Gitea remote triggers the
homelab Tekton pipeline `gitea-docker-build-multiarch`, which builds and pushes a
multi-arch (amd64 + arm64) image plus a signed SBOM:

```
git.cyrilix.bzh/cyrilix/domogeek:<tag>
```

```sh
git tag v0.4.0
git push --tags
```

Watch the run: <https://tekton.banquise.cyrilix.bzh>

