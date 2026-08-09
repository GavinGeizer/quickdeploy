# quickdeploy

`quickdeploy` is a CLI for finding and launching cheap GPU hosting for AI models.

The long-term goal is a command like:

```bash
quickdeploy --model "deepseek/V4-Flash" --fakelocal
```

That should resolve the model's VRAM requirements, find the cheapest reliable host, deploy a minimal Docker image, start the remote service, and point a local port at the endpoint.

## MVP Status

Implemented:

- `--fetch "{gpu name}"`: find the top 5 cheapest matching GPU offers.
- `--deploy`: launch a Docker image on the cheapest fetched offer or a specified offer ID.

Immediately after the MVP, running plain `quickdeploy` should open a TUI for browsing models, comparing offers, and starting deploy flows.

## Usage

Set your Vast API key in the process environment:

```bash
export VAST_API_KEY="your-key"
```

Fetch the cheapest offers for a GPU:

```bash
go run ./cmd/quickdeploy --fetch "RTX 5090"
```

GPU names with spaces should be quoted. Results include the Vast offer ID needed for explicit deployment.

Deploy an image on the cheapest matching offer:

```bash
go run ./cmd/quickdeploy --fetch "RTX 5090" --deploy --image "ubuntu:22.04"
```

Deploy a specific offer without fetching first:

```bash
go run ./cmd/quickdeploy --deploy --offer-id 12345678 --image "ubuntu:22.04"
```

`--deploy` immediately creates a billable Vast instance. The command prints the instance ID after Vast accepts the request; it does not wait for the container to become ready or stop the instance later. If an error says the deployment outcome is unknown, inspect your Vast instances before retrying to avoid duplicate rentals.

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
