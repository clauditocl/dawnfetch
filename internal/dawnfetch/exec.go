// this file runs external commands with timeouts and safe defaults.
package dawnfetch

import (
	"context"
	"errors"
	"os/exec"
	"strings"
	"sync"
	"time"
)

var commandExistsCache sync.Map

func runCmd(timeout time.Duration, name string, args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, name, args...)
	out, err := cmd.CombinedOutput()
	if err != nil && name == "powershell" {
		cmd = exec.CommandContext(ctx, "pwsh", args...)
		out, err = cmd.CombinedOutput()
	}
	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		return "", ctx.Err()
	}
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func commandExists(name string) bool {
	if v, ok := commandExistsCache.Load(name); ok {
		return v.(bool)
	}
	_, err := exec.LookPath(name)
	exists := err == nil
	commandExistsCache.Store(name, exists)
	return exists
}
