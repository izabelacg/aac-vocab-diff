.PHONY: build build-arm64 test lint fmt fmt-html setup-dev clean \
        deploy-pi

BIN        := bin/aac-vocab-diff
PI_HOST    ?= raspberrypi.local

GIT_DESCRIBE := $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
GIT_COMMIT   := $(shell git rev-parse --short HEAD 2>/dev/null || echo unknown)
BUILD_TIME   := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)

LDFLAGS := -X github.com/izabelacg/aac-vocab-diff/internal/version.Version=$(GIT_DESCRIBE) \
	-X github.com/izabelacg/aac-vocab-diff/internal/version.Commit=$(GIT_COMMIT) \
	-X github.com/izabelacg/aac-vocab-diff/internal/version.BuildTime=$(BUILD_TIME)

# ── Build ────────────────────────────────────────────────────────────────────────

build:
	mkdir -p bin
	go build -ldflags "$(LDFLAGS)" -o $(BIN) ./cmd/aac-vocab-diff

build-arm64:
	mkdir -p bin
	GOOS=linux GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o $(BIN)-arm64 ./cmd/aac-vocab-diff

# ── Clean ───────────────────────────────────────────────────────────────────────

clean:
	rm -rf bin/

# ── Tests ───────────────────────────────────────────────────────────────────────

test:
	go test ./... -v

# ── Lint ────────────────────────────────────────────────────────────────────────

lint:
	go vet ./...

# ── Format ──────────────────────────────────────────────────────────────────────

fmt: fmt-go fmt-html

fmt-go:
	gofmt -w .

fmt-html:
	@test -x .venv/bin/djlint || { echo "Run: make setup-dev"; exit 1; }
	.venv/bin/djlint report/templates/ server/templates/ --reformat

# Python venv + djlint — formats embedded HTML templates (report/server).
# Requires Python 3.10+. PEP 668–safe: never installs into system site-packages.
setup-dev:
	python3 -m venv .venv
	.venv/bin/pip install -U pip
	.venv/bin/pip install -r requirements-dev.txt

# ── Deploy ──────────────────────────────────────────────────────────────────

# Cross-compile and deploy to the Raspberry Pi over SSH.
# Usage: make deploy-pi PI_HOST=raspberrypi.local
deploy-pi: build-arm64
	@echo "==> Deploying to pi@$(PI_HOST) …"
	scp $(BIN)-arm64 pi@$(PI_HOST):/tmp/aac-vocab-diff
	scp deploy/aac-vocab-diff.service pi@$(PI_HOST):/tmp/aac-vocab-diff.service
	ssh pi@$(PI_HOST) '\
	  sudo mv /tmp/aac-vocab-diff /usr/local/bin/aac-vocab-diff && \
	  sudo chmod +x /usr/local/bin/aac-vocab-diff && \
	  sudo cp /tmp/aac-vocab-diff.service /etc/systemd/system/aac-vocab-diff.service && \
	  sudo mkdir -p /var/log/aac-vocab-diff && \
	  sudo chown pi:pi /var/log/aac-vocab-diff && \
	  sudo systemctl daemon-reload && \
	  sudo systemctl enable --now aac-vocab-diff && \
	  sudo systemctl restart aac-vocab-diff && \
	  sleep 2 && curl -fsS http://localhost:8888/health || { echo "ERROR: health check failed"; exit 1; }'
	@echo "==> Deploy complete."
