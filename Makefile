GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOARCH=$(shell go env GOARCH)
GOOS=$(shell go env GOOS )

BASE_PATH := $(shell pwd)
BUILD_PATH = $(BASE_PATH)/build
WEB_PATH=$(BASE_PATH)/frontend
ASSERT_PATH= $(BASE_PATH)/core/cmd/server/web/assets

CORE_PATH=$(BASE_PATH)/core
CORE_MAIN=$(CORE_PATH)/cmd/server/main.go
CORE_NAME=1panel-core

AGENT_PATH=$(BASE_PATH)/agent
AGENT_MAIN=$(AGENT_PATH)/cmd/server/main.go
AGENT_NAME=1panel-agent


clean_assets:
	rm -rf $(ASSERT_PATH)

upx_bin:
	upx $(BUILD_PATH)/$(CORE_NAME)
	upx $(BUILD_PATH)/$(AGENT_NAME)

build_frontend:
	cd $(WEB_PATH) && npm install && npm run build:pro

build_core_on_linux:
	cd $(CORE_PATH) \
	&& GOOS=$(GOOS) GOARCH=$(GOARCH) $(GOBUILD) -trimpath -ldflags '-s -w' -o $(BUILD_PATH)/$(CORE_NAME) $(CORE_MAIN)

build_agent_on_linux:
	cd $(AGENT_PATH) \
    && GOOS=$(GOOS) GOARCH=$(GOARCH) $(GOBUILD) -trimpath -ldflags '-s -w' -o $(BUILD_PATH)/$(AGENT_NAME) $(AGENT_MAIN)

build_core_on_darwin:
	cd $(CORE_PATH) \
	&& GOOS=linux GOARCH=amd64 $(GOBUILD) -trimpath -ldflags '-s -w'  -o $(BUILD_PATH)/$(CORE_NAME) $(CORE_MAIN)

build_agent_on_darwin:
	cd $(AGENT_PATH) \
    && GOOS=linux GOARCH=amd64 $(GOBUILD) -trimpath -ldflags '-s -w'  -o $(BUILD_PATH)/$(AGENT_NAME) $(AGENT_MAIN)

build_all: build_frontend build_core_on_linux build_agent_on_linux

build_on_local: clean_assets build_frontend build_core_on_darwin build_agent_on_darwin upx_bin
