package offers

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"
)

type Offer struct {
	ID          int     `json:"id"`
	CPUName     string  `json:"cpu_name"`
	CPURAMGB    float64 `json:"-"`
	GPUName     string  `json:"gpu_name"`
	RAMGB       float64 `json:"-"`
	HourlyPrice float64 `json:"dph_total"`
	Reliability float64 `json:"reliability"`
}

type vastOffer struct {
	ID          int     `json:"id"`
	CPUName     string  `json:"cpu_name"`
	CPURAMMB    float64 `json:"cpu_ram"`
	GPUName     string  `json:"gpu_name"`
	GPURAMMB    float64 `json:"gpu_ram"`
	HourlyPrice float64 `json:"dph_total"`
	Reliability float64 `json:"reliability"`
}

type response struct {
	Offers []vastOffer `json:"offers"`
}

func Decode(r io.Reader) ([]Offer, error) {
	var decoded response
	if err := json.NewDecoder(r).Decode(&decoded); err != nil {
		return nil, err
	}

	offers := make([]Offer, 0, len(decoded.Offers))
	for _, raw := range decoded.Offers {
		offers = append(offers, Offer{
			ID:          raw.ID,
			CPUName:     strings.TrimSpace(raw.CPUName),
			CPURAMGB:    mbToGB(raw.CPURAMMB),
			GPUName:     strings.TrimSpace(raw.GPUName),
			RAMGB:       mbToGB(raw.GPURAMMB),
			HourlyPrice: raw.HourlyPrice,
			Reliability: raw.Reliability,
		})
	}

	return offers, nil
}

func TopCheapest(offers []Offer, limit int) []Offer {
	if limit <= 0 {
		return nil
	}

	sorted := append([]Offer(nil), offers...)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].HourlyPrice < sorted[j].HourlyPrice
	})

	if len(sorted) > limit {
		sorted = sorted[:limit]
	}
	return sorted
}

func Format(offers []Offer) string {
	lines := make([]string, 0, len(offers))
	for i, offer := range offers {
		lines = append(lines, fmt.Sprintf(
			"%d. GPU: %s, RAM: %.0fGB, CPU_RAM: %.0fGB, CPU: %s, Reliability: %.2f%%, Price: $%.2f/hr",
			i+1,
			offer.GPUName,
			offer.RAMGB,
			offer.CPURAMGB,
			offer.CPUName,
			offer.Reliability*100,
			offer.HourlyPrice,
		))
	}
	return strings.Join(lines, "\n")
}

func mbToGB(mb float64) float64 {
	return mb / 1024
}
