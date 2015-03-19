package worker

import (
	"fmt"
	"net"
	"time"
)

type PortResp struct {
	// Error error
	Error   string
	IsOpen  bool
	TimeDur float64
}

func PortScan(dest *string, ptype *string, verbose bool) (result PortResp) {

	t1 := time.Now()

	switch *ptype {
	case "tcp":
		_, errdial := net.Dial(*ptype, *dest)
		if errdial != nil {
			result.IsOpen = false
		} else {
			result.IsOpen = true
		}
	case "udp":
		_, errdial := net.Dial(*ptype, *dest)
		if errdial != nil {
			result.IsOpen = false
		} else {
			result.IsOpen = true
		}
	default:
		// result.Error = fmt.Errorf("port type [%s] unsupport", *ptype)
		result.Error = fmt.Sprintf("port type [%s] unsupport", *ptype)
		return
	}

	t2 := time.Now()

	duration := t2.Sub(t1).Seconds()
	result.TimeDur = duration

	return
}
