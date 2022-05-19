package main

import (
	"net"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

type mdnsPacket struct {
	packet     gopacket.Packet
	srcMAC     *net.HardwareAddr
	dstMAC     *net.HardwareAddr
	isIPv6     bool
	srcIP     *net.IP
	vlanTag    *uint16
	isDNSQuery bool
}

func parsePacketsLazily(source *gopacket.PacketSource) chan mdnsPacket {
	// Process packets, and forward Bonjour traffic to the returned channel

	// Set decoding to Lazy
	source.DecodeOptions = gopacket.DecodeOptions{Lazy: true}

	packetChan := make(chan mdnsPacket, 100)

	go func() {
		for packet := range source.Packets() {
			tag := parseVLANTag(packet)

			// Get source and destination mac addresses
			srcMAC, dstMAC := parseEthernetLayer(packet)

			// Check IP protocol version and get srcIP
			isIPv6, srcIP := parseIPLayer(packet)

			// Get UDP payload
			payload := parseUDPLayer(packet)

			isDNSQuery := parseDNSPayload(payload)

			// Pass on the packet for its next adventure
			packetChan <- mdnsPacket{
				packet:     packet,
				vlanTag:    tag,
				srcMAC:     srcMAC,
				dstMAC:     dstMAC,
				isIPv6:     isIPv6,
				srcIP:     srcIP,
				isDNSQuery: isDNSQuery,
			}
		}
	}()

	return packetChan
}

func parseEthernetLayer(packet gopacket.Packet) (srcMAC, dstMAC *net.HardwareAddr) {
	if parsedEth := packet.Layer(layers.LayerTypeEthernet); parsedEth != nil {
		srcMAC = &parsedEth.(*layers.Ethernet).SrcMAC
		dstMAC = &parsedEth.(*layers.Ethernet).DstMAC
	}
	return
}

func parseVLANTag(packet gopacket.Packet) (tag *uint16) {
	if parsedTag := packet.Layer(layers.LayerTypeDot1Q); parsedTag != nil {
		tag = &parsedTag.(*layers.Dot1Q).VLANIdentifier
	}
	return
}

func parseIPLayer(packet gopacket.Packet) (isIPv6 bool, srcIP *net.IP) {
	if parsedIP := packet.Layer(layers.LayerTypeIPv4); parsedIP != nil {
		isIPv6 = false
		srcIP = &parsedIP.(*layers.IPv4).SrcIP
	}
	if parsedIP := packet.Layer(layers.LayerTypeIPv6); parsedIP != nil {
		isIPv6 = true
		srcIP = &parsedIP.(*layers.IPv6).SrcIP
	}

	return
}

func parseUDPLayer(packet gopacket.Packet) (payload []byte) {
	if parsedUDP := packet.Layer(layers.LayerTypeUDP); parsedUDP != nil {
		payload = parsedUDP.(*layers.UDP).Payload
	}
	return
}

func parseDNSPayload(payload []byte) (isDNSQuery bool) {
	packet := gopacket.NewPacket(payload, layers.LayerTypeDNS, gopacket.Default)
	if parsedDNS := packet.Layer(layers.LayerTypeDNS); parsedDNS != nil {
		isDNSQuery = !parsedDNS.(*layers.DNS).QR
	}
	return
}

type packetWriter interface {
	WritePacketData([]byte) error
}

func sendMdnsPacket(
	handle packetWriter,
	mdnsPacket *mdnsPacket,
	tag uint16,
	brMACAddress net.HardwareAddr,
	srcIPAddress net.IP,
	spoofsrcIP bool,
	dstMACAddress net.HardwareAddr) {

	// Change the VLAN Tag
	*mdnsPacket.vlanTag = tag
	// Change the source MAC to the interface MAC
	*mdnsPacket.srcMAC = brMACAddress
	// Change the dest MAC to the one of the target
	*mdnsPacket.dstMAC = dstMACAddress

	buf := gopacket.NewSerializeBuffer()
	serializeOptions := gopacket.SerializeOptions{}

	// We change the Source IP address of the mDNS query since Chromecasts ignore
	// packets coming from outside their subnet.
	if spoofsrcIP{
		serializeOptions = gopacket.SerializeOptions{ComputeChecksums: true}
		*mdnsPacket.srcIP = srcIPAddress
		// We recalculate the checksum since the IP was modified
		if parsedIP := mdnsPacket.packet.Layer(layers.LayerTypeIPv4); parsedIP != nil {
			if parsedUDP := mdnsPacket.packet.Layer(layers.LayerTypeUDP); parsedUDP != nil {
				parsedUDP.(*layers.UDP).SetNetworkLayerForChecksum(parsedIP.(*layers.IPv4))
			}
		}
	}

	gopacket.SerializePacket(buf, serializeOptions, mdnsPacket.packet)
	handle.WritePacketData(buf.Bytes())
}
