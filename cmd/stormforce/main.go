package main

import (
	"log"

	"stormforce/internal/config"
	"stormforce/internal/loadtest"
	"stormforce/internal/ui"
)

func main() {
	ui.PrintLogo()

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Error loading configuration: %v", err)
	}

	err = config.SetupLogging(cfg)
	if err != nil {
		log.Fatalf("Error setting up logging: %v", err)
	}

	results, err := loadtest.Run(cfg)
	if err != nil {
		log.Fatalf("Error running load test: %v", err)
	}

	ui.DisplayResults(results, cfg)
	err = ui.GenerateCharts(results, cfg)
	if err != nil {
		log.Printf("Error generating charts: %v", err)
	}

	if cfg.JSONOutput {
		err = results.OutputJSON()
		if err != nil {
			log.Printf("Error outputting JSON: %v", err)
		}
	}

	config.Cleanup()
}
