#!/bin/bash

golint ./... \
&& go fmt -w ./... \
&& make \
&& build/dp-content-resolver
