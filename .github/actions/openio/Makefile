ROOT=../../..

.PHONY: all
all:	bin image run

my_exporter:
	GOOS=linux GOARCH=amd64 go build -v $(ROOT)/.
	cp $(ROOT)/my_exporter .

.PHONY: image
image: Dockerfile entrypoint.sh my_exporter
	docker build --rm -t my-openio .

.PHONY: run
run:
	docker run -it my-openio:latest
