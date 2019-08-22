#!/bin/bash
set -eux

source scripts/common.sh

go build -o "$go2chef_output" ./bin
