package config

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	N                int
	Threads          int
	URL              string
	Method           string
	BearerToken      string
	Headers          map[string]string
	RetryLimit       int
	ThresholdTime    float64
	ThresholdSuccess float64
	EmailEnabled     bool
	EmailTo          string
	CurlMaxTime      int
	LogFile          string
	DisableLogging   bool
	RequestBody      string
	ResponsePattern  string
	JSONOutput       bool
}

func LoadConfig() (Config, error) {
	err := godotenv.Load()
	if err != nil {
		return Config{}, fmt.Errorf("error loading .env file: %w", err)
	}

	config := Config{
		N:                getEnvAsInt("REQUESTS", 1000),
		Threads:          getEnvAsInt("THREADS", 10),
		URL:              os.Getenv("URL"),
		Method:           os.Getenv("METHOD"),
		Headers:          make(map[string]string),
		RetryLimit:       getEnvAsInt("RETRY_LIMIT", 3),
		ThresholdTime:    getEnvAsFloat("THRESHOLD_TIME", 1.0),
		ThresholdSuccess: getEnvAsFloat("THRESHOLD_SUCCESS", 95.0),
		EmailEnabled:     getEnvAsBool("EMAIL_ENABLED", false),
		EmailTo:          os.Getenv("EMAIL_TO"),
		CurlMaxTime:      getEnvAsInt("CURL_MAX_TIME", 10),
		LogFile:          os.Getenv("LOG_FILE"),
		DisableLogging:   getEnvAsBool("DISABLE_LOGGING", false),
		RequestBody:      os.Getenv("REQUEST_BODY"),
		ResponsePattern:  os.Getenv("RESPONSE_PATTERN"),
		JSONOutput:       getEnvAsBool("JSON_OUTPUT", false),
	}

	return promptForOverrides(config)
}

func promptForOverrides(config Config) (Config, error) {
	fmt.Println("\n🔧 Enter configuration values. Press Enter to keep default.")

	config.URL = promptString("URL to test", config.URL)
	config.Method = promptString("HTTP Method (GET or POST)", config.Method)
	if config.Method == "POST" {
		config.RequestBody = promptString("Request body (leave empty if not needed)", config.RequestBody)
	}

	config.N = promptInt("Number of requests", config.N)
	config.Threads = promptInt("Number of concurrent threads", config.Threads)
	config.RetryLimit = promptInt("Retry limit for failed requests", config.RetryLimit)
	config.ThresholdTime = promptFloat("Response time threshold in seconds", config.ThresholdTime)
	config.ThresholdSuccess = promptFloat("Success rate threshold in percentage", config.ThresholdSuccess)
	config.CurlMaxTime = promptInt("Curl max-time in seconds", config.CurlMaxTime)

	config.ResponsePattern = promptString("Response validation pattern (regex, leave empty if not needed)", config.ResponsePattern)

	config.BearerToken = promptString("Bearer Token (leave empty if not needed)", config.BearerToken)
	headersInput := promptString("Custom headers (key1:value1,key2:value2)", "")
	if headersInput != "" {
		headerPairs := strings.Split(headersInput, ",")
		for _, pair := range headerPairs {
			kv := strings.SplitN(pair, ":", 2)
			if len(kv) == 2 {
				config.Headers[strings.TrimSpace(kv[0])] = strings.TrimSpace(kv[1])
			}
		}
	}

	config.EmailEnabled = promptBool("Enable email notifications", config.EmailEnabled)
	if config.EmailEnabled {
		config.EmailTo = promptString("Email address for notifications", config.EmailTo)
	}

	config.LogFile = promptString("Log file path", config.LogFile)
	config.DisableLogging = promptBool("Disable logging", config.DisableLogging)
	config.JSONOutput = promptBool("Enable JSON output", config.JSONOutput)

	return config, nil
}

func SetupLogging(config Config) error {
	if config.DisableLogging {
		log.SetOutput(io.Discard)
		return nil
	}

	writers := []io.Writer{os.Stdout}

	if config.LogFile != "" {
		file, err := os.OpenFile(config.LogFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			return fmt.Errorf("failed to open log file: %v", err)
		}
		writers = append(writers, file)
	}

	log.SetOutput(io.MultiWriter(writers...))
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile)
	return nil
}

func Cleanup() {
	log.Println("Performing cleanup...")
	// Add cleanup logic here
	log.Println("Cleanup completed.")
}

func getEnvAsInt(name string, defaultVal int) int {
	valueStr := os.Getenv(name)
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultVal
}

func getEnvAsFloat(name string, defaultVal float64) float64 {
	valueStr := os.Getenv(name)
	if value, err := strconv.ParseFloat(valueStr, 64); err == nil {
		return value
	}
	return defaultVal
}

func getEnvAsBool(name string, defaultVal bool) bool {
	valueStr := os.Getenv(name)
	if value, err := strconv.ParseBool(valueStr); err == nil {
		return value
	}
	return defaultVal
}

func promptString(prompt string, defaultValue string) string {
	fmt.Printf("%s (default: %s): ", prompt, defaultValue)
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)
	if input == "" {
		return defaultValue
	}
	return input
}

func promptInt(prompt string, defaultValue int) int {
	input := promptString(prompt, strconv.Itoa(defaultValue))
	value, err := strconv.Atoi(input)
	if err != nil {
		return defaultValue
	}
	return value
}

func promptFloat(prompt string, defaultValue float64) float64 {
	input := promptString(prompt, fmt.Sprintf("%.2f", defaultValue))
	value, err := strconv.ParseFloat(input, 64)
	if err != nil {
		return defaultValue
	}
	return value
}

func promptBool(prompt string, defaultValue bool) bool {
	input := promptString(prompt, fmt.Sprintf("%v", defaultValue))
	value, err := strconv.ParseBool(input)
	if err != nil {
		return defaultValue
	}
	return value
}
