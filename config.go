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
	CcIp		string	`json:"cc_ip"`
	CcMac		string	`json:"cc_mac"`
	ClientIp	string	`json:"client_ip"`
	ClientMac	string	`json:"client_mac"`
}

type CcClientMap struct {
	CcIp		net.IP				`json:"cc_ip"`
	CcMac		net.HardwareAddr	`json:"cc_mac"`
	ClientIp	net.IP				`json:"client_ip"`
	ClientMac	net.HardwareAddr	`json:"client_mac"`
}

func LoadCcClientMaps(file_path string) ([]CcClientMap){

	// We read the file
    jsonFile, err := os.Open(file_path)

    if err != nil {
        log.Fatal(err)
    }

    defer jsonFile.Close()
    byteValue, _ := ioutil.ReadAll(jsonFile)

    var CcClientMapsRead []CcClientMapRead
    json.Unmarshal(byteValue, &CcClientMapsRead)

	var ccClientMaps []CcClientMap
	// Only complete mappings (cc mac, cc ip, client mac, client ip) are returned
	for _, ccClientMapRead := range CcClientMapsRead {
		if ccClientMapRead.CcMac != "" && ccClientMapRead.ClientMac != "" && ccClientMapRead.CcIp != "" && ccClientMapRead.ClientIp != "" {
			// Parsing MAC to make sure they are valid
			ccMac, err := net.ParseMAC(ccClientMapRead.CcMac)
			if err != nil {
				log.Println("Invalid CC MAC: " + ccClientMapRead.CcMac)
				continue
			}
			clientMac, err := net.ParseMAC(ccClientMapRead.ClientMac)
			if err != nil {
				log.Println("Invalid Client MAC: " + ccClientMapRead.ClientMac)
				continue
			}
			// Parsing IP to make sure they are valid
			ccIp := net.ParseIP(ccClientMapRead.CcIp)
			if ccIp == nil {
				log.Println("Invalid CC MAC: " + ccClientMapRead.CcIp)
				continue
			}
			clientIp := net.ParseIP(ccClientMapRead.ClientIp)
			if clientIp == nil {
				log.Println("Invalid CC MAC: " + ccClientMapRead.ClientIp)
				continue
			}
			ccClientMap := CcClientMap{ccIp, ccMac, clientIp, clientMac}
			ccClientMaps = append(ccClientMaps, ccClientMap)
		}
    }

	return ccClientMaps

}

func main() {
	mappings := LoadCcClientMaps("test.json")
	for _, mapping := range mappings {
		fmt.Println("CC IP: " + mapping.CcIp.String())
		fmt.Println("CC MAC: " + mapping.CcMac.String())
		fmt.Println("Client IP: " + mapping.ClientIp.String())
		fmt.Println("Client MAC: " + mapping.ClientMac.String())
	}
}
