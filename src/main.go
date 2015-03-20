package main

import (
	"fmt"
	mo "github.com/gosexy/gettext"
	"os"

	"inc"
)

func init() {
	// gettext settings
	mo.BindTextdomain(inc.AppName, inc.GetAppRealDirPath()+"/share/locale/")
	mo.Textdomain(inc.AppName)
	os.Setenv("LANGUAGE", "zh_CN.UTF8")
	mo.SetLocale(mo.LC_ALL, "zh_CN.UTF8")
}

func main() {

	if os.Geteuid() != 0 {
		output("require root privileges!")
		os.Exit(1)
	}

	if inc.GetSysLoadavg() >= 0 {
		output("system overload!")
		os.Exit(1)
	}

	os.Exit(0)
}

func output(s string) {
	fmt.Println(trans(s))
}

func trans(s string) string {
	return mo.Gettext(s)
}
