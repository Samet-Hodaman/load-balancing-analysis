package main

import (
	"bufio"
	"context"
	"load-balancing-analysis/utils"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"strings"
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

	logFile   *os.File
	logWriter *bufio.Writer
	logger    *log.Logger
)

func GetLoadBalancerAlgorithm(alg string) utils.LoadBalancer {
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

func initLogger() error {
	logDir := "/app/logs"
	// Create logs directory if it doesn't exist
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return err
	}

	logFileName := logDir + "/load_balancer.log"
	var err error
	logFile, err = os.OpenFile(logFileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return err
	}

	// Using buffered writer to improve performance
	logWriter = bufio.NewWriterSize(logFile, 4096)
	logger = log.New(logWriter, "", log.LstdFlags)

	// Starting a goroutine to periodically flush the buffer
	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			if logWriter != nil {
				logWriter.Flush()
			}
		}
	}()

	return nil
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
	lb = GetLoadBalancerAlgorithm(alg)

	log.Printf("Load balancing algorithm: %s", lb.Name())

	// Initialize file logger
	if err := initLogger(); err != nil {
		log.Printf("Warning: Failed to initialize file logger: %v", err)
		log.Printf("Logging will continue to stdout only")
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

	// Write to log file (buffered, low overhead)
	if logger != nil {
		logger.Printf("server=%s latency=%v", srv.URL.Host, elapsed)
	}
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

	// Flush and close log file
	if logWriter != nil {
		logWriter.Flush()
	}
	if logFile != nil {
		logFile.Close()
	}

	// Shutdown HTTP server
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Server shutdown error: %v", err)
	}

	log.Println("Server stopped")
}
