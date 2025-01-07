FROM golang:1.23

# Set destination for COPY
WORKDIR /app

# Add go module files
COPY go.mod go.sum ./

# Download and cache dependencies in a dedicated layer.
RUN --mount=type=secret,id=gh_token,required=true \
    git config --global url."https://$(cat /run/secrets/gh_token):x-oauth-basic@github.com/snyk".insteadOf "https://github.com/snyk" && \
    go env -w GOPRIVATE=github.com/snyk && \
    go mod download && \
    git config --global --unset url."https://$(cat /run/secrets/gh_token):x-oauth-basic@github.com/snyk".insteadOf

# Add source code & build
COPY . .
RUN --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 go build -v -o /npmjs-deps-fetcher


EXPOSE 8080

# Run
CMD ["/npmjs-deps-fetcher"]