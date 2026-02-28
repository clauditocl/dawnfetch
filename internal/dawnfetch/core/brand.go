// this file defines built-in palettes and palette name normalization.
package core

import "strings"

type ThemeConfig struct {
	Palettes map[string][]string `json:"palettes"`
}

func ResolvePaletteName(requested string, palettes map[string][]string) string {
	name := NormalizePaletteName(requested)
	if _, ok := palettes[name]; ok {
		return name
	}
	return DefaultPalette
}

func ResolvePaletteNameStrict(requested string, palettes map[string][]string) (string, bool) {
	name := NormalizePaletteName(requested)
	_, ok := palettes[name]
	return name, ok
}

func NormalizePaletteName(requested string) string {
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

func DefaultBrandConfig() BrandConfig {
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
