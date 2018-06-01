#!/usr/bin/env sh

binary="chainid-$1-$2"

mkdir -p dist

docker run --rm -tv $(pwd)/api:/src -e BUILD_GOOS="$1" -e BUILD_GOARCH="$2" chainid/golang-builder:cross-platform /src/cmd/chainid

mv "api/cmd/chainid/$binary" dist/
#sha256sum "dist/$binary" > chainid-checksum.txt
