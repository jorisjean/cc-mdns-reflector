package main

import (
	"flag"
	"fmt"
	// "log"
	// "net"
	// "time"

	// "github.com/google/gopacket"
	// "github.com/google/gopacket/pcap"
)


func main() {
	// Read config file and generate mDNS forwarding maps
	configPath := flag.String("config", "", "Config file in INI")
	statesPath := flag.String("states", "", "State file in JSON")

	// debug := flag.Bool("debug", false, "Enable console debugs")
	flag.Parse()

	ccToClient, clientToCc := LoadCcClientMappings(*statesPath)
	fmt.Println(ccToClient, clientToCc)

	config := LoadConfig(*configPath)
	fmt.Println(config)

// 	// Get a handle on the network interface
// 	rawTraffic, err := pcap.OpenLive(*capInt, 65536, true, time.Second)
// 	if err != nil {
// 		log.Fatalf("Could not find network interface: %v", *capInt)
// 	}


// 	// Get the local MAC address, to filter out Bonjour packet generated locally
// 	intf, err := net.InterfaceByName(*capInt)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	brMACAddress := intf.HardwareAddr

// 	// Filter tagged bonjour traffic
// 	filterTemplate := "not (ether src %s) and vlan and dst net (224.0.0.251 or ff02::fb) and udp dst port 5353"
// 	err = rawTraffic.SetBPFFilter(fmt.Sprintf(filterTemplate, brMACAddress))
// 	if err != nil {
// 		log.Fatalf("Could not apply filter on network interface: %v", err)
// 	}

// 	// Get a channel of Bonjour packets to process
// 	decoder := gopacket.DecodersByLayerName["Ethernet"]
// 	source := gopacket.NewPacketSource(rawTraffic, decoder)
// 	bonjourPackets := parsePacketsLazily(source)

// 	// Process Bonjours packets
// 	for bonjourPacket := range bonjourPackets {
// 		if *debug {
// 			fmt.Println(bonjourPacket.packet.String())
// 		}

// 		// Forward the mDNS query or response to appropriate VLANs
// 		if bonjourPacket.isDNSQuery {
// 			tags, ok := poolsMap[*bonjourPacket.vlanTag]
// 			if !ok {
// 				continue
// 			}

// 			for _, tag := range tags {
// 				sendBonjourPacket(rawTraffic, &bonjourPacket, tag, brMACAddress, ccSubnetIP, true, *bonjourPacket.dstMAC, false)
// 			}
// 		} else {
// 			device, ok := cfg.Devices[macAddress(bonjourPacket.srcMAC.String())]
// 			if !ok {
// 				continue
// 			}
// 			for _, tag := range device.SharedPools {
// 				// if we have a MAC stored for this vlan we also send the response packet directly to it
// 				if clientMAC, ok := lastquery[tag]; ok {
// 					fmt.Printf("Sending direct packet to MAC %v \n", clientMAC)
// 					sendBonjourPacket(rawTraffic, &bonjourPacket, tag, brMACAddress, *bonjourPacket.srcIP, false, clientMAC, true)
// 				}
// 				// we always forward the multicast answer
// 				sendBonjourPacket(rawTraffic, &bonjourPacket, tag, brMACAddress, *bonjourPacket.srcIP, false, *bonjourPacket.dstMAC, false)
// 			}
// 		}
// 	}
}
