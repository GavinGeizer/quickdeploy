package main

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"

	"quickdeploy/internal/offers"
	"quickdeploy/internal/vast"
)

type offerFetcher interface {
	FindOffers(gpuName string) ([]offers.Offer, error)
}

type fetchFunc func(gpuName string) ([]offers.Offer, error)

func (f fetchFunc) FindOffers(gpuName string) ([]offers.Offer, error) {
	return f(gpuName)
}

func run(args []string, out io.Writer, fetcher offerFetcher) error {
	flags := flag.NewFlagSet("quickdeploy", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	fetchGPU := flags.String("fetch", "", "GPU name to search for")
	if err := flags.Parse(args); err != nil {
		return err
	}

	if *fetchGPU == "" {
		return fmt.Errorf("usage: quickdeploy --fetch {gpu name}")
	}

	found, err := fetcher.FindOffers(*fetchGPU)
	if err != nil {
		return err
	}
	if len(found) == 0 {
		fmt.Fprintf(out, "no offers found for %s\n", *fetchGPU)
		return nil
	}

	fmt.Fprintln(out, offers.Format(found))
	return nil
}

func main() {
	client := vast.NewClient(vast.DefaultBaseURL, http.DefaultClient, os.Getenv("VAST_API_KEY"))
	if err := run(os.Args[1:], os.Stdout, client); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
