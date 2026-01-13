# Load Balancing Algorithms Analysis

[![Go Version](https://img.shields.io/badge/Go-1.21.5-blue.svg)](https://golang.org/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

This project provides a theoretical and experimental analysis of classical load balancing algorithms under realistic dynamic conditions such as heterogeneous server capacities, traffic bursts, and server failures.

## Table of Contents

- [About](#about)
- [Features](#features)
- [Supported Algorithms](#supported-algorithms)
- [Architecture](#architecture)
- [Installation](#installation)
- [Usage](#usage)
- [Performance Metrics](#performance-metrics)
- [Project Structure](#project-structure)
- [Contributing](#contributing)

## About

Load balancing is a fundamental mechanism for distributing client requests across multiple backend servers to improve system performance, availability, and scalability. While classical load balancing algorithms such as Round Robin, Weighted Round Robin, and Least Connections perform well under ideal conditions, their performance often degrades when servers have heterogeneous capacities, traffic is bursty, or network latency fluctuates.

This project aims to fill the gap in existing literature by systematically comparing the behavior of classical algorithms under realistic dynamic conditions.

### Problem Definition

Most classical load balancing algorithms rely on static assumptions about backend performance, which can lead to:

- Load imbalance when server capabilities vary (e.g., CPU/memory differences)
- High tail latency (p95/p99) under sudden traffic bursts
- Unstable or unfair assignment when network delay fluctuates
- Poor failover behavior when servers become unavailable

## Features

- **Three Classical Algorithms**: Round Robin, Weighted Round Robin, and Least Connections
- **Comprehensive Metrics**: Latency tracking with p50, p90, p95, p99 percentiles
- **Docker Support**: Docker Compose for easy setup and execution
- **Visualization**: Python-based latency distribution analysis and CDF plots
- **Heterogeneous Servers**: Realistic test environment with different CPU capacities and memory limits
- **Experimental Scenarios**: Multiple test scenarios with different traffic patterns
- **High-Performance Logging**: Async, non-blocking log system optimized for high throughput
- **Connection Pooling**: Optimized HTTP transport with connection reuse

## Supported Algorithms

### 1. Round Robin (RR)
A simple and fair distribution algorithm that selects servers in rotation. Sends requests to each server with equal frequency.

### 2. Weighted Round Robin (WRR)
A weighted version of Round Robin based on server capacities. Server weights are set proportional to their processing capacities.

### 3. Least Connections (LC)
Selects the server with the fewest active connections. Provides dynamic load distribution.


## Architecture

```
┌─────────────┐
│   Client    │
└──────┬──────┘
       │
       ▼
┌─────────────────┐
│  Load Balancer  │ ◄─── Algorithm Selection
│    (Port 8000)  │      (RR/WRR/LC)
│                 │      via ALGORITHM env var
└──────┬──────────┘
       │
       ├──────────┬──────────┬──────────┐
       ▼          ▼          ▼          ▼
┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐
│Backend 1 │ │Backend 2 │ │Backend 3 │ │Backend 4 │
│CPU: 0.5  │ │CPU: 1.0  │ │CPU: 0.3  │ │CPU: 2.0  │
│Mem: 256M │ │Mem: 512M │ │Mem: 128M │ │Mem: 1G   │
└──────────┘ └──────────┘ └──────────┘ └──────────┘
```

## Installation

### Requirements

- [Go](https://golang.org/dl/) 1.21.5 or higher
- [Docker](https://www.docker.com/get-started) and Docker Compose
- [Python](https://www.python.org/downloads/) 3.7+ (for analysis and visualization)

### Steps

1. **Clone the repository:**
```bash
git clone https://github.com/Samet-Hodaman/load-balancing-analysis.git
cd load-balancing-analysis
```

2. **Start the system with Docker Compose:**
```bash
# Default algorithm (Round Robin)
docker-compose up --build

# Veya belirli bir algoritma ile başlatmak için:
ALGORITHM=weighted_round_robin docker-compose up --build
ALGORITHM=least_connections docker-compose up --build
```

Bu komut şunları başlatır:
- Load balancer (port 8000)
- 4 backend server (farklı CPU kapasiteleri ve bellek limitleri ile)

## Usage

### Basic Test

After the load balancer is running, send requests:

```bash
# Simple request
curl http://localhost:8000/

# Load test (example: 1000 requests)
for i in {1..1000}; do curl http://localhost:8000/ & done; wait
```

### Algoritma Değiştirme

Algoritmayı değiştirmek için `ALGORITHM` environment variable'ını kullanın:

```bash
# Round Robin (varsayılan)
ALGORITHM=round_robin docker-compose up

# Weighted Round Robin
ALGORITHM=weighted_round_robin docker-compose up

# Least Connections
ALGORITHM=least_connections docker-compose up
```

Desteklenen algoritma değerleri:
- `round_robin` (varsayılan)
- `weighted_round_robin`
- `least_connections`

### Metrik Toplama

Load balancer her istek için latency bilgilerini log dosyasına yazar:
- Log dosyası: `logs/load_balancer.log`
- Her backend server'ın kendi log dosyası: `logs/backend1.log`, `logs/backend2.log`, vb.

Log formatı:
```
2026/01/13 21:25:24 server=backend1:8080 latency=465.542ms
```

Test sonuçları için `experiments/` klasöründeki senaryoları kullanabilirsiniz. Her senaryo için JSON formatında sonuçlar ve HTML grafikler mevcuttur.

### Görselleştirme

Test sonuçlarını görselleştirmek için `experiments/` klasöründeki hazır grafikleri kullanabilirsiniz. Her senaryo için HTML formatında interaktif grafikler mevcuttur.

**Manuel grafik oluşturma:**

Eğer Vegeta gibi bir load testing tool kullanarak JSON formatında sonuçlar topladıysanız:

```bash
# JSON dosyasından grafik oluştur
python3 load-balancer/plot_latency.py results.json round_robin
```

Not: `plot_latency.py` scripti JSON formatında latency verilerini bekler. Vegeta gibi araçlarla toplanan sonuçlar JSON formatında olmalıdır.

## Performance Metrics

Proje aşağıdaki performans metriklerini ölçer:

- **Average Latency**: Tüm isteklerin ortalama gecikme süresi
- **p50 (Median)**: 50. yüzdelik dilim latency
- **p90**: 90. yüzdelik dilim latency
- **p95**: 95. yüzdelik dilim latency (tail latency)
- **p99**: 99. yüzdelik dilim latency (extreme tail latency)
- **CDF (Cumulative Distribution Function)**: Latency dağılımının kümülatif dağılım fonksiyonu
- **Throughput**: Saniye başına işlenen istek sayısı

### Test Senaryoları

Proje üç farklı test senaryosu içerir (`experiments/` klasöründe):

1. **Scenario 1**: Heterojen server kapasiteleri ile temel performans testi
2. **Scenario 2**: Trafik patlaması senaryosu
3. **Scenario 3**: Server hata senaryosu (failover davranışı)

Her senaryo için Round Robin, Weighted Round Robin ve Least Connections algoritmalarının sonuçları mevcuttur.

## Project Structure

```
load-balancing-analysis/
├── backend/                    # Backend server implementation
│   ├── Dockerfile
│   └── main.go                 # HTTP backend server with logging
├── load-balancer/              # Load balancer implementation
│   ├── Dockerfile
│   ├── main.go                 # Load balancer with algorithm selection
│   └── plot_latency.py        # Latency visualization script (JSON input)
├── utils/                      # Helper functions and models
│   ├── latency.go             # Latency calculation functions
│   ├── models.go              # Data models and interfaces
│   ├── round_robin.go         # Round Robin implementation
│   ├── weighted_round_robin.go # Weighted Round Robin implementation
│   └── least_connections.go   # Least Connections implementation
├── experiments/                # Test scenarios and results
│   ├── scenario-1/            # Heterogeneous server capacities
│   ├── scenario-2/            # Traffic burst scenario
│   └── scenario-3/            # Server failure scenario
├── logs/                       # Log files directory
├── results/                    # Test results directory
├── docker-compose.yml         # Docker Compose configuration
├── go.mod                     # Go module definitions
├── plot.sh                    # Simple script to generate plots
├── system_design.png          # System architecture diagram
└── README.md                  # This file
```

## Experimental Results

Proje aşağıdaki soruları yanıtlamayı amaçlar:

1. Heterojen server kapasiteleri altında hangi algoritma daha iyi yük dağılımı sağlar?
2. Trafik patlamaları sırasında hangi algoritma daha düşük tail latency gösterir?
3. Server hataları durumunda hangi algoritma daha iyi failover sağlar?
4. Hangi algoritmalar network latency değişkenliğinden daha çok etkilenir?

Detaylı sonuçlar için `experiments/` klasöründeki senaryo sonuçlarına bakabilirsiniz. Her senaryo için:
- JSON formatında ham sonuçlar (`results.json`)
- HTML formatında interaktif grafikler (`latency_plot.html`)

## Technical Details

### Algoritma Özellikleri

- **Round Robin**: Thread-safe atomic operasyonlar kullanarak yüksek throughput sağlar
- **Weighted Round Robin**: GCD tabanlı lock-free implementasyon, server ağırlıklarına göre dağıtım yapar
- **Least Connections**: Aktif bağlantı sayısına göre dinamik server seçimi

### Performans Optimizasyonları

- **Connection Pooling**: HTTP transport connection reuse ile optimize edilmiş
- **Async Logging**: Non-blocking, batch processing ile yüksek throughput için optimize edilmiş log sistemi
- **Pre-created Proxies**: Her backend için önceden oluşturulmuş reverse proxy'ler
- **Graceful Shutdown**: Güvenli kapanma mekanizması

## References

Bu proje aşağıdaki akademik çalışmalardan ilham almıştır:

- Mitzenmacher (Randomized Load Balancing)
- Cardellini et al. (Dynamic Load Balancing Overview)
- Wierman et al. (Fairness and Performance Analysis)
- Gandhi et al. (Heterogeneous Server Environments)
- Google Maglev
- Facebook Katran
- Xu & Li (Latency-aware Load Balancing)

## Contributing

Contributions are welcome! Please:

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## Author

**CENG505 - Advanced Networking Project**
