# Quickdeploy MVP Design

## Goal

Finish `--fetch` so it reliably returns the five cheapest Vast offers, then add the smallest billable `--deploy` flow. Model resolution, custom image building, startup configuration, endpoint polling, local proxying, and the TUI remain out of scope.

## CLI behavior

- `quickdeploy --fetch "RTX 5090"` prints the five cheapest matching offers, including each Vast offer ID.
- `quickdeploy --fetch "RTX 5090" --deploy --image ubuntu:22.04` prints those offers and deploys the cheapest one.
- `quickdeploy --deploy --offer-id 123 --image ubuntu:22.04` deploys that offer without fetching.
- When `--fetch` and `--offer-id` are both present, fetch output is still printed and the explicit ID wins.
- `--deploy` requires `--image` and either `--fetch` or a positive `--offer-id`. Deployment is immediate because the flag itself is the explicit confirmation of a paid action.
- Reject `--image` or `--offer-id` unless `--deploy` is present, reject positional arguments, and treat a whitespace-only image as missing.
- A successful request prints `deployed instance <instance-id> from offer <offer-id>`.

## Implementation

- Add ascending `dph_total` ordering and a server limit of five to the existing Vast offer request; keep the local sort as a defensive guarantee.
- Extend the existing injected provider boundary with `Deploy(offerID, image)`, implemented by `internal/vast.Client` using `PUT /api/v0/asks/{offer_id}/` with an `image` JSON field.
- Read `VAST_API_KEY` in `main`, inject it into the Vast client, and send `Authorization: Bearer <key>`. Return a clear error before network access when the key is absent.
- Decode Vast's `new_contract` response as the instance ID. A created contract is MVP success; do not poll for readiness.
- Add `.env` to `.gitignore`. The executable reads process environment only; local development may source `.env` without adding a dotenv dependency.

## Errors and safety

- Reject the invalid flag combinations above and non-positive explicit offer IDs before any billable request.
- If an automatic deployment finds no offers, print the existing no-offers message and make no deployment request.
- Propagate network, non-2xx Vast responses, malformed responses, and an absent instance ID as errors.
- Warn that the outcome is unknown after a deployment request may have reached Vast, so users inspect existing instances before retrying.
- Never print or commit the API key.
- For the live smoke test, rent the cheapest suitable offer below $1/hour with `ubuntu:22.04`, verify an instance ID is returned, and immediately destroy that test instance to stop billing.

## Tests

- CLI tests cover fetch-only, automatic cheapest selection, explicit offer override, missing image/offer selection, no offers, and deploy errors.
- Vast client tests verify offer ordering/limit, Bearer auth, deployment method/path/body, response decoding, and API failures with a fake `http.RoundTripper`.
- Run `go test -count=1 ./...` and `go vet ./...` with writable caches, then perform the bounded live smoke test.
