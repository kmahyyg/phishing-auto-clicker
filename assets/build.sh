#!/bin/bash

export CGO_ENABLED=0
export GOOS=$(uname -s)
export GOARCH=$(uname -m)

BUILDRAND=$(openssl rand -hex 4)

go mod download
go install mvdan.cc/garble@latest

garble -literals -tiny -seed=random build -o bin/xorer_${BUILDRAND} -trimpath -ldflags="-s -w -X common.Version=$(git describe --always --long --dirty --tags) -X common.EndUserID=$1 -X common.EndUserNonce=$2 -X common.EndUserPublicKey=$(cat ~/.ylic-root.pub) -X common.EndUserLicenseType=T2" cmd/xorer.go

garble -literals -tiny -seed=random build -o bin/clicker_${BUILDRAND} -trimpath -ldflags="-s -w -X common.Version=$(git describe --always --tags --dirty --long) -X common.EndUserID=$1 -X common.EndUserNonce=$2 -X common.EndUserPublicKey=$(cat ~/.ylic-root.pub) -X common.EndUserLicenseType=T2" cmd/clicker.go

