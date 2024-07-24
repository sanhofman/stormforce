package ui

import (
	"fmt"
	"log"
	"time"

	"stormforce/internal/config"
	"stormforce/internal/results"
)

func PrintLogo() {
	logo := `
    ⚡️ icapps StormForce ⚡️
    Load Testing Tool
    ==================
    `
	fmt.Println(logo)
	fmt.Println("Welcome to StormForce - Your Ultimate Load Testing Companion!")
	fmt.Println("Version 1.0.0")
	fmt.Println("====================================================")
}

func DisplayResults(results results.Results, config config.Config) {
	fmt.Println("\n========================================")
	fmt.Println("RESULTS:")
	fmt.Printf("- Total requests: %d\n", results.TotalRequests)
	fmt.Printf("- Successful requests: %d\n", results.SuccessfulRequests)
	fmt.Printf("- Failed requests: %d\n", results.FailedRequests)
	fmt.Printf("- Average response time: %.2f seconds\n", results.AverageTime)
	fmt.Printf("- Median response time: %.2f seconds\n", results.MedianTime)
	fmt.Printf("- 90th percentile response time: %.2f seconds\n", results.PercentileTime90)
	fmt.Printf("- Min response time: %.2f seconds\n", results.MinTime)
	fmt.Printf("- Max response time: %.2f seconds\n", results.MaxTime)

	successRate := float64(results.SuccessfulRequests) / float64(results.TotalRequests) * 100
	fmt.Printf("- Success rate: %.2f%%\n", successRate)

	if results.AverageTime > config.ThresholdTime {
		fmt.Println("- Average response time is higher than the threshold. 🚩")
	} else {
		fmt.Println("- Average response time is within the acceptable range. 👍")
	}

	if successRate < config.ThresholdSuccess {
		fmt.Println("- Success rate is lower than the threshold. 🚩")
	} else {
		fmt.Println("- Success rate meets the threshold. ✔️")
	}
	fmt.Println("========================================")

	log.Println("Test Results:")
	log.Printf("Total requests: %d\n", results.TotalRequests)
	log.Printf("Successful requests: %d\n", results.SuccessfulRequests)
	log.Printf("Failed requests: %d\n", results.FailedRequests)
	log.Printf("Average response time: %.2f seconds\n", results.AverageTime)
	log.Printf("Median response time: %.2f seconds\n", results.MedianTime)
	log.Printf("90th percentile response time: %.2f seconds\n", results.PercentileTime90)
	log.Printf("Min response time: %.2f seconds\n", results.MinTime)
	log.Printf("Max response time: %.2f seconds\n", results.MaxTime)
	log.Printf("Success rate: %.2f%%\n", successRate)
}

func Spinner(done chan bool) {
	spinChars := []string{"⚡", "🌩️", "⚡", "🌪️"}
	delay := 100 * time.Millisecond
	i := 0
	for {
		select {
		case <-done:
			fmt.Print("\r") // Clear the spinner line
			return
		default:
			fmt.Printf("\r[%s] StormForce is unleashing chaos... ", spinChars[i])
			i = (i + 1) % len(spinChars)
			time.Sleep(delay)
		}
	}
}
