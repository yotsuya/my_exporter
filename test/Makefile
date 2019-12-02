PROJECT_DIR := $(shell dirname $(dir $(realpath $(firstword $(MAKEFILE_LIST)))))
NAME := $(shell basename $(PROJECT_DIR))
IMAGE := local/$(NAME):test

.PHONY: test
test: build run

.PHONY: build
build: $(NAME)
	docker build -t $(IMAGE) .

.PHONY: run
run:
	docker run -it --rm $(IMAGE)

.PHONY: clean
clean:
	rm -f $(NAME)

$(NAME):
	cd .. && make clean && GOOS=linux GOARCH=amd64 make $@ && cp $@ ./test
