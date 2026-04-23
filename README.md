# mcp-holded

MCP server exposing the Holded Invoicing API as Model Context Protocol tools.

## Tool Surface

This server exposes 72 tools using lowercase dot-separated names:

- `holded.contacts.*`
- `holded.documents.*`
- `holded.products.*`
- `holded.treasuries.*`
- `holded.expense_accounts.*`
- `holded.numbering_series.*`
- `holded.sales_channels.*`
- `holded.warehouses.*`
- `holded.payments.*`
- `holded.payment_methods.list`
- `holded.taxes.list`
- `holded.contact_groups.*`
- `holded.remittances.*`
- `holded.services.*`

Read tools are enabled by default. Write tools are implemented but fail closed
unless `HOLDED_ALLOW_WRITE=true`.

## Transports

- `stdio` (default) — for Claude Desktop, Codex CLI, local subprocesses
- `sse` — HTTP Server-Sent Events
- `streamable-http` — modern HTTP streaming (recommended for remote)

## Configuration

All configuration is via environment variables. HTTP transports additionally accept per-request headers that override the env defaults.

| Env var | Default | Purpose |
|---|---|---|
| `HOLDED_API_KEY` | — | Upstream credential (required) |
| `HOLDED_API_BASE` | `https://api.holded.com/api/invoicing/v1` | Upstream base URL |
| `HOLDED_ALLOW_WRITE` | `false` | Enable write tools |
| `HOLDED_ALLOWED_TOOLS` | (empty = all) | Comma-separated allowlist |
| `HOLDED_TIMEOUT_MS` | `30000` | HTTP client timeout (ms) |
| `HOLDED_DEBUG` | `false` | Verbose logging + args in spans |
| `HOLDED_RATE_LIMIT_DISABLE` | `false` | Skip the rate limiter |

HTTP headers (SSE / streamable-http):

| Header | Overrides |
|---|---|
| `X-HOLDED-URL` | `HOLDED_API_BASE` |
| `X-HOLDED-API-Key` | `HOLDED_API_KEY` |

Example allowlist:

```bash
HOLDED_ALLOWED_TOOLS=holded.contacts.list,holded.documents.get,holded.taxes.list
```

## Running

### Docker (recommended)

```bash
docker run -i --rm -e HOLDED_API_KEY=xxx luisra51/mcp-holded:latest -t stdio
```

### Local

```bash
HOLDED_API_KEY=xxx go run ./cmd/mcp-holded -t stdio
```

## Development

```bash
cp mcp-holded-dev.env.example mcp-holded-dev.env
task dev:up                                 # start dev container
task dev:exec CMD='go build ./...'          # build
task dev:exec CMD='go test ./...'           # test
task dev:exec CMD='go run ./cmd/mcp-holded -t stdio'
task dev:down                               # stop
```

## Release

Tag a semver commit to trigger the Docker Hub build:

```bash
git tag v0.1.0
git push origin v0.1.0
```

## License

MIT.
