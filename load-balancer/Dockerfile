FROM golang:1.22-alpine

WORKDIR /app
COPY go.mod .
COPY main.go .
RUN go build -o lb

EXPOSE 8000
CMD ["./lb"]