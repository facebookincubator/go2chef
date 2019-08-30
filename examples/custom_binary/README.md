# Custom Binaries

Need to build with custom plugins? `go2chef` is designed to make this super simple:

1. Copy this directory to a new location and `cd` to it
2. Edit `go.mod` to set the module name to something informative
3. Add `_` imports for your custom plugins in `bin/plugins.go`
4. Run `make go2chef` to build your custom `go2chef` binary
5. Go...to...Chef.
