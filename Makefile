ORG = projectcontour
PROJECT = ir2proxy
MODULE = github.com/$(ORG)/$(PROJECT)
REGISTRY ?= projectcontour
IMAGE := $(REGISTRY)/$(PROJECT)
SRCDIRS := ./cmd

TAG_LATEST ?= false

# Sets GIT_REF to a tag if it's present, otherwise the short rev.
GIT_REF = $(shell git describe --tags || git rev-parse --short=8 --verify HEAD)
VERSION ?= $(GIT_REF)
# Used for the tag-latest action.
# The tag-latest action will be a noop unless this is explicitly
# set outside this Makefile, as a safety valve.
LATEST_VERSION ?= NOLATEST

export GO111MODULE=on

Check_Targets := \
	check-test \
	check-test-race \
	check-vet \
	check-gofmt \
	check-staticcheck \
	check-misspell \
	check-unconvert \
	check-unparam \
	check-ineffassign \

.PHONY: check
check: install $(Check_Targets) ## Run tests and CI checks

.PHONY: pedantic
pedantic: check check-errcheck ## Run pedantic CI checks

install: ## Build and install the contour binary
	go install -mod=readonly -v -tags "oidc gcp" $(MODULE)/cmd/$(PROJECT)

race:
	go install -mod=readonly -v -race -tags "oidc gcp" $(MODULE)/cmd/$(PROJECT)

download: ## Download Go modules
	go mod download

.PHONY: check-test
check-test:
	go test -cover -mod=readonly $(MODULE)/...

.PHONY: check-test-race
check-test-race: | check-test
	go test -race -mod=readonly $(MODULE)/...

.PHONY: check-staticcheck
check-staticcheck:
	go install honnef.co/go/tools/cmd/staticcheck
	staticcheck \
		-checks all,-ST1003 \
		$(MODULE)/{cmd,internal}/...

.PHONY: check-misspell
check-misspell:
	go install github.com/client9/misspell/cmd/misspell
	misspell \
		-i clas \
		-locale US \
		-error \
		cmd/* internal/* docs/* design/* site/*.md site/_{guides,posts,resources} *.md

.PHONY: check-unconvert
check-unconvert:
	go install github.com/mdempsky/unconvert
	unconvert -v $(MODULE)/{cmd,internal}/...

.PHONY: check-ineffassign
check-ineffassign:
	go install github.com/gordonklaus/ineffassign
	find $(SRCDIRS) -name '*.go' | xargs ineffassign

.PHONY: check-unparam
check-unparam:
	go install mvdan.cc/unparam
	unparam -exported $(MODULE)/{cmd,internal}/...

.PHONY: check-errcheck
check-errcheck:
	go install github.com/kisielk/errcheck
	errcheck $(MODULE)/...

.PHONY: check-gofmt
check-gofmt:
	@echo Checking code is gofmted
	@test -z "$(shell gofmt -s -l -d -e $(SRCDIRS) | tee /dev/stderr)"

.PHONY: check-vet
check-vet: | check-test
	go vet $(MODULE)/...

help: ## Display this help
	@echo Contour high performance Ingress controller for Kubernetes
	@echo
	@echo Targets:
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z0-9._-]+:.*?## / {printf "  %-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST) | sort
