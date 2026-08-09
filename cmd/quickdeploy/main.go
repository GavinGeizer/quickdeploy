package main

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"quickdeploy/internal/offers"
	"quickdeploy/internal/vast"
)

type offerFetcher interface {
	FindOffers(gpuName string) ([]offers.Offer, error)
}

type offerDeployer interface {
	Deploy(offerID int, image string) (string, error)
}

func run(args []string, out io.Writer, fetcher offerFetcher, deployer offerDeployer) error {
	flags := flag.NewFlagSet("quickdeploy", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	fetchGPU := flags.String("fetch", "", "GPU name to search for")
	deploy := flags.Bool("deploy", false, "deploy an offer")
	image := flags.String("image", "", "Docker image to deploy")
	offerID := flags.Int("offer-id", 0, "specific Vast offer ID")
	if err := flags.Parse(args); err != nil {
		return err
	}

	set := make(map[string]bool)
	flags.Visit(func(f *flag.Flag) { set[f.Name] = true })
	if flags.NArg() != 0 {
		return fmt.Errorf("positional arguments are not supported")
	}
	if !*deploy && (set["image"] || set["offer-id"]) {
		return fmt.Errorf("--image and --offer-id require --deploy")
	}
	cleanImage := strings.TrimSpace(*image)
	if *deploy && cleanImage == "" {
		return fmt.Errorf("--deploy requires --image")
	}
	if set["offer-id"] && *offerID <= 0 {
		return fmt.Errorf("--offer-id must be positive")
	}
	if *deploy && *fetchGPU == "" && !set["offer-id"] {
		return fmt.Errorf("--deploy requires --fetch or --offer-id")
	}
	if *fetchGPU == "" && !*deploy {
		return fmt.Errorf("usage: quickdeploy --fetch {gpu name}")
	}

	var found []offers.Offer
	if *fetchGPU != "" {
		var err error
		found, err = fetcher.FindOffers(*fetchGPU)
		if err != nil {
			return err
		}
		if len(found) == 0 {
			fmt.Fprintf(out, "no offers found for %s\n", *fetchGPU)
		} else {
			fmt.Fprintln(out, offers.Format(found))
		}
	}

	if !*deploy {
		return nil
	}
	selectedOfferID := *offerID
	if !set["offer-id"] {
		if len(found) == 0 {
			return nil
		}
		selectedOfferID = found[0].ID
	}
	instanceID, err := deployer.Deploy(selectedOfferID, cleanImage)
	if err != nil {
		return err
	}
	fmt.Fprintf(out, "deployed instance %s from offer %d\n", instanceID, selectedOfferID)
	return nil
}

func main() {
	client := vast.NewClient(vast.DefaultBaseURL, &http.Client{Timeout: 30 * time.Second}, os.Getenv("VAST_API_KEY"))
	if err := run(os.Args[1:], os.Stdout, client, client); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
