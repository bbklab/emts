package inc

import (
	"os/exec"
)

var (
	CheckerDirPath string = GetAppRealDirPath() + "/c"
)

var Checker map[string]string

func init() {
	Checker = map[string]string{
		"cpu":       CheckerDirPath + "/cpu",
		"dnsbl":     CheckerDirPath + "/dnsbl",
		"emgmqueue": CheckerDirPath + "/emgmqueue",
		"emqueue":   CheckerDirPath + "/emqueue",
		"fsio":      CheckerDirPath + "/fsio",
		"http":      CheckerDirPath + "/http",
		"pop":       CheckerDirPath + "/pop",
		"smtp":      CheckerDirPath + "/smtp",
		"imap":      CheckerDirPath + "/imap",
		"mysqlping": CheckerDirPath + "/mysqlping",
		"mysqlrepl": CheckerDirPath + "/mysqlrepl",
		"tcpconn":   CheckerDirPath + "/tcpconn",
		"iostat":    CheckerDirPath + "/iostat",
	}
}

func Caller(name, args string) string {
	cmd := exec.Command(name, args)
	if output, err := cmd.Output(); err != nil {
		return ""
	} else {
		return string(output)
	}
}
