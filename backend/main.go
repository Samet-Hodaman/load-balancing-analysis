package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"sync/atomic"
	"time"
)

var reqCount uint64

func main() {
	id := os.Getenv("SERVER_ID")
	delayMs := os.Getenv("DELAY_MS")

	delay := 0
	if delayMs != "" {
		fmt.Sscanf(delayMs, "%d", &delay)
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		if delay > 0 {
			time.Sleep(time.Duration(delay) * time.Millisecond)
		}

		count := atomic.AddUint64(&reqCount, 1)
		elapsed := time.Since(start)

		log.Printf("req=%d elapsed=%v", count, elapsed)
		fmt.Fprintf(w, "Hello from %s | req=%d | latency=%v\n", id, count, elapsed)
	})

	log.Printf("Backend %s running (delay=%dms)", id, delay)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
