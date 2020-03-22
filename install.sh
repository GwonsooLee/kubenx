#!/bin/bash
set -e

go build

if [[ -f kubenx ]]; then
    mv kubenx /usr/local/bin
fi

