// this file defines shared data types used across collectors and rendering.
package dawnfetch

type Field struct {
	Label string `json:"label"`
	Value string `json:"value"`
}

type BrandConfig struct {
	Palettes map[string][]string `json:"palettes"`
	Logos    map[string]LogoSet  `json:"logos"`
}

type LogoSet struct {
	Normal  []LogoLine `json:"normal"`
	Compact []LogoLine `json:"compact"`
	Tiny    []LogoLine `json:"tiny"`
}

type LogoLine struct {
	Text       string `json:"text"`
	ColorIndex int    `json:"color_index"`
	RawANSI    bool   `json:"raw_ansi,omitempty"`
}

type RenderedLine struct {
	Raw    string
	Styled string
}

type osReleaseInfo struct {
	ID         string
	IDLike     string
	PrettyName string
}

const defaultPalette = "transgender"
