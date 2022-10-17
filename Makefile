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

build_web:
	cd $(WEB_PATH) && npm install && npm run build:dev

build_bin:
	cd $(SERVER_PATH) \
    && GOOS=$(GOOS) GOARCH=$(GOARCH)  $(GOBUILD) -trimpath  -ldflags "-s -w"  -o $(BUILD_PATH)/$(APP_NAME) $(MAIN)

build_all: build_web  build_bin

