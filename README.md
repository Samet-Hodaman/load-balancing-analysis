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
- [Experimental Results](#experimental-results)
- [Technical Details](#technical-details)
- [References](#references)
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

# Or start with a specific algorithm:
ALGORITHM=weighted_round_robin docker-compose up --build
ALGORITHM=least_connections docker-compose up --build
```

This command starts:
- Load balancer (port 8000)
- 4 backend servers (with different CPU capacities and memory limits)

## Usage

### Basic Test

After the load balancer is running, send requests:

```bash
# Simple request
curl http://localhost:8000/

# Load test (example: 1000 requests)
for i in {1..1000}; do curl http://localhost:8000/ & done; wait
```

### Changing Algorithms

To change the algorithm, use the `ALGORITHM` environment variable:

```bash
# Round Robin (default)
ALGORITHM=round_robin docker-compose up

# Weighted Round Robin
ALGORITHM=weighted_round_robin docker-compose up

# Least Connections
ALGORITHM=least_connections docker-compose up
```

Supported algorithm values:
- `round_robin` (default)
- `weighted_round_robin`
- `least_connections`

### Collecting Metrics

The load balancer writes latency information to log files for each request:
- Log file: `logs/load_balancer.log`
- Each backend server has its own log file: `logs/backend1.log`, `logs/backend2.log`, etc.

Log format:
```
2026/01/13 21:25:24 server=backend1:8080 latency=465.542ms
```

You can use the scenarios in the `experiments/` folder for test results. Each scenario includes results in JSON format and HTML graphs.

### Visualization

You can use the ready-made graphs in the `experiments/` folder to visualize test results. Each scenario includes interactive graphs in HTML format.

**Manual graph creation:**

If you have collected results in JSON format using a load testing tool like Vegeta:

```bash
# Generate graph from JSON file
python3 load-balancer/plot_latency.py results.json round_robin
```

Note: The `plot_latency.py` script expects latency data in JSON format. Results collected with tools like Vegeta should be in JSON format.

### Load Testing Tools

For comprehensive load testing and analysis, you can use Vegeta:

```bash
# High-rate load test with Vegeta
echo "GET http://localhost:8000/" | vegeta attack -rate=50000 -duration=30s -timeout=5s | tee results.bin | vegeta report
```

This command:
- Sends requests at 50,000 requests/second for 30 seconds
- Saves results to `results.bin`
- Displays a real-time report with latency percentiles and throughput metrics

## Performance Metrics

The project measures the following performance metrics:

- **Average Latency**: Average delay time of all requests
- **p50 (Median)**: 50th percentile latency
- **p90**: 90th percentile latency
- **p95**: 95th percentile latency (tail latency)
- **p99**: 99th percentile latency (extreme tail latency)
- **CDF (Cumulative Distribution Function)**: Cumulative distribution function of latency distribution
- **Throughput**: Number of requests processed per second

### Test Scenarios

The project includes three different test scenarios (in the `experiments/` folder):

1. **Scenario 1**: Basic performance test with heterogeneous server capacities
2. **Scenario 2**: Traffic burst scenario
3. **Scenario 3**: Server failure scenario (failover behavior)

Results for Round Robin, Weighted Round Robin, and Least Connections algorithms are available for each scenario.

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

The project aims to answer the following questions:

1. Which algorithm provides better load distribution under heterogeneous server capacities?
2. Which algorithm shows lower tail latency during traffic bursts?
3. Which algorithm provides better failover in case of server failures?
4. Which algorithms are more affected by network latency variability?

For detailed results, you can check the scenario results in the `experiments/` folder. Each scenario includes:
- Raw results in JSON format (`results.json`)
- Interactive graphs in HTML format (`latency_plot.html`)

## Technical Details

### Algorithm Features

- **Round Robin**: Provides high throughput using thread-safe atomic operations
- **Weighted Round Robin**: GCD-based lock-free implementation that distributes requests according to server weights
- **Least Connections**: Dynamic server selection based on active connection count

### Performance Optimizations

- **Connection Pooling**: Optimized with HTTP transport connection reuse
- **Async Logging**: Non-blocking log system optimized for high throughput with batch processing
- **Pre-created Proxies**: Pre-created reverse proxies for each backend
- **Graceful Shutdown**: Safe shutdown mechanism

## References

This project is inspired by the following academic works:

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
