package main

import (
	"bytes"
	"strings"
	"testing"

	"quickdeploy/internal/offers"
)

func TestRunFetchPrintsFormattedTopCheapestOffers(t *testing.T) {
	fetcher := fetchFunc(func(gpuName string) ([]offers.Offer, error) {
		if gpuName != "RTX 5090" {
			t.Fatalf("gpuName = %q, want RTX 5090", gpuName)
		}
		return []offers.Offer{
			{ID: 2, GPUName: "RTX 5090", RAMGB: 32, CPURAMGB: 64, CPUName: "AMD EPYC", Reliability: 0.992, HourlyPrice: 0.10},
		}, nil
	})

	var out bytes.Buffer
	if err := run([]string{"--fetch", "RTX 5090"}, &out, fetcher); err != nil {
		t.Fatalf("run() error = %v", err)
	}

	got := strings.TrimSpace(out.String())
	want := "1. GPU: RTX 5090, RAM: 32GB, CPU_RAM: 64GB, CPU: AMD EPYC, Reliability: 99.20%, Price: $0.10/hr"
	if got != want {
		t.Fatalf("output = %q, want %q", got, want)
	}
}
