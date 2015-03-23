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
