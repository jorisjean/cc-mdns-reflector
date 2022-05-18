package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net"
	"os"

	"gopkg.in/ini.v1"
)


type CcClientRead struct {
	CcMac		string	`json:"cc_mac"`
	ClientMac	string	`json:"client_mac"`
}


type MrConfig struct {
	capInt		string
	ccSubnetIP	net.IP
	ccVlan		int
	clientVlan	int
}

func LoadConfig(filePath string) (MrConfig){
	msg := "Error while parsing confing file " + filePath + ": %v"

	cfg, err := ini.Load(filePath)
	if err != nil {
        log.Fatalf(msg, err)
    }

	ccVlan, err := cfg.Section("mdns-reflector").Key("chromecast_vlan").Int()
	if err != nil {
        log.Fatalf(msg, err)
    }

	clientVlan, err := cfg.Section("mdns-reflector").Key("client_vlan").Int()
	if err != nil {
        log.Fatalf(msg, err)
    }

	capInt := cfg.Section("mdns-reflector").Key("mdns_reflector_int").String()
	if len(capInt) == 0 {
		err := errors.New("Missing mdns_reflector_int value")
		log.Fatalf(msg, err)
	}

	ip := cfg.Section("mdns-reflector").Key("chromecast_spoof_ip").String()
	ccSubnetIP := net.ParseIP(ip)
	if ccSubnetIP == nil {
		err := errors.New("Invalid IP in config file chromecast_spoof_ip=" + ip)
		log.Fatalf(msg, err)
	}

	return MrConfig{capInt, ccSubnetIP, ccVlan, clientVlan}
}

func LoadCcClientMappings(filePath string) (map[string]string, map[string]string){

	// We read the file
    jsonFile, err := os.Open(filePath)

    if err != nil {
        log.Fatal(err)
    }

    defer jsonFile.Close()
    byteValue, _ := ioutil.ReadAll(jsonFile)

    var CcClientRead []CcClientRead
    json.Unmarshal(byteValue, &CcClientRead)

	ccToClient := make(map[string]string)
	clientToCc := make(map[string]string)
	// Only complete mappings (cc mac, cc ip, client mac, client ip) are returned
	for _, ccClientRead := range CcClientRead {
		if ccClientRead.CcMac != "" && ccClientRead.ClientMac != "" {
			// Parsing MAC to make sure they are valid
			ccMac, err := net.ParseMAC(ccClientRead.CcMac)
			if err != nil {
				log.Println("Invalid CC MAC: " + ccClientRead.CcMac)
				continue
			}
			clientMac, err := net.ParseMAC(ccClientRead.ClientMac)
			if err != nil {
				log.Println("Invalid Client MAC: " + ccClientRead.ClientMac)
				continue
			}
			ccToClient[ccMac.String()] = clientMac.String()
			clientToCc[clientMac.String()] = ccMac.String()
		}
    }

	return ccToClient, clientToCc

}
