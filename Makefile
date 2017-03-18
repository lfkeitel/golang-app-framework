# variable definitions
NAME := My App
DESC := A description
VERSION := $(shell git describe --tags --always --dirty)
GOVERSION := $(shell go version)
BUILDTIME := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
BUILDDATE := $(shell date -u +"%B %d, %Y")
BUILDER := $(shell echo "`git config user.name` <`git config user.email`>")
PKG_RELEASE ?= 1

CGO_ENABLED ?= 1
PWD := $(shell pwd)
GOBIN := $(PWD)/bin

ifeq ($(shell uname -o), Cygwin)
PWD := $(shell cygpath -w -a `pwd`)
GOBIN := $(PWD)\bin
endif

PROJECT_URL := "https://repo.example.com/$(NAME)"
BUILDTAGS ?= dball
LDFLAGS := -X 'main.version=$(VERSION)' \
			-X 'main.buildTime=$(BUILDTIME)' \
			-X 'main.builder=$(BUILDER)' \
			-X 'main.goversion=$(GOVERSION)' \
			-X 'main.appName=$(NAME)'

.PHONY: all doc fmt alltests test coverage benchmark lint vet app dist clean

all: test app

# development tasks
doc:
	@godoc -http=:6060 -index

fmt:
	@go fmt $$(go list ./src/...)

alltests: test lint vet

test:
	@go test -race $$(go list ./src/...)

coverage:
	@go test -cover $$(go list ./src/...)

benchmark:
	@echo "Running tests..."
	@go test -bench=. $$(go list ./src/...)

# https://github.com/golang/lint
# go get github.com/golang/lint/golint
lint:
	@golint ./src/...

vet:
	@go vet $$(go list ./src/...)

app:
	@mkdir -p bin
	go install -v -ldflags "$(LDFLAGS)" -tags '$(BUILDTAGS)' ./cmd/app

dist: vet test app
	@echo "Distribution task needs to be defined"

clean:
	rm $(TARGETS)
	rm -rf ./logs/*
	rm -rf ./sessions/*
