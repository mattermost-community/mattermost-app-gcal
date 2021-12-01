# Include custom targets and environment variables here
default: all

BUILD_DATE = $(shell date -u)
BUILD_HASH = $(shell git rev-parse HEAD)
BUILD_HASH_SHORT = $(shell git rev-parse --short HEAD)
LDFLAGS += -X "github.com/mattermost/mattermost-plugin-apps/server/config.BuildDate=$(BUILD_DATE)"
LDFLAGS += -X "github.com/mattermost/mattermost-plugin-apps/server/config.BuildHash=$(BUILD_HASH)"
LDFLAGS += -X "github.com/mattermost/mattermost-plugin-apps/server/config.BuildHashShort=$(BUILD_HASH_SHORT)"
GO_BUILD_FLAGS += -ldflags '$(LDFLAGS)'
GO_TEST_FLAGS += -ldflags '$(LDFLAGS)'

BUNDLE_NAME = $(PLUGIN_ID)-$(PLUGIN_VERSION)-plugin.tar.gz
AWS_BUNDLE_NAME ?= $(PLUGIN_ID)-$(PLUGIN_VERSION)-aws.zip

## run: runs the app locally
.PHONY: run
run:
	cd http-server ; $(GO) run $(GO_BUILD_FLAGS) .

## dist-aws: creates the bundle file for AWS Lambda deployments
.PHONY: dist-aws
dist-aws:
	rm -rf dist/aws && mkdir -p dist/aws
	cd aws ; \
		CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build $(GO_BUILD_FLAGS) -o ../dist/aws/main .
	cp manifest.json dist/aws
	cp -r static dist/aws
	cd dist/aws ; \
		zip -m gcal.zip main ; \
		zip -rm ../$(AWS_BUNDLE_NAME) manifest.json static gcal.zip
	rm -r dist/aws
