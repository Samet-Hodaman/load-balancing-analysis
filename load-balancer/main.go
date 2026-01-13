package main

import (
	"bufio"
	"context"
	"fmt"
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

	logFile   *os.File
	logWriter *bufio.Writer
	logger    *log.Logger
	logChan   chan string
	logWg     sync.WaitGroup

	// Shared HTTP transport with connection pooling
	sharedTransport *http.Transport
	// Pre-created proxies for each backend
	proxies map[string]*httputil.ReverseProxy
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

	// Using larger buffered writer for better performance
	logWriter = bufio.NewWriterSize(logFile, 128*1024) // 128KB buffer
	logger = log.New(logWriter, "", log.LstdFlags)

	// Create smaller buffered channel to minimize memory usage
	// Buffer size: 5000 entries - if full, drop immediately (non-blocking)
	logChan = make(chan string, 5000)

	// Start async log writer goroutine with batch processing
	// This goroutine runs continuously and processes logs from channel
	// When channel has space, it automatically accepts new log entries
	logWg.Add(1)
	go func() {
		defer logWg.Done()
		ticker := time.NewTicker(50 * time.Millisecond) // Flush every 50ms
		defer ticker.Stop()

		batch := make([]string, 0, 100) // Batch buffer
		batchTimeout := time.NewTicker(10 * time.Millisecond)
		defer batchTimeout.Stop()

		// Continuously process logs - automatically resumes when load decreases
		for {
			select {
			case msg, ok := <-logChan:
				if !ok {
					// Channel closed, write remaining batch and flush
					if len(batch) > 0 && logWriter != nil {
						for _, m := range batch {
							logWriter.WriteString(m)
							logWriter.WriteString("\n")
						}
						logWriter.Flush()
					}
					return
				}
				// Add to batch - channel has space, logging resumes automatically
				batch = append(batch, msg)
				// Write batch if it reaches capacity
				if len(batch) >= 100 {
					if logWriter != nil {
						for _, m := range batch {
							logWriter.WriteString(m)
							logWriter.WriteString("\n")
						}
						logWriter.Flush()
					}
					batch = batch[:0] // Reset batch
				}
			case <-batchTimeout.C:
				// Write batch periodically even if not full
				// This ensures logs are written even during low load
				if len(batch) > 0 && logWriter != nil {
					for _, m := range batch {
						logWriter.WriteString(m)
						logWriter.WriteString("\n")
					}
					logWriter.Flush()
					batch = batch[:0] // Reset batch
				}
			case <-ticker.C:
				// Periodic flush (extra safety)
				if logWriter != nil {
					logWriter.Flush()
				}
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

	// Initialize shared HTTP transport with connection pooling
	sharedTransport = &http.Transport{
		MaxIdleConns:          1000,             // Maximum idle connections
		MaxIdleConnsPerHost:   250,              // Maximum idle connections per backend
		MaxConnsPerHost:       500,              // Maximum total connections per backend
		IdleConnTimeout:       90 * time.Second, // Timeout for idle connections
		DisableKeepAlives:     false,            // Enable connection reuse
		ResponseHeaderTimeout: 10 * time.Second, // Timeout for response headers
		ExpectContinueTimeout: 1 * time.Second,
	}

	// Initialize proxies map
	proxies = make(map[string]*httputil.ReverseProxy)

	for i, u := range urls {
		parsed, _ := url.Parse(u)
		backends = append(backends, &utils.Server{
			URL:        parsed,
			Weight:     weights[i],
			CurrentWRR: weights[i],
		})

		// Create and configure reverse proxy for this backend
		proxy := httputil.NewSingleHostReverseProxy(parsed)
		proxy.Transport = sharedTransport
		proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
			log.Printf("Proxy error for %s: %v", parsed.Host, err)
			http.Error(w, "Service unavailable", http.StatusServiceUnavailable)
		}
		proxies[parsed.Host] = proxy
	}

	// Algorithm selection from environment variable
	alg := strings.ToLower(os.Getenv("ALGORITHM"))
	lb = GetLoadBalancerAlgorithm(alg)

	log.Printf("Load balancing algorithm: %s", lb.Name())
	log.Printf("HTTP Transport configured: MaxIdleConns=1000, MaxIdleConnsPerHost=250, MaxConnsPerHost=500")

	// Initialize file logger
	if err := initLogger(); err != nil {
		log.Printf("Warning: Failed to initialize file logger: %v", err)
		log.Printf("Logging will continue to stdout only")
	}
}

func handler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	srv := lb.SelectServer(backends)
	if srv == nil {
		http.Error(w, "No backend available", http.StatusServiceUnavailable)
		return
	}

	atomic.AddInt64(&srv.Conns, 1)
	defer atomic.AddInt64(&srv.Conns, -1)

	// Use pre-created proxy instead of creating a new one
	proxy := proxies[srv.URL.Host]
	if proxy == nil {
		// Fallback: create proxy if not found (shouldn't happen)
		proxy = httputil.NewSingleHostReverseProxy(srv.URL)
		proxy.Transport = sharedTransport
		proxies[srv.URL.Host] = proxy
	}

	// Add timeout context to the request
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	r = r.WithContext(ctx)

	proxy.ServeHTTP(w, r)

	elapsed := time.Since(start)

	// Ultra-fast async logging: non-blocking, drop if queue full
	// Handler NEVER blocks - requests are always served
	// When load decreases, logging automatically resumes
	if logChan != nil {
		// Fast path: try to send without blocking
		// This ensures handler continues serving requests even under extreme load
		select {
		case logChan <- fmt.Sprintf("%s server=%s latency=%v",
			time.Now().Format("2006/01/02 15:04:05"),
			srv.URL.Host,
			elapsed):
			// Successfully queued - will be written by background goroutine
		default:
			// Channel full - drop immediately to avoid blocking handler
			// Handler continues serving requests without delay
			// When load decreases and channel has space, logging automatically resumes
		}
	}
}

func main() {
	http.HandleFunc("/", handler)

	// Setup graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	server := &http.Server{
		Addr:           ":8000",
		Handler:        nil,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		IdleTimeout:    120 * time.Second,
		MaxHeaderBytes: 1 << 20, // 1MB
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

	// Close log channel and wait for writer to finish
	if logChan != nil {
		close(logChan)
		logWg.Wait() // Wait for all pending logs to be written
	}

	// Final flush and close log file
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
