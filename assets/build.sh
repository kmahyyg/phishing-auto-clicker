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

rm -rf bin/clicker_* bin/xorer_*
garble -tiny -seed=random -literals build -o bin/xorer_${BUILDRAND} -trimpath -ldflags "-s -w -X phishingAutoClicker/common.VERSION=$(git describe --always --long --dirty --tags) -X phishingAutoClicker/common.EndUserID=${USERID_B} -X phishingAutoClicker/common.EndUserNonce=${USERNONCE_B} -X phishingAutoClicker/common.LicensePublicKey=$(cat ~/.ylic-root.pub | tr -d "=") -X phishingAutoClicker/common.EndUserLicenseType=${USERLIC_TYPE_B}" cmd/xorer.go
garble -tiny -seed=random -literals build -o bin/clicker_${BUILDRAND} -trimpath -ldflags "-s -w -X phishingAutoClicker/common.VERSION=$(git describe --always --tags --dirty --long) -X phishingAutoClicker/common.EndUserID=${USERID_B} -X phishingAutoClicker/common.EndUserNonce=${USERNONCE_B} -X phishingAutoClicker/common.LicensePublicKey=$(cat ~/.ylic-root.pub | tr -d "=") -X phishingAutoClicker/common.EndUserLicenseType=${USERLIC_TYPE_B}" cmd/clicker.go

