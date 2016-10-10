package dhcp

import (
	"bufio"
	"strings"
	"testing"
)

func TestParser(t *testing.T) {
	// TODO: Actually check the underlying config to make sure it matches the parsed string
	_, err := newParser(bufio.NewReader(strings.NewReader(testConfig))).parse()
	if err != nil {
		t.Fatal(err)
	}
}
