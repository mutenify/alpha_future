FROM golang:1.24-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /api ./cmd/api

FROM alpine:3.21
RUN addgroup -S app && adduser -S -G app app && mkdir -p /data/documents && chown -R app:app /data
COPY --from=build /api /api
USER app
EXPOSE 8080
HEALTHCHECK --interval=5s --timeout=3s --retries=15 CMD wget -q -O /dev/null http://localhost:8080/api/v1/ready || exit 1
ENTRYPOINT ["/api"]
