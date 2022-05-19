package main

import (
	"net"
	"reflect"
	"testing"
)

var _ccToClient = map[string]string{"aa:bb:cc:dd:ee:ff": "00:11:22:33:44:55"}
var _clientToCc = map[string]string{"00:11:22:33:44:55": "aa:bb:cc:dd:ee:ff"}

func TestLoadCcClientTable(t *testing.T) {
	computedCcToClient, computedClientToCc := LoadCcClientMappings("states_test.json")

	if !reflect.DeepEqual(computedCcToClient, _ccToClient) {
		t.Error("Error in LoadCcClientMappings(): unexpected mapping returned")
	}
	if !reflect.DeepEqual(computedClientToCc, _clientToCc) {
		t.Error("Error in LoadCcClientMappings(): unexpected mapping returned")
	}
}

var config = MrConfig{"eth0", net.ParseIP("192.168.1.100"), 10, 20}
func TestLoadConfig(t *testing.T) {
	computedConfig := LoadConfig("config_test.ini")

	if !reflect.DeepEqual(computedConfig, config) {
		t.Error("Error in LoadConfig(): unexpected config returned")
	}
}
