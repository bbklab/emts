package worker

import (
	"fmt"
	"net"
	"strconv"
	"time"
)

type ResolveResp struct {
	// Error   error
	Error   string
	Body    string
	TimeDur float64
}

func Resolve(dest *string, rtype *string, verbose bool) (result ResolveResp) {

	var errresv error

	t1 := time.Now()

	switch *rtype {
	case "a":
		temp := make([]net.IP, 0)
		temp, errresv = net.LookupIP(*dest)
		if errresv != nil {
			result.Error = errresv.Error()
			return
		}
		if verbose {
			result.Body = func([]net.IP) string { // !!! directory use anon function
				body := ""
				for _, ip := range temp {
					body += ip.String() + "\n" // net.IP
				}
				return body
			}(temp) // passing temp as anon function args
		}
	case "ns":
		temp := make([]*net.NS, 0)
		temp, errresv = net.LookupNS(*dest)
		if errresv != nil {
			result.Error = errresv.Error()
			return
		}
		if verbose {
			result.Body = func([]*net.NS) string {
				body := ""
				for _, ns := range temp {
					body += ns.Host + "\n" // net.NS
				}
				return body
			}(temp)
		}
	case "mx":
		temp := make([]*net.MX, 0)
		temp, errresv = net.LookupMX(*dest)
		if errresv != nil {
			result.Error = errresv.Error()
			return
		}
		if verbose {
			result.Body = func([]*net.MX) string {
				body := ""
				for _, mx := range temp {
					body += strconv.Itoa(int(mx.Pref)) + ":" + mx.Host + "\n" // net.MX
				}
				return body
			}(temp)
		}
	case "ptr":
		temp := make([]string, 0)
		temp, errresv = net.LookupAddr(*dest)
		if errresv != nil {
			result.Error = errresv.Error()
			return
		}
		if verbose {
			result.Body = func([]string) string {
				body := ""
				for _, ptr := range temp {
					body += ptr + "\n" // []string
				}
				return body
			}(temp)
		}
	case "txt":
		temp := make([]string, 0)
		temp, errresv = net.LookupTXT(*dest)
		if errresv != nil {
			result.Error = errresv.Error()
			return
		}
		if verbose {
			result.Body = func([]string) string {
				body := ""
				for _, txt := range temp {
					body += txt + "\n"
				}
				return body
			}(temp)
		}
	case "cname":
		var temp string
		temp, errresv = net.LookupCNAME(*dest)
		if errresv != nil {
			result.Error = errresv.Error()
			return
		}
		if verbose {
			result.Body = temp // string
		}
	default:
		result.Error = fmt.Sprintf("dns record type [%s] unsupport", *rtype)
	}

	t2 := time.Now()

	duration := t2.Sub(t1).Seconds()
	result.TimeDur = duration

	return
}
