FROM golang:1.22-alpine

WORKDIR /app
COPY go.mod .
COPY backend/main.go .
RUN go build -o backend

# Create logs directory for log files
RUN mkdir -p /app/logs && chmod 755 /app/logs

EXPOSE 8080
CMD ["./backend"]
