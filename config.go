package main

import (
    "encoding/json"
    "fmt"
    "io/ioutil"
    "os"
	"net"
	"log"
)


type CcClientMapRead struct {
	CcIp		net.IP				`json:"cc_ip"`
	CcMac		string			`json:"cc_mac"`
	ClientIp	net.IP				`json:"client_ip"`
	ClientMac	string			`json:"client_mac"`
}

type CcClientMap struct {
	CcIp		net.IP				`json:"cc_ip"`
	CcMac		net.HardwareAddr	`json:"cc_mac"`
	ClientIp	net.IP				`json:"client_ip"`
	ClientMac	net.HardwareAddr	`json:"client_mac"`
}

func LoadCcClientMaps() ([]CcClientMap){
    // Open our jsonFile
    jsonFile, err := os.Open("test.json")
    // if we os.Open returns an error then handle it
    if err != nil {
        log.Fatal(err)
    }

    defer jsonFile.Close()

    byteValue, _ := ioutil.ReadAll(jsonFile)
    var CcClientMapsRead []CcClientMapRead
    json.Unmarshal(byteValue, &CcClientMapsRead)

	var ccClientMaps []CcClientMap
	for _, ccClientMapRead := range CcClientMapsRead {
		if ccClientMapRead.CcMac != "" && ccClientMapRead.ClientMac != "" && ccClientMapRead.CcIp != nil && ccClientMapRead.ClientIp != nil {
			ccMac, err := net.ParseMAC(ccClientMapRead.CcMac)
			if err != nil {
				log.Fatal(err)
			}
			clientMac, err := net.ParseMAC(ccClientMapRead.ClientMac)
			if err != nil {
				log.Fatal(err)
			}
			ccClientMap := CcClientMap{ccClientMapRead.CcIp, ccMac, ccClientMapRead.ClientIp, clientMac}
			ccClientMaps = append(ccClientMaps, ccClientMap)
		}
    }

	return ccClientMaps

}

func main() {
	mappings := LoadCcClientMaps()
	for _, mapping := range mappings {
		fmt.Println("CC IP: " + mapping.CcIp.String())
		fmt.Println("CC MAC: " + mapping.CcMac.String())
		fmt.Println("Client IP: " + mapping.ClientIp.String())
		fmt.Println("Client MAC: " + mapping.ClientMac.String())
	}
}
