#!/bin/bash

set -e
PRJROOT=$(dirname "${BASH_SOURCE[0]}")/..
cd $PRJROOT

MODNAME="github.com/OutOfBedlam/ots"

if [ "$1" == "" ]; then
    echo "error: missing argument (target name)"
    exit 1
fi

# Check the Go installation
if [ "$(which go)" == "" ]; then
	echo "error: Go is not installed. Please download and follow installation"\
		 "instructions at https://golang.org/dl to continue."
	exit 1
fi

# Version from 'version.txt' or git tags
if [ -f "./version.txt" ]; then
    VERSION=`cat ./version.txt`
elif [ -d ".git" ]; then
    VERSION=$(git describe --tags --abbrev=0)
fi

# Hardcode some values to the core package.
if [ -d ".git" ]; then
	GITSHA=$(git rev-parse --short HEAD)
	LDFLAGS="$LDFLAGS -X $MODNAME/banner.versionString=${VERSION}"
	LDFLAGS="$LDFLAGS -X $MODNAME/banner.versionGitSHA=${GITSHA}"
fi
GOVERSTR=$(go version | sed -r 's/go version go(.*)\ .*/\1/')
LDFLAGS="$LDFLAGS -X $MODNAME/banner.goVersionString=${GOVERSTR}"
LDFLAGS="$LDFLAGS -X $MODNAME/banner.buildTimestamp=$(date "+%Y-%m-%dT%H:%M:%S")"

# Set final Go environment options
LDFLAGS="$LDFLAGS -extldflags '-static'"
export CGO_ENABLED=0

if [ "$NOMODULES" != "1" ]; then
	export GO111MODULE=on
	export GOFLAGS=-mod=vendor
	go mod vendor
fi

# Build and store objects into original directory.
go build -ldflags "$LDFLAGS" -o $PRJROOT/tmp/$1 $1/*.go
