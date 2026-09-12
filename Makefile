# NFR-009 scale bench. The 1 GiB corpus is generated under /tmp (or
# LOGIFY_NFR009_DIR) and must not be committed.
GO ?= go
NFR009_DIR ?= /tmp/logify-nfr009

.PHONY: bench-nfr009-smoke bench-nfr009

# CI / local smoke: 8 MiB fixture + short -bench (AC4). Does not claim AC2/AC3.
bench-nfr009-smoke:
	$(GO) test ./internal/analyzer -run TestNFR009ScaleSmoke -count=1
	$(GO) test ./internal/analyzer -bench BenchmarkNFR009ScaleSmoke -benchtime=1x -count=1

# Manual/nightly 1 GiB run (AC1–AC4). Peak RSS is measured in a child process.
bench-nfr009:
	LOGIFY_NFR009_FULL=1 LOGIFY_NFR009_DIR=$(NFR009_DIR) \
		$(GO) test ./internal/analyzer -run TestNFR009Scale1GiB -count=1 -timeout 15m
