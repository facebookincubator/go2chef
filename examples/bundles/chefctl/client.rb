#
# Copyright (c) Facebook, Inc. and its affiliates. All Rights Reserved
#
chef_repo_root = ENV.include?('CHEF_REPO') ? ENV['CHEF_REPO'] : '/etc/chef/repo'

STDERR.puts("Chef repo: #{chef_repo_root}")

cookbook_root = File.join(chef_repo_root, 'cookbooks')

cookbook_paths = Dir.entries(cookbook_root).map { |family|
  File.join(cookbook_root, family)
}.select { |path|
  File.directory?(path) && !['.', '..'].include?(File.basename(path))
}
puts cookbook_paths
cookbook_path cookbook_paths
node_path File.join(chef_repo_root, 'nodes')
file_cache_path (Gem.win_platform? ? 'C:/chef/cache' : '/var/cache/chef')

local_mode true
chef_zero.enabled true
