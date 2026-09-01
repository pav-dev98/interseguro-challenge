FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /go-api .

FROM scratch

COPY --from=builder /go-api /go-api

EXPOSE 3001

ENTRYPOINT ["/go-api"]
