# This repo's root import path (under GOPATH).
ROOT := github.com/funstory-ai/gobun
# Target binaries. You can build multiple binaries for a single project.
TARGETS := gobun
# Disable CGO by default.
CGO_ENABLED ?= 0
#
# These variables should not need tweaking.
#

# It's necessary to set this because some environments don't link sh -> bash.
export SHELL := bash
export SHELLOPTS := errexit
# Project main package location (can be multiple ones).
CMD_DIR := ./cmd
# Project output directory.
OUTPUT_DIR := ./bin
DEBUG_DIR := ./debug-bin
# Build directory.
BUILD_DIR := ./build
# Current version of the project.
VERSION ?= $(shell git describe --match 'v[0-9]*' --always --tags --abbrev=0)
BUILD_DATE=$(shell date -u +'%Y-%m-%dT%H:%M:%SZ')
GIT_COMMIT=$(shell git rev-parse HEAD)
GIT_TAG ?= $(shell if [ -z "`git status --porcelain`" ]; then git describe --exact-match --tags HEAD 2>/dev/null; fi)
GIT_TREE_STATE=$(shell if [ -z "`git status --porcelain`" ]; then echo "clean" ; else echo "dirty"; fi)
GITSHA ?= $(shell git rev-parse --short HEAD)
GIT_LATEST_TAG ?= $(shell git describe --tags --abbrev=0)

# Golang standard bin directory.
GOPATH ?= $(shell go env GOPATH)
GOROOT ?= $(shell go env GOROOT)
BIN_DIR := $(GOPATH)/bin
GOLANGCI_LINT := $(BIN_DIR)/golangci-lint
MOCKGEN := $(BIN_DIR)/mockgen

# Default golang flags used in build and test
# -mod=vendor: force go to use the vendor files instead of using the `$GOPATH/pkg/mod`
# -p: the number of programs that can be run in parallel
# -count: run each test and benchmark 1 times. Set this flag to disable test cache
export GOFLAGS ?= -count=1

build-release:
	@for target in $(TARGETS); do                                                      \
	  CGO_ENABLED=$(CGO_ENABLED) go build -trimpath -o $(OUTPUT_DIR)/$${target}     \
	    -ldflags "-s -w -X $(ROOT)/pkg/version.version=$(VERSION) \
		-X $(ROOT)/pkg/version.buildDate=$(BUILD_DATE) \
		-X $(ROOT)/pkg/version.gitCommit=$(GIT_COMMIT) \
		-X $(ROOT)/pkg/version.gitTreeState=$(GIT_TREE_STATE)                     \
		-X $(ROOT)/pkg/version.gitTag=$(GIT_TAG)" \
	    $(CMD_DIR)/$${target};                                                         \
	done