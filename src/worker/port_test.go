package worker

import (
	"testing"
)

func Test_PortScan_1(t *testing.T) {
	addr, ptype := "127.0.0.1:22", "tcp"
	result := PortScan(&addr, &ptype, true)
	if result.Error != "" {
		t.Error(result.Error)
		return
	}
	if !result.IsOpen {
		t.Errorf("expect port %s(%s) is open", addr, ptype)
		return
	}
	t.Logf("pass tcp port scan, %s(%s) open", addr, ptype)
}

func Test_PortScan_2(t *testing.T) {
	addr, ptype := "8.8.8.8:53", "udp"
	result := PortScan(&addr, &ptype, true)
	if result.Error != "" {
		t.Error(result.Error)
		return
	}
	if !result.IsOpen {
		t.Errorf("expect port %s(%s) is open", addr, ptype)
		return
	}
	t.Logf("pass udp port scan, %s(%s) open", addr, ptype)
}

func Test_PortScan_3(t *testing.T) {
	addr, ptype := "127.0.0.1:22", "aaa"
	result := PortScan(&addr, &ptype, true)
	if result.Error == "" {
		t.Error("expect error complain on port type unsupport!")
		return
	}
	t.Logf("pass port invalid check: %s", result.Error)
}
