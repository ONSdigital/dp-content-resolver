#!/bin/bash

golint ./... && make && build/dp-content-resolver
