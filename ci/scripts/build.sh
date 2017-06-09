#!/bin/bash -eux

export BINPATH=$(pwd)/build
export GOPATH=$(pwd)/go

pushd $GOPATH/src/github.com/ONSdigital/dp-content-resolver
  go build -o $BINPATH/dp-content-resolver && cp Dockerfile.concourse $BINPATH/
popd
