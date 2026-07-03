package vast

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
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
		if got := r.Header.Get("Authorization"); got != "$VAST_API_KEY" {
			t.Fatalf("Authorization = %q, want %q", got, "$VAST_API_KEY")
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

	client := NewClient("https://vast.test", httpClient)
	offers, err := client.FindOffers("RTX 5090")
	if err != nil {
		t.Fatalf("FindOffers() error = %v", err)
	}

	gpuName := gotPayload["gpu_name"].(map[string]any)["in"].([]any)[0]
	if gpuName != "RTX 5090" {
		t.Fatalf("gpu_name in request = %q, want RTX 5090", gpuName)
	}
	if got := int(gotPayload["limit"].(float64)); got != 10 {
		t.Fatalf("request limit = %d, want 10", got)
	}
	if len(offers) != 5 {
		t.Fatalf("len(offers) = %d, want 5", len(offers))
	}
	if offers[0].ID != 2 || offers[4].ID != 6 {
		t.Fatalf("offers sorted IDs start/end = %d/%d, want 2/6", offers[0].ID, offers[4].ID)
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}
