package inc

var (
	CheckerDirPath string = GetAppRealDirPath() + "/c"
	SinfoDirPath   string = GetAppRealDirPath() + "/sinfo"
	Sinfo          string = SinfoDirPath + "/sinfo"
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
		"memcache":  CheckerDirPath + "/memcache",
	}
}
