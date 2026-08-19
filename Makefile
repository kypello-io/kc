PWD := $(shell pwd)
GOPATH := $(shell go env GOPATH)
LDFLAGS := $(shell go run buildscripts/gen-ldflags.go)

TARGET_GOARCH ?= $(shell go env GOARCH)
TARGET_GOOS ?= $(shell go env GOOS)

VERSION ?= $(shell git describe --tags)
TAG ?= "ghcr.io/kypello-io/kc:$(VERSION)"

GOLANGCI = $(GOPATH)/bin/golangci-lint

all: build

checks:
	@echo "Checking dependencies"
	@(env bash $(PWD)/buildscripts/checkdeps.sh)

getdeps:
	@mkdir -p ${GOPATH}/bin

crosscompile:
	@(env bash $(PWD)/buildscripts/cross-compile.sh)

verifiers: getdeps vet

docker: build
	@docker build -t $(TAG) . -f Dockerfile.dev

vet:
	@echo "Running $@"
	@GO111MODULE=on go vet github.com/kypello-io/kc/...

lint-fix: getdeps ## runs golangci-lint suite of linters with automatic fixes
	@echo "Running $@ check"
	@$(GOLANGCI) run --build-tags kqueue --timeout=10m --config ./.golangci.yml --fix

lint: getdeps
	@echo "Running $@ check"
	$(GOLANGCI) run --build-tags kqueue --timeout=10m --config ./.golangci.yml

# Builds kc, runs the verifiers then runs the tests.
check: test
test: verifiers build
	@echo "Running unit tests"
	@GO111MODULE=on CGO_ENABLED=0 go test -tags kqueue ./... 1>/dev/null
	@echo "Running functional tests"
	@GO111MODULE=on MC_TEST_RUN_FULL_SUITE=true go test -race -v --timeout 20m ./... -run Test_FullSuite

test-race: verifiers build
	@echo "Running unit tests under -race"
	@GO111MODULE=on go test -race -v --timeout 20m ./... 1>/dev/null

# Verify kc binary
verify:
	@echo "Verifying build with race"
	@GO111MODULE=on CGO_ENABLED=1 go build -race -tags kqueue -trimpath --ldflags "$(LDFLAGS)" -o $(PWD)/kc 1>/dev/null
	@echo "Running functional tests"
	@GO111MODULE=on MC_TEST_RUN_FULL_SUITE=true go test -race -v --timeout 20m ./... -run Test_FullSuite

# Builds kc locally.
build: checks
	@echo "Building kc binary to './kc'"
	@GO111MODULE=on GOOS=$(TARGET_GOOS) GOARCH=$(TARGET_GOARCH) CGO_ENABLED=0 go build -trimpath -tags kqueue --ldflags "$(LDFLAGS)" -o $(PWD)/kc

# Builds kc and installs it to $GOPATH/bin.
install: build
	@echo "Installing kc binary to '$(GOPATH)/bin/kc'"
	@mkdir -p $(GOPATH)/bin && cp -f $(PWD)/kc $(GOPATH)/bin/kc
	@echo "Installation successful. To learn more, try \"kc --help\"."

clean:
	@echo "Cleaning up all the generated files"
	@find . -name '*.test' | xargs rm -fv
	@find . -name '*~' | xargs rm -fv
	@rm -rvf kc
	@rm -rvf build
	@rm -rvf release
