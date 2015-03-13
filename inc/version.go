package inc

import (
	"fmt"
)

const (
	// node version
	VERSION = "1.0.0-alpha"
)

func ShowVersion() {
	fmt.Println(VERSION)
}
