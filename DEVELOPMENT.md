# Developing testdbs

## Overview

testdbs is a single Go package at the repository root. `testdbs.go` holds the
public API and DB selection, with one file per target (`sqlite.go`, `mysql.go`,
`mariadb.go`, `postgres.go`) and the connection parameters in `connparams.go`.
MariaDB and MySQL share one target type in `mysql.go`; `mariadb.go` only sets
its flavor. The server targets start containers with testcontainers-go. The
package's own `TestMain` adds MySQL to the defaults, so tests cover both.

## Prerequisites

- Go 1.26 (from `go.mod`). Keep the `go` directive and dependency versions
  low: downstream apps inherit them as their floor.
- Docker, for the MariaDB, MySQL, Postgres, and CGo SQLite targets. Every `make` test
  target runs the full `-alldbs` set.
- A C compiler and `CGO_ENABLED=1`, for the CGo SQLite target.
- [golangci-lint](https://github.com/golangci/golangci-lint) v2.12 for `make lint`.
- [go-licence-detector](https://github.com/elastic/go-licence-detector) for
  `make license-check`:
  `go install go.elastic.co/go-licence-detector@v0.10.0`.

## Common tasks

```sh
make help           # list the targets
make test           # tests on every DB; writes the cover profile coverage.out
make test-race      # same, with the race detector
make test-consumer  # the consumer module's tests, plain and with TESTDBS_ALL
make lint           # golangci-lint
make license-check  # fail on dependencies outside allowedLicenses.json
make coverage       # fail if coverage.out is below 80% (COVERAGE_THRESHOLD=…); runs no tests
make verify         # test-race, test-consumer, license-check, lint, coverage; run it before opening a PR
```

`make coverage` only reads the profile the last `make test` or `make test-race`
wrote, so `verify` runs the package's own suite once.

For a quick run without Docker, `go test ./...` tests only the pure-Go SQLite
target.

Two kinds of tests check testdbs as a user sees it:

- `example_test.go` holds the README snippets as `Example` functions, so they
  keep compiling and show on pkg.go.dev. Update both together.
- `testdata/consumer/` is a separate module that imports testdbs through a
  `replace`, with one package that calls `InitDBS()` from `TestMain` and one
  that doesn't import testdbs at all. `make test-consumer` runs it with a plain
  `go test` (SQLite only) and with `TESTDBS_ALL=1` (every DB). Go tooling
  skips `testdata/`, so the main module's tests, lint, and license check don't
  see it. After changing the main `go.mod`, run `go mod tidy` in it too.

CI runs `make test` with the coverage gate, `make test-consumer`, the lint, and
the license check on every push and pull request (`.github/workflows/`).

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
