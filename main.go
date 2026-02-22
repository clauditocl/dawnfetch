// this file is the cli entrypoint and delegates to the dawnfetch runner.
package main

import (
	"os"

	"dawnfetch/internal/dawnfetch"
)

func main() {
	code := dawnfetch.Run()
	dawnfetch.MaybePauseOnExit(code)
	os.Exit(code)
}
