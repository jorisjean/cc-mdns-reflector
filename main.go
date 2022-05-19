package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/pcap"
)


func main() {
	// Read config file and generate mDNS forwarding maps
	configPath := flag.String("config", "", "Config file in INI")
	statesPath := flag.String("states", "", "State file in JSON")
	debug := flag.Bool("debug", false, "Enable console debugs")
	flag.Parse()

	ccToClient, clientToCc := LoadCcClientMappings(*statesPath)
	fmt.Println(ccToClient, clientToCc)

	config := LoadConfig(*configPath)
	fmt.Println(config)

	// Get a handle on the network interface
	rawTraffic, err := pcap.OpenLive(config.capInt, 65536, true, time.Second)
	if err != nil {
		log.Fatalf("Could not find network interface: %v", config.capInt)
	}

	// Get the local MAC address, to filter out Bonjour packet generated locally
	intf, err := net.InterfaceByName(config.capInt)
	if err != nil {
		log.Fatal(err)
	}
	brMACAddress := intf.HardwareAddr

	// Filter tagged bonjour traffic
	filterTemplate := "not (ether src %s) and vlan and dst net (224.0.0.251 or ff02::fb) and udp dst port 5353"
	err = rawTraffic.SetBPFFilter(fmt.Sprintf(filterTemplate, brMACAddress))
	if err != nil {
		log.Fatalf("Could not apply filter on network interface: %v", err)
	}

	// Get a channel of Bonjour packets to process
	decoder := gopacket.DecodersByLayerName["Ethernet"]
	source := gopacket.NewPacketSource(rawTraffic, decoder)
	mdnsPackets := parsePacketsLazily(source)

	// Process Bonjours packets
	for mdnsPacket := range mdnsPackets {
		if *debug {
			fmt.Println(mdnsPacket.packet.String())
		}

		// No IPv6 support yet
		if mdnsPacket.isIPv6 {
			continue
		}

		// Forward the mDNS query to the appropriate Chromecast
		if mdnsPacket.isDNSQuery {
			// Looking for the client MAC in the states
			ccMacString, ok := clientToCc[mdnsPacket.srcMAC.String()]
			if !ok {
				continue
			}
			ccMAC, err := net.ParseMAC(ccMacString)
			if err != nil{
				continue
			}
			sendMdnsPacket(rawTraffic, &mdnsPacket, uint16(config.ccVlan), brMACAddress, ccMAC, config.ccSubnetIP)
		// Forward the mDNS response to the appropriate Client
		} else {
			clientMacString, ok := ccToClient[mdnsPacket.srcMAC.String()]
			if !ok {
				continue
			}
			clientMac, err := net.ParseMAC(clientMacString)
			if err != nil{
				continue
			}
			sendMdnsPacket(rawTraffic, &mdnsPacket, uint16(config.clientVlan), brMACAddress, clientMac, nil)
		}
	}
}
