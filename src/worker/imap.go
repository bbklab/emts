package worker

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"time"
)

type ImapResp struct {
	// Error   error
	Error   string
	Flag    string
	Message string
	TimeDur float64
}

func parseWelcomeLine(s string) (flag string, message string, err error) {
	if s[0] != '*' || s[1] != ' ' {
		err = fmt.Errorf(s)
		return
	}
	if string(s[2:4]) == "OK" {
		flag = string(s[0:4])
		message = string(s[4:])
	} else {
		err = fmt.Errorf(s)
		return
	}
	return
}

func ImapPing(addr *string, verbose bool) (result ImapResp) {

	t1 := time.Now()

	conn, errdial := net.Dial("tcp", *addr)
	if errdial != nil {
		result.Error = errdial.Error()
		return
	}
	defer func() {
		conn.Close()
	}()

	welcome, errread := bufio.NewReader(conn).ReadString('\n') // !!! must be '\n' as ReadString() want byte
	if errread != nil {
		result.Error = fmt.Sprintf("read line on server response failed: [%s]", errread.Error())
		return
	}

	welcome = strings.TrimSpace(welcome)

	flag, message, errparse := parseWelcomeLine(welcome)
	if errparse != nil {
		result.Error = fmt.Sprintf("parse welcome response failed: [%s]", errparse.Error())
		return
	}
	if flag != "* OK" {
		result.Error = fmt.Sprintf(welcome)
		return
	}

	t2 := time.Now()

	duration := t2.Sub(t1).Seconds()
	result.TimeDur = duration

	result.Flag = flag
	if verbose {
		result.Message = strings.TrimRight(message, "\r\n")
	}

	return
}
