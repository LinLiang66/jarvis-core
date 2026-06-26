FROM golang:1.23-alpine AS builder
WORKDIR /src
COPY scheduler/go.mod scheduler/go.sum ./
RUN go mod download
COPY scheduler/ ./
RUN CGO_ENABLED=0 go build -o /scheduler ./cmd/server

FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=builder /scheduler /app/scheduler
EXPOSE 9000
ENTRYPOINT ["/app/scheduler"]
