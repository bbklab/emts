package inc

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

type Config struct {
	LogLevel int
	PidFile  string
	LogFile  string
	IpAllow  []string
	Tcpd     *Tcpd
	Httpd    *Httpd
	JobStore *JobStore
}

type Tcpd struct {
	Enable bool
	Listen []string
}

type Httpd struct {
	Enable bool
	Listen []string
}

type JobStore struct {
	File   *JobStoreFile
	Sqlite *JobStoreSqlite
}

type JobStoreFile struct {
	Rqst string
	Resp string
}

type JobStoreSqlite struct {
}

func NewConfig(cfile string) (*Config, error) {
	config := new(Config)
	if content, err := ioutil.ReadFile(cfile); err != nil {
		return nil, fmt.Errorf("Read Config File Error: %s", err.Error())
	} else {
		if err := json.Unmarshal(content, &config); err != nil {
			return nil, fmt.Errorf("Json Unmarshal Error: %s", err.Error())
		}
	}
	return config, nil
}
