package main

import (
	"fmt"
	"os"
)

func debugOutput(file *os.File, bytes []byte) {
	dbgLen := 16
	if len(bytes) < 16 {
		dbgLen = len(bytes)
	}

	file.Write([]byte(fmt.Sprintf("Block %d: %x\n\n", len(bytes), bytes[0:dbgLen])))
}
