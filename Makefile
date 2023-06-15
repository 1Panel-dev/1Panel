GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOARCH=$(shell go env GOARCH)
GOOS=$(shell go env GOOS )

BASE_PAH := $(shell pwd)
BUILD_PATH = $(BASE_PAH)/build
WEB_PATH=$(BASE_PAH)/frontend
SERVER_PATH=$(BASE_PAH)/backend
MAIN= $(BASE_PAH)/cmd/server/main.go
APP_NAME=1panel
ASSERT_PATH= $(BASE_PAH)/cmd/server/web/assets

clean_assets:
	rm -rf $(ASSERT_PATH)

upx_bin:
	upx $(BUILD_PATH)/$(APP_NAME)

build_frontend:
	cd $(WEB_PATH) && npm install && npm run build:pro

build_backend_on_linux:
	cd $(SERVER_PATH) \
    && GOOS=$(GOOS) GOARCH=$(GOARCH) $(GOBUILD) -trimpath -ldflags '-s -w' -o $(BUILD_PATH)/$(APP_NAME) $(MAIN)

build_backend_on_darwin:
	cd $(SERVER_PATH) \
    && GOOS=linux GOARCH=amd64 $(GOBUILD) -trimpath -ldflags '-s -w'  -o $(BUILD_PATH)/$(APP_NAME) $(MAIN)

build_backend_on_archlinux:
	cd $(SERVER_PATH) \
    && GOOS=$(GOOS) GOARCH=$(GOARCH) $(GOBUILD) -trimpath -ldflags '-s -w' -o $(BUILD_PATH)/$(APP_NAME) $(MAIN)

build_all: build_frontend build_backend_on_linux

build_on_local: clean_assets build_frontend build_backend_on_darwin upx_bin
