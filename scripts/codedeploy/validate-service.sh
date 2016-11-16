#!/bin/bash

if [[ $(docker inspect --format="{{ .State.Running}}" content-resolver) == "false" ]]; then
  exit 1;
fi
