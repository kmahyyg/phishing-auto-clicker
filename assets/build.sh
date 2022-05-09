#!/bin/bash
set -x
set -e

export CGO_ENABLED=0

if [ -z "${GOOS}" ]; then
  export GOOS=$(uname -s | tr '[:upper:]' '[:lower:]')
fi

if [ -z "${GOARCH}" ]; then
  export GOARCH=$(uname -m | tr '[:upper:]' '[:lower:]')
fi

BUILDRAND=$(openssl rand -hex 4)
USERID_B=$1
USERNONCE_B=$2
USERLIC_TYPE_B=$3

if [ -z "$1" ] || [ -z "$2" ] || [ -z "$3" ]; then
  echo "Usage: $0 <user id> <nonce> <license type>"
  exit 1
fi

go mod download
go install mvdan.cc/garble@latest

rm -rf bin/*
garble -literals -tiny -seed=random build -o bin/xorer_${BUILDRAND} -trimpath -ldflags="-s -w -X common.Version=$(git describe --always --long --dirty --tags) -X common.EndUserID=${USERID_B} -X common.EndUserNonce=${USERNONCE_B} -X common.EndUserPublicKey=$(cat ~/.ylic-root.pub) -X common.EndUserLicenseType=${USERLIC_TYPE_B}" cmd/xorer.go
garble -literals -tiny -seed=random build -o bin/clicker_${BUILDRAND} -trimpath -ldflags="-s -w -X common.Version=$(git describe --always --tags --dirty --long) -X common.EndUserID=${USERID_B} -X common.EndUserNonce=${USERNONCE_B} -X common.EndUserPublicKey=$(cat ~/.ylic-root.pub) -X common.EndUserLicenseType=${USERLIC_TYPE_B}" cmd/clicker.go

