package main

import (
	"io"
	"net"
	"reflect"
	"testing"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

var (
	srcMACTest          = net.HardwareAddr{0xFF, 0xAA, 0xFA, 0xAA, 0xFF, 0xAA}
	dstMACTest          = net.HardwareAddr{0xBD, 0xBD, 0xBD, 0xBD, 0xBD, 0xBD}
	brMACTest           = net.HardwareAddr{0xF2, 0xAA, 0xFA, 0xAA, 0xFF, 0xAA}
	vlanIdentifierTest  = uint16(30)
	srcIPv4Test         = net.IP{127, 0, 0, 1}
	srcIPv4SpoofTest    = net.IP{192, 168, 1, 100}
	dstIPv4Test         = net.IP{224, 0, 0, 251}
	srcIPv6Test         = net.ParseIP("::1")
	dstIPv6Test         = net.ParseIP("ff02::fb")
	srcUDPPortTest      = layers.UDPPort(5353)
	dstUDPPortTest      = layers.UDPPort(5353)
	questionPayloadTest = []byte{0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 7, 101, 120, 97,
		109, 112, 108, 101, 3, 99, 111, 109, 0, 0, 1, 0, 1}
	spoofAddrTest       = net.IP{192, 168, 1, 1}
)

func createMockmDNSPacket(isDNSQuery bool) []byte {
	return createRawPacket( isDNSQuery, vlanIdentifierTest, srcIPv4Test, dstIPv4Test, srcMACTest, dstMACTest, dstUDPPortTest)
}

func createRawPacket(
	isDNSQuery bool,
	vlanTag uint16,
	srcIP net.IP,
	dstIP net.IP,
	srcMAC net.HardwareAddr,
	dstMAC net.HardwareAddr,
	dstPort layers.UDPPort) []byte {
	var ethernetLayer, dot1QLayer, ipLayer, udpLayer, dnsLayer gopacket.SerializableLayer

	ethernetLayer = &layers.Ethernet{
		SrcMAC:       srcMAC,
		DstMAC:       dstMAC,
		EthernetType: layers.EthernetTypeDot1Q,
	}

	dot1QLayer = &layers.Dot1Q{
		VLANIdentifier: vlanTag,
		Type:           layers.EthernetTypeIPv4,
	}

	ipLayer = &layers.IPv4{
		SrcIP:    srcIP,
		DstIP:    dstIP,
		Version:  4,
		Protocol: layers.IPProtocolUDP,
		Length:   146,
		IHL:      5,
		TOS:      0,
	}

	udpLayer = &layers.UDP{
		SrcPort: srcUDPPortTest,
		DstPort: dstPort,
	}

	if isDNSQuery {
		dnsLayer = &layers.DNS{
			Questions: []layers.DNSQuestion{layers.DNSQuestion{
				Name:  []byte("example.com"),
				Type:  layers.DNSTypeA,
				Class: layers.DNSClassIN,
			}},
			QDCount: 1,
		}
	} else {
		dnsLayer = &layers.DNS{
			Answers: []layers.DNSResourceRecord{layers.DNSResourceRecord{
				Name:  []byte("example.com"),
				Type:  layers.DNSTypeA,
				Class: layers.DNSClassIN,
				TTL:   1024,
				IP:    net.IP([]byte{1, 2, 3, 4}),
			}},
			ANCount: 1,
			QR:      true,
		}
	}

	buffer := gopacket.NewSerializeBuffer()
	gopacket.SerializeLayers(
		buffer,
		gopacket.SerializeOptions{},
		ethernetLayer,
		dot1QLayer,
		ipLayer,
		udpLayer,
		dnsLayer,
	)
	return buffer.Bytes()
}

func TestParseEthernetLayer(t *testing.T) {
	decoder := gopacket.DecodersByLayerName["Ethernet"]
	options := gopacket.DecodeOptions{Lazy: true}

	packet := gopacket.NewPacket(createMockmDNSPacket(true), decoder, options)

	expectedResult1, expectedResult2 := &srcMACTest, &dstMACTest
	computedResult1, computedResult2 := parseEthernetLayer(packet)
	if !reflect.DeepEqual(expectedResult1, computedResult1) || !reflect.DeepEqual(expectedResult2, computedResult2) {
		t.Error("Error in parseEthernetLayer()")
	}
}

func TestParseVLANTag(t *testing.T) {
	decoder := gopacket.DecodersByLayerName["Ethernet"]
	options := gopacket.DecodeOptions{Lazy: true}

	packet := gopacket.NewPacket(createMockmDNSPacket(true), decoder, options)

	expectedLayer := &layers.Dot1Q{
		VLANIdentifier: vlanIdentifierTest,
		Type:           layers.EthernetTypeIPv4,
	}
	expectedResult := &expectedLayer.VLANIdentifier
	computedResult := parseVLANTag(packet)
	if !reflect.DeepEqual(expectedResult, computedResult) {
		t.Error("Error in parseEthernetLayer()")
	}
}

func TestParseIPLayer(t *testing.T) {
	decoder := gopacket.DecodersByLayerName["Ethernet"]
	options := gopacket.DecodeOptions{Lazy: true}

	ipv4Packet := gopacket.NewPacket(createMockmDNSPacket(true), decoder, options)

	computedIsIPv6, srcIP := parseIPLayer(ipv4Packet)
	if computedIsIPv6 == true || srcIP == nil {
		t.Error("Error in parseIPLayer() for IPv4 addresses")
	}
}

func TestParseUDPLayer(t *testing.T) {
	decoder := gopacket.DecodersByLayerName["Ethernet"]
	options := gopacket.DecodeOptions{Lazy: true}

	packet := gopacket.NewPacket(createMockmDNSPacket(true), decoder, options)

	questionPacketPayload := parseUDPLayer(packet)
	if !reflect.DeepEqual(questionPayloadTest, questionPacketPayload) {
		t.Error("Error in parseUDPLayer()")
	}
}

func TestParseDNSPayload(t *testing.T) {
	decoder := gopacket.DecodersByLayerName["Ethernet"]
	options := gopacket.DecodeOptions{Lazy: true}

	questionPacket := gopacket.NewPacket(createMockmDNSPacket(true), decoder, options)

	questionPacketPayload := parseUDPLayer(questionPacket)

	questionExpectedResult := true
	questionComputedResult := parseDNSPayload(questionPacketPayload)
	if !reflect.DeepEqual(questionExpectedResult, questionComputedResult) {
		t.Error("Error in parseDNSPayload() for DNS queries")
	}

	answerPacket := gopacket.NewPacket(createMockmDNSPacket(false), decoder, options)

	answerPacketPayload := parseUDPLayer(answerPacket)

	answerExpectedResult := false
	answerComputedResult := parseDNSPayload(answerPacketPayload)
	if !reflect.DeepEqual(answerExpectedResult, answerComputedResult) {
		t.Error("Error in parseDNSPayload() for DNS answers")
	}
}

type dataSource struct {
	sentPackets int
	data        [][]byte
}

func (dataSource *dataSource) ReadPacketData() (data []byte, ci gopacket.CaptureInfo, err error) {
	// Return one packet for each call.
	// If all the expected packets have already been returned in the past, return an EOF error
	// to end the reading of packets from this source.
	if dataSource.sentPackets >= len(dataSource.data) {
		return nil, ci, io.EOF
	}
	data = dataSource.data[dataSource.sentPackets]
	ci = gopacket.CaptureInfo{
		Timestamp:      time.Time{},
		CaptureLength:  len(data),
		Length:         ci.CaptureLength,
		InterfaceIndex: 0,
	}
	dataSource.sentPackets++
	return data, ci, nil
}

func createMockPacketSource() (packetSource *gopacket.PacketSource, packet gopacket.Packet) {
	// send one legitimate packet
	// Return the packetSource and the legitimate packet
	data := [][]byte{
		createMockmDNSPacket(true)}
	dataSource := &dataSource{
		sentPackets: 0,
		data:        data,
	}
	decoder := gopacket.DecodersByLayerName["Ethernet"]
	packetSource = gopacket.NewPacketSource(dataSource, decoder)
	packet = gopacket.NewPacket(data[len(data)-1], decoder, gopacket.DecodeOptions{Lazy: true})
	return
}

func areMdnsPacketsEqual(a, b mdnsPacket) (areEqual bool) {
	areEqual = (*a.vlanTag == *b.vlanTag) && (a.srcMAC.String() == b.srcMAC.String()) && (a.isDNSQuery == b.isDNSQuery)
	// While comparing Mdns packets, we do not want to compare packets entirely.
	// In particular, packet.metadata may be slightly different, we do not need them to be the same.
	// So we only compare the layers part of the packets.
	areEqual = areEqual && reflect.DeepEqual(a.packet.Layers(), b.packet.Layers())
	return
}

func TestFilterMdnsPacketsLazily(t *testing.T) {
	mockPacketSource, packet := createMockPacketSource()
	packetChan := parsePacketsLazily(mockPacketSource)

	expectedResult := mdnsPacket{
		packet:     packet,
		vlanTag:    &vlanIdentifierTest,
		srcMAC:     &srcMACTest,
		isDNSQuery: true,
	}

	computedResult := <-packetChan
	if !areMdnsPacketsEqual(expectedResult, computedResult) {
		t.Error("Error in parsePacketsLazily()")
	}
}

type mockPacketWriter struct {
	packet gopacket.Packet
}

func (pw *mockPacketWriter) WritePacketData(bytes []byte) (err error) {
	decoder := gopacket.DecodersByLayerName["Ethernet"]
	pw.packet = gopacket.NewPacket(bytes, decoder, gopacket.DecodeOptions{Lazy: true})
	return
}

func TestSendMdnsPacket(t *testing.T) {
	// Craft a test packet
	initialDataIPv4 := createMockmDNSPacket(true)
	decoder := gopacket.DecodersByLayerName["Ethernet"]
	initialPacketIPv4 := gopacket.NewPacket(initialDataIPv4, decoder, gopacket.DecodeOptions{Lazy: true})

	srcMACv4, dstMACv4 := parseEthernetLayer(initialPacketIPv4)
	isIPv6, srcIP := parseIPLayer(initialPacketIPv4)
	mdnsTestPacketIPv4 := mdnsPacket{
		packet:     initialPacketIPv4,
		vlanTag:    parseVLANTag(initialPacketIPv4),
		srcMAC:     srcMACv4,
		dstMAC:     dstMACv4,
		srcIP:     srcIP,
		isDNSQuery: true,
		isIPv6:     isIPv6,
	}

	newVlanTag := uint16(29)

	// Test without changing the source IP
	expectedDstMACv4 := net.HardwareAddr{0x01, 0x02, 0x03, 0x04, 0x05, 0x06}
	expectedDataIPv4 := createRawPacket(true, newVlanTag, srcIPv4Test, dstIPv4Test, brMACTest, expectedDstMACv4, dstUDPPortTest)
	expectedPacketIPv4 := gopacket.NewPacket(expectedDataIPv4, decoder, gopacket.DecodeOptions{Lazy: true})

	pw := &mockPacketWriter{packet: nil}

	sendMdnsPacket(pw, &mdnsTestPacketIPv4, newVlanTag, brMACTest, expectedDstMACv4, nil)
	if !reflect.DeepEqual(expectedPacketIPv4.Layers(), pw.packet.Layers()) {
		t.Error("Error in sendMdnsPacket() for IPv4")
	}

	// Test with changing the source IP
	expectedDstMACv4 = net.HardwareAddr{0x01, 0x02, 0x03, 0x04, 0x05, 0x06}
	expectedDataIPv4 = createRawPacket(true, newVlanTag, srcIPv4Test, dstIPv4Test, brMACTest, expectedDstMACv4, dstUDPPortTest)
	expectedPacketIPv4 = gopacket.NewPacket(expectedDataIPv4, decoder, gopacket.DecodeOptions{Lazy: true})

	// When we change the src IP address of mDNS Query we expect the IPv4 layer to be different
	sendMdnsPacket(pw, &mdnsTestPacketIPv4, newVlanTag, brMACTest, dstMACTest, spoofAddrTest)
	if reflect.DeepEqual(expectedPacketIPv4.Layers(), pw.packet.Layers()) {
		t.Error("Error in sendMdnsPacket() for IPv4 with spoof address")
	}

}
