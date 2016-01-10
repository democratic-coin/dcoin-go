// +build windows

package pb

import (
	"github.com/c-darwin/dcoin-go/vendor/src/github.com/olekukonko/ts"
	"os"
)

var tty = os.Stdin

func terminalWidth() (int, error) {
	size, err := ts.GetSize()
	return size.Col(), err
}
