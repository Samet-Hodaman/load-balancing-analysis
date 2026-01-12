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

Load balancing is a fundamental mechanism for distributing client requests across multiple backend servers to improve system performance, availability, and scalability. While classical load balancing algorithms such as Round Robin, Least Connections, and Consistent Hashing perform well under ideal conditions, their performance often degrades when servers have heterogeneous capacities, traffic is bursty, or network latency fluctuates.

This project aims to fill the gap in existing literature by systematically comparing the behavior of classical algorithms under realistic dynamic conditions.

### Problem Definition

Most classical load balancing algorithms rely on static assumptions about backend performance, which can lead to:

- Load imbalance when server capabilities vary (e.g., CPU/IO differences)
- High tail latency (p95/p99) under sudden traffic bursts
- Unstable or unfair assignment when network delay fluctuates
- Poor failover behavior, especially in hash-based schemes

## Features

- **Four Classical Algorithms**: Round Robin, Weighted Round Robin, Least Connections, and Consistent Hashing
- **Comprehensive Metrics**: Average latency, p50, p95, p99 percentiles
- **Docker Support**: Docker Compose for easy setup and execution
- **Visualization**: Python-based latency distribution analysis and CDF plots
- **Heterogeneous Servers**: Realistic test environment with different CPU capacities and delay times
- **Experimental Analysis**: Traffic burst and server failure scenarios

## Supported Algorithms

### 1. Round Robin (RR)
A simple and fair distribution algorithm that selects servers in rotation. Sends requests to each server with equal frequency.

### 2. Weighted Round Robin (WRR)
A weighted version of Round Robin based on server capacities. Server weights are set proportional to their processing capacities.

### 3. Least Connections (LC)
Selects the server with the fewest active connections. Provides dynamic load distribution.

### 4. Consistent Hashing
Hash-based server selection. Ensures minimal redistribution when servers are added or removed.

## Architecture

```
┌─────────────┐
│   Client    │
└──────┬──────┘
       │
       ▼
┌─────────────────┐
│  Load Balancer  │ ◄─── Algorithm Selection
│    (Port 8000)  │      (RR/WRR/LC/CH)
└──────┬──────────┘
       │
       ├──────────┬──────────┬──────────┐
       ▼          ▼          ▼          ▼
┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐
│Backend 1 │ │Backend 2 │ │Backend 3 │ │Backend 4 │
│CPU: 0.5  │ │CPU: 1.0  │ │CPU: 0.3  │ │CPU: 2.0  │
│Delay:10ms│ │Delay:30ms│ │Delay:60ms│ │Delay: 0ms│
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
docker-compose up --build
```

This command starts:
- Load balancer (port 8000)
- 4 backend servers (with different capacities and delays)

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

You can change the algorithm by using `selectRoundRobin()` or `selectLeastConn()` functions in the load balancer code.

### Collecting Metrics

After 40 seconds, the load balancer automatically:
- Writes latency data to `latency.csv` file
- Prints p50, p95, p99 percentiles to console

### Visualization

To visualize latency data:

```bash
# Install required Python packages
pip install pandas numpy matplotlib scipy

# Generate plot
python plot_latency.py
```

This command creates the `latency_cdf_lc_vs_rr.png` file.

## Performance Metrics

The project measures the following performance metrics:

- **Average Latency**: Average delay time of all requests
- **p50 (Median)**: 50th percentile latency
- **p95**: 95th percentile latency (tail latency)
- **p99**: 99th percentile latency (extreme tail latency)
- **CDF (Cumulative Distribution Function)**: Cumulative distribution function of latency distribution

### Test Scenarios

1. **Heterogeneous Server Capacities**: Different CPU limits and delay times
2. **Traffic Bursts**: Performance under sudden load increases
3. **Server Failures**: Failover behavior in case of server crashes

## Project Structure

```
load-balancing-analysis/
├── backend/                 # Backend server implementation
│   ├── Dockerfile
│   └── main.go
├── load-balancer/          # Load balancer implementation
│   ├── Dockerfile
│   └── main.go
├── utils/                  # Helper functions and models
│   ├── latency.go         # Latency calculation functions
│   ├── models.go          # Data models and interfaces
│   ├── round_robin.go     # Round Robin implementation
│   └── weighted_round_robin.go  # Weighted Round Robin implementation
├── docker-compose.yml     # Docker Compose configuration
├── go.mod                 # Go module definitions
├── plot_latency.py        # Latency visualization script
└── README.md              # This file
```

## Experimental Results

The project aims to answer the following questions:

1. Which algorithm provides better load distribution under heterogeneous server capacities?
2. Which algorithm shows lower tail latency during traffic bursts?
3. Which algorithm provides better failover in case of server failures?
4. Which algorithms are more affected by network latency variability?

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

## License

This project is licensed under the MIT License.

## Author

**CENG505 - Advanced Networking Project**
