package bundle

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/facebookincubator/go2chef/util"
)

func chefRuby() string {
	if runtime.GOOS == "windows" {
		return "C:/opscode/chef/embedded/bin/ruby.exe"
	}
	return "/opt/chef/embedded/bin/ruby"
}

type BundleEntrypointFinder func(fullPath string, ctx context.Context) *exec.Cmd

var unixLoadOrder = []string{
	"bundle.rb",
	"bundle.sh",
}
var loadOrder = map[string][]string{
	"windows": []string{
		"bundle.rb",
		"bundle.ps1",
		"bundle.bat",
	},
	"linux":  unixLoadOrder,
	"darwin": unixLoadOrder,
}

func findEntrypoint(dir string) (string, error) {
	var realLoadOrder []string
	if lo, ok := loadOrder[runtime.GOOS]; ok {
		realLoadOrder = lo
	} else {
		realLoadOrder = unixLoadOrder
	}
	for _, lo := range realLoadOrder {
		loPath := filepath.Join(dir, lo)
		if util.PathExists(loPath) {
			return loPath, nil
		}
	}
	return "", os.ErrNotExist
}

func commandForPath(path string, ctx context.Context) *exec.Cmd {
	switch filepath.Ext(path) {
	case ".ps1":
		return exec.CommandContext(ctx, "powershell.exe", "-ExecutionPolicy", "Bypass", "-File", path)
	case ".bat", ".cmd":
		return exec.CommandContext(ctx, "cmd", "/c", path)
	case ".rb":
		return exec.CommandContext(ctx, chefRuby(), path)
	case ".sh":
		return exec.CommandContext(ctx, "sh", path)
	case ".bash":
		return exec.CommandContext(ctx, "bash", path)
	default:
		return nil
	}
}
