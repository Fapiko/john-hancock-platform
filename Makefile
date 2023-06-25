.PHONY: build
VERSION := $(shell cat VERSION)
DOCKER_IMAGE := docker.io/ljandrew/john-hancock-platform

build:
	@rm -rf build/target
	mkdir -p build/target
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o build/target/john-hancock ./app

build-docker:
	docker build -t $(DOCKER_IMAGE):$(VERSION) -t $(DOCKER_IMAGE):latest .

push-docker:
	docker push $(DOCKER_IMAGE):$(VERSION)
	docker push $(DOCKER_IMAGE):latest

build-push-docker: build build-docker push-docker
