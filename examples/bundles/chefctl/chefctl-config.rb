# This config file is located at /etc/chefctl-config.rb
# You can change this location by passing `-C/--config` to `chefctl`
# Default options and descriptions are in comments below.

# Allow the chef run to provide colored output.
#color false

# Whether or not chefctl should provide verbose output.
verbose false

root_path = Gem.win_platform? ? 'C:/opscode/chef' : '/opt/chef'

# The chef-client process to use.
if Dir.exist? root_path
  chef_client File.join(root_path, '/bin/chef-client')
elsif Dir.exist? root_path+"dk"
  chef_client File.join(root_path+"dk", '/bin/chef-client')
end

# Whether or not chef-client should provide debug output.
#debug false

# Default options to pass to chef-client.
chef_options [
    '--no-fork',
    '-c', '/etc/chef/client.rb',
    '-z',
]

# Whether or not to provide human-readable output.
#human false

# If set, ignore the splay and stop pending chefctl processes before
# running. This is intended for interactive runs of chef
# (i.e. started by a human).
immediate false

# The lock file to use for chefctl.
lock_file (Gem.win_platform? ? 'C:/chef/chefctl.lock' : '/var/lock/subsys/chefctl')

# How long to wait for the lock to become available.
#lock_time 1800

# Directory where per-run chef logs should be placed.
log_dir (Gem.win_platform? ? 'C:/chef/outputs' : '/var/chef/outputs')

# If set, will not copy chef log to stdout.
quiet false

# The default splay to use. Ignored if `immediate` is set to true.
#splay 870

# How many chef-client retries to attempt before failing.
# See Chefctl::Plugin.rerun_chef?
#max_retries 1

# The testing timestamp.
# See https://github.com/facebook/taste-tester
#testing_timestamp '/etc/chef/test_timestamp'

# Whether or not to run chef in whyrun mode.
#whyrun false

# The default location of the chefctl plugin file.
#plugin_path '/etc/chef/chefctl_hooks.rb'

# The default PATH environment variable to use for chef-client.
puts "env", ENV.inspect
if Gem.win_platform?
  pth = %w{
    C:/Windows/System32
  }
else
  pth = %w{
    /usr/sbin
    /usr/bin
  }
end
path pth

# Whether or not to symlink output files for chef.cur.out and chef.last.out
#symlink_output true

