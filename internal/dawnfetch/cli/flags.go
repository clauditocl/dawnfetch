// this file contains cli argument normalization helpers.
package cli

import "strings"

func normalizePreviewArgs(args []string) []string {
	if len(args) < 2 {
		return args
	}
	if !strings.EqualFold(strings.TrimSpace(args[1]), "preview-theme") {
		return args
	}
	out := make([]string, 0, len(args)+2)
	out = append(out, args[0], "--preview-theme")
	rest := args[2:]
	if len(rest) > 0 && !strings.HasPrefix(strings.TrimSpace(rest[0]), "-") {
		out = append(out, "--theme", strings.TrimSpace(rest[0]))
		rest = rest[1:]
	}
	out = append(out, rest...)
	return out
}
