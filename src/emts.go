package main

import (
	// "encoding/json"
	"bufio"
	"flag"
	"fmt"
	sjson "github.com/bitly/go-simplejson"
	mo "github.com/gosexy/gettext"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"

	"inc"
)

var LevelMap map[string]int = map[string]int{
	"Succ": 0,
	"Note": 0,
	"Warn": 0,
	"Crit": 0,
}

func init() {
	// gettext settings
	mo.BindTextdomain(inc.AppName, inc.GetAppRealDirPath()+"/share/locale/")
	mo.Textdomain(inc.AppName)
	// os.Setenv("LANGUAGE", "en_US.UTF8") // this is not necessary
	mo.SetLocale(mo.LC_ALL, "zh_CN.UTF8")
}

func main() {

	if os.Geteuid() != 0 {
		output("E_Require_Privilege")
		os.Exit(1)
	}

	cfgfile := flag.String("c", inc.GetAppRealDirPath()+"/conf/config.json", "config file path")
	flag.Parse()
	config, err := inc.NewConfig(*cfgfile)
	if err != nil {
		fmt.Println(trans("E_UnMarshal_FAIL on CfgFile"), err.Error())
		os.Exit(1)
	}

	if inc.GetSysLoadavg() >= config.SysLoadUplimit {
		output("E_Sys_OverLoad")
		os.Exit(1)
	}

	output("Collecting ...")
	sinfo := inc.Caller(inc.Sinfo, []string{})
	if sinfo == "" {
		output("E_Collect_FAIL on Sinfo")
		os.Exit(1)
	}
	/* replace by go-simplejson
	var StructSinfo interface{}
	if err := json.Unmarshal([]byte(sinfo), &StructSinfo); err != nil {
		output("E_UnMarshal_FAIL on Sinfo: " + err.Error())
		os.Exit(1)
	}
	*/
	jsonsinfo, err := sjson.NewJson([]byte(sinfo))
	if err != nil {
		fmt.Println(trans("E_UnMarshal_FAIL on Sinfo"), err.Error())
		os.Exit(1)
	}

	// Processing Result
	process(jsonsinfo, config)

	printResultSummary(LevelMap)

	os.Exit(0)
}

func process(sinfo *sjson.Json, config *inc.Config) {
	c := make(chan string)
	n := 0

	if sysStartups, err := sinfo.Get("startups").StringArray(); err == nil {
		must := []string{"network", "sshd"}
		go checkSysStartups(c, sysStartups, must)
		n++
	}

	sysSuperUsers := sinfo.Get("supuser").MustArray()
	go checkSuperUser(c, sysSuperUsers)
	n++

	sysSelinux := sinfo.Get("selinux").MustMap()
	go checkSelinux(c, sysSelinux)
	n++

	sysHostName := sinfo.Get("os_info").Get("hostname").MustString()
	go checkHostName(c, sysHostName)
	n++

	sysBitMode := sinfo.Get("os_info").Get("os_bitmode").MustString()
	sysMemSize := sinfo.Get("mem_info").Get("os_mem_total").MustString()
	sysKernelRls := sinfo.Get("os_info").Get("kernel_release").MustString()
	go checkMemorySize(c, sysBitMode, sysMemSize, sysKernelRls)
	n++

	sysSwapSize := sinfo.Get("mem_info").Get("os_swap_total").MustString()
	go checkSwapSize(c, sysSwapSize)
	n++

	tcp_statistics := sinfo.Get("netstat").Get("tcp_statistics").MustMap()
	go checkSeqRetransRate(c, tcp_statistics, config.SeqRetransRate)
	n++

	udp_statistics := sinfo.Get("netstat").Get("udp_statistics").MustMap()
	go checkUdpLostRate(c, udp_statistics, config.UdpLostRate)
	n++

	processSum := sinfo.Get("process").Get("totalnum").MustInt()
	go checkProcessNum(c, processSum, config.ProcessSum)
	n++

	runTime := sinfo.Get("systime").Get("runtime").MustString()
	go checkRuntime(c, runTime, config.RecentRestart)
	n++

	idleRate := sinfo.Get("systime").Get("idlerate").MustString()
	go checkIdlerate(c, idleRate, config.IdleRate)
	n++

	loadNow := sinfo.Get("sysload").Get("1min").MustString()
	go checkLoadnow(c, loadNow, config.Load)
	n++

	memUsage := sinfo.Get("mem_usage").MustMap()
	go checkMemUsage(c, memUsage, config.MemUsage)
	n++

	cpuUsage := sinfo.Get("cpu_usage").MustMap()
	go checkCpuUsage(c, cpuUsage, config.CpuUsage)
	n++

	diskUsage := sinfo.Get("disk_space").MustArray()
	go checkDiskUsage(c, diskUsage, config.DiskUsage)
	n++

	diskFsio := sinfo.Get("disk_fsio").MustMap()
	go checkDiskFsio(c, diskFsio)
	n++

	i := 0
	for {
		s := <-c
		i++
		if len(s) > 0 {
			fmt.Println(s)
		}
		if i >= n {
			break
		}
	}

	/* Following using inc.Caller to run other command
	   If run by goroutine, will lead to nothing returned
	*/
	exposedAddr := sinfo.Get("epinfo").Get("common").Get("exposed").MustString()
	checkDnsbl(exposedAddr, config.ExposedIP)

	mailIsInstalled := sinfo.Get("epinfo").Get("mail").Get("is_installed").MustInt()
	if mailIsInstalled == 0 {
		return
	}

	/*
		begin eyou mail related check
	*/
	if sysStartups, err := sinfo.Get("startups").StringArray(); err == nil {
		checkMailStartups(sysStartups, []string{"eyou_mail"})
	}

	checkSudoTTY()

	mailSvrAddr := sinfo.Get("epinfo").Get("mail").Get("config").Get("svraddr").MustMap()
	checkMailSvr(mailSvrAddr)

	// get eyou mail startups
	arrMailStartups, err := sinfo.Get("epinfo").Get("mail").Get("startups").StringArray()
	if err != nil {
		return
	}
	strMailStartups := strings.Join(arrMailStartups, " ")

	// if mail startups contains phpd
	if strings.Contains(strMailStartups, "phpd") {
		mailConfigs := sinfo.Get("epinfo").Get("mail").MustMap()
		checkMailPhpd(mailConfigs, config.GMQueueLimit)
	}

	// if mail startups contains remote or local
	if strings.Contains(strMailStartups, "remote") || strings.Contains(strMailStartups, "local") {
		checkMailQueue(config.QueueLimit)
	}

	// if mail startups contains mysql backedn
	if strings.Contains(strMailStartups, "mysql") ||
		strings.Contains(strMailStartups, "mysql_index") ||
		strings.Contains(strMailStartups, "mysql_log") {
		checkMailMysqlRepl() // don't check if is slave, as caller return nothing if not slave
	}
}

func checkSysStartups(c chan string, ss []string, must []string) {
	lost := make([]string, 0)
	for _, v := range must {
		isLost := true
		for _, s := range ss {
			if v == s {
				isLost = false
				break
			}
		}
		if isLost {
			lost = append(lost, v)
		}
	}
	n := len(lost)
	if n > 0 {
		c <- _warn(fmt.Sprintf(trans("Lost %d System Startups: %v"), n, lost))
	} else {
		c <- _succ(fmt.Sprintf(trans("%d System Startups Ready"), len(must)))
	}
}

func checkSuperUser(c chan string, ss []interface{}) {
	n := len(ss)
	if n > 1 {
		c <- _warn(fmt.Sprintf(trans("%d System Super Privileged Users"), n))
	} else {
		c <- _succ(fmt.Sprintf(trans("System Super User")))
	}
}

func checkSelinux(c chan string, ss map[string]interface{}) {
	if ss["status"] == "enforcing" {
		c <- _crit(fmt.Sprintf(trans("Selinux Enforcing")))
	} else {
		c <- _succ(fmt.Sprintf(trans("Selinux Closed")))
	}
}

func checkHostName(c chan string, hostname string) {
	if hostname == "localhost" || hostname == "localhost.localdomain" {
		c <- _note(fmt.Sprintf(trans("ReName the host a Better Name other than [%s]"), hostname))
	} else {
		c <- _succ(fmt.Sprintf(trans("Hostname [%s]"), hostname))
	}
}

func checkMemorySize(c chan string, bitmode, memsize, kernelrls string) {
	if msize, err := strconv.ParseInt(memsize, 10, 64); err == nil {
		memSize := int64(msize / 1024 / 1024)
		if memSize >= 4 && bitmode == "32" {
			if strings.Contains(kernelrls, "PAE") {
				c <- _succ(fmt.Sprintf(trans("%sbit OS with %dGB Memory and PAE Kernel"), bitmode, memSize))
			} else {
				c <- _note(fmt.Sprintf(trans("%sbit OS with %dGB Memory and without PAE Kernel"), bitmode, memSize))
			}
		} else {
			c <- _succ(fmt.Sprintf(trans("%sbit OS with %dGB Memory"), bitmode, memSize))
		}
	} else {
		c <- ""
	}
}

func checkSwapSize(c chan string, swapsize string) {
	if size, err := strconv.ParseFloat(swapsize, 64); err == nil {
		if size <= 0 {
			c <- _warn(fmt.Sprintf(trans("Swap Size %0.2fGB"), size/1024/1024))
		} else {
			c <- _succ(fmt.Sprintf(trans("Swap Size %0.2fGB"), size/1024/1024))
		}
	}
	c <- ""
}

func checkSeqRetransRate(c chan string, ss map[string]interface{}, limit float64) {
	s := ss["seg_retrans_rate"]
	switch value := s.(type) {
	case string:
		if rate, err := strconv.ParseFloat((strings.TrimRight(value, "%")), 64); err == nil {
			if rate >= limit {
				c <- _note(fmt.Sprintf(trans("Tcp Sequence Retransfer Rate %0.2f%s"), rate, "%"))
			} else {
				c <- _succ(fmt.Sprintf(trans("Tcp Sequence Retransfer Rate %0.2f%s"), rate, "%"))
			}
		} else {
			c <- ""
		}
	default:
		c <- ""
	}
}

func checkUdpLostRate(c chan string, ss map[string]interface{}, limit float64) {
	s := ss["packet_lostrate"]
	switch value := s.(type) {
	case string:
		if rate, err := strconv.ParseFloat((strings.TrimRight(value, "%")), 64); err == nil {
			if rate >= limit {
				c <- _note(fmt.Sprintf(trans("Udp Packet Lost Rate %0.2f%s"), rate, "%"))
			} else {
				c <- _succ(fmt.Sprintf(trans("Udp Packet Lost Rate %0.2f%s"), rate, "%"))
			}
		} else {
			c <- ""
		}
	default:
		c <- ""
	}
}

func checkProcessNum(c chan string, s int, limit int) {
	if s >= limit {
		c <- _warn(fmt.Sprintf(trans("Running Process Sum %d"), s))
	} else {
		c <- _succ(fmt.Sprintf(trans("Running Process Sum %d"), s))
	}
}

func checkRuntime(c chan string, s string, limit int) {
	if rtime, err := strconv.ParseFloat(s, 64); err == nil {
		d := int(rtime / float64(3600*24))
		if d <= limit {
			c <- _note(fmt.Sprintf(trans("OS Restart Recently ?")))
		} else {
			c <- _succ(fmt.Sprintf(trans("OS has been Running for %d days"), d))
		}
	} else {
		c <- ""
	}
}

func checkIdlerate(c chan string, s string, limit float64) {
	if rate, err := strconv.ParseFloat(strings.TrimRight(s, "%"), 64); err == nil {
		if rate <= limit {
			c <- _note(fmt.Sprintf(trans("System is Busy, Avg Idle Rate %0.2f%s"), rate, "%"))
		} else {
			c <- _succ(fmt.Sprintf(trans("System is Idle, Avg Idle Rate %0.2f%s"), rate, "%"))
		}
	} else {
		c <- ""
	}
}

func checkLoadnow(c chan string, s string, limit float64) {
	if load, err := strconv.ParseFloat(s, 64); err == nil {
		if load >= limit {
			c <- _warn(fmt.Sprintf(trans("System Load Avg %0.2f"), load))
		} else {
			c <- _succ(fmt.Sprintf(trans("System Load Avg %0.2f"), load))
		}
	} else {
		c <- ""
	}
}

func checkMemUsage(c chan string, ss map[string]interface{}, limit float64) {
	var memused int64
	var memtotal int64
	var memusage float64

	switch value := ss["mem_total"].(type) {
	case string:
		if temp, err := strconv.ParseInt(value, 10, 64); err == nil {
			memtotal = temp
			memused = temp
		} else {
			goto Exit
		}
	default:
		goto Exit
	}
	for _, v := range []string{"mem_free", "mem_cache", "mem_buffer"} {
		switch value := ss[v].(type) {
		case string:
			if temp, err := strconv.ParseInt(value, 10, 64); err == nil {
				memused -= temp
			} else {
				goto Exit
			}
		default:
			goto Exit
		}
	}

	memusage = float64(memused * 100 / memtotal)
	if memusage >= limit {
		c <- _warn(fmt.Sprintf(trans("Memory Usage %0.2f%s"), memusage, "%"))
	} else {
		c <- _succ(fmt.Sprintf(trans("Memory Usage %0.2f%s"), memusage, "%"))
	}

Exit:
	c <- ""
}

func checkCpuUsage(c chan string, ss map[string]interface{}, limit float64) {
	idle := ss["id"]
	switch value := idle.(type) {
	case string:
		if id, err := strconv.ParseFloat(value, 64); err == nil {
			usage := float64(100 - id)
			if usage >= limit {
				c <- _warn(fmt.Sprintf(trans("CPU Usage %0.2f%s"), usage, "%"))
			} else {
				c <- _succ(fmt.Sprintf(trans("CPU Usage %0.2f%s"), usage, "%"))
			}
		} else {
			goto Exit
		}
	default:
		goto Exit
	}

Exit:
	c <- ""
}

func checkDiskUsage(c chan string, ss []interface{}, limit *inc.DiskUsage) {
	result := ""
	warn := 0

	for _, s := range ss {

		disk := make(map[string]interface{}) // s is an empty interface{}
		switch a := s.(type) {
		case map[string]interface{}:
			disk = a
		default:
			goto Exit
		}

		mount := disk["mount"] // mount is an empty interface{}
		switch mountValue := mount.(type) {
		case string:
			if mountValue == "/boot" {
				continue
			}
			diskInfo := make(map[string]int64)
			// convert all filed as float64 and saved in map diskInfo
			for _, v := range []string{"msize", "mused", "isize", "iused"} {
				switch value := disk[v].(type) {
				case string:
					if temp, err := strconv.ParseInt(value, 10, 64); err == nil {
						diskInfo[v] = temp
					} else {
						goto Exit
					}
				default:
					goto Exit
				}
			}
			spaceUsed := float64(100 * diskInfo["mused"] / diskInfo["msize"])
			inodeUsed := float64(100 * diskInfo["iused"] / diskInfo["isize"])
			tempstr := ""
			if spaceUsed >= limit.Space {
				warn++
			}
			tempstr += fmt.Sprintf(trans(" SpaceUsed %0.2f%s,"), spaceUsed, "%")
			if inodeUsed >= limit.Inode {
				warn++
			}
			tempstr += fmt.Sprintf(trans(" InodeUsed %0.2f%s"), inodeUsed, "%")
			result += "\n\t" + mountValue + tempstr
		default:
			goto Exit
		}
	}

	if warn > 0 {
		c <- _warn(fmt.Sprintf(trans("Local Disk Space/Inode Usage"))) + result
	} else {
		c <- _succ(fmt.Sprintf(trans("Local Disk Space/Inode Usage")))
	}

Exit:
	c <- ""
}

func checkDiskFsio(c chan string, ss map[string]interface{}) {
	result := ""
	warn := 0

	fsstat := ss["fsstat"]
	iotest := ss["iotest"]
	switch a := fsstat.(type) {
	case map[string]interface{}:
		for dev, stat := range a {
			switch v := stat.(type) {
			case string:
				if v != "clean" {
					warn++
					result += "\n\t" + dev + " " + v
				}
			default:
				goto Exit
			}
		}
	default:
		goto Exit
	}

	switch a := iotest.(type) {
	case map[string]interface{}:
		for mount, stat := range a {
			switch v := stat.(type) {
			case string:
				if v != "succ" {
					warn++
					result += "\n\t" + mount + " " + v
				}
			default:
				goto Exit
			}
		}
	default:
		goto Exit
	}

	if warn > 0 {
		c <- _warn(trans("Local Disk FSstat/IOTest")) + result
	} else {
		c <- _succ(trans("Local Disk FSstat/IOTest"))
	}

Exit:
	c <- ""
}

func checkDnsbl(s string, cs []string) {
	result := ""
	ips := []string{}
	if len(cs) > 0 { // if exposed address specified by config file
		ips = cs
	} else if len(s) > 0 { // if auto detected exposed address
		ips = append(ips, s)
	} else {
		return
	}
	result = inc.Caller(inc.Checker["dnsbl"], ips)

	warn, rest := parseCheckerOutput(result)
	if warn > 0 {
		fmt.Printf(_warn(trans("%d Exposed IPAddress Listed in DNSBL\n%s\n")), warn, rest)
	} else {
		if len(rest) > 0 {
			fmt.Printf(_succ(trans("%d Exposed IPAddress NOT Listed in DNSBL\n")), len(ips))
		}
	}
}

func checkMailStartups(ss []string, must []string) {
	lost := make([]string, 0)
	for _, v := range must {
		isLost := true
		for _, s := range ss {
			if v == s {
				isLost = false
				break
			}
		}
		if isLost {
			lost = append(lost, v)
		}
	}
	n := len(lost)
	if n > 0 {
		fmt.Printf(_warn(trans("Lost %d eYou Product as System Startups: %v\n")), n, lost)
	} else {
		fmt.Printf(_succ(trans("%d eYou Product as System Startups Ready\n")), len(must))
	}
}

func checkSudoTTY() {
	if reg, err := regexp.Compile("^Defaults[ \t]*requiretty"); err == nil {
		file := "/etc/sudoers"
		if inc.FGrepBool(file, reg) {
			fmt.Printf(_warn(trans("sudo Require TTY\n")))
		} else {
			fmt.Printf(_succ(trans("sudo Ignore TTY\n")))
		}
	}
}

func checkMailSvr(ss map[string]interface{}) {
	for svr, addr := range ss {
		switch addrs := addr.(type) {
		case string:
			checkMtaSvr(svr, addrs)
		}
	}
}

func checkMtaSvr(svr string, addrs string) {
	args := strings.SplitN(addrs, " ", -1)
	if svr == "smtp" || svr == "pop" || svr == "imap" || svr == "http" {
		result := inc.Caller(inc.Checker[svr], args)
		warn, rest := parseCheckerOutput(result)
		if warn > 0 {
			fmt.Printf(_warn(trans("%d %s Mail Service Fail\n%s\n")), warn, strings.ToUpper(svr), rest)
		} else {
			fmt.Printf(_succ(trans("%s Mail Service\n")), strings.ToUpper(svr))
		}
	}
}

func checkMailPhpd(mailcfg map[string]interface{}, GMQueueLimit int64) {
	mailcfg_tools := mailcfg["tools"]
	mailcfg_config := mailcfg["config"]
	var mailMysqlCLI, mailMysqlAdmin string
	var mailUsrMysql, mailIdxMysql, mailLogMysql map[string]interface{}
	switch tools := mailcfg_tools.(type) {
	case map[string]interface{}:
		switch v := tools["mysqlcli"].(type) {
		case string:
			mailMysqlCLI = v
		}
		switch v := tools["mysqladmin"].(type) {
		case string:
			mailMysqlAdmin = v
		}
	}
	switch config := mailcfg_config.(type) {
	case map[string]interface{}:
		switch v := config["usrdb"].(type) {
		case map[string]interface{}:
			mailUsrMysql = v
		}
		switch v := config["idxdb"].(type) {
		case map[string]interface{}:
			mailIdxMysql = v
		}
		switch v := config["logdb"].(type) {
		case map[string]interface{}:
			mailLogMysql = v
		}
	}
	checkMailDBSvr(mailMysqlAdmin, mailUsrMysql, mailIdxMysql, mailLogMysql)
	checkMailGMSvr(mailMysqlCLI, mailUsrMysql, GMQueueLimit)
}

func checkMailDBSvr(mysqladmin string, userdb, idxdb, logdb map[string]interface{}) {
	if mysqladmin == "" {
		return
	}
	args := make([]string, 0)
	args = append(args, mysqladmin)
	dbcfg := map[string][]string{
		"usr": []string{"db_mysql_host", "db_mysql_port", "db_mysql_user", "db_mysql_pass",
			"mta_db_mysql_host", "mta_db_mysql_port", "mta_db_mysql_user", "mta_db_mysql_pass",
		},
		"idx": []string{"dbumi_mysql_dsn", "dbumi_mysql_user", "dbumi_mysql_pass",
			"mta_dbumi_mysql_dsn", "mta_dbumi_mysql_user", "mta_dbumi_mysql_pass",
		},
		"log": []string{"dblog_mysql_host", "dblog_mysql_port", "dblog_mysql_user", "dblog_mysql_pass"},
	}
	for name, conf := range dbcfg {
		temp := ""
		switch name {
		case "usr":
			for i, _ := range conf {
				switch v := userdb[conf[i]].(type) {
				case string:
					if i%4 == 3 {
						temp += v
						args = append(args, temp)
						temp = ""
					} else {
						temp += v + ","
					}
				}
			}
		case "log":
			for i, _ := range conf {
				switch v := logdb[conf[i]].(type) {
				case string:
					if i%4 == 3 {
						temp += v
						args = append(args, temp)
						temp = ""
					} else {
						temp += v + ","
					}
				}
			}
		case "idx":
			dsnhead := []string{}
			user := ""
			pass := ""
			for i, _ := range conf {
				if i%3 == 0 { // parse dsn
					switch v := idxdb[conf[i]].(type) {
					case []interface{}:
						for _, dsn := range v {
							switch vdsn := dsn.(type) {
							case string:
								if strings.Contains(vdsn, "host=") {
									arr := strings.Split(vdsn, ";")
									if len(arr) >= 3 {
										host := strings.Replace(arr[0], "host=", "", -1)
										port := strings.Replace(arr[1], "port=", "", -1)
										dsnhead = append(dsnhead, host+","+port)
									}
								} else if strings.Contains(vdsn, "unix_socket=") {
									arr := strings.Split(vdsn, ";")
									if len(arr) >= 2 {
										unixsock := strings.Replace(arr[0], "unix_socket=", "", -1)
										dsnhead = append(dsnhead, unixsock)
									}
								}
							}
						}
					}
				} else if i%3 == 1 { // parse user
					switch v := idxdb[conf[i]].(type) {
					case string:
						user = v
					}
				} else if i%3 == 2 { // parse pass
					switch v := idxdb[conf[i]].(type) {
					case string:
						pass = v
						for _, head := range dsnhead {
							args = append(args, head+","+user+","+pass)
						}
						dsnhead = []string{} // emtpy dsnhead []
					}
				}
			}
		}
	}
	// Oh! finally finished! WTF!
	result := inc.Caller(inc.Checker["mysqlping"], args)
	warn, rest := parseCheckerOutput(result)
	if warn > 0 {
		fmt.Printf(_crit(trans("%d Mysql Backend Connection Fail\n%s\n")), warn, rest)
	} else {
		fmt.Printf(_succ(trans("%d Mysql Backend Connection\n")), len(args)-1)
	}
}

func checkMailGMSvr(mysqlcli string, userdb map[string]interface{}, limit int64) {
	if mysqlcli == "" {
		return
	}
	args := make([]string, 0)
	args = append(args, mysqlcli)
	dbcfg := []string{"db_mysql_host", "db_mysql_port", "db_mysql_user", "db_mysql_pass", "db_name"}
	temp := ""
	for _, v := range dbcfg {
		switch vv := userdb[v].(type) {
		case string:
			if len(temp) > 0 {
				temp += "," + vv
			} else {
				temp += vv
			}
		}
	}
	args = append(args, temp)
	result := inc.Caller(inc.Checker["emgmqueue"], args)
	arr := strings.SplitN(result, " ", 2)
	if len(arr) >= 2 {
		if num, err := strconv.ParseInt(arr[0], 10, 64); err == nil {
			if num >= limit {
				fmt.Printf(_warn(trans("Gearman Backend Queue %d\n")), num)
				details := strings.SplitN(arr[1], ",", -1)
				for _, v := range details {
					if len(strings.TrimSpace(v)) > 0 {
						fmt.Printf("\t%s\n", v)
					}
				}
			} else {
				fmt.Printf(_succ(trans("Gearman Backend Queue %d\n")), num)
			}
		}
	}
}

func checkMailQueue(limit int64) {
	result := inc.Caller(inc.Checker["emqueue"], []string{})
	arr := strings.SplitN(result, " ", -1)
	if len(arr) >= 1 {
		if num, err := strconv.ParseInt(arr[0], 10, 64); err == nil {
			if num >= limit {
				fmt.Printf(_warn(trans("Mail Queue %d\n")), num)
			} else {
				fmt.Printf(_succ(trans("Mail Queue %d\n")), num)
			}
		}
	}
}

func checkMailMysqlRepl() {
	// use fix default settings here
	args := []string{"/usr/local/eyou/mail/opt/mysql/bin/mysql",
		"127.0.0.1,3306,eyou,eyou",
		"127.0.0.1,3316,eyou,eyou",
		"127.0.0.1,3326,eyou,eyou",
	}
	result := inc.Caller(inc.Checker["mysqlrepl"], args)
	warn, rest := parseCheckerOutput(result)
	if warn > 0 {
		fmt.Printf(_crit(trans("%d Mysql Replication Fail\n%s\n")), warn, rest)
	} else {
		if len(rest) > 0 { // if indeed have result
			fmt.Printf(_succ(trans("%d Mysql Replication\n")), len(args)-1)
		}
	}
}

func parseCheckerOutput(s string) (int, string) {
	warn := 0
	result := ""
	r := strings.NewReader(s) // returned string.Reader implement io.Reader
	buf := bufio.NewReader(r) // use bufio to scan the result
	for {
		line, err := buf.ReadBytes('\n')
		if err != nil { // include io.EOF
			if err == io.EOF {
				break
			} else {
				continue
			}
		}
		sline := strings.TrimRight(string(line), "\n")
		if len(sline) > 0 {
			arrline := strings.SplitN(sline, " ", 3)
			if len(arrline) >= 2 {
				if arrline[1] == "warn" {
					warn++
					if len(arrline) >= 3 {
						if len(result) > 0 {
							result += "\n\t" + arrline[0] + " - " + arrline[2]
						} else {
							result += "\t" + arrline[0] + " - " + arrline[2]
						}
					}
				}
			}
		} else {
			continue
		}
	}
	return warn, result
}

func output(s string) {
	fmt.Printf("%s\n", trans(s))
}

func trans(s string) string {
	return mo.Gettext(s)
}

func _succ(s string) string {
	LevelMap["Succ"]++
	return trans("SUCC: ") + s
}
func _note(s string) string {
	LevelMap["Note"]++
	return _yellow(trans("NOTE: ") + s)
}
func _warn(s string) string {
	LevelMap["Warn"]++
	return _red(trans("WARN: ") + s)
}
func _crit(s string) string {
	LevelMap["Crit"]++
	return _purple(trans("CRIT: ") + s)
}
func _yellow(s string) string {
	return "\033[1;33m" + s + "\033[0m"
}
func _red(s string) string {
	return "\033[1;31m" + s + "\033[0m"
}
func _purple(s string) string {
	return "\033[1;35m" + s + "\033[0m"
}
func _green(s string) string {
	return "\033[1;32m" + s + "\033[0m"
}

func printResultSummary(s map[string]int) {
	score := 100 - 40*s["Crit"] - 20*s["Warn"] - 5*s["Note"]
	if score < 0 {
		score = 0
	}
	fmt.Printf("\n------\n")
	fmt.Printf(trans("Result: %s:%s, %s:%s, %s:%s, %s:%s\nScore: %d\n"),
		trans("SUCC"), _green(strconv.Itoa(s["Succ"])),
		trans("NOTE"), _yellow(strconv.Itoa(s["Note"])),
		trans("WARN"), _red(strconv.Itoa(s["Warn"])),
		trans("CRIT"), _purple(strconv.Itoa(s["Crit"])),
		score,
	)
}
