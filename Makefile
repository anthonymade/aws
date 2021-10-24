SHELL := /usr/bin/env bash

GO_VERSION=$(shell sed -n -e '/^go /s/^go //p' go.mod)

GO_XCOMPILE_VARS = GOOS=$(word 2, $(subst _, ,$*)) GOARCH=$(lastword $(subst _, ,$(basename $*)))
GO_XCOMPILE_CMD_NAME = $(firstword $(subst _, ,$*))

CHECKSUMS := checksums_sha256.txt
COVERAGE := coverage.out
DOCKER_IMAGE_NAME := aws_tools_image

all: target/$(CHECKSUMS)

test:
	go test -v ./...

test_with_coverage:
	go test -v -coverprofile=$(COVERAGE) ./...

cover: test_with_coverage
	go tool cover -func=$(COVERAGE)

cover_html: test_with_coverage
	go tool cover -html=$(COVERAGE)

fmt:
	goimports -l -w .

vet:
	go vet ./...

build: fmt
	go build ./...
	cd cmd/awsi && go build .

clean:
	rm -rf target $(COVERAGE)

target:
	mkdir target

target/%: cover vet target
	$(GO_XCOMPILE_VARS) go build ./...
	cd cmd/$(GO_XCOMPILE_CMD_NAME) && $(GO_XCOMPILE_VARS) go build -ldflags "-s -w" -o ../../$@ .
	xz -k $@

xcompile: \
	target/awsi_linux_amd64 \
	target/awsi_linux_arm64 \
	target/awsi_darwin_amd64 \
	target/awsi_darwin_arm64 \
	target/awsi_windows_amd64.exe

target/$(CHECKSUMS): xcompile
	cd target && sha256sum $$( ls | grep -v \.txt ) > $(CHECKSUMS)

build_docker_image:
	time docker build \
		--build-arg GO_VERSION=$(GO_VERSION) \
		--build-arg USER_ID=$(shell id -u) \
		--build-arg GROUP_ID=$(shell id -g) \
		--tag $(DOCKER_IMAGE_NAME) \
		.

docker_%: build_docker_image
	time docker run --rm --user $(shell id -u):$(shell id -g) -v "$(PWD)":/usr/src/myapp -w /usr/src/myapp $(DOCKER_IMAGE_NAME) make $*


.PHONY: all test test_with_coverage cover cover_html clean build xcompile build_docker_image fmt vet
