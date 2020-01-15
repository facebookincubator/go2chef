#!/bin/bash
# Copyright (c) Facebook, Inc. and its affiliates. All Rights Reserved
set -eux
goos="$GOOS"
goarch="$GOARCH"
go2chef="go2chef"

if [[ "$goos" == "windows" ]]; then
	go2chef="$go2chef.exe"
fi

go build -o "build/$goos/$goarch/$go2chef"
