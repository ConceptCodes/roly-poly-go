FROM golang:1.25-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/server ./main.go
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/migrate github.com/pressly/goose/v3/cmd/goose

FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=builder /app/server /server
COPY --from=builder /app/migrate /migrate
COPY --from=builder /app/infra/migrations /migrations

EXPOSE 8080

ENTRYPOINT ["/server"]
