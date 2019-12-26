#!/bin/bash
#
# Copyright (c) Facebook, Inc. and its affiliates. All Rights Reserved
#
set -eux
install -m 755 -d /etc/chef/repo
install -m 644 config.json /etc/chef/config.json
cp -r ./* /etc/chef/repo/
