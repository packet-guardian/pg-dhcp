// +build !appengine,!js,!windows,!nacl,!plan9

package tty

import (
	"io"
	"os"
)

func CheckIfTerminal(w io.Writer) bool {
	switch v := w.(type) {
	case *os.File:
		return isTerminal(int(v.Fd()))
	default:
		return false
	}
}
