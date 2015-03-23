package main

import (
	"encoding/json"
	"flag"
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
		output("E_Require_Privilege")
		os.Exit(1)
	}

	cfgfile := flag.String("c", "conf/config.json", "config file path")
	config, err := inc.NewConfig(*cfgfile)
	if err != nil {
		output(err.Error())
		os.Exit(1)
	}

	if inc.GetSysLoadavg() >= config.SysLoadUplimit {
		output("E_Sys_OverLoad")
		os.Exit(1)
	}

	sinfo := inc.Caller(inc.Sinfo, "")
	var StructSinfo interface{}
	if err := json.Unmarshal([]byte(sinfo), &StructSinfo); err != nil {
		output("E_UnMarshal_FAIL on Sinfo: " + err.Error())
		os.Exit(1)
	}
	fmt.Println(StructSinfo)

	os.Exit(0)
}

func output(s string) {
	fmt.Println(trans(s))
}

func trans(s string) string {
	return mo.Gettext(s)
}
