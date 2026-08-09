package main

import (
	"bytes"
	"errors"
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
	if err := run([]string{"--fetch", "RTX 5090"}, &out, fetcher, nil); err != nil {
		t.Fatalf("run() error = %v", err)
	}

	got := strings.TrimSpace(out.String())
	want := "1. ID: 2, GPU: RTX 5090, RAM: 32GB, CPU_RAM: 64GB, CPU: AMD EPYC, Reliability: 99.20%, Price: $0.10/hr"
	if got != want {
		t.Fatalf("output = %q, want %q", got, want)
	}
}

func TestRunDeploysCheapestFetchedOffer(t *testing.T) {
	fetcher := fetchFunc(func(gpuName string) ([]offers.Offer, error) {
		return []offers.Offer{
			{ID: 7, GPUName: "RTX 5090", HourlyPrice: 0.10},
			{ID: 8, GPUName: "RTX 5090", HourlyPrice: 0.20},
		}, nil
	})
	var gotOfferID int
	var gotImage string
	deployer := deployFunc(func(offerID int, image string) (string, error) {
		gotOfferID = offerID
		gotImage = image
		return "99", nil
	})

	var out bytes.Buffer
	err := run([]string{"--fetch", "RTX 5090", "--deploy", "--image", "ubuntu:22.04"}, &out, fetcher, deployer)
	if err != nil {
		t.Fatalf("run() error = %v", err)
	}
	if gotOfferID != 7 || gotImage != "ubuntu:22.04" {
		t.Fatalf("Deploy() got offer/image = %d/%q, want 7/ubuntu:22.04", gotOfferID, gotImage)
	}
	if got := out.String(); !strings.Contains(got, "1. ID: 7") || !strings.Contains(got, "deployed instance 99 from offer 7") {
		t.Fatalf("output = %q, want offers and deployment confirmation", got)
	}
}

func TestRunDeploysExplicitOfferWithoutFetch(t *testing.T) {
	fetchCalled := false
	fetcher := fetchFunc(func(gpuName string) ([]offers.Offer, error) {
		fetchCalled = true
		return nil, nil
	})
	var gotOfferID int
	deployer := deployFunc(func(offerID int, image string) (string, error) {
		gotOfferID = offerID
		return "99", nil
	})

	var out bytes.Buffer
	err := run([]string{"--deploy", "--offer-id", "42", "--image", "ubuntu:22.04"}, &out, fetcher, deployer)
	if err != nil {
		t.Fatalf("run() error = %v", err)
	}
	if fetchCalled {
		t.Fatal("run() fetched offers for an explicit deployment")
	}
	if gotOfferID != 42 {
		t.Fatalf("Deploy() offer ID = %d, want 42", gotOfferID)
	}
	if got := strings.TrimSpace(out.String()); got != "deployed instance 99 from offer 42" {
		t.Fatalf("output = %q, want deployment confirmation", got)
	}
}

func TestRunExplicitOfferOverridesFetchedOffer(t *testing.T) {
	fetcher := fetchFunc(func(gpuName string) ([]offers.Offer, error) {
		return []offers.Offer{{ID: 7, GPUName: "RTX 5090", HourlyPrice: 0.10}}, nil
	})
	var gotOfferID int
	deployer := deployFunc(func(offerID int, image string) (string, error) {
		gotOfferID = offerID
		return "99", nil
	})

	var out bytes.Buffer
	err := run([]string{"--fetch", "RTX 5090", "--deploy", "--offer-id", "42", "--image", "ubuntu:22.04"}, &out, fetcher, deployer)
	if err != nil {
		t.Fatalf("run() error = %v", err)
	}
	if gotOfferID != 42 {
		t.Fatalf("Deploy() offer ID = %d, want explicit ID 42", gotOfferID)
	}
}

func TestRunDoesNotDeployWhenNoOffersFound(t *testing.T) {
	fetcher := fetchFunc(func(gpuName string) ([]offers.Offer, error) { return nil, nil })
	deployCalled := false
	deployer := deployFunc(func(offerID int, image string) (string, error) {
		deployCalled = true
		return "", nil
	})

	var out bytes.Buffer
	err := run([]string{"--fetch", "RTX 5090", "--deploy", "--image", "ubuntu:22.04"}, &out, fetcher, deployer)
	if err != nil {
		t.Fatalf("run() error = %v", err)
	}
	if deployCalled {
		t.Fatal("run() deployed without an offer")
	}
	if got := strings.TrimSpace(out.String()); got != "no offers found for RTX 5090" {
		t.Fatalf("output = %q, want no-offers message", got)
	}
}

func TestRunRejectsInvalidDeployFlagsBeforeNetworkCalls(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{name: "positional argument", args: []string{"--fetch", "RTX 5090", "extra"}},
		{name: "deploy without image", args: []string{"--deploy", "--offer-id", "42"}},
		{name: "deploy without selector", args: []string{"--deploy", "--image", "ubuntu:22.04"}},
		{name: "negative offer ID", args: []string{"--deploy", "--offer-id=-1", "--image", "ubuntu:22.04"}},
		{name: "blank image", args: []string{"--deploy", "--offer-id", "42", "--image", "  "}},
		{name: "image without deploy", args: []string{"--fetch", "RTX 5090", "--image", "ubuntu:22.04"}},
		{name: "offer ID without deploy", args: []string{"--fetch", "RTX 5090", "--offer-id", "42"}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fetcher := fetchFunc(func(gpuName string) ([]offers.Offer, error) {
				t.Fatal("FindOffers() called for invalid flags")
				return nil, nil
			})
			deployer := deployFunc(func(offerID int, image string) (string, error) {
				t.Fatal("Deploy() called for invalid flags")
				return "", nil
			})

			if err := run(test.args, &bytes.Buffer{}, fetcher, deployer); err == nil {
				t.Fatal("run() error = nil, want flag validation error")
			}
		})
	}
}

func TestRunReturnsDeployError(t *testing.T) {
	wantErr := errors.New("deploy failed")
	deployer := deployFunc(func(offerID int, image string) (string, error) {
		return "", wantErr
	})

	err := run(
		[]string{"--deploy", "--offer-id", "42", "--image", "ubuntu:22.04"},
		&bytes.Buffer{},
		fetchFunc(func(gpuName string) ([]offers.Offer, error) { return nil, nil }),
		deployer,
	)
	if !errors.Is(err, wantErr) {
		t.Fatalf("run() error = %v, want deploy failed", err)
	}
}

type deployFunc func(offerID int, image string) (string, error)

func (f deployFunc) Deploy(offerID int, image string) (string, error) {
	return f(offerID, image)
}

type fetchFunc func(gpuName string) ([]offers.Offer, error)

func (f fetchFunc) FindOffers(gpuName string) ([]offers.Offer, error) {
	return f(gpuName)
}
