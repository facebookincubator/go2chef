#!/bin/bash
find . -type f -name '*.go' \
	| grep -v ^./examples \
	| grep -v ^./scripts \
	| grep -v _test.go \
	| grep -v testutil \
	| grep -v doc.go \
	| grep -v plugin/step/install \
	| grep -v '^\s*//' \
	| xargs wc -l
