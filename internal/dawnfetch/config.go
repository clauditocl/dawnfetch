// this file handles persisted user config and default theme state.
package dawnfetch

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

type UserConfig struct {
	DefaultTheme string `json:"default_theme"`
	Initialized  bool   `json:"initialized"`
}

const defaultConfigFileName = "dawnfetch_config.json"

func loadPersistedDefaultTheme(fallback string) string {
	cfg, err := loadUserConfig()
	if err != nil {
		return fallback
	}
	if v := strings.TrimSpace(cfg.DefaultTheme); v != "" {
		return v
	}
	return fallback
}

func loadUserConfig() (UserConfig, error) {
	for _, path := range userConfigLoadCandidates() {
		b, err := os.ReadFile(path)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return UserConfig{}, err
		}
		var cfg UserConfig
		if err := json.Unmarshal(b, &cfg); err != nil {
			return UserConfig{}, fmt.Errorf("invalid config file %s: %w", path, err)
		}
		return cfg, nil
	}
	return UserConfig{}, nil
}

func userConfigLoadCandidates() []string {
	out := make([]string, 0, 4)
	seen := map[string]struct{}{}
	add := func(path string, ok bool) {
		if !ok || strings.TrimSpace(path) == "" {
			return
		}
		if _, exists := seen[path]; exists {
			return
		}
		seen[path] = struct{}{}
		out = append(out, path)
	}

	add(executableConfigPath())
	add(userScopedConfigPath())
	return out
}

func userConfigPath() (string, error) {
	exePath, hasExe := executableConfigPath()
	userPath, hasUser := userScopedConfigPath()

	// If an executable-local config already exists and is writable,
	// keep using it for portable installs.
	if hasExe {
		if _, err := os.Stat(exePath); err == nil && canWritePath(exePath) {
			return exePath, nil
		}
	}

	// Prefer per-user writable config for system installs (/usr/bin, Program Files).
	if hasUser && canWritePath(userPath) {
		return userPath, nil
	}

	// Fallback to executable-local if writable.
	if hasExe && canWritePath(exePath) {
		return exePath, nil
	}

	// Last-resort best effort so caller gets a concrete target path.
	if hasUser {
		return userPath, nil
	}
	if hasExe {
		return exePath, nil
	}
	return "", fmt.Errorf("failed to resolve config path")
}

func executableConfigPath() (string, bool) {
	exePath, err := os.Executable()
	if err != nil {
		return "", false
	}
	return filepath.Join(filepath.Dir(exePath), defaultConfigFileName), true
}

func userScopedConfigPath() (string, bool) {
	if base, err := os.UserConfigDir(); err == nil && strings.TrimSpace(base) != "" {
		return filepath.Join(base, "dawnfetch", defaultConfigFileName), true
	}

	home, err := os.UserHomeDir()
	if err != nil || strings.TrimSpace(home) == "" {
		return "", false
	}
	if runtime.GOOS == "windows" {
		return filepath.Join(home, "AppData", "Roaming", "dawnfetch", defaultConfigFileName), true
	}
	return filepath.Join(home, ".config", "dawnfetch", defaultConfigFileName), true
}

func canWritePath(path string) bool {
	path = strings.TrimSpace(path)
	if path == "" {
		return false
	}

	if fi, err := os.Stat(path); err == nil && !fi.IsDir() {
		f, err := os.OpenFile(path, os.O_WRONLY|os.O_APPEND, 0)
		if err != nil {
			return false
		}
		_ = f.Close()
		return true
	}

	dir := filepath.Dir(path)
	if dir == "" {
		dir = "."
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return false
	}
	f, err := os.CreateTemp(dir, ".dawnfetch-write-*")
	if err != nil {
		return false
	}
	name := f.Name()
	_ = f.Close()
	_ = os.Remove(name)
	return true
}

func saveUserConfig(cfg UserConfig) error {
	path, err := userConfigPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	b, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, append(b, '\n'), 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}
