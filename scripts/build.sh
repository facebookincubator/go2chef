#!/bin/bash
# Copyright (c) Facebook, Inc. and its affiliates. All Rights Reserved
set -eux

source scripts/common.sh

go build -o "$go2chef_output" ./bin
