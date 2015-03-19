package worker

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
	"time"
)

// Icmp Packet
type Icmp struct {
	Type        uint8
	Code        uint8
	CheckSum    uint16
	Identifier  uint16
	SequenceNum uint16
}

// ping response result
type PingResp struct {
	// Error       error
	Error       string
	SendSum     int
	RecvSum     int
	RecvTime    []float64
	RecvTimeSum float64
	AvgTime     float64
}

func PingIP4(dest *string, sendnum *int, verbose bool) (result PingResp) {
	ldest := net.IPAddr{IP: net.ParseIP("0.0.0.0")}
	rdest, errresv := net.ResolveIPAddr("ip4", *dest) // returned {rdest} is a pointer
	if errresv != nil {
		result.Error = errresv.Error()
		return
	}

	conn, errdial := net.DialIP("ip4:icmp", &ldest, rdest)
	if errdial != nil {
		result.Error = errdial.Error()
		return
	}

	defer conn.Close()

	icmp := Icmp{8, 0, 0, 0, 0} // init a Icmp struct and CheckSum
	var buffer bytes.Buffer
	binary.Write(&buffer, binary.BigEndian, icmp)
	icmp.CheckSum = checkSum(buffer.Bytes()) // set icmp.CheckSum
	buffer.Reset()
	binary.Write(&buffer, binary.BigEndian, icmp) // and rewrite icmp binary into buffer

	// begin to ping
	receive := make([]byte, 1024)
	result.SendSum = 0

	for i := 1; i <= *sendnum; i++ {
		_, errwrite := conn.Write(buffer.Bytes())
		if errwrite != nil { // write request failed
			continue
		}

		result.SendSum++
		t1 := time.Now()

		conn.SetReadDeadline(time.Now().Add(time.Second * 5)) // set timeout as 5 seconds
		_, errread := conn.Read(receive)
		if errread != nil { // receive response failed
			continue
		}

		result.RecvSum++
		t2 := time.Now()

		duration := t2.Sub(t1).Seconds()
		if verbose {
			result.RecvTime = append(result.RecvTime, duration)
		}
		result.RecvTimeSum += duration
	}

	if result.RecvSum > 0 {
		result.AvgTime = float64(result.RecvTimeSum / float64(result.RecvSum))
	} else {
		result.Error = fmt.Sprintf("%s unreachable!", *dest)
		return
	}

	return
}

func checkSum(data []byte) uint16 {
	var (
		sum    uint32
		length int = len(data)
		index  int
	)
	for length > 1 {
		sum += uint32(data[index])<<8 + uint32(data[index+1])
		index += 2
		length -= 2
	}
	if length > 0 {
		sum += uint32(data[index])
	}
	sum += (sum >> 16)

	return uint16(^sum)
}
