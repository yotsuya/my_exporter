NAME := my_exporter
sources := $(shell find . -type d -name tmp -prune -o -type f -name '*.go' -print)
GOBIN := $(shell go env GOPATH)/bin
export CGO_ENABLED := 0

$(NAME): $(sources)
	go build -v -o $@

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

.PHONY: long-test
long-test:
	cd ./test && make test

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

# See .goreleaser.yml for what these means
git_branch := $(shell git symbolic-ref --short HEAD 2>/dev/null)
export BRANCH := $(if $(git_branch),$(git_branch),HEAD)
git_user_name := $(shell git config user.name 2>/dev/null)
git_user_email := $(shell git config user.email 2>/dev/null)
git_user := $(strip $(git_user_name) $(if $(git_user_email),<$(git_user_email)>,))
export BUILD_USER := $(if $(git_user),$(git_user),$(shell whoami)@$(shell hostname))
export BUILD_DATE := $(shell date +"%Y%m%d-%H:%M:%S")

.PHONY: dist
dist: GORELEASER_FLAGS ?= --snapshot --skip-publish --rm-dist
dist:
	goreleaser $(GORELEASER_FLAGS)

.PHONY: echo_branch echo_build_user echo_build_date
echo_branch:
	@echo $(BRANCH)
echo_build_user:
	@echo $(BUILD_USER)
echo_build_date:
	@echo $(BUILD_DATE)


# Shortcuts to control the development container
DEV_TARGETS = start stop exec
.PHONY: $(DEV_TARGETS)
$(DEV_TARGETS):
	cd ./dev && make $@
