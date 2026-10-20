.DEFAULT_GOAL := help
.FOO: help run dev build

BUILD_TIME := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)

help:
	@echo "targets: run, dev, build"

generate:
	go generate ./...

run: generate
	go run -ldflags "-X main.generateSecrets=true -X main.buildTime=$(BUILD_TIME)" ./cmd/onestopshop

dev: generate
	go build -ldflags "-X main.generateSecrets=true -X main.buildTime=$(BUILD_TIME)" -o bin/onestopshop-dev ./cmd/onestopshop

build:
	go build -ldflags="-s -w -X main.buildTime=$(BUILD_TIME)" -o bin/onestopshop ./cmd/onestopshop