VERSION := 1.0.2
LOCALPATH := /usr/local/bin/
ARTIFACT_PATH := bin
SCRIPT_FILE=release.sh

build:
	go build -o ${ARTIFACT_PATH}/kubenx kubenx.go
	cp ${ARTIFACT_PATH}/kubenx ${LOCALPATH}

clean:
	rm -rf ${ARTIFACT_PATH}

upload:
	./${SCRIPT_FILE} ${VERSION}

all: build upload clean
