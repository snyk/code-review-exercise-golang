# Snyk code review exercise (Golang) ![CICD Status](https://github.com/snyk/code-review-exercise-golang/actions/workflows/cicd.yml/badge.svg?branch=main)

A web server that provides a basic HTTP API for querying the dependency tree of an [npm](https://npmjs.org) package.

## Prerequisites

- [Golang >= 1.24](https://go.dev/doc/install)
- [GNU Make](https://www.gnu.org/software/make/)
- (Optionally) [Docker](https://docs.docker.com/engine/install/)

## Getting Started

For convenience, a `Makefile` is provided to perform different tasks. For more details, run:

```sh
make help
```

### Running Go program

```sh
make run
```

### Running in a Docker container

```sh
make docker-run
```

### API

The server will now be running on an available port (defaulting to 8080).

The server contains two endpoints
- `/healthcheck`
- `/package/{packageName}/{packageVersion}`

Here is an example that uses `curl` and `jq` to fetch the dependencies for `react@16.13.0`

```sh
curl -s http://localhost:8080/package/react/16.13.0 | jq .
```

## Formatting

The code is formatted using [golangci-lint](https://golangci-lint.run/), you can run this via:

```sh
make fmt
```

## Linting

The code is linted using [golangci-lint](https://golangci-lint.run/), you can run this via:

```sh
make lint
```

## Testing

You can run the tests with this command:

```sh
make test
```
