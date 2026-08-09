package vast

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"quickdeploy/internal/offers"
)

const DefaultBaseURL = "https://console.vast.ai"

type Client struct {
	baseURL    string
	httpClient *http.Client
	apiKey     string
}

func NewClient(baseURL string, httpClient *http.Client, apiKey string) *Client {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	return &Client{
		baseURL:    strings.TrimRight(baseURL, "/"),
		httpClient: httpClient,
		apiKey:     strings.TrimSpace(apiKey),
	}
}

func (c *Client) FindOffers(gpuName string) ([]offers.Offer, error) {
	if c.apiKey == "" {
		return nil, fmt.Errorf("VAST_API_KEY is not set")
	}

	body, err := json.Marshal(map[string]any{
		"gpu_name":    map[string][]string{"in": []string{gpuName}},
		"num_gpus":    map[string]int{"gte": 1},
		"reliability": map[string]float64{"gte": 0.99},
		"verified":    map[string]bool{"eq": true},
		"rentable":    map[string]bool{"eq": true},
		"type":        "ondemand",
		"order":       [][]string{{"dph_total", "asc"}},
		"limit":       5,
	})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, c.baseURL+"/api/v0/bundles/", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("vast api returned %s", resp.Status)
	}

	decoded, err := offers.Decode(resp.Body)
	if err != nil {
		return nil, err
	}
	return offers.TopCheapest(decoded, 5), nil
}
