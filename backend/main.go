package main

import (
	"bufio"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync/atomic"
	"time"
)

var reqCount uint64
var logger *log.Logger
var logFile *os.File
var logWriter *bufio.Writer

func initLogger(id string) error {
	logDir := "/app/logs"
	// Create logs directory if it doesn't exist
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return err
	}

	logFileName := fmt.Sprintf("%s/%s.log", logDir, id)
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
			logWriter.Flush()
		}
	}()

	return nil
}

func main() {
	id := os.Getenv("SERVER_ID")

	// Initialize logger
	if err := initLogger(id); err != nil {
		log.Fatalf("Failed to create log file: %v", err)
	}
	defer logFile.Close()
	defer logWriter.Flush()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		count := atomic.AddUint64(&reqCount, 1)
		elapsed := time.Since(start)

		logger.Printf("req=%d elapsed=%v", count, elapsed)
		fmt.Fprintf(w, "Hello from %s | req=%d | latency=%v\n", id, count, elapsed)
	})

	log.Fatal(http.ListenAndServe(":8080", nil))
}
