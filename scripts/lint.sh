#!/bin/bash
# Copyright (c) Facebook, Inc. and its affiliates. All Rights Reserved
dirs=( $(find . -type f -name '*.go' | xargs -n1 dirname | sort | uniq) )

for d in "${dirs[@]}"; do
	echo "=========== $d"
	find $d -mindepth 1 -maxdepth 1  -type f -name '*.go' | xargs golint
done
