#!/bin/bash
set -eux

. scripts/common.sh

target="$1"
config="$2"
sshtgt="$(echo -n "$target" | tr -d '[]')"

bin="$(basename "$go2chef_output")"
pth="/tmp"

if [[ "$goos" == "windows" ]]; then
	pth=""
fi

scripts/build.sh

scp "$go2chef_output" "$target:$pth/$bin"
ssh "$sshtgt" -- "$pth/$bin"
