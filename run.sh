#!/bin/bash

golint ./... \
&& go fmt ./... \
&& go test ./... \
&& make \
&& build/dp-content-resolver
