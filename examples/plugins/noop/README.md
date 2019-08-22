# Writing Plugins

## Fundamentals

At a high level, the plugin process consists of:

1. Implementing the interface(s) you want your plugin to provide
2. Registering the plugin in your plugin module's `init()` function using the `go2chef.Register*` method for each interface implemented
3. Importing the plugin module into your `go2chef` build
4. Configuring the plugin in your config file

## Example: `noop`

This directory contains the complete implementation of a `noop` plugin.

If you `cd` to this directory and run:

```
$ go run ./bin --local-config config.json
```

You'll see:

```
2019/07/20 14:25:42 INFO: executing step 0: noop
2019/07/20 14:25:42 INFO: noop noop: Download() called
2019/07/20 14:25:42 INFO: noop noop: Execute() called
```

### How does it work?

#### `noop.go`

`noop.go` provides the concrete implementation of `go2chef.Step` and has its `init()` function set up to register the plugin under the name `noop`. This structure is essentially the same for `Log`, `Source`, and `Step` plugins. See the comments in the file for some guidance on general implementation principles.

This is #1 and #2 in the **Fundamentals** section at the top.

#### `bin/go2chef.go`

`go2chef.go` is all that's required to make a custom build of `go2chef`. Note the side-effect import of `noop`:

```
import (
    ...
    _ "noop"
)
```

This is all that's required to build the plugin, and because we register it in the `init()` function of `noop.go` it's automatically loaded.

#### `config.json`

`config.json` demonstrates how to use the `noop` plugin in your configuration file. To make a `noop` step just create a step block with the `type` set to `noop`.

#### Building

You can just build the `bin` folder to get a `go2chef` binary, or use the top-level directory's `make examples`:

```
$ go build -o /tmp/go2chef ./bin
$ /tmp/go2chef --help
```
