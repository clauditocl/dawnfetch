// this file loads and normalizes color palette themes.
package dawnfetch

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
)

type ThemeConfig struct {
	Palettes map[string][]string `json:"palettes"`
}

func resolvePaletteName(requested string, palettes map[string][]string) string {
	name := normalizePaletteName(requested)
	if _, ok := palettes[name]; ok {
		return name
	}
	return defaultPalette
}

func resolvePaletteNameStrict(requested string, palettes map[string][]string) (string, bool) {
	name := normalizePaletteName(requested)
	_, ok := palettes[name]
	return name, ok
}

func normalizePaletteName(requested string) string {
	name := strings.ToLower(strings.TrimSpace(requested))
	aliases := map[string]string{
		"trans":       "transgender",
		"transgender": "transgender",
		"transfem":    "transfeminine",
		"transmasc":   "transmasculine",
		"nb":          "nonbinary",
	}
	if a, ok := aliases[name]; ok {
		return a
	}
	return name
}

func defaultBrandConfig() BrandConfig {
	return BrandConfig{
		Palettes: map[string][]string{
			"transgender":    {"38;2;91;206;250", "38;2;245;169;184", "38;2;255;255;255", "38;2;245;169;184", "38;2;91;206;250"},
			"transfeminine":  {"38;2;245;169;184", "38;2;255;255;255", "38;2;170;224;255", "38;2;255;255;255", "38;2;245;169;184"},
			"transmasculine": {"38;2;95;184;255", "38;2;255;200;120", "38;2;255;255;255", "38;2;255;200;120", "38;2;95;184;255"},
			"nonbinary":      {"38;2;255;244;48", "38;2;255;255;255", "38;2;156;89;209", "38;2;40;40;40", "38;2;156;89;209"},
			"genderfluid":    {"38;2;255;117;162", "38;2;255;255;255", "38;2;190;0;255", "38;2;32;32;32", "38;2;70;130;255"},
			"agender":        {"38;2;32;32;32", "38;2;190;190;190", "38;2;255;255;255", "38;2;181;245;131", "38;2;255;255;255"},
		},
		Logos: map[string]LogoSet{
			"generic": {
				Normal:  []LogoLine{{Text: "dawnfetch", ColorIndex: 0}},
				Compact: []LogoLine{{Text: "dawnfetch", ColorIndex: 0}},
				Tiny:    []LogoLine{{Text: "dawnfetch", ColorIndex: 0}},
			},
		},
	}
}

func loadThemePalettes(path string) (map[string][]string, error) {
	paths := themeFileCandidates(path)
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
	var cfg ThemeConfig
	if err := json.Unmarshal(b, &cfg); err != nil {
		return nil, fmt.Errorf("themes file %q is invalid json", path)
	}
	if len(cfg.Palettes) == 0 {
		return nil, fmt.Errorf("themes file %q has no palettes", path)
	}
	return cfg.Palettes, nil
}

func shouldWarnThemeLoad(path string, err error) bool {
	if err == nil {
		return false
	}
	if strings.EqualFold(strings.TrimSpace(path), "themes.json") && errors.Is(err, os.ErrNotExist) {
		return false
	}
	return true
}
