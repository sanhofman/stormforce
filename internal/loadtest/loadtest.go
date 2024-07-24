package loadtest

import (
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"stormforce/internal/config"
	"stormforce/internal/results"
	"stormforce/pkg/httpclient"
)

func Run(cfg config.Config) (results.Results, error) {
	startTime := time.Now()

	res := results.Results{
		ResponseTimes:     make([]float64, 0, cfg.N),
		FailedStatusCodes: make([]int, 0),
		MinTime:           math.MaxFloat64,
		MaxTime:           0,
	}

	client := httpclient.NewClient(cfg.CurlMaxTime)

	log.Printf("Starting load test with %d requests and %d threads", cfg.N, cfg.Threads)
	log.Printf("URL: %s", cfg.URL)
	log.Printf("Method: %s", cfg.Method)
	log.Printf("Retry limit: %d", cfg.RetryLimit)
	log.Printf("Threshold time: %.2f seconds", cfg.ThresholdTime)
	log.Printf("Threshold success rate: %.2f%%", cfg.ThresholdSuccess)

	// Warmup necessity depends on loadtesting services; enable/disable in env?
	fmt.Println("Starting warmup...")
	for i := 0; i < cfg.Threads; i++ {
		makeRequest(client, cfg, &res)
	}

	fmt.Println("> WARMUP COMPLETE, STARTING UP THE STORM")
	fmt.Println("========================================")

	pool := NewWorkerPool(cfg.Threads)
	pool.Start()

	var wg sync.WaitGroup
	for i := 0; i < cfg.N; i++ {
		wg.Add(1)
		pool.Submit(func() {
			defer wg.Done()
			makeRequest(client, cfg, &res)
		})
	}

	wg.Wait()
	pool.Stop()

	// Calculate statistics
	sort.Float64s(res.ResponseTimes)
	res.MedianTime = res.ResponseTimes[len(res.ResponseTimes)/2]
	res.PercentileTime90 = res.ResponseTimes[int(float64(len(res.ResponseTimes))*0.9)]
	res.AverageTime = calculateAverage(res.ResponseTimes)
	res.TotalDuration = time.Since(startTime).Seconds()

	return res, nil
}

func makeRequest(client *http.Client, cfg config.Config, res *results.Results) {
	var req *http.Request
	var err error

	for attempt := 0; attempt < cfg.RetryLimit; attempt++ {
		start := time.Now()

		if cfg.Method == "POST" && cfg.RequestBody != "" {
			req, err = http.NewRequest(cfg.Method, cfg.URL, strings.NewReader(cfg.RequestBody))
		} else {
			req, err = http.NewRequest(cfg.Method, cfg.URL, nil)
		}

		if err != nil {
			log.Printf("Error creating request: %v\n", err)
			continue
		}

		for key, value := range cfg.Headers {
			req.Header.Set(key, value)
		}

		if cfg.BearerToken != "" {
			req.Header.Set("Authorization", "Bearer "+cfg.BearerToken)
		}

		resp, err := client.Do(req)
		duration := time.Since(start).Seconds()

		if err != nil {
			log.Printf("Error: %v (Attempt %d/%d)\n", err, attempt+1, cfg.RetryLimit)
			continue
		}

		body, _ := ioutil.ReadAll(resp.Body)
		resp.Body.Close()

		if resp.StatusCode >= 400 {
			res.FailedStatusCodes = append(res.FailedStatusCodes, resp.StatusCode)
			log.Printf("Failed request: Status %d, Time: %.2fs (Attempt %d/%d)\n", resp.StatusCode, duration, attempt+1, cfg.RetryLimit)
			res.ResponseTimes = append(res.ResponseTimes, -1) // Indicate a failed request
			return
		} else {
			log.Printf("Successful request: Status %d, Time: %.2fs\n", resp.StatusCode, duration)

			if cfg.ResponsePattern != "" {
				matched, _ := regexp.Match(cfg.ResponsePattern, body)
				if !matched {
					log.Printf("Response doesn't match the expected pattern\n")
					res.FailedStatusCodes = append(res.FailedStatusCodes, resp.StatusCode)
					res.ResponseTimes = append(res.ResponseTimes, -1) // Indicate a failed request
					return
				}
			}
		}

		res.ResponseTimes = append(res.ResponseTimes, duration)
		if duration < res.MinTime {
			res.MinTime = duration
		}
		if duration > res.MaxTime {
			res.MaxTime = duration
		}
		res.TotalRequests++
		res.SuccessfulRequests++
		return
	}

	log.Printf("Request failed after %d attempts\n", cfg.RetryLimit)
	res.ResponseTimes = append(res.ResponseTimes, -1) // Indicate a failed request after all retries
	res.FailedRequests++
}

func calculateAverage(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	sum := 0.0
	for _, v := range values {
		if v >= 0 {
			sum += v
		}
	}
	return sum / float64(len(values))
}
