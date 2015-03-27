package inc

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

type Config struct {
	SysLoadUplimit int
	SeqRetransRate float64
	UdpLostRate    float64
	ProcessSum     int
	RecentRestart  int
	IdleRate       float64
	Load           float64
	MemUsage       float64
	CpuUsage       float64
	DiskUsage      *DiskUsage
	ExposedIP      []string
	GMQueueLimit   int64
	QueueLimit     int64
}

type DiskUsage struct {
	Inode float64
	Space float64
}

func NewConfig(cfile string) (*Config, error) {
	config := new(Config)
	if content, err := ioutil.ReadFile(cfile); err != nil {
		return nil, fmt.Errorf("E_Read_FAIL on CfgFile: %s", err.Error())
	} else {
		if err := json.Unmarshal(content, &config); err != nil {
			return nil, fmt.Errorf("E_UnMarshal_FAIL on CfgContent: %s", err.Error())
		}
	}
	return config, nil
}
