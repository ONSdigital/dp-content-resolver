#!/bin/bash

golint ./... \
&& go fmt ./... \
&& make \
&& build/dp-content-resolver
