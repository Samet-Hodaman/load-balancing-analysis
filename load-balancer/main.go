package main

import (
	"context"
	"encoding/csv"
	"load-balancing-analysis/utils"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
)

const (
	RoundRobin         string = "round_robin"
	WeightedRoundRobin string = "weighted_round_robin"
	LeastConnections   string = "least_connections"
)

var (
	lb       utils.LoadBalancer
	backends []*utils.Server

	csvFile   *os.File
	csvWriter *csv.Writer
	csvMu     sync.Mutex
)

func GetLoadBalancer(alg string) utils.LoadBalancer {
	switch alg {
	case RoundRobin:
		return utils.NewRoundRobin()
	case WeightedRoundRobin:
		return utils.NewWeightedRoundRobin()
	case LeastConnections:
		return utils.NewLeastConnections()
	default:
		return utils.NewRoundRobin()
	}
}

func init() {
	urls := []string{
		"http://backend1:8080",
		"http://backend2:8080",
		"http://backend3:8080",
		"http://backend4:8080",
	}

	// Default weights for Weighted Round Robin
	weights := []int64{5, 10, 3, 2}

	for i, u := range urls {
		parsed, _ := url.Parse(u)
		backends = append(backends, &utils.Server{
			URL:        parsed,
			Weight:     weights[i],
			CurrentWRR: weights[i],
		})
	}

	// Algorithm selection from environment variable
	alg := strings.ToLower(os.Getenv("ALGORITHM"))
	lb = GetLoadBalancer(alg)

	log.Printf("Load balancing algorithm: %s", lb.Name())

	// Get current working directory for debugging
	wd, err := os.Getwd()
	if err != nil {
		log.Printf("Warning: Could not get working directory: %v", err)
	} else {
		log.Printf("Current working directory: %s", wd)
	}

	// Initialize CSV file for real-time logging
	resultsDir := "results"
	if err := os.MkdirAll(resultsDir, 0755); err != nil {
		log.Printf("Error creating results directory '%s': %v", resultsDir, err)
		log.Printf("CSV logging will be disabled")
	} else {
		log.Printf("Results directory '%s' created/verified successfully", resultsDir)

		csvPath := "results/latency.csv"
		var err error
		csvFile, err = os.Create(csvPath)
		if err != nil {
			log.Printf("Error: Could not create CSV file '%s': %v", csvPath, err)
			log.Printf("CSV logging will be disabled")
		} else {
			log.Printf("CSV file created successfully: %s", csvPath)
			csvWriter = csv.NewWriter(csvFile)
			if err := csvWriter.Write([]string{"latency_ms"}); err != nil {
				log.Printf("Error writing CSV header: %v", err)
			} else {
				csvWriter.Flush()
				if err := csvWriter.Error(); err != nil {
					log.Printf("Error flushing CSV header: %v", err)
				} else {
					log.Printf("CSV file initialized successfully at %s", csvPath)
				}
			}
		}
	}
}

func handler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	srv := lb.SelectServer(backends)
	atomic.AddInt64(&srv.Conns, 1)
	defer atomic.AddInt64(&srv.Conns, -1)

	proxy := httputil.NewSingleHostReverseProxy(srv.URL)
	proxy.ServeHTTP(w, r)

	elapsed := time.Since(start)

	// Write to CSV in real-time
	csvMu.Lock()
	if csvWriter != nil && csvFile != nil {
		if err := csvWriter.Write([]string{formatMs(elapsed)}); err != nil {
			log.Printf("Error writing to CSV: %v", err)
		} else {
			csvWriter.Flush()
			if err := csvWriter.Error(); err != nil {
				log.Printf("Error flushing CSV: %v", err)
			}
		}
	}
	// Removed warning log to reduce noise - CSV may not be initialized intentionally
	csvMu.Unlock()
}

func writeCSV() {
	csvMu.Lock()
	defer csvMu.Unlock()

	if csvWriter != nil {
		csvWriter.Flush()
		if err := csvWriter.Error(); err != nil {
			log.Printf("Error flushing CSV during shutdown: %v", err)
		} else {
			log.Println("CSV file flushed successfully")
		}
	} else {
		log.Println("Warning: csvWriter is nil, nothing to flush")
	}
}

func closeCSV() {
	csvMu.Lock()
	defer csvMu.Unlock()

	if csvWriter != nil {
		csvWriter.Flush()
		if err := csvWriter.Error(); err != nil {
			log.Printf("Error flushing CSV before close: %v", err)
		}
		csvWriter = nil
	}
	if csvFile != nil {
		if err := csvFile.Close(); err != nil {
			log.Printf("Error closing CSV file: %v", err)
		} else {
			log.Println("CSV file closed successfully")
		}
		csvFile = nil
	}
}

func formatMs(d time.Duration) string {
	// Write Duration directly as string to maintain nanosecond precision
	// Go automatically selects the most appropriate unit (ns, µs, ms, s)
	return d.String()
}

func main() {
	http.HandleFunc("/", handler)

	// Setup graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	server := &http.Server{
		Addr:    ":8000",
		Handler: nil,
	}

	// Start HTTP server in a goroutine
	go func() {
		log.Printf("Load balancer (%s + metrics) running on :8000", lb.Name())
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	// Wait for interrupt signal
	<-sigChan
	log.Println("Shutting down gracefully...")

	// Flush and close CSV file
	writeCSV()
	closeCSV()

	// Shutdown HTTP server
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Server shutdown error: %v", err)
	}

	log.Println("Server stopped")
}
