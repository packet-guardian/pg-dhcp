// +build js nacl plan9

package tty

import (
	"io"
)

func CheckIfTerminal(w io.Writer) bool {
	return false
}
