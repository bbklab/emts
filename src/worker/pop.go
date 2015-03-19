package worker

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"time"
	// "regexp"
)

type PopResp struct {
	// Error   error
	Error   string
	Flag    string
	Message string
	TimeDur float64
}

/*
func isLikePopResp(s *string) bool {
	news := strings.TrimRight(*s, "\n") // !!! trim first
	popregex := regexp.MustCompile(`^(\+OK)|(\-ERR) .+$`)
	return popregex.MatchString(news)
}
*/

func parseFlagLine(s string) (flag string, message string, err error) {
	if string(s[0:3]) == "+OK" {
		flag = string(s[0:3])
		message = string(s[3:])
	} else if string(s[0:4]) == "-ERR" {
		flag = string(s[0:4])
		message = string(s[4:])
	} else {
		err = fmt.Errorf(s)
		return
	}
	return
}

func PopPing(addr *string, verbose bool) (result PopResp) {

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

	flag, message, errparse := parseFlagLine(welcome)
	if errparse != nil {
		result.Error = fmt.Sprintf("parse welcome response failed: [%s]", errparse.Error())
		return
	}
	if flag != "+OK" {
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
