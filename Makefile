GO ?= $(shell command -v go 2> /dev/null)
MM_DEBUG ?=
MANIFEST_FILE ?= plugin.json
GOPATH ?= $(shell go env GOPATH)
GO_TEST_FLAGS ?= -race
GO_BUILD_FLAGS ?=

export GO111MODULE=on

# You can include assets this directory into the plugin bundle. This can be e.g.
# used to include profile pictures.
ASSETS_DIR ?= assets

## Define the default target (make all)
.PHONY: default
default: all

# Verify environment, and define PLUGIN_ID, PLUGIN_VERSION, HAS_SERVER and
# HAS_WEBAPP as needed.
include build/setup.mk
include build/legacy.mk

PLUGIN_BUNDLE_NAME ?= $(PLUGIN_ID)-$(PLUGIN_VERSION)-plugin.tar.gz
AWS_BUNDLE_NAME ?= $(PLUGIN_ID)-$(PLUGIN_VERSION)-aws.zip

# Include custom makefile, if present
ifneq ($(wildcard build/custom.mk),)
	include build/custom.mk
endif

## Checks the code style, tests, builds and bundles the app.
.PHONY: all
all: check-style test dist

## Runs eslint and golangci-lint
.PHONY: check-style
check-style: 
	@echo Checking for style guide compliance

ifneq ($(HAS_SERVER),)
	@if ! [ -x "$$(command -v golangci-lint)" ]; then \
		echo "golangci-lint is not installed. Please see https://github.com/golangci/golangci-lint#install for installation instructions."; \
		exit 1; \
	fi; \

	@echo Running golangci-lint
	golangci-lint run ./...
endif

.PHONY: run
## run: runs the app locally
run: 
	cd http-server ; $(GO) run $(GO_BUILD_FLAGS) .

.PHONY: dist-aws
## dist-aws: creates the bundle file for AWS Lambda deployments
dist-aws: 
	rm -rf dist/aws && mkdir -p dist/aws
	cd aws ; \
	 	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ../dist/aws/main . 
	cp manifest.json dist/aws
	cp -r static dist/aws
	cd dist/aws ; \
		zip -m gcal.zip main ; \
		zip -rm ../$(AWS_BUNDLE_NAME) manifest.json static gcal.zip 
	rm -r dist/aws

## Builds the server, if it exists, for all supported architectures.
.PHONY: server
server:
ifneq ($(HAS_SERVER),)
	mkdir -p server/dist;
ifeq ($(MM_DEBUG),)
	cd server && env GOOS=linux GOARCH=amd64 $(GO) build $(GO_BUILD_FLAGS) -o dist/plugin-linux-amd64;
	cd server && env GOOS=darwin GOARCH=amd64 $(GO) build $(GO_BUILD_FLAGS) -o dist/plugin-darwin-amd64;
	cd server && env GOOS=darwin GOARCH=arm64 $(GO) build $(GO_BUILD_FLAGS) -o dist/plugin-darwin-arm64;
	cd server && env GOOS=windows GOARCH=amd64 $(GO) build $(GO_BUILD_FLAGS) -o dist/plugin-windows-amd64.exe;
else
	$(info DEBUG mode is on; to disable, unset MM_DEBUG)

	cd server && env GOOS=darwin GOARCH=amd64 $(GO) build $(GO_BUILD_FLAGS) -gcflags "all=-N -l" -o dist/plugin-darwin-amd64;
	cd server && env GOOS=darwin GOARCH=arm64 $(GO) build $(GO_BUILD_FLAGS) -gcflags "all=-N -l" -o dist/plugin-darwin-arm64;
	cd server && env GOOS=linux GOARCH=amd64 $(GO) build $(GO_BUILD_FLAGS) -gcflags "all=-N -l" -o dist/plugin-linux-amd64;
	cd server && env GOOS=windows GOARCH=amd64 $(GO) build $(GO_BUILD_FLAGS) -gcflags "all=-N -l" -o dist/plugin-windows-amd64.exe;
endif
endif

## Generates a tar bundle of the plugin for install.
.PHONY: dist-plugin
dist-plugin:
	rm -rf dist/
	mkdir -p dist/$(PLUGIN_ID)
	cp $(MANIFEST_FILE) dist/$(PLUGIN_ID)/
ifneq ($(wildcard $(ASSETS_DIR)/.),)
	cp -r $(ASSETS_DIR) dist/$(PLUGIN_ID)/
endif
ifneq ($(HAS_PUBLIC),)
	cp -r public dist/$(PLUGIN_ID)/
endif
ifneq ($(HAS_SERVER),)
	mkdir -p dist/$(PLUGIN_ID)/server
	cp -r server/dist dist/$(PLUGIN_ID)/server/
endif
	cd dist && tar -cvzf $(PLUGIN_BUNDLE_NAME) $(PLUGIN_ID)

	@echo plugin built at: dist/$(PLUGIN_BUNDLE_NAME)

## Builds and bundles the plugin.
.PHONY: dist
dist:	server dist-plugin dist-aws

## Builds and installs the plugin to a server.
.PHONY: deploy
deploy: dist
	./build/bin/pluginctl deploy $(PLUGIN_ID) dist/$(PLUGIN_BUNDLE_NAME)

## Runs any lints and unit tests.
.PHONY: test
test: 
ifneq ($(HAS_SERVER),)
	$(GO) test -v $(GO_TEST_FLAGS) ./server/...
endif
ifneq ($(wildcard ./build/sync/plan/.),)
	cd ./build/sync && $(GO) test -v $(GO_TEST_FLAGS) ./...
endif

## Creates a coverage report for the server code.
.PHONY: coverage
coverage: 
ifneq ($(HAS_SERVER),)
	$(GO) test $(GO_TEST_FLAGS) -coverprofile=server/coverage.txt ./server/...
	$(GO) tool cover -html=server/coverage.txt
endif

## Disable the plugin.
.PHONY: disable
disable: detach
	./build/bin/pluginctl disable $(PLUGIN_ID)

## Enable the plugin.
.PHONY: enable
enable:
	./build/bin/pluginctl enable $(PLUGIN_ID)

## Clean removes all build artifacts.
.PHONY: clean
clean:
	rm -fr dist/
ifneq ($(HAS_SERVER),)
	rm -fr server/coverage.txt
	rm -fr server/dist
endif
	rm -fr build/bin/

## Sync directory with a starter template
sync:
ifndef STARTERTEMPLATE_PATH
	@echo STARTERTEMPLATE_PATH is not set.
	@echo Set STARTERTEMPLATE_PATH to a local clone of https://github.com/mattermost/mattermost-plugin-starter-template and retry.
	@exit 1
endif
	cd ${STARTERTEMPLATE_PATH} && go run ./build/sync/main.go ./build/sync/plan.yml $(PWD)

# Help documentation à la https://marmelab.com/blog/2016/02/29/auto-documented-makefile.html
help:
	@cat Makefile build/*.mk | grep -v '\.PHONY' |  grep -v '\help:' | grep -B1 -E '^[a-zA-Z0-9_.-]+:.*' | sed -e "s/:.*//" | sed -e "s/^## //" |  grep -v '\-\-' | sed '1!G;h;$$!d' | awk 'NR%2{printf "\033[36m%-30s\033[0m",$$0;next;}1' | sort
