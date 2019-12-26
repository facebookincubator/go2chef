#!/bin/bash
#
# Copyright (c) Facebook, Inc. and its affiliates. All Rights Reserved
#
set -eux
install -m 755 -d /etc/chef
install -m 755 -o root chefctl.rb /usr/local/bin/chefctl
install -m 644 -o root chefctl-config.rb /etc/chefctl-config.rb
install -m 644 -o root chefctl_hooks.rb /etc/chef/chefctl_hooks.rb
install -m 644 -o root client.rb /etc/chef/client.rb
