PROJECT_DIR := $(shell dirname $(dir $(realpath $(firstword $(MAKEFILE_LIST)))))
NAME := $(shell basename $(PROJECT_DIR))
IMAGE := local/$(NAME):test

.PHONY: test
test: build run

.PHONY: build
build:
	docker build -t $(IMAGE) .

.PHONY: run
run:
	docker run -it --rm $(IMAGE)
