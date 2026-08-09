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

	resp, err := c.do(http.MethodPost, "/api/v0/bundles/", body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	decoded, err := offers.Decode(resp.Body)
	if err != nil {
		return nil, err
	}
	return offers.TopCheapest(decoded, 5), nil
}

func (c *Client) Deploy(offerID int, image string) (string, error) {
	if offerID <= 0 {
		return "", fmt.Errorf("offer ID must be positive")
	}
	image = strings.TrimSpace(image)
	if image == "" {
		return "", fmt.Errorf("image is required")
	}
	body, err := json.Marshal(map[string]string{"image": image})
	if err != nil {
		return "", err
	}
	resp, err := c.do(http.MethodPut, fmt.Sprintf("/api/v0/asks/%d/", offerID), body)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result struct {
		NewContract json.Number `json:"new_contract"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	if result.NewContract == "" {
		return "", fmt.Errorf("vast api response missing new_contract")
	}
	return result.NewContract.String(), nil
}

func (c *Client) do(method, path string, body []byte) (*http.Response, error) {
	if c.apiKey == "" {
		return nil, fmt.Errorf("VAST_API_KEY is not set")
	}
	req, err := http.NewRequest(method, c.baseURL+path, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		resp.Body.Close()
		return nil, fmt.Errorf("vast api returned %s", resp.Status)
	}
	return resp, nil
}
