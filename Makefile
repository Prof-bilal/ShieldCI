GOCACHE ?= /tmp/shieldci-go-cache
WARDEN_DIR := engine/warden/warden-starter/warden

.PHONY: backend-test backend-race backend-vet backend-build warden-test warden-build landing-build verify

backend-test:
	cd backend && GOCACHE=$(GOCACHE) GOMAXPROCS=2 go test ./...

backend-race:
	cd backend && GOCACHE=$(GOCACHE) GOMAXPROCS=2 go test -race ./...

backend-vet:
	cd backend && GOCACHE=$(GOCACHE) GOMAXPROCS=2 go vet ./...

backend-build:
	cd backend && GOCACHE=$(GOCACHE) GOMAXPROCS=2 go build -buildvcs=false ./cmd/shieldci

warden-test:
	cd $(WARDEN_DIR) && GOCACHE=$(GOCACHE) GOMAXPROCS=2 go test -p 1 ./internal/policy/... ./internal/sandbox/...

warden-build:
	cd $(WARDEN_DIR) && GOCACHE=$(GOCACHE) GOMAXPROCS=2 go build -buildvcs=false ./cmd/warden

landing-build:
	cd landing && npm run build

verify: backend-test backend-race backend-vet backend-build warden-test warden-build landing-build
