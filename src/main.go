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

	cfgfile := flag.String("c", "conf/config.json", "config file path")
	flag.Parse()
	config, err := inc.NewConfig(*cfgfile)
	if err != nil {
		fmt.Println(err.Error())
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
		output("E_UnMarshal_FAIL on Sinfo: " + err.Error())
		os.Exit(1)
	}

	// Processing Result
	process(jsonsinfo, config)

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
	go checkMemorySize(c, sysBitMode, sysMemSize)
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
			fmt.Printf("%s\n", s)
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
	} else {

		if sysStartups, err := sinfo.Get("startups").StringArray(); err == nil {
			checkMailStartups(sysStartups, []string{"eyou_mail"})
		}

		checkSudoTTY()

		mailSvrAddr := sinfo.Get("epinfo").Get("mail").Get("config").Get("svraddr").MustMap()
		checkMailSvr(mailSvrAddr)

		if mailStartups, err := sinfo.Get("epinfo").Get("mail").Get("startups").StringArray(); err == nil {
			strMailStartups := strings.Join(mailStartups, " ")

			if strings.Contains(strMailStartups, "phpd") { // if mail startups contains phpd,
				mailConfigs := sinfo.Get("epinfo").Get("mail").MustMap()
				checkMailPhpd(mailConfigs, config.GMQueueLimit)
			}

			if strings.Contains(strMailStartups, "remote") || strings.Contains(strMailStartups, "local") {
				checkMailQueue(config.QueueLimit)
			}

			if strings.Contains(strMailStartups, "mysql") ||
				strings.Contains(strMailStartups, "mysql_index") ||
				strings.Contains(strMailStartups, "mysql_log") {
				checkMailMysqlRepl() // don't check if is slave, as caller return nothing if not slave
			}
		}
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
		c <- fmt.Sprintf("WARN: Lost %d System Startups: %v", n, lost)
	} else {
		c <- fmt.Sprintf("SUCC: %d System Startups Ready", len(must))
	}
}

func checkSuperUser(c chan string, ss []interface{}) {
	n := len(ss)
	if n > 1 {
		c <- fmt.Sprintf("WARN: %d System Super Privileged Users", n)
	} else {
		c <- fmt.Sprintf("SUCC: System Super User")
	}
}

func checkSelinux(c chan string, ss map[string]interface{}) {
	if ss["status"] == "enforcing" {
		c <- fmt.Sprintf("CRIT: Selinux Enforcing")
	} else {
		c <- fmt.Sprintf("SUCC: Selinux Closed")
	}
}

func checkHostName(c chan string, hostname string) {
	if hostname == "localhost" || hostname == "localhost.localdomain" {
		c <- fmt.Sprintf("NOTE: ReName the host a Better Name other than [%s]", hostname)
	} else {
		c <- fmt.Sprintf("SUCC: Hostname %s", hostname)
	}
}

func checkMemorySize(c chan string, bitmode string, memsize string) {
	if msize, err := strconv.ParseInt(memsize, 10, 64); err == nil {
		memSize := int64(msize / 1024 / 1024)
		if memSize >= 4 && bitmode == "32" {
			c <- fmt.Sprintf("NOTE: %sbit OS with %dGB Memory", bitmode, memSize)
		} else {
			c <- fmt.Sprintf("SUCC: %sbit OS with %dGB Memory", bitmode, memSize)
		}
	} else {
		c <- ""
	}
}

func checkSwapSize(c chan string, swapsize string) {
	if size, err := strconv.ParseFloat(swapsize, 64); err == nil {
		if size <= 0 {
			c <- fmt.Sprintf("WARN: Swap Size %0.2fGB", size/1024/1024)
		} else {
			c <- fmt.Sprintf("SUCC: Swap Size %0.2fGB", size/1024/1024)
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
				c <- fmt.Sprintf("NOTE: Tcp Sequence Retransfer Rate %0.2f%s", rate, "%")
			} else {
				c <- fmt.Sprintf("SUCC: Tcp Sequence Retransfer Rate %0.2f%s", rate, "%")
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
				c <- fmt.Sprintf("NOTE: Udp Packet Lost Rate %0.2f%s", rate, "%")
			} else {
				c <- fmt.Sprintf("SUCC: Udp Packet Lost Rate %0.2f%s", rate, "%")
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
		c <- fmt.Sprintf("WARN: Running Process Sum %d ", s)
	} else {
		c <- fmt.Sprintf("SUCC: Runing Process Sum %d", s)
	}
}

func checkRuntime(c chan string, s string, limit int) {
	if rtime, err := strconv.ParseFloat(s, 64); err == nil {
		d := int(rtime / float64(3600*24))
		if d <= limit {
			c <- fmt.Sprintf("NOTE: OS Restart Recently ?")
		} else {
			c <- fmt.Sprintf("SUCC: OS has been Running for %d days", d)
		}
	} else {
		c <- ""
	}
}

func checkIdlerate(c chan string, s string, limit float64) {
	if rate, err := strconv.ParseFloat(strings.TrimRight(s, "%"), 64); err == nil {
		if rate <= limit {
			c <- fmt.Sprintf("NOTE: System is Busy, Avg Idle Rate %0.2f%s", rate, "%")
		} else {
			c <- fmt.Sprintf("SUCC: System is Idle, Avg Idle Rate %0.2f%s", rate, "%")
		}
	} else {
		c <- ""
	}
}

func checkLoadnow(c chan string, s string, limit float64) {
	if load, err := strconv.ParseFloat(s, 64); err == nil {
		if load >= limit {
			c <- fmt.Sprintf("WARN: System Load Avg %0.2f", load)
		} else {
			c <- fmt.Sprintf("SUCC: System Load Avg %0.2f", load)
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
		c <- fmt.Sprintf("WARN: Memory Usage %0.2f%s", memusage, "%")
	} else {
		c <- fmt.Sprintf("SUCC: Memory Usage %0.2f%s", memusage, "%")
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
				c <- fmt.Sprintf("WARN: CPU Usage %0.2f%s", usage, "%")
			} else {
				c <- fmt.Sprintf("SUCC: CPU Usage %0.2f%s", usage, "%")
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
			tempstr += fmt.Sprintf(" SpaceUsed %0.2f%s,", spaceUsed, "%")
			if inodeUsed >= limit.Inode {
				warn++
			}
			tempstr += fmt.Sprintf(" InodeUsed %0.2f%s", inodeUsed, "%")
			result += "\n\t" + mountValue + tempstr
		default:
			goto Exit
		}
	}

	if warn > 0 {
		c <- "WARN: Local Disk Space/Inode Usage" + result
	} else {
		c <- "SUCC: Local Disk Space/Inode Usage"
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
		c <- "WARN: Local Disk FSstat/IOTest" + result
	} else {
		c <- "SUCC: Local Disk FSstat/IOTest"
	}

Exit:
	c <- ""
}

func checkDnsbl(s string, cs []string) {
	result := ""
	if len(cs) > 0 { // if exposed address specified by config file
		result = inc.Caller(inc.Checker["dnsbl"], cs)
	} else if len(s) > 0 { // if auto detected exposed address
		result = inc.Caller(inc.Checker["dnsbl"], []string{s})
	}

	warn, rest := parseCheckerOutput(result)
	if warn > 0 {
		fmt.Printf("WARN: %d IPAddress Listed in DNSBL\n%s\n", warn, rest)
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
		fmt.Printf("WARN: Lost %d eYou Product as System Startups: %v\n", n, lost)
	} else {
		fmt.Printf("SUCC: %d eYou Product as System Startups Ready\n", len(must))
	}
}

func checkSudoTTY() {
	if reg, err := regexp.Compile("^Defaults[ \t]*requiretty"); err == nil {
		file := "/etc/sudoers"
		if inc.FGrepBool(file, reg) {
			fmt.Printf("WARN: sudo Require TTY\n")
		} else {
			fmt.Printf("SUCC: sudo Ignore TTY\n")
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
			fmt.Printf("WARN: %d %s Service Fail\n%s\n", warn, strings.ToUpper(svr), rest)
		} else {
			fmt.Printf("SUCC: %s Service\n", strings.ToUpper(svr))
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
		"log": []string{"dblog_mysql_host", "dblog_mysql_port", "dblog_mysql_user", "dblog_mysql_user"},
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
		fmt.Printf("CRIT: %d Mysql Backend Connection Fail\n%s\n", warn, rest)
	} else {
		fmt.Printf("SUCC: Mysql Backend Connection\n")
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
				fmt.Printf("WARN: Gearman Backend Queue %d\n", num)
				details := strings.SplitN(arr[1], ",", -1)
				for _, v := range details {
					if len(strings.TrimSpace(v)) > 0 {
						fmt.Printf("\t%s\n", v)
					}
				}
			} else {
				fmt.Printf("SUCC: Gearman Backend Queue %d\n", num)
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
				fmt.Printf("WARN: Mail Queue %d\n", num)
			} else {
				fmt.Printf("SUCC: Mail Queue %d\n", num)
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
		fmt.Printf("CRIT: %d Mysql Replication Fail\n%s\n", warn, rest)
	} else {
		if len(rest) > 0 { // if indeed have result
			fmt.Printf("SUCC: Mysql Replication\n")
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
	fmt.Println(trans(s))
}

func trans(s string) string {
	return mo.Gettext(s)
}
