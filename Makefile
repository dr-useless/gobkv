## Thanks to https://github.com/chemidy/smallest-secured-golang-docker-image

VERSION=`git rev-parse HEAD`
BUILD=`date +%FT%T%z`
LDFLAGS=-ldflags "-X main.Version=${VERSION} -X main.Build=${BUILD}"
DOCKER_IMAGE=rocketkv

## - Show help
.PHONY: help
help:
	@printf "\033[32m\xE2\x9c\x93 usage: make [target]\n\n\033[0m"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

## - Docker pull latest images
.PHONY: docker-pull
docker-pull:
	@printf "\033[32m\xE2\x9c\x93 docker pull latest images\n\033[0m"
	@docker pull golang:alpine

## - Build
.PHONY: build
build:docker-pull
	@printf "\033[32m\xE2\x9c\x93 Build rocketkv\n\033[0m"
	$(eval BUILDER_IMAGE=$(shell docker inspect --format='{{index .RepoDigests 0}}' golang:alpine))
	@export DOCKER_CONTENT_TRUST=1
	@docker build --build-arg "BUILDER_IMAGE=$(BUILDER_IMAGE)" -t rocketkv .

## - List docker images
.PHONY: ls
ls:
	@printf "\033[32m\xE2\x9c\x93 Look at the size dude !\n\033[0m"
	@docker image ls rocketkv

## - Run
.PHONY: run
run:
	@printf "\033[32m\xE2\x9c\x93 Run rocketkv\n\033[0m"
	@docker run -d -p 8100 -v $(PWD):/etc/rocketkv rocketkv -c /etc/rocketkv/config.json

## - Scan for known vulnerabilities
.PHONY: scan
scan:
	@printf "\033[32m\xE2\x9c\x93 Scan for known vulnerabilities in rocketkv\n\033[0m"
	@docker scan -f Dockerfile rocketkv
