NAME := my_exporter
sources := $(shell find . -type d -name tmp -prune -o -type f -name '*.go' -print)

GOBIN := $(shell go env GOPATH)/bin

# Embed build metadata into the binary.
# see https://godoc.org/github.com/prometheus/common/version
export VERSION := $(shell git describe --tags --abbrev=0 2>/dev/null)
export REVISION := $(shell git rev-parse HEAD 2>/dev/null)
export BRANCH := $(shell git symbolic-ref --short HEAD 2>/dev/null)
GIT_USER_NAME := $(shell git config user.name 2>/dev/null)
GIT_USER_EMAIL := $(shell git config user.email 2>/dev/null)
GIT_USER := $(strip $(GIT_USER_NAME) $(if $(GIT_USER_EMAIL),<$(GIT_USER_EMAIL)>,))
export BUILD_USER := $(if $(GIT_USER),$(GIT_USER),$(shell whoami)@$(shell hostname))
export BUILD_DATE := $(shell date +"%Y%m%d-%H:%M:%S")
LDFLAGS := \
	-X 'github.com/prometheus/common/version.Version=$(VERSION)' \
	-X 'github.com/prometheus/common/version.Revision=$(REVISION)' \
	-X 'github.com/prometheus/common/version.Branch=$(BRANCH)' \
	-X 'github.com/prometheus/common/version.BuildUser=$(BUILD_USER)' \
	-X 'github.com/prometheus/common/version.BuildDate=$(BUILD_DATE)'

$(NAME): $(sources)
	go build -v -ldflags "$(LDFLAGS)" -o $@

.PHONY: all
all: check $(NAME)

.PHONY: check
check: lint vet test

.PHONY: lint
lint:
	$(GOBIN)/golint ./...

.PHONY: vet
vet:
	go vet ./...

.PHONY: test
test:
	go test -v ./...

.PHONY: setup
setup: deps
	go get golang.org/x/lint/golint

.PHONY: deps
deps:
	go get -v -t -d ./...

.PHONY: clean
clean:
	go clean
	rm -f $(NAME)

# TODO: create target for test with openio container
.PHONY: full-test
full-test:
	@echo "Not implemented yet."

GORELEASER_FLAGS := --rm-dist --skip-publish
.PHONY: dist
dist:
	goreleaser $(GORELEASER_FLAGS)

# TODO: fail if the environment is not linux/amd64
.PHONY: install
install:
	go install -ldflags "$(LDFLAGS)"

# Shortcuts to control the development container
DEV_TARGETS = start stop exec
.PHONY: $(DEV_TARGETS)
$(DEV_TARGETS):
	cd ./dev && make $@
