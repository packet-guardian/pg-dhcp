package utils

import (
	"encoding/binary"
	"os"
)

func FileExists(file string) bool {
	_, err := os.Stat(file)
	return !os.IsNotExist(err)
}

func Itob(i int64) []byte {
	out := make([]byte, 8)
	binary.PutVarint(out, i)
	return out
}

func Btoi(in []byte) int64 {
	out, _ := binary.Varint(in)
	return out
}
