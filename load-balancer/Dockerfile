FROM golang:1.22-alpine

WORKDIR /app
COPY go.mod .
COPY utils/ ./utils/
COPY load-balancer/main.go ./load-balancer/
RUN go build -o lb ./load-balancer/main.go

# Create results directory with proper permissions
RUN mkdir -p /app/results && chmod 755 /app/results

EXPOSE 8000
CMD ["./lb"]