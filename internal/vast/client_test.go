package vast

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestClientFindOffersPostsGPUNameAndReturnsTopFive(t *testing.T) {
	var gotPayload map[string]any

	httpClient := &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		if r.Method != http.MethodPost {
			t.Fatalf("method = %s, want POST", r.Method)
		}
		if r.URL.Path != "/api/v0/bundles/" {
			t.Fatalf("path = %s, want /api/v0/bundles/", r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer test-key" {
			t.Fatalf("Authorization = %q, want %q", got, "Bearer test-key")
		}
		if err := json.NewDecoder(r.Body).Decode(&gotPayload); err != nil {
			t.Fatalf("decode request body: %v", err)
		}

		return &http.Response{
			StatusCode: http.StatusOK,
			Status:     "200 OK",
			Header:     http.Header{"Content-Type": []string{"application/json"}},
			Body: io.NopCloser(bytes.NewBufferString(`{
			"offers": [
				{"id": 1, "gpu_name": "RTX 5090", "gpu_ram": 32607, "cpu_ram": 64467, "cpu_name": "AMD EPYC 1", "dph_total": 0.90, "reliability": 0.991},
				{"id": 2, "gpu_name": "RTX 5090", "gpu_ram": 32607, "cpu_ram": 64467, "cpu_name": "AMD EPYC 2", "dph_total": 0.10, "reliability": 0.992},
				{"id": 3, "gpu_name": "RTX 5090", "gpu_ram": 32607, "cpu_ram": 64467, "cpu_name": "AMD EPYC 3", "dph_total": 0.20, "reliability": 0.993},
				{"id": 4, "gpu_name": "RTX 5090", "gpu_ram": 32607, "cpu_ram": 64467, "cpu_name": "AMD EPYC 4", "dph_total": 0.30, "reliability": 0.994},
				{"id": 5, "gpu_name": "RTX 5090", "gpu_ram": 32607, "cpu_ram": 64467, "cpu_name": "AMD EPYC 5", "dph_total": 0.40, "reliability": 0.995},
				{"id": 6, "gpu_name": "RTX 5090", "gpu_ram": 32607, "cpu_ram": 64467, "cpu_name": "AMD EPYC 6", "dph_total": 0.50, "reliability": 0.996}
			]
		}`)),
		}, nil
	})}

	client := NewClient("https://vast.test", httpClient, "test-key")
	offers, err := client.FindOffers("RTX 5090")
	if err != nil {
		t.Fatalf("FindOffers() error = %v", err)
	}

	gpuName := gotPayload["gpu_name"].(map[string]any)["in"].([]any)[0]
	if gpuName != "RTX 5090" {
		t.Fatalf("gpu_name in request = %q, want RTX 5090", gpuName)
	}
	if got := int(gotPayload["limit"].(float64)); got != 5 {
		t.Fatalf("request limit = %d, want 5", got)
	}
	order := gotPayload["order"].([]any)[0].([]any)
	if order[0] != "dph_total" || order[1] != "asc" {
		t.Fatalf("order = %#v, want dph_total ascending", order)
	}
	if got := gotPayload["reliability"].(map[string]any)["gte"]; got != 0.99 {
		t.Fatalf("minimum reliability = %v, want 0.99", got)
	}
	if got := gotPayload["verified"].(map[string]any)["eq"]; got != true {
		t.Fatalf("verified filter = %v, want true", got)
	}
	if got := gotPayload["rentable"].(map[string]any)["eq"]; got != true {
		t.Fatalf("rentable filter = %v, want true", got)
	}
	if got := gotPayload["num_gpus"].(map[string]any)["gte"]; got != float64(1) {
		t.Fatalf("minimum GPUs = %v, want 1", got)
	}
	if got := gotPayload["type"]; got != "ondemand" {
		t.Fatalf("type = %v, want ondemand", got)
	}
	if len(offers) != 5 {
		t.Fatalf("len(offers) = %d, want 5", len(offers))
	}
	if offers[0].ID != 2 || offers[4].ID != 6 {
		t.Fatalf("offers sorted IDs start/end = %d/%d, want 2/6", offers[0].ID, offers[4].ID)
	}
}

func TestClientFindOffersRequiresAPIKey(t *testing.T) {
	called := false
	httpClient := &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		called = true
		return nil, nil
	})}

	client := NewClient("https://vast.test", httpClient, "")
	_, err := client.FindOffers("RTX 5090")
	if err == nil {
		t.Fatal("FindOffers() error = nil, want missing API key error")
	}
	if called {
		t.Fatal("FindOffers() called HTTP transport without an API key")
	}
}

func TestClientDeployCreatesInstance(t *testing.T) {
	var gotPayload struct {
		Image string `json:"image"`
	}

	httpClient := &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		if r.Method != http.MethodPut {
			t.Fatalf("method = %s, want PUT", r.Method)
		}
		if r.URL.Path != "/api/v0/asks/42/" {
			t.Fatalf("path = %s, want /api/v0/asks/42/", r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer test-key" {
			t.Fatalf("Authorization = %q, want %q", got, "Bearer test-key")
		}
		if got := r.Header.Get("Content-Type"); got != "application/json" {
			t.Fatalf("Content-Type = %q, want application/json", got)
		}
		if err := json.NewDecoder(r.Body).Decode(&gotPayload); err != nil {
			t.Fatalf("decode request body: %v", err)
		}

		return &http.Response{
			StatusCode: http.StatusOK,
			Status:     "200 OK",
			Body:       io.NopCloser(bytes.NewBufferString(`{"new_contract":1234}`)),
		}, nil
	})}

	client := NewClient("https://vast.test", httpClient, "test-key")
	instanceID, err := client.Deploy(42, "ubuntu:22.04")
	if err != nil {
		t.Fatalf("Deploy() error = %v", err)
	}
	if gotPayload.Image != "ubuntu:22.04" {
		t.Fatalf("image = %q, want ubuntu:22.04", gotPayload.Image)
	}
	if instanceID != "1234" {
		t.Fatalf("instance ID = %q, want 1234", instanceID)
	}
}

func TestClientDeployRejectsInvalidInputBeforeRequest(t *testing.T) {
	tests := []struct {
		name    string
		offerID int
		image   string
		apiKey  string
	}{
		{name: "zero offer ID", offerID: 0, image: "ubuntu:22.04", apiKey: "test-key"},
		{name: "negative offer ID", offerID: -1, image: "ubuntu:22.04", apiKey: "test-key"},
		{name: "blank image", offerID: 42, image: "  ", apiKey: "test-key"},
		{name: "missing API key", offerID: 42, image: "ubuntu:22.04"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			called := false
			httpClient := &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
				called = true
				return nil, nil
			})}

			client := NewClient("https://vast.test", httpClient, test.apiKey)
			if _, err := client.Deploy(test.offerID, test.image); err == nil {
				t.Fatal("Deploy() error = nil, want validation error")
			}
			if called {
				t.Fatal("Deploy() called HTTP transport for invalid input")
			}
		})
	}
}

func TestClientDeployRejectsInvalidResponses(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		status     string
		body       string
		uncertain  bool
	}{
		{name: "API error", statusCode: http.StatusBadRequest, status: "400 Bad Request", body: `{}`},
		{name: "malformed JSON", statusCode: http.StatusOK, status: "200 OK", body: `{`, uncertain: true},
		{name: "missing instance ID", statusCode: http.StatusOK, status: "200 OK", body: `{}`, uncertain: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			httpClient := &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: test.statusCode,
					Status:     test.status,
					Body:       io.NopCloser(bytes.NewBufferString(test.body)),
				}, nil
			})}

			client := NewClient("https://vast.test", httpClient, "test-key")
			_, err := client.Deploy(42, "ubuntu:22.04")
			if err == nil {
				t.Fatal("Deploy() error = nil, want response error")
			}
			if test.uncertain && !strings.Contains(err.Error(), "deployment outcome unknown") {
				t.Fatalf("Deploy() error = %q, want uncertain-outcome warning", err)
			}
		})
	}
}

func TestClientDeployReturnsTransportError(t *testing.T) {
	wantErr := errors.New("network failed")
	httpClient := &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		return nil, wantErr
	})}

	client := NewClient("https://vast.test", httpClient, "test-key")
	_, err := client.Deploy(42, "ubuntu:22.04")
	if !errors.Is(err, wantErr) {
		t.Fatalf("Deploy() error = %v, want network failed", err)
	}
	if !strings.Contains(err.Error(), "deployment outcome unknown") {
		t.Fatalf("Deploy() error = %q, want uncertain-outcome warning", err)
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}
