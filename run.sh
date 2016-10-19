#!/bin/bash

golint  -set_exit_status ./... && make && build/dp-content-resolver
