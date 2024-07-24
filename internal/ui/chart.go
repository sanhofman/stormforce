package ui

import (
	"fmt"
	"math"
	"os"
	"sort"

	"stormforce/internal/config"
	"stormforce/internal/results"

	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/components"
	"github.com/go-echarts/go-echarts/v2/opts"
)

func GenerateCharts(results results.Results, cfg config.Config) error {
	page := components.NewPage()
	page.PageTitle = "Load Test Results"

	page.AddCharts(
		generateHistogram(results),
		generateCDF(results),
		generateResponseTimeSeries(results),
		generateRequestRateVsResponseTime(results),
		generateErrorRateChart(results),
		generatePercentileDistribution(results),
		generateStatusCodeDistribution(results),
		generateConcurrentUsersVsResponseTime(results, cfg),
	)

	f, err := os.Create("load_test_results.html")
	if err != nil {
		return fmt.Errorf("error creating output file: %v", err)
	}
	defer f.Close()

	return page.Render(f)
}

func generateHistogram(results results.Results) *charts.Bar {
	histChart := charts.NewBar()
	histChart.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{Title: "Response Time Distribution"}),
		charts.WithXAxisOpts(opts.XAxis{Name: "Response Time (s)"}),
		charts.WithYAxisOpts(opts.YAxis{Name: "Frequency"}),
	)

	buckets := 20
	histData := make([]opts.BarData, buckets)
	bucketSize := (results.MaxTime - results.MinTime) / float64(buckets)

	for i := range histData {
		histData[i] = opts.BarData{Value: 0}
	}

	for _, time := range results.ResponseTimes {
		if time >= 0 {
			bucket := int((time - results.MinTime) / bucketSize)
			if bucket == buckets {
				bucket--
			}
			histData[bucket].Value = histData[bucket].Value.(int) + 1
		}
	}

	xAxisData := make([]string, buckets)
	for i := 0; i < buckets; i++ {
		xAxisData[i] = fmt.Sprintf("%.2f", results.MinTime+float64(i)*bucketSize)
	}

	histChart.SetXAxis(xAxisData).AddSeries("Response Time", histData)
	return histChart
}

func generateCDF(results results.Results) *charts.Line {
	cdfChart := charts.NewLine()
	cdfChart.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{Title: "Cumulative Distribution Function"}),
		charts.WithXAxisOpts(opts.XAxis{Name: "Response Time (s)"}),
		charts.WithYAxisOpts(opts.YAxis{Name: "Cumulative Probability"}),
	)

	validTimes := []float64{}
	for _, time := range results.ResponseTimes {
		if time >= 0 {
			validTimes = append(validTimes, time)
		}
	}
	sort.Float64s(validTimes)

	cdfData := make([]opts.LineData, len(validTimes))
	for i, time := range validTimes {
		cdfData[i] = opts.LineData{Value: []float64{time, float64(i+1) / float64(len(validTimes))}}
	}

	cdfChart.AddSeries("CDF", cdfData)
	return cdfChart
}

func generateResponseTimeSeries(results results.Results) *charts.Line {
	timeSeriesChart := charts.NewLine()
	timeSeriesChart.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{Title: "Response Time Series"}),
		charts.WithXAxisOpts(opts.XAxis{Name: "Request Number"}),
		charts.WithYAxisOpts(opts.YAxis{Name: "Response Time (s)"}),
	)

	data := make([]opts.LineData, len(results.ResponseTimes))
	for i, time := range results.ResponseTimes {
		if time >= 0 {
			data[i] = opts.LineData{Value: time}
		} else {
			data[i] = opts.LineData{Value: nil}
		}
	}

	timeSeriesChart.AddSeries("Response Time", data)
	return timeSeriesChart
}

func generateRequestRateVsResponseTime(results results.Results) *charts.Scatter {
	scatterChart := charts.NewScatter()
	scatterChart.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{Title: "Request Rate vs. Response Time"}),
		charts.WithXAxisOpts(opts.XAxis{Name: "Request Rate (requests/second)"}),
		charts.WithYAxisOpts(opts.YAxis{Name: "Response Time (s)"}),
	)

	data := make([]opts.ScatterData, 0, len(results.ResponseTimes))
	for i, time := range results.ResponseTimes {
		if time >= 0 {
			requestRate := float64(i+1) / results.TotalDuration
			data = append(data, opts.ScatterData{Value: []float64{requestRate, time}})
		}
	}

	scatterChart.AddSeries("Request Rate vs. Response Time", data)
	return scatterChart
}

func generateErrorRateChart(results results.Results) *charts.Pie {
	pieChart := charts.NewPie()
	pieChart.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{Title: "Error Rate"}),
	)

	successfulRequests := results.SuccessfulRequests
	failedRequests := results.FailedRequests

	data := []opts.PieData{
		{Name: "Successful", Value: successfulRequests},
		{Name: "Failed", Value: failedRequests},
	}

	pieChart.AddSeries("Error Rate", data)
	pieChart.SetSeriesOptions(charts.WithLabelOpts(
		opts.Label{
			Show:      opts.Bool(true),
			Formatter: "{b}: {c} ({d}%)",
		},
	))

	return pieChart
}

func generatePercentileDistribution(results results.Results) *charts.Line {
	lineChart := charts.NewLine()
	lineChart.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{Title: "Percentile Distribution"}),
		charts.WithXAxisOpts(opts.XAxis{Name: "Percentile"}),
		charts.WithYAxisOpts(opts.YAxis{Name: "Response Time (s)"}),
	)

	validTimes := []float64{}
	for _, time := range results.ResponseTimes {
		if time >= 0 {
			validTimes = append(validTimes, time)
		}
	}
	sort.Float64s(validTimes)

	percentiles := []float64{50, 75, 90, 95, 99}
	data := make([]opts.LineData, len(percentiles))
	xAxis := make([]string, len(percentiles))

	for i, p := range percentiles {
		index := int(float64(len(validTimes)-1) * p / 100)
		data[i] = opts.LineData{Value: validTimes[index]}
		xAxis[i] = fmt.Sprintf("P%.0f", p)
	}

	lineChart.SetXAxis(xAxis).AddSeries("Percentile", data)
	return lineChart
}

func generateStatusCodeDistribution(results results.Results) *charts.Bar {
	barChart := charts.NewBar()
	barChart.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{Title: "Status Code Distribution"}),
		charts.WithXAxisOpts(opts.XAxis{Name: "Status Code"}),
		charts.WithYAxisOpts(opts.YAxis{Name: "Count"}),
	)

	statusCodes := make(map[int]int)
	for _, code := range results.FailedStatusCodes {
		statusCodes[code]++
	}
	statusCodes[200] = results.SuccessfulRequests

	var keys []int
	for k := range statusCodes {
		keys = append(keys, k)
	}
	sort.Ints(keys)

	xAxis := make([]string, len(keys))
	data := make([]opts.BarData, len(keys))
	for i, code := range keys {
		xAxis[i] = fmt.Sprintf("%d", code)
		data[i] = opts.BarData{Value: statusCodes[code]}
	}

	barChart.SetXAxis(xAxis).AddSeries("Status Codes", data)
	return barChart
}

func generateConcurrentUsersVsResponseTime(results results.Results, cfg config.Config) *charts.Scatter {
	scatterChart := charts.NewScatter()
	scatterChart.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{Title: "Concurrent Users vs. Response Time"}),
		charts.WithXAxisOpts(opts.XAxis{Name: "Concurrent Users"}),
		charts.WithYAxisOpts(opts.YAxis{Name: "Response Time (s)"}),
	)

	data := make([]opts.ScatterData, 0, len(results.ResponseTimes))
	for i, time := range results.ResponseTimes {
		if time >= 0 {
			concurrentUsers := int(math.Min(float64(i+1), float64(cfg.Threads)))
			data = append(data, opts.ScatterData{Value: []float64{float64(concurrentUsers), time}})
		}
	}

	scatterChart.AddSeries("Concurrent Users vs. Response Time", data)
	return scatterChart
}
