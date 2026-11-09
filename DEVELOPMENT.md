# Developing testdbs

## Overview

testdbs is a single Go package at the repository root. `testdbs.go` holds the
public API and DB selection, with one file per target (`sqlite.go`, `mysql.go`,
`postgres.go`) and the connection parameters in `connparams.go`. The MySQL and
Postgres targets start containers with testcontainers-go.

## Prerequisites

- Go 1.24 (from `go.mod`). Keep the `go` directive and dependency versions
  low: downstream apps inherit them as their floor.
- Docker, for the MySQL, Postgres, and CGo SQLite targets. Every `make` test
  target runs the full `-alldbs` set.
- A C compiler and `CGO_ENABLED=1`, for the CGo SQLite target.
- [golangci-lint](https://github.com/golangci/golangci-lint) v2.12 for `make lint`.
- [go-licence-detector](https://github.com/elastic/go-licence-detector) for
  `make license-check`:
  `go install go.elastic.co/go-licence-detector@v0.10.0`.

## Common tasks

```sh
make help           # list the targets
make test           # tests on every DB, with coverage
make test-race      # same, with the race detector
make lint           # golangci-lint
make license-check  # fail on dependencies outside allowedLicenses.json
make coverage       # fail if coverage is below 80% (COVERAGE_THRESHOLD=…)
make verify         # all of the above; run it before opening a PR
```

For a quick run without Docker, `go test ./...` tests only the pure-Go SQLite
target.

CI runs `make test`, the lint, and the license check on every push and pull
request (`.github/workflows/`).

## Release

Releases are Git tags on `main`; the Go module proxy picks them up. There is
no build step. Run `make verify` on a clean `main`, then tag and push:

```sh
git tag -a v1.2.3 -m "Release version: v1.2.3"
git push origin v1.2.3
```

`make tag version=v1.2.3` does the same after `make verify`, but it first
deletes the tag locally and on `origin` and pushes it again. Never use it for
a version that is already published: re-pushing a module tag breaks the module
proxy and checksum database for consumers.
