// this file is the root package facade for the cli entrypoint.
package dawnfetch

import (
	"dawnfetch/internal/dawnfetch/cli"
	"dawnfetch/internal/dawnfetch/platform"
)

func Run() int {
	return cli.Run()
}

func MaybePauseOnExit(code int) {
	platform.MaybePauseOnExit(code)
}
