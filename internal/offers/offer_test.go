package offers

import (
	"strings"
	"testing"
)

func TestTopCheapestParsesSortsLimitsAndFormats(t *testing.T) {
	input := strings.NewReader(`{
		"offers": [
			{"id": 1, "gpu_name": "RTX 4090", "gpu_ram": 24576, "cpu_ram": 65536, "cpu_name": "AMD EPYC 1", "dph_total": 0.90, "reliability": 0.991},
			{"id": 2, "gpu_name": "RTX 4090", "gpu_ram": 24576, "cpu_ram": 32768, "cpu_name": "AMD EPYC 2", "dph_total": 0.12, "reliability": 0.992},
			{"id": 3, "gpu_name": "RTX 4090", "gpu_ram": 24576, "cpu_ram": 49152, "cpu_name": "AMD EPYC 3", "dph_total": 0.51, "reliability": 0.993},
			{"id": 4, "gpu_name": "RTX 4090", "gpu_ram": 24576, "cpu_ram": 8192, "cpu_name": "AMD EPYC 4", "dph_total": 0.08, "reliability": 0.994},
			{"id": 5, "gpu_name": "RTX 4090", "gpu_ram": 24576, "cpu_ram": 16384, "cpu_name": "AMD EPYC 5", "dph_total": 0.31, "reliability": 0.995},
			{"id": 6, "gpu_name": "RTX 4090", "gpu_ram": 24576, "cpu_ram": 24576, "cpu_name": "AMD EPYC 6", "dph_total": 0.44, "reliability": 0.996}
		]
	}`)

	decoded, err := Decode(input)
	if err != nil {
		t.Fatalf("Decode() error = %v", err)
	}

	cheapest := TopCheapest(decoded, 5)
	if len(cheapest) != 5 {
		t.Fatalf("len(cheapest) = %d, want 5", len(cheapest))
	}

	wantIDs := []int{4, 2, 5, 6, 3}
	for i, wantID := range wantIDs {
		if cheapest[i].ID != wantID {
			t.Fatalf("cheapest[%d].ID = %d, want %d", i, cheapest[i].ID, wantID)
		}
	}

	got := Format(cheapest)
	want := strings.Join([]string{
		"1. ID: 4, GPU: RTX 4090, RAM: 24GB, CPU_RAM: 8GB, CPU: AMD EPYC 4, Reliability: 99.40%, Price: $0.08/hr",
		"2. ID: 2, GPU: RTX 4090, RAM: 24GB, CPU_RAM: 32GB, CPU: AMD EPYC 2, Reliability: 99.20%, Price: $0.12/hr",
		"3. ID: 5, GPU: RTX 4090, RAM: 24GB, CPU_RAM: 16GB, CPU: AMD EPYC 5, Reliability: 99.50%, Price: $0.31/hr",
		"4. ID: 6, GPU: RTX 4090, RAM: 24GB, CPU_RAM: 24GB, CPU: AMD EPYC 6, Reliability: 99.60%, Price: $0.44/hr",
		"5. ID: 3, GPU: RTX 4090, RAM: 24GB, CPU_RAM: 48GB, CPU: AMD EPYC 3, Reliability: 99.30%, Price: $0.51/hr",
	}, "\n")

	if got != want {
		t.Fatalf("Format() =\n%s\nwant\n%s", got, want)
	}
}

func TestDecodeRejectsMissingOffers(t *testing.T) {
	if _, err := Decode(strings.NewReader(`{}`)); err == nil {
		t.Fatal("Decode() error = nil, want missing offers error")
	}
}

func TestDecodeAcceptsEmptyOffers(t *testing.T) {
	offers, err := Decode(strings.NewReader(`{"offers":[]}`))
	if err != nil {
		t.Fatalf("Decode() error = %v", err)
	}
	if len(offers) != 0 {
		t.Fatalf("len(offers) = %d, want 0", len(offers))
	}
}
