package main

import (
	"encoding/csv"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

type Backend struct {
	URL   *url.URL
	Conns int64
}

var (
	backends []*Backend
	mu       sync.Mutex

	latenciesMu sync.Mutex
	latencies   []time.Duration
)

func init() {
	urls := []string{
		"http://backend1:8080",
		"http://backend2:8080",
		"http://backend3:8080",
		"http://backend4:8080",
	}

	for _, u := range urls {
		parsed, _ := url.Parse(u)
		backends = append(backends, &Backend{URL: parsed})
	}
}

var rrIdx uint64

func selectRoundRobin() *Backend {
	i := atomic.AddUint64(&rrIdx, 1)
	return backends[i%uint64(len(backends))]
}

func selectLeastConn() *Backend {
	mu.Lock()
	defer mu.Unlock()

	var chosen *Backend
	min := int64(^uint64(0) >> 1)

	for _, b := range backends {
		c := atomic.LoadInt64(&b.Conns)
		if c < min {
			min = c
			chosen = b
		}
	}
	return chosen
}

func handler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	b := selectRoundRobin()
	atomic.AddInt64(&b.Conns, 1)
	defer atomic.AddInt64(&b.Conns, -1)

	proxy := httputil.NewSingleHostReverseProxy(b.URL)
	proxy.ServeHTTP(w, r)

	elapsed := time.Since(start)

	latenciesMu.Lock()
	latencies = append(latencies, elapsed)
	latenciesMu.Unlock()
}

func writeCSV() {
	file, err := os.Create("latency.csv")
	if err != nil {
		log.Println("CSV create error:", err)
		return
	}
	defer file.Close()

	w := csv.NewWriter(file)
	defer w.Flush()

	w.Write([]string{"latency_ms"})

	for _, l := range latencies {
		w.Write([]string{
			formatMs(l),
		})
	}

	log.Println("Latency CSV written")
}

func formatMs(d time.Duration) string {
	// Duration'ı direkt string olarak yaz, böylece nanosaniye hassasiyetinde kalır
	// Go otomatik olarak en uygun birimi seçer (ns, µs, ms, s)
	return d.String()
}

func percentiles() {
	latenciesMu.Lock()
	defer latenciesMu.Unlock()

	sort.Slice(latencies, func(i, j int) bool {
		return latencies[i] < latencies[j]
	})

	n := len(latencies)
	if n == 0 {
		return
	}

	p50 := latencies[n*50/100]
	p95 := latencies[n*95/100]
	p99 := latencies[n*99/100]

	log.Printf("p50=%v p95=%v p99=%v total=%d\n",
		p50, p95, p99, n)
}

func main() {
	http.HandleFunc("/", handler)

	go func() {
		<-time.After(40 * time.Second)
		percentiles()
		writeCSV()
	}()

	log.Println("Load balancer (LC + metrics) running on :8000")
	log.Fatal(http.ListenAndServe(":8000", nil))
}
