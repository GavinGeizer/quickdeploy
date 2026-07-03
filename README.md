# quickdeploy

`quickdeploy` is a CLI for finding and launching cheap GPU hosting for AI models.

The long-term goal is a command like:

```bash
quickdeploy --model "deepseek/V4-Flash" --fakelocal
```

That should resolve the model's VRAM requirements, find the cheapest reliable host, deploy a minimal Docker image, start the remote service, and point a local port at the endpoint.

## MVP Status

Current focus:

- `--fetch "{gpu name}"`: find the top 5 cheapest matching GPU offers.
- `--deploy`: basic deployment flow, still to be built.

Immediately after the MVP, running plain `quickdeploy` should open a TUI for browsing models, comparing offers, and starting deploy flows.

## Usage

Fetch the cheapest offers for a GPU:

```bash
go run ./cmd/quickdeploy --fetch "RTX 5090"
```

GPU names with spaces should be quoted.

## Project Layout

- `cmd/quickdeploy/`: CLI entrypoint and flag handling.
- `internal/offers/`: offer decoding, sorting, and display formatting.
- `internal/vast/`: Vast API client.
- `AGENTS.md`: project direction and guidance for future agents.

## Testing

Use a writable Go build cache in sandboxed environments:

```bash
GOCACHE=/tmp/quickdeploy-go-cache go test -count=1 ./...
```
