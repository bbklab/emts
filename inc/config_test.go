package inc

import (
	"encoding/json"
	"testing"
)

func Test_NewConfig_1(t *testing.T) {
	config, err := NewConfig("../conf/config.json")
	if err != nil {
		t.Error(err)
		return
	}

	json, err := json.Marshal(config)
	if err != nil {
		t.Error(err)
		return
	}

	t.Log(string(json))
}
