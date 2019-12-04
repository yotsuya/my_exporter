PROJECT_DIR := $(shell dirname $(dir $(realpath $(firstword $(MAKEFILE_LIST)))))
NAME := $(shell basename $(PROJECT_DIR))
IMAGE := local/$(NAME):dev
CONTAINER := $(NAME).dev
PORT := 11010
WORK_DIR := /work

.PHONY: build
build:
	docker build -t $(IMAGE) .

.PHONY: run
run:
	docker run -d -p $(PORT):$(PORT) -v $(PROJECT_DIR):$(WORK_DIR) --name $(CONTAINER) $(IMAGE)

.PHONY: start
start:
	docker start $(CONTAINER)

.PHONY: stop
stop:
	docker stop $(CONTAINER)

.PHONY: attach exec
attach exec:
	docker exec -it $(CONTAINER) bash

.PHONY: rm
rm:
	docker rm $(CONTAINER)
