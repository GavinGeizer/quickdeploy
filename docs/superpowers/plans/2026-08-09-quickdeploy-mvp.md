# Quickdeploy MVP Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Guarantee the five cheapest Vast offers and add an explicit, testable deployment command that creates one Vast instance.

**Architecture:** Keep `cmd/quickdeploy` responsible for flag validation and orchestration, `internal/offers` responsible for display, and `internal/vast.Client` responsible for authenticated HTTP. Extend the existing injected CLI boundary rather than adding a package or dependency.

**Tech Stack:** Go 1.26.4 standard library, Vast.ai REST API, fake `http.RoundTripper` tests.

## Global Constraints

- Preserve a flag-driven CLI and quoted multi-word GPU names.
- Keep provider behavior out of `cmd/quickdeploy` and behind the existing test seam.
- Read `VAST_API_KEY` from the process environment; never print or commit it.
- Do not add model resolution, image building, startup configuration, polling, proxying, or a TUI.
- Use writable caches: `GOPATH=/tmp/quickdeploy-gopath GOMODCACHE=/tmp/quickdeploy-go-mod-cache GOCACHE=/tmp/quickdeploy-go-cache`.

---

### Task 1: Secure credentials and finish cheapest-offer fetching

**Files:**
- Create: `.gitignore`
- Modify: `internal/vast/client.go`
- Modify: `internal/vast/client_test.go`
- Modify: `cmd/quickdeploy/main.go`
- Modify: `internal/offers/offer.go`
- Test: `internal/offers/offer_test.go`

**Interfaces:**
- Produces: `vast.NewClient(baseURL string, httpClient *http.Client, apiKey string) *Client`
- Preserves: `FindOffers(gpuName string) ([]offers.Offer, error)`

- [ ] **Step 1: Protect the development key**

Add exactly this ignore rule:

```gitignore
.env
```

- [ ] **Step 2: Write failing request and formatting assertions**

Update the client test to construct `NewClient("https://vast.test", httpClient, "test-key")` and assert:

```go
if got := r.Header.Get("Authorization"); got != "Bearer test-key" {
    t.Fatalf("Authorization = %q, want %q", got, "Bearer test-key")
}
if got := int(gotPayload["limit"].(float64)); got != 5 {
    t.Fatalf("limit = %d, want 5", got)
}
order := gotPayload["order"].([]any)[0].([]any)
if order[0] != "dph_total" || order[1] != "asc" {
    t.Fatalf("order = %#v, want dph_total ascending", order)
}
```

Update the offers formatting expectation so every row begins with its deployable ID:

```text
1. ID: 4, GPU: RTX 4090, ...
```

- [ ] **Step 3: Run the focused tests and confirm failure**

Run:

```bash
GOPATH=/tmp/quickdeploy-gopath GOMODCACHE=/tmp/quickdeploy-go-mod-cache GOCACHE=/tmp/quickdeploy-go-cache go test -count=1 ./internal/offers ./internal/vast
```

Expected: failures for the old constructor/header/limit/order and old formatted output.

- [ ] **Step 4: Implement the minimum changes**

Store the trimmed key on `Client`, set the Bearer header on requests, and build the offer payload with:

```go
"order": [][]string{{"dph_total", "asc"}},
"limit": 5,
```

Read the key at the composition root:

```go
client := vast.NewClient(vast.DefaultBaseURL, http.DefaultClient, os.Getenv("VAST_API_KEY"))
```

Prefix formatted rows with `ID: %d` using the existing `Offer.ID`.

- [ ] **Step 5: Run focused tests**

Run the Task 1 command again. Expected: PASS.

### Task 2: Add the Vast deployment request

**Files:**
- Modify: `internal/vast/client.go`
- Test: `internal/vast/client_test.go`

**Interfaces:**
- Consumes: authenticated `*vast.Client`
- Produces: `func (c *Client) Deploy(offerID int, image string) (string, error)`

- [ ] **Step 1: Write failing deployment tests**

Use the existing fake transport to assert a deployment sends:

```text
PUT /api/v0/asks/42/
Authorization: Bearer test-key
Content-Type: application/json
{"image":"ubuntu:22.04"}
```

Return `{"new_contract":"1234"}` and require `Deploy` to return `"1234"`. Add table cases for non-positive offer ID, blank image, missing key, non-2xx status, malformed JSON, and missing `new_contract`; assert validation failures do not call the transport.

- [ ] **Step 2: Run the focused test and confirm failure**

Run:

```bash
GOPATH=/tmp/quickdeploy-gopath GOMODCACHE=/tmp/quickdeploy-go-mod-cache GOCACHE=/tmp/quickdeploy-go-cache go test -count=1 ./internal/vast
```

Expected: compile failure because `Deploy` does not exist.

- [ ] **Step 3: Implement deployment**

Validate inputs, encode only the image field, send `PUT /api/v0/asks/{offerID}/`, reject non-2xx responses, and decode:

```go
var result struct {
    NewContract json.Number `json:"new_contract"`
}
```

Return an error when `NewContract.String()` is empty. Reuse one small request helper for the shared base URL, Bearer header, content type, and missing-key check.

- [ ] **Step 4: Run the focused test**

Run the Task 2 command again. Expected: PASS.

### Task 3: Orchestrate deployment from CLI flags

**Files:**
- Modify: `cmd/quickdeploy/main.go`
- Test: `cmd/quickdeploy/main_test.go`

**Interfaces:**
- Consumes: `FindOffers(string) ([]offers.Offer, error)` and `Deploy(int, string) (string, error)`
- Produces flags: `--deploy`, `--image`, and `--offer-id`

- [ ] **Step 1: Write failing CLI tests**

Use a fake provider with function fields to cover:

```text
--fetch "RTX 5090"                                      -> print offers only
--fetch "RTX 5090" --deploy --image ubuntu:22.04       -> deploy found[0].ID
--deploy --offer-id 42 --image ubuntu:22.04              -> skip fetch and deploy 42
--fetch "RTX 5090" --deploy --offer-id 42 --image ...   -> print offers and deploy 42
```

Assert automatic no-offer results make no deploy call. Add table cases for positional arguments, deploy without image/selector, non-positive explicit ID, and image/offer ID without deploy.

- [ ] **Step 2: Run the focused test and confirm failure**

Run:

```bash
GOPATH=/tmp/quickdeploy-gopath GOMODCACHE=/tmp/quickdeploy-go-mod-cache GOCACHE=/tmp/quickdeploy-go-cache go test -count=1 ./cmd/quickdeploy
```

Expected: compile or assertion failures because the flags and provider method are absent.

- [ ] **Step 3: Implement orchestration**

Extend the injected provider interface with `Deploy`. Validate all flags before calling it. Fetch and print offers when `--fetch` is present; select `found[0].ID` unless a positive `--offer-id` overrides it; then print:

```go
fmt.Fprintf(out, "deployed instance %s from offer %d\n", instanceID, selectedOfferID)
```

- [ ] **Step 4: Run the focused test**

Run the Task 3 command again. Expected: PASS.

### Task 4: Document and verify the MVP

**Files:**
- Modify: `README.md`
- Modify: `docs/superpowers/specs/2026-08-09-quickdeploy-mvp-design.md`

**Interfaces:**
- Documents the final CLI commands and `VAST_API_KEY` requirement.

- [ ] **Step 1: Update user documentation**

Document exporting `VAST_API_KEY`, fetch-only usage, cheapest deployment, explicit offer deployment, the immediate billing behavior, and the absence of readiness polling. Amend the design to state that fetch output includes offer IDs.

- [ ] **Step 2: Run all local checks**

Run:

```bash
GOPATH=/tmp/quickdeploy-gopath GOMODCACHE=/tmp/quickdeploy-go-mod-cache GOCACHE=/tmp/quickdeploy-go-cache gofmt -w cmd/quickdeploy/*.go internal/offers/*.go internal/vast/*.go
GOPATH=/tmp/quickdeploy-gopath GOMODCACHE=/tmp/quickdeploy-go-mod-cache GOCACHE=/tmp/quickdeploy-go-cache go test -count=1 ./...
GOPATH=/tmp/quickdeploy-gopath GOMODCACHE=/tmp/quickdeploy-go-mod-cache GOCACHE=/tmp/quickdeploy-go-cache go vet ./...
git diff --check
```

Expected: every command succeeds.

- [ ] **Step 3: Perform the paid smoke test with guaranteed cleanup**

Source `.env` only in the smoke-test shell. Fetch a low-cost common GPU, abort if the selected offer exceeds $1/hour, deploy `ubuntu:22.04`, record only the returned instance ID, and immediately call:

```text
DELETE https://console.vast.ai/api/v0/instances/<created-instance-id>
Authorization: Bearer <VAST_API_KEY>
```

Require a `200` response with `{"success":true}`. Do not delete any pre-existing instance or retry by creating another instance automatically.

- [ ] **Step 4: Commit the implementation**

Stage `.gitignore`, source, tests, README, and the amended docs—but never `.env`—and commit with:

```bash
git commit -m "feat: add basic Vast deployment"
```
