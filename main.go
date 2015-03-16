package main

import (
	"fmt"
	//mo "github.com/gosexy/gettext"
	"os"

	"inc"
)

func main() {
	fmt.Println(inc.GetAppRealPath())
	inc.ShowVersion()
	os.Exit(0)
}
