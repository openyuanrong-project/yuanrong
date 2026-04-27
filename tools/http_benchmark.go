// HTTP Benchmark Tool with TLS Mutual Authentication Support
//
// Compilation:
//
//	go build -o http_benchmark http_benchmark.go
//
// Usage Examples:
//
//	# Basic HTTP benchmark with request count
//	./http_benchmark -url http://localhost:8080 -c 10 -n 100
//
//	# Benchmark with duration (30 seconds)
//	./http_benchmark -url http://localhost:8080 -c 10 -d 30s
//
//	# HTTPS benchmark with TLS mutual authentication
//	./http_benchmark -url https://example.com -c 20 -n 1000 \
//	  -cert /path/to/client.crt -key /path/to/client.key \
//	  -ca /path/to/ca.crt -server-name example.com
//
//	# Using environment variables for certificates
//	export YR_CERT_FILE=/path/to/client.crt
//	export YR_PRIVATE_KEY_FILE=/path/to/client.key
//	export YR_VERIFY_FILE=/path/to/ca.crt
//	export YR_SERVER_NAME=example.com
//	./http_benchmark -url https://example.com -c 50 -d 1m
//
//	# POST request with custom method
//	./http_benchmark -url https://api.example.com/endpoint -method POST -c 10 -n 100
//
// Parameters:
//
//	-url string         Target URL (default "http://localhost:8080")
//	-method string      HTTP method (default "GET")
//	-c int              Number of concurrent workers (default 10)
//	-n int              Total number of requests (mutually exclusive with -d)
//	-d duration         Test duration (e.g., 30s, 5m, 1h) (mutually exclusive with -n)
//	-body string        Request body (optional)
//	-cert string        Client certificate file path (or set YR_CERT_FILE)
//	-key string         Client private key file path (or set YR_PRIVATE_KEY_FILE)
//	-ca string          CA certificate file path (or set YR_VERIFY_FILE)
//	-server-name string Server name for SNI (or set YR_SERVER_NAME)
//
// Output includes:
//   - Total time, requests, success/failure counts
//   - First error message (if any failures occurred)
//   - Requests per second
//   - Latency statistics: Average, P50, P90, P99, Min, Max
package main

import (
	"crypto/tls"
	"crypto/x509"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"sync"
	"time"
)

type Result struct {
	Duration time.Duration
	Success  bool
	Error    error
}

type Stats struct {
	TotalRequests   int
	SuccessRequests int
	FailedRequests  int
	TotalDuration   time.Duration
	Latencies       []time.Duration
	FirstError      error
	once            sync.Once
}

func (s *Stats) AddResult(r Result) {
	s.TotalRequests++
	if r.Success {
		s.SuccessRequests++
		s.Latencies = append(s.Latencies, r.Duration)
	} else {
		s.FailedRequests++
		if r.Error != nil {
			s.once.Do(func() {
				s.FirstError = r.Error
			})
		}
	}
}

func (s *Stats) Calculate() {
	sort.Slice(s.Latencies, func(i, j int) bool {
		return s.Latencies[i] < s.Latencies[j]
	})
}

func (s *Stats) GetPercentile(p float64) time.Duration {
	if len(s.Latencies) == 0 {
		return 0
	}
	index := int(float64(len(s.Latencies)) * p)
	if index >= len(s.Latencies) {
		index = len(s.Latencies) - 1
	}
	return s.Latencies[index]
}

func (s *Stats) GetAverage() time.Duration {
	if len(s.Latencies) == 0 {
		return 0
	}
	var total time.Duration
	for _, d := range s.Latencies {
		total += d
	}
	return total / time.Duration(len(s.Latencies))
}

func createHTTPClient(clientCert, clientKey, caCert, serverName string) (*http.Client, error) {
	tlsConfig := &tls.Config{}

	// Load client certificate and key if provided
	if clientCert != "" && clientKey != "" {
		cert, err := tls.LoadX509KeyPair(clientCert, clientKey)
		if err != nil {
			return nil, fmt.Errorf("failed to load client certificate: %v", err)
		}
		tlsConfig.Certificates = []tls.Certificate{cert}
	}

	// Load CA certificate if provided
	if caCert != "" {
		caCertPool := x509.NewCertPool()
		caCertBytes, err := os.ReadFile(caCert)
		if err != nil {
			return nil, fmt.Errorf("failed to read CA certificate: %v", err)
		}
		if !caCertPool.AppendCertsFromPEM(caCertBytes) {
			return nil, fmt.Errorf("failed to parse CA certificate")
		}
		tlsConfig.RootCAs = caCertPool
	}

	// Set server name for SNI if provided
	if serverName != "" {
		tlsConfig.ServerName = serverName
	}

	transport := &http.Transport{
		TLSClientConfig: tlsConfig,
	}

	return &http.Client{
		Timeout:   30 * time.Second,
		Transport: transport,
	}, nil
}

func doRequest(url string, method string, body string, client *http.Client) Result {
	start := time.Now()

	var req *http.Request
	var err error

	if body != "" {
		req, err = http.NewRequest(method, url, nil)
	} else {
		req, err = http.NewRequest(method, url, nil)
	}

	if err != nil {
		return Result{
			Duration: time.Since(start),
			Success:  false,
			Error:    err,
		}
	}

	resp, err := client.Do(req)
	if err != nil {
		return Result{
			Duration: time.Since(start),
			Success:  false,
			Error:    err,
		}
	}
	defer resp.Body.Close()

	// Read and discard response body
	_, _ = io.ReadAll(resp.Body)

	duration := time.Since(start)
	success := resp.StatusCode >= 200 && resp.StatusCode < 300

	return Result{
		Duration: duration,
		Success:  success,
		Error:    nil,
	}
}

func worker(id int, url string, method string, body string, client *http.Client, jobs <-chan int, results chan<- Result, wg *sync.WaitGroup) {
	defer wg.Done()

	for range jobs {
		result := doRequest(url, method, body, client)
		results <- result
	}
}

func workerDuration(id int, url string, method string, body string, client *http.Client, stopCh <-chan struct{}, results chan<- Result, wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		select {
		case <-stopCh:
			return
		default:
			result := doRequest(url, method, body, client)
			results <- result
		}
	}
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func main() {
	url := flag.String("url", "http://localhost:8080", "Target URL")
	method := flag.String("method", "GET", "HTTP method")
	concurrency := flag.Int("c", 10, "Number of concurrent workers")
	totalRequests := flag.Int("n", 0, "Total number of requests (mutually exclusive with -d)")
	duration := flag.Duration("d", 0, "Test duration (e.g., 30s, 5m, 1h) (mutually exclusive with -n)")
	body := flag.String("body", "", "Request body")
	clientCert := flag.String("cert", getEnvOrDefault("YR_CERT_FILE", ""), "Client certificate file path")
	clientKey := flag.String("key", getEnvOrDefault("YR_PRIVATE_KEY_FILE", ""), "Client private key file path")
	caCert := flag.String("ca", getEnvOrDefault("YR_VERIFY_FILE", ""), "CA certificate file path")
	serverName := flag.String("server-name", getEnvOrDefault("YR_SERVER_NAME", ""), "Server name for SNI")

	flag.Parse()

	if *url == "" {
		fmt.Println("URL is required")
		return
	}

	// Check mutual exclusivity of -n and -d
	if *totalRequests > 0 && *duration > 0 {
		fmt.Println("Error: -n and -d are mutually exclusive, please specify only one")
		return
	}
	if *totalRequests == 0 && *duration == 0 {
		*totalRequests = 100 // Default to 100 requests
	}

	useRequestCount := *totalRequests > 0

	// Create HTTP client with TLS configuration
	client, err := createHTTPClient(*clientCert, *clientKey, *caCert, *serverName)
	if err != nil {
		fmt.Printf("Failed to create HTTP client: %v\n", err)
		return
	}

	fmt.Printf("Starting benchmark:\n")
	fmt.Printf("  URL: %s\n", *url)
	fmt.Printf("  Method: %s\n", *method)
	fmt.Printf("  Concurrency: %d\n", *concurrency)
	if useRequestCount {
		fmt.Printf("  Total Requests: %d\n", *totalRequests)
	} else {
		fmt.Printf("  Duration: %v\n", *duration)
	}
	if *clientCert != "" {
		fmt.Printf("  Client Certificate: %s\n", *clientCert)
	}
	if *caCert != "" {
		fmt.Printf("  CA Certificate: %s\n", *caCert)
	}
	if *serverName != "" {
		fmt.Printf("  Server Name (SNI): %s\n", *serverName)
	}
	fmt.Println()

	var wg sync.WaitGroup
	var results chan Result
	var startTime time.Time
	var totalTime time.Duration

	if useRequestCount {
		// Request count based benchmark
		jobs := make(chan int, *totalRequests)
		results = make(chan Result, *totalRequests)

		startTime = time.Now()
		for i := 0; i < *concurrency; i++ {
			wg.Add(1)
			go worker(i, *url, *method, *body, client, jobs, results, &wg)
		}

		for i := 0; i < *totalRequests; i++ {
			jobs <- i
		}
		close(jobs)

		wg.Wait()
		close(results)
		totalTime = time.Since(startTime)
	} else {
		// Duration based benchmark
		stopCh := make(chan struct{})
		results = make(chan Result, 10000)

		startTime = time.Now()
		for i := 0; i < *concurrency; i++ {
			wg.Add(1)
			go workerDuration(i, *url, *method, *body, client, stopCh, results, &wg)
		}

		// Wait for duration
		time.Sleep(*duration)
		close(stopCh)

		wg.Wait()
		close(results)
		totalTime = time.Since(startTime)
	}

	// Collect results
	initialCapacity := *totalRequests
	if !useRequestCount {
		initialCapacity = 10000
	}
	stats := &Stats{
		Latencies: make([]time.Duration, 0, initialCapacity),
	}

	for result := range results {
		stats.AddResult(result)
	}

	// Calculate statistics
	stats.Calculate()

	// Print results
	fmt.Printf("Results:\n")
	fmt.Printf("  Total Time: %v\n", totalTime)
	fmt.Printf("  Total Requests: %d\n", stats.TotalRequests)
	fmt.Printf("  Successful: %d\n", stats.SuccessRequests)
	fmt.Printf("  Failed: %d\n", stats.FailedRequests)
	if stats.FirstError != nil {
		fmt.Printf("  First Error: %v\n", stats.FirstError)
	}
	fmt.Printf("  Requests/sec: %.2f\n\n", float64(stats.SuccessRequests)/totalTime.Seconds())

	if len(stats.Latencies) > 0 {
		fmt.Printf("Latency Statistics:\n")
		fmt.Printf("  Average: %v\n", stats.GetAverage())
		fmt.Printf("  P50: %v\n", stats.GetPercentile(0.50))
		fmt.Printf("  P90: %v\n", stats.GetPercentile(0.90))
		fmt.Printf("  P99: %v\n", stats.GetPercentile(0.99))
		fmt.Printf("  Min: %v\n", stats.Latencies[0])
		fmt.Printf("  Max: %v\n", stats.Latencies[len(stats.Latencies)-1])
	}
}
