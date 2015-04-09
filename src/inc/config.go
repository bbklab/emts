package inc

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

type Config struct {
	SysLoadUplimit int
	JobTimeOut     int // no use for now
	SuperUserNum   int
	SeqRetransRate float64
	UdpLostRate    float64
	SysProcess     *SysProcess
	RecentRestart  int
	IdleRate       float64
	Load           float64
	MemUsage       float64
	CpuUsage       float64
	DiskUsage      *DiskUsage
	ExposedIP      []string
	GMQueueLimit   int64
	QueueLimit     int64
	MailLicense    *MailLicense
	GwLicense      *GwLicense
}

type SysProcess struct {
	TotalSum int
	StateD   int
	StateZ   int
}

type DiskUsage struct {
	Inode float64
	Space float64
}

type MailLicense struct {
	RemainRate float64
	RemainSum  int64
	RemainDay  int64
}

type GwLicense struct {
	RemainDay int64
}

func NewConfig(cfile string) (*Config, error) {
	config := new(Config)
	if content, err := ioutil.ReadFile(cfile); err != nil {
		return nil, fmt.Errorf("E_Read_FAIL on CfgFile: %s", err.Error())
	} else {
		if err := json.Unmarshal(content, &config); err != nil {
			return nil, fmt.Errorf(err.Error())
		}
	}
	return config, nil
}
