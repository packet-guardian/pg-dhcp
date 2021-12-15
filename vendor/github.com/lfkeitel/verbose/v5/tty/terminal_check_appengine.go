// +build appengine

package tty

import (
	"io"
)

func CheckIfTerminal(w io.Writer) bool {
	return true
}
