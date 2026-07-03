# Quickdeploy Agent Guide

## Project Idea

`quickdeploy` is a CLI executable for finding and launching cheap GPU hosting for AI models.

The intended user experience is flag-driven:

```bash
quickdeploy --model "deepseek/V4-Flash" --fakelocal
```

In the fuller product, that command should:

1. Resolve the model's hosting requirements, especially VRAM.
2. Find the cheapest reliable GPU offer that can host it.
3. Deploy the model using a custom minimal Docker image.
4. Start hosting it on the selected provider.
5. Point a local port at the remote endpoint so the deployment feels local.

The CLI should prefer explicit flags over hidden behavior. If a feature choice is ambiguous, ask clarifying questions instead of guessing.

## Current MVP

The immediate MVP is intentionally smaller:

1. Finish `--fetch`, which searches for a quoted GPU name and prints the top 5 cheapest results.
2. Add a basic `--deploy` feature.

Do not jump ahead into the full model resolver, Docker image builder, or local-port proxy unless the user asks for that slice directly.

## Immediately After MVP

After `--fetch` and a basic `--deploy` exist, add a TUI for the basic executable experience.

When the user runs:

```bash
quickdeploy
```

the app should open a terminal UI that helps browse/select models, compare hosting offers, and start deploy flows without requiring the user to remember every flag.

Keep the flag-driven CLI as the scriptable interface. The TUI should sit on top of the same internal packages rather than duplicating provider, offer, or deployment logic.

## Current Layout

- `cmd/quickdeploy/`: CLI entrypoint, flag parsing, process exit behavior.
- `internal/offers/`: offer model, Vast response decoding, cheapest-offer sorting, CLI snippet formatting.
- `internal/vast/`: Vast API client and request construction.
- `offer.json`: sample Vast API response for reference.

Prefer adding small focused packages over growing one large file.

## CLI Direction

Known or planned flags:

- `--fetch "{gpu name}"`: fetch top 5 cheapest matching GPU offers.
- `--model "{provider/model}"`: future model requirement lookup input.
- `--deploy`: MVP deployment path.
- `--fakelocal`: future local-port-to-remote-endpoint experience.

Assume users will quote multi-word GPU names, for example:

```bash
quickdeploy --fetch "RTX 5090"
```

## Implementation Notes

- Leave the existing Vast auth header behavior alone unless the user explicitly asks to change it.
- Keep network-facing code behind interfaces so CLI behavior can be tested without real network calls.
- Keep formatting in `internal/offers` unless output grows into multiple presentation modes.
- Keep provider-specific behavior out of `cmd/quickdeploy`; the command package should orchestrate packages, not contain provider logic.
- Use package-level tests for new behavior.

## Testing

Use a writable Go build cache in this environment:

```bash
GOCACHE=/tmp/quickdeploy-go-cache go test -count=1 ./...
```

Avoid tests that require opening local sockets unless necessary; sandboxed environments may block them. Prefer fake `http.RoundTripper` implementations for HTTP client tests.
