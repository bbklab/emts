package main

import (
	// "encoding/json"
	"flag"
	"fmt"
	sjson "github.com/bitly/go-simplejson"
	mo "github.com/gosexy/gettext"
	"os"
	"strconv"
	"strings"

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
	sinfo := inc.Caller(inc.Sinfo, "")
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
		go checkSysStartups(c, sysStartups)
		n++
	}

	sysSuperUsers := sinfo.Get("supuser").MustArray()
	go checkSuperUser(c, sysSuperUsers)
	n++

	sysSelinux := sinfo.Get("selinux").MustMap()
	go checkSelinux(c, sysSelinux)
	n++

	sysHostName := sinfo.Get("os_info").MustMap()
	go checkHostName(c, sysHostName)
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

	diskUsage := sinfo.Get("disk_space").MustArray()
	go checkDiskUsage(c, diskUsage, config.DiskUsage)
	n++

	i := 0
	for {
		s := <-c
		if len(s) > 0 {
			fmt.Printf("%-2d: %s\n", i, s)
		}
		i++
		if i >= n {
			break
		}
	}
}

func checkSysStartups(c chan string, ss []string) {
	must := []string{"eyou_mail", "sshd", "network"}
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

func checkHostName(c chan string, ss map[string]interface{}) {
	if ss["hostname"] == "localhost" || ss["hostname"] == "localhost.localdomain" {
		c <- fmt.Sprintf("NOTE: ReName the host a Better Name other than [%s]", ss["hostname"])
	} else {
		c <- fmt.Sprintf("SUCC: Hostname %s", ss["hostname"])
	}
}

func checkSeqRetransRate(c chan string, ss map[string]interface{}, limit float64) {
	s := ss["seg_retrans_rate"]
	switch value := s.(type) {
	case string:
		if rate, err := strconv.ParseFloat((strings.TrimRight(value, "%")), 64); err == nil {
			if rate >= limit {
				c <- fmt.Sprintf("NOTE: Tcp Sequence Retransfer Rate %0.2f", rate)
			} else {
				c <- fmt.Sprintf("SUCC: Tcp Sequence Retransfer Rate %0.2f", rate)
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
				c <- fmt.Sprintf("NOTE: Udp Packet Lost Rate %0.2f", rate)
			} else {
				c <- fmt.Sprintf("SUCC: Udp Packet Lost Rate %0.2f", rate)
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
		if rate >= limit {
			c <- fmt.Sprintf("NOTE: System Idle Rate %0.2f", rate)
		} else {
			c <- fmt.Sprintf("SUCC: System Idle Rate %0.2f", rate)
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
		c <- fmt.Sprintf("WARN: Memory Usage %0.2f", memusage)
	} else {
		c <- fmt.Sprintf("SUCC: Memory Usage %0.2f", memusage)
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
			result += "\t" + mountValue + tempstr + "\n"
		default:
			goto Exit
		}
	}

	if warn > 0 {
		c <- "WARN: Disk Usage\n" + result
	} else {
		c <- "SUCC: Disk Usage\n" + result
	}

Exit:
	c <- ""
}

func output(s string) {
	fmt.Println(trans(s))
}

func trans(s string) string {
	return mo.Gettext(s)
}
