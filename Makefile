COMMIT_SHA_SHORT ?= $(shell git rev-parse --short=12 HEAD)
PWD_DIR := ${CURDIR}

default: help

#==========================================================================================
##@ Testing
#==========================================================================================
# Cover profile written by `make test` / `make test-race` and read by `make coverage`, so verify runs the suite once.
COVER_PROFILE ?= coverage.out

# -alldbs must follow the package list: placed before it, go test passes ./...
# to the test binary and only tests the current directory.
test: ## run go tests across all DBs (testcontainers, needs Docker), writing the cover profile
	@rm -f $(COVER_PROFILE)
	@go test ./... -alldbs -covermode=atomic -coverprofile=$(COVER_PROFILE)

test-race: ## same as test, with the race detector
	@rm -f $(COVER_PROFILE)
	@go test ./... -alldbs -race -covermode=atomic -coverprofile=$(COVER_PROFILE)

CONSUMER_DIR := testdata/consumer

# A separate module that imports testdbs like a user's project, with a package
# that uses it (InitDBS in TestMain) and one that doesn't. It runs a plain
# `go test` (embedded SQLite only) and one with TESTDBS_ALL (every DB, needs
# Docker); -count=1 because go test caches results that read env vars.
test-consumer: ## run the consumer module's tests, as a user would: plain and with TESTDBS_ALL
	@cd $(CONSUMER_DIR) && go test ./... -count=1
	@cd $(CONSUMER_DIR) && TESTDBS_ALL=1 go test ./... -count=1

lint: ## run go linter
	# check lining, depends on https://github.com/golangci/golangci-lint
	@golangci-lint run

license-check: ## check for invalid licenses
	# Check licenses, depends on https://github.com/elastic/go-licence-detector
	@go list -m -mod=readonly  -json all  | go-licence-detector -includeIndirect -rules allowedLicenses.json \
	-overrides overrideLicenses.json

# Default coverage threshold is 80
COVERAGE_THRESHOLD ?= 80

# Reads the profile of `make test` / `make test-race` (no test run of its own): each package must reach the
# threshold. A profile block is "dir/file.go:pos numStmts count", so a package is the dir of its blocks.
.PHONY: coverage
coverage: ## gate per-package coverage on the cover profile of `make test` / `make test-race`
	@if [ ! -f $(COVER_PROFILE) ]; then \
		echo "❌ $(COVER_PROFILE) not found: run make test or make test-race first"; \
		exit 1; \
	fi
	@go list ./... | awk -v threshold=$(COVERAGE_THRESHOLD) ' \
		FNR == 1 && NR == 1 { next } \
		NR == FNR { \
			pkg = $$1; sub(/\/[^\/]*:.*$$/, "", pkg); \
			if (!($$1 in stmts)) { stmts[$$1] = $$2; total[pkg] += $$2 } \
			if ($$3 > 0 && !($$1 in hit)) { hit[$$1] = 1; covered[pkg] += $$2 } \
			next \
		} \
		{ \
			if (!total[$$1]) { printf "⚠️  %-70s no coverage data\n", $$1; fail = 1; next } \
			pct = 100 * covered[$$1] / total[$$1]; \
			if (pct < threshold) { printf "❌ %-70s %.1f%% (below %s%%)\n", $$1, pct, threshold; fail = 1 } \
			else printf "✅ %-70s %.1f%%\n", $$1, pct \
		} \
		END { \
			if (fail) printf "❌ coverage below threshold (%s%%)\n", threshold; \
			else printf "✅ coverage: all packages >= %s%%\n", threshold; \
			exit fail \
		}' $(COVER_PROFILE) -

.PHONY: verify
verify: ## run all checks (full DB matrix); runs every check and fails if any fail
	@fail=0; \
	for target in test-race test-consumer license-check lint coverage; do \
		echo "==================== make $$target ===================="; \
		$(MAKE) --no-print-directory $$target || fail=1; \
	done; \
	if [ $$fail -ne 0 ]; then \
		echo "❌ verify failed (see above)"; \
		exit 1; \
	fi; \
	echo "✅ verify passed"

#==========================================================================================
##@ Release
#==========================================================================================

.PHONY: check-git-clean
check-git-clean: # check if git repo is clen
	@git diff --quiet

.PHONY: check-branch
check-branch:
	@current_branch=$$(git symbolic-ref --short HEAD) && \
	if [ "$$current_branch" != "main" ]; then \
		echo "Error: You are on branch '$$current_branch'. Please switch to 'main'."; \
		exit 1; \
	fi

check_env: # check for needed envs
	@[ "${version}" ] || ( echo ">> version is not set, usage: make release version=\"v1.2.3\" "; exit 1 )

tag: check_env check-branch check-git-clean verify ## create a tag and push to git
	@git diff --quiet || ( echo 'git is in dirty state' ; exit 1 )
	@[ "${version}" ] || ( echo ">> version is not set, usage: make release version=\"v1.2.3\" "; exit 1 )
	@git tag -d $(version) || true
	@git tag -a $(version) -m "Release version: $(version)"
	@git push --delete origin $(version) || true
	@git push origin $(version) || true

#==========================================================================================
#  Help
#==========================================================================================
.PHONY: help
help: # Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)
