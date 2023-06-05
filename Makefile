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
ASSERT_PATH= $(BASE_PAH)/cmd/server/web/asserts

clean_asserts:
	rm -rf $(ASSERT_PATH)

upx_bin:
	upx $(BUILD_PATH)/$(APP_NAME)

build_frontend:
	cd $(WEB_PATH) && npm install && npm run build:dev

build_backend_on_linux:
	cd $(SERVER_PATH) \
    && CGO_ENABLED=1 GOOS=$(GOOS) GOARCH=$(GOARCH) $(GOBUILD) -trimpath -ldflags '-s -w --extldflags "-static -fpic"' -tags 'osusergo,netgo' -o $(BUILD_PATH)/$(APP_NAME) $(MAIN)

build_backend_on_darwin:
	cd $(SERVER_PATH) \
    && CGO_ENABLED=1 GOOS=linux GOARCH=amd64 CC=x86_64-linux-musl-gcc CXX=x86_64-linux-musl-g++ $(GOBUILD) -trimpath -ldflags '-s -w --extldflags "-static -fpic"'  -o $(BUILD_PATH)/$(APP_NAME) $(MAIN)

build_backend_on_archlinux:
	cd $(SERVER_PATH) \
    && CGO_ENABLED=1 GOOS=$(GOOS) GOARCH=$(GOARCH) $(GOBUILD) -trimpath -ldflags '-s -w --extldflags "-fpic"' -tags osusergo -o $(BUILD_PATH)/$(APP_NAME) $(MAIN)

build_all: build_frontend  build_backend_on_linux

build_on_local: clean_asserts build_frontend  build_backend_on_darwin upx_bin
