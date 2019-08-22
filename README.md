# `go2chef`: "just enough Go to get to Chef"

## What is `go2chef`?
`go2chef` is a Go tool for bootstrapping Chef installations in a flexible and self-contained way. With `go2chef`, our goal is to make bootstrapping any node in a Chef deployment as simple as "get `go2chef` onto a machine and run it.

## Design

`go2chef` has four basic building blocks, all of which are implemented using a plugin model:

* **Configuration Sources (`go2chef.ConfigSource`):** fetch `go2chef` configuration from remote sources
* **Loggers (`go2chef.Logger`):** send log messages and structured events to logging backends via a common plugin API
* **Steps (`go2chef.Step`):** implement the building blocks of a `go2chef` workflow. Every action that needs to be taken to set up your Chef environment can be integrated into a `go2chef` step. See "Steps" for more details
* **Sources (`go2chef.Source`):** implement a common API for retrieval of remote resources needed for `Step` execution

### Configuration Sources
Configuration sources are the plugins which allow you to customize how `go2chef` retrieves its runtime configuration. We provide a couple configuration plugins out-of-the-box:

* `go2chef.config_source.local`: loads configuration from a JSON file accessible on the filesystem. (*this is the default configuration source*)
* `go2chef.config_source.http`: loads configuration source in JSON format from an HTTP(S) endpoint. Enable using `go2chef --config-source go2chef.config_source.http`
* `go2chef.config_source.embed`: loads configuration source from an embedded variable. This probably isn't what you want, but if it is, have it.

New configuration sources can be registered with `go2chef.RegisterConfigSource`.

### Loggers
Loggers are the plugins which allow `go2chef` users to report run information for monitoring and analysis, and provide plugin authors with a single API for logging and events.

#### For Users
Logging plugins are configured using the `loggers` key in `go2chef` configuration. An example configuration setting up the default `go2chef.logger.stdlib` looks like:

```json
{
  "loggers": [
    {
      "type": "go2chef.logger.stdlib",
      "name": "stdlib",
      "level": "DEBUG",
      "debugging": 1,
      "verbosity": 1
    }
  ]
}
```

The `loggers` key is an array so that you can log to multiple places, which may be useful for the following scenarios:

1. You want your raw log messages to go to syslog, but you also want to send specific events to a separate logging service using a custom plugin to trigger some downstream action (i.e. changing asset service state).
2. You want to log to file and stderr and syslog at varying levels of verbosity (and so on and so forth)

#### For Developers
Logging plugins may skip parts of the interface specification by stubbing out the unneeded methods as no-ops.

The `go2chef.MultiLogger` implementation synchronously dumps messages out to backends at the moment, so delays in message sending in a `Logger` plugin may slow down execution of `go2chef` as well.

### Steps
Steps are the plugins which actually "do stuff" in `go2chef`. These can do pretty much anything you want if you implement it, but we've intentionally limited the built-in plugins to the following initially:

* **Sanity checking:** make sure that the runtime environment is sane before trying to install Chef -- are we `root`? Is the clock set right? Is there disk space?
* **Bundle exec:** provide a simple abstraction for fetching and running some arbitrary scripts/binaries before/after installation. Do things like set up required certs, install `chefctl.rb`, etc.
* **Installers:** provide installer implementations for each platform (and sub-platforms thereof, if necessary).

Many `Step` implementations will require some sort of remote resource retrieval; rather than leaving it up to each implementation to bring its own support code for downloads, we provide it to you using `Sources` (described next).

### Sources
Source plugins implement a common API for resource retrieval for `go2chef`. This allows all steps to configure remote resource retrieval with the same idiom:

```json
{
  "steps": [
    {
      "type": "go2chef.step.install.linux.apt",
      "name": "install chef",
      "source": {
        "type": "go2chef.source.http",
        "url": "https://example.com/chef-15.deb"
      }
    }
  ]
}
```

A `source` key inside a step configuration block defines how the remote resources for that step should be retrieved.

## Code Layout

```
bin/        # go2chef binary source code
build/      # temporary directory for build outputs
cli/        # CLI implementation
plugin/     # plugins directory
  config/   # configuration source plugins
  logger/   # logger plugins
  source/   # source plugins
  step/     # step plugins
*.go        # go code for the base go2chef module
```

## Building

```
$ make all      # build all binaries supported
$ make linux    # build only linux binary
$ make test     # run tests
```

## Improvements Planned

* Basic templating in config (path substitution etc)
  * Store a map of step name => step download dir?# go2chef

go2chef is a tool to boostrap a system from first principles so that it's able
to run Chef to be managed. It takes care of downloading and installing Chef in
a generic and multi-platform way.

## Usage
TODO

## Requirements
go2chef has no runtime dependencies.

## Contribute
See the CONTRIBUTING file for how to help out.

## License
go2chef is Apache 2.0 licensed.
