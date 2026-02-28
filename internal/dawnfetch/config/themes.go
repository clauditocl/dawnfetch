// this file loads and validates theme palette files.
package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"dawnfetch/internal/dawnfetch/core"
)

func LoadThemePalettes(path string) (map[string][]string, error) {
	paths := ThemeFileCandidates(path)
	var b []byte
	var err error
	seen := map[string]struct{}{}
	lastReadErr := error(nil)
	for _, p := range paths {
		if _, ok := seen[p]; ok {
			continue
		}
		seen[p] = struct{}{}
		b, err = os.ReadFile(p)
		if err == nil {
			break
		}
		lastReadErr = err
	}
	if err != nil {
		if errors.Is(lastReadErr, os.ErrNotExist) {
			return nil, fmt.Errorf("themes file %q was not found", path)
		}
		return nil, fmt.Errorf("themes file %q could not be read", path)
	}
	var cfg core.ThemeConfig
	if err := json.Unmarshal(b, &cfg); err != nil {
		return nil, fmt.Errorf("themes file %q is invalid json", path)
	}
	if len(cfg.Palettes) == 0 {
		return nil, fmt.Errorf("themes file %q has no palettes", path)
	}
	return cfg.Palettes, nil
}

func ShouldWarnThemeLoad(path string, err error) bool {
	if err == nil {
		return false
	}
	if strings.EqualFold(strings.TrimSpace(path), "themes.json") && errors.Is(err, os.ErrNotExist) {
		return false
	}
	return true
}
