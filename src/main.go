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

	runTime := sinfo.Get("systime").Get("runtime").MustFloat64()
	fmt.Println("a:", runTime)
	go checkRuntime(c, runTime, config.RecentRestart)
	n++

	i := 0
	for {
		s := <-c
		if len(s) > 0 {
			fmt.Println(i, ":", s)
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
		c <- fmt.Sprintf("Lost %d System Startups: %v", n, lost)
	} else {
		c <- fmt.Sprintf("%d System Startups Ready", len(must))
	}
}

func checkSuperUser(c chan string, ss []interface{}) {
	n := len(ss)
	if n > 1 {
		c <- fmt.Sprintf("%d System Super Privileged Users", n)
	} else {
		c <- fmt.Sprintf("System Super User OK")
	}
}

func checkSelinux(c chan string, ss map[string]interface{}) {
	if ss["status"] == "enforcing" {
		c <- fmt.Sprintf("Selinux Enforcing")
	} else {
		c <- fmt.Sprintf("Selinux OK")
	}
}

func checkHostName(c chan string, ss map[string]interface{}) {
	if ss["hostname"] == "localhost" || ss["hostname"] == "localhost.localdomain" {
		c <- fmt.Sprintf("ReName the host a Better Name other than [%s]", ss["hostname"])
	} else {
		c <- fmt.Sprintf("Hostname OK")
	}
}

func checkSeqRetransRate(c chan string, ss map[string]interface{}, limit float64) {
	s := ss["seg_retrans_rate"]
	switch value := s.(type) {
	case string:
		if rate, err := strconv.ParseFloat((strings.TrimRight(value, "%")), 64); err == nil {
			if rate >= limit {
				c <- fmt.Sprintf("WARN: Tcp Sequence Lost Rate %f", rate)
			} else {
				c <- fmt.Sprintf("OK: Tcp Sequence Lost Rate %f", rate)
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
				c <- fmt.Sprintf("WARN: Udp Packet Lost Rate %f", rate)
			} else {
				c <- fmt.Sprintf("OK: Udp Packet Lost Rate %f", rate)
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
		c <- fmt.Sprintf("OK: Runing Process Sum %d", s)
	}
}

func checkRuntime(c chan string, s float64, limit int) {
	fmt.Println("running", s)
	fmt.Println("limit", limit)
	d := int(s / float64(3600*24))
	fmt.Println("running-day", d)
	if d <= limit {
		c <- fmt.Sprintf("WARN: OS Restart Recently ?")
	} else {
		c <- fmt.Sprintf("OK: OS has been Running for %d days", d)
	}
}

func output(s string) {
	fmt.Println(trans(s))
}

func trans(s string) string {
	return mo.Gettext(s)
}
