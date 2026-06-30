package main

import "fmt"



/*
should come out like this:
 -> 1. GPU: NVIDIA RTX 4090, VRAM: 24GB, CPU: Intel i9-10900K, RAM: 32GB, Price: $3.50/hr
*/
type Offer struct {
    ID          int // Unique identifier for the offer, Can be hidden from the user
	CPUName     string
	RAMGB       float64
    GPUName     string
    VRAMGB      float64
    HourlyPrice float64
    Reliability float64 //dont need to show this either
}

func FindOffers() ([]Offer, error) {
    // Implementation for finding offers
    return nil, nil
}

func main() {
    fmt.Println("Hello, World!")
}