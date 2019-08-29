require 'tempfile'
require 'json'
require 'chef/mixin/deep_merge'

CONFIG_JSON = '/etc/chef/config.json'
CONFIG_JSON_D = '/etc/chef/config.json.d'
module OkoChefctl
  def pre_run(output)
    config = JSON.parse(File.read(CONFIG_JSON))
    if Dir.exist? CONFIG_JSON_D
      Dir.glob(File.join(CONFIG_JSON_D, '**.json')).each do |configp|
        config = Chef::Mixin::DeepMerge.deep_merge(
          config, JSON.parse(File.read(configp)))
      end
    end
    @tempfile = Tempfile.new
    @tempfile.write(JSON.pretty_generate(config))
    @tempfile.flush
    @tempfile.rewind

    Chefctl::Config.chef_options << '-j' << @tempfile.path
  end
end
Chefctl::Plugin.register OkoChefctl

