# ────────────────────────────────────────────────────────────
# Addressbook — multi-stage Docker build
#
# Stage 1 builds the frontend (Vite) static assets.
# Stage 2 compiles the Go server and admin CLI.
# Stage 3 assembles a minimal runtime image.
# ────────────────────────────────────────────────────────────

# ── Stage 1: frontend ──────────────────────────────────────
# Debian-based node image (not alpine) so rollup's native binary
# (@rollup/rollup-linux-x64-gnu) matches the glibc lockfile.
FROM node:22 AS frontend

WORKDIR /web

# Install dependencies first to leverage Docker layer caching.
# package.json + lockfile are copied separately so a source-only
# change does not invalidate the npm cache layer.
COPY web/package.json web/package-lock.json ./
RUN npm ci

# Copy the rest of the app and produce the production bundle.
COPY web/ .
RUN npm run build

# ── Stage 2: Go builder ────────────────────────────────────
FROM golang:1.26-alpine AS builder

# Honor go.mod's toolchain directive automatically.
ENV GOTOOLCHAIN=auto

# gcc + musl-dev are required for the cgo SQLite driver (mattn/go-sqlite3).
RUN apk add --no-cache gcc musl-dev

WORKDIR /src

# Cache module downloads separately from source changes.
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# sqlite_fts5 enables full-text search in the SQLite driver.
RUN go build -tags sqlite_fts5 -o /build/server ./cmd/server/ && \
    go build -o /build/admin ./cmd/admin/

# ── Stage 3: runtime ───────────────────────────────────────
FROM alpine:3.21

# ca-certificates for outbound HTTPS (mail/Stripe), tzdata for timezones.
RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

# Static frontend bundle and both binaries.
COPY --from=frontend /web/dist ./web/dist
COPY --from=builder /build/server ./server
COPY --from=builder /build/admin ./admin

EXPOSE 8080

CMD ["./server"]
