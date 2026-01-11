FROM golang:1.22-alpine

WORKDIR /app
COPY go.mod .
COPY main.go .
RUN go build -o backend

EXPOSE 8080
CMD ["./backend"]
