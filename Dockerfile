# ── Build stage ────────────────────────────────────────────
FROM node:22-alpine AS frontend
WORKDIR /web
COPY web/package.json web/package-lock.json ./
RUN npm ci
COPY web/ .
RUN npm run build

FROM golang:1.26-alpine AS builder

ENV GOTOOLCHAIN=auto

RUN apk add --no-cache gcc musl-dev

WORKDIR /src
COPY go.mod go.sum vendor/ ./
COPY . .
RUN go build -mod=vendor -tags sqlite_fts5 -o /build/server ./cmd/server/ && \
    go build -mod=vendor -o /build/admin ./cmd/admin/

# ── Runtime stage ──────────────────────────────────────────
FROM alpine:3.21

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=builder /build/server ./server
COPY --from=builder /build/admin ./admin
COPY --from=frontend /web/dist ./web/dist

EXPOSE 8080

CMD ["./server"]
