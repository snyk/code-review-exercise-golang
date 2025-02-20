# syntax=docker/dockerfile:1
ARG GO_VERSION=1.24.0

###############
# Build stage #
###############
FROM golang:${GO_VERSION} AS builder
ARG APP

# Set working directory
WORKDIR /go/src/${APP}

# Download and cache dependencies in a dedicated layer.
COPY go.mod go.sum ./
RUN go mod download

# Add source code & build
COPY . .
RUN --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 go build -v -o /go/bin/app ./cmd/${APP}

#################
# Runtime stage #
#################
FROM gcr.io/distroless/static-debian12

COPY --from=builder /go/bin/app .
COPY config.json .

EXPOSE 8080
CMD ["/app"]
