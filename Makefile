LOCALPATH := /usr/local/bin/
ARTIFACT_PATH := bin/macos
PROJECT ?= cmd/kubenx
FILE ?= $(PROJECT)/main.go
LINUX_ARTIFACT_PATH := bin/linux
SCRIPT_FILE=release.sh
SERVICE := kubenx

build:
	go build -o $(ARTIFACT_PATH)/$(SERVICE) $(FILE)

linux-build:
	GOOS=linux go build -o ${LINUX_ARTIFACT_PATH}/${SERVICE} kubenx.go

clean:
	rm -rf ${ARTIFACT_PATH}

upload:
	./${SCRIPT_FILE}

install: build clean

format:
	go fmt ./...

all: build linux-build upload
