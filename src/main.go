package main

import (
	"fmt"
	mo "github.com/gosexy/gettext"
	"os"

	"inc"
)

func main() {
	mo.BindTextdomain(inc.AppName, "./share/locale/")
	mo.Textdomain(inc.AppName)
	os.Setenv("LANGUAGE", "zh_CN.UTF8")
	mo.SetLocale(mo.LC_ALL, "zh_CN.UTF8")

	fmt.Println(inc.GetAppRealPath())
	inc.ShowVersion()

	if os.Geteuid() != 0 {
		fmt.Println(mo.Gettext("require root privileges!"))
		os.Exit(1)
	}

	os.Exit(0)
}
