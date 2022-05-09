#!/bin/bash

export CGO_ENABLED=0
export GOOS=$(uname -s)
export GOARCH=$(uname -m)

BUILDRAND=$(openssl rand -hex 4)

go mod download
go install mvdan.cc/garble@latest

garble -literals -tiny -seed=random build -o bin/xorer_${BUILDRAND} -trimpath -ldflags="-s -w -Xcommon.Version=$(git describe --always --dirty --long --tags) -Xcommon.EndUserID=$1 -Xcommon.EndUserNonce=$2 -Xcommon.EndUserPublicKey=$(cat ~/.ylic-root.pub) -Xcommon.EndUserLicenseType=T2" cmd/xorer.go

garble -literals -tiny -seed=random build -o bin/clicker_${BUILDRAND} -trimpath -ldflags="-s -w -Xcommon.Version=$(git describe --always --dirty --tags) -Xcommon.EndUserID=$1 -Xcommon.EndUserNonce=$2 -Xcommon.EndUserPublicKey=$(cat ~/.ylic-root.pub) -Xcommon.EndUserLicenseType=T2" cmd/clicker.go

