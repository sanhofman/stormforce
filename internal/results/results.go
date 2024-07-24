package results

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

type Results struct {
	TotalRequests      int
	SuccessfulRequests int
	FailedRequests     int
	ResponseTimes      []float64
	FailedStatusCodes  []int
	MinTime            float64
	MaxTime            float64
	MedianTime         float64
	PercentileTime90   float64
	AverageTime        float64
	TotalDuration      float64
}

func (r *Results) OutputJSON() error {
	jsonData, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshalling JSON: %v", err)
	}

	err = ioutil.WriteFile("results.json", jsonData, 0644)
	if err != nil {
		return fmt.Errorf("error writing JSON file: %v", err)
	}

	fmt.Println("Results saved to results.json")
	return nil
}
