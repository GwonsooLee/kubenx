#!/bin/bash
set -e

curl -LO https://feeldayone-public.s3.ap-northeast-2.amazonaws.com/release/latest/kubenx.tar.gz
gzip -d kubenx.tar.gz
tar -xzvf kubenx.tar
mv kubenx /usr/local/bin
kubenx version

rm -rf kubenx.tar*