// this file contains shared pause behavior used on all platforms.
package platform

import (
	"bufio"
	"fmt"
	"os"
)

func MaybePauseOnExit(_ int) {
	if !shouldPauseOnExit() {
		return
	}
	fmt.Print("\nPress Enter to exit...")
	_, _ = bufio.NewReader(os.Stdin).ReadString('\n')
}
