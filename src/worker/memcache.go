package worker

import (
	"bufio"
	"io"
	"net"
	"strings"
	"time"
)

type MemcacheResp struct {
	Error   string
	Stat    bool
	TimeDur float64
}

func MemcachePing(addr *string, verbose bool) (result MemcacheResp) {

	t1 := time.Now()

	conn, errdial := net.Dial("tcp", *addr)
	if errdial != nil {
		result.Error = errdial.Error()
		return
	}
	defer func() {
		conn.Close()
	}()

	if _, err := conn.Write([]byte("stats\r\n")); err != nil {
		result.Error = err.Error()
		return
	}

	response := ""
	buf := bufio.NewReader(conn)
	for {
		line, err := buf.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			result.Error = err.Error()
			return
		}
		response += line
		if strings.TrimSpace(line) == "END" ||
			strings.TrimSpace(line) == "ERROR" {
			break
		}
	}

	t2 := time.Now()

	duration := t2.Sub(t1).Seconds()
	result.TimeDur = duration

	if strings.Contains(response, "ERROR") {
		result.Error = string(response)
		result.Stat = false
		return
	} else {
		result.Stat = true
	}

	return
}
