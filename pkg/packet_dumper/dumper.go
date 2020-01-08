package packet_dumper

import (
	"github.com/google/gopacket"
	"github.com/google/gopacket/pcap"
	"strconv"
	"time"
)

type PacketData struct {
	Timestamp       time.Time `json:"timestamp"`
	Protocol        string    `json:"protocol"`
	SourceMAC       string    `json:"source_mac"`
	DestinationMAC  string    `json:"destination_mac"`
	SourceIP        string    `json:"source_ip"`
	DestinationIP   string    `json:"destination_ip"`
	SourcePort      int       `json:"source_port"`
	DestinationPort int       `json:"destination_port"`
	Length          int       `json:"length"`
}

type Output struct {
	Timestamp  time.Time  `json:"timestamp"`
	PacketData PacketData `json:"packet_data"`
}

func handlePacket(packet gopacket.Packet, callback func(output Output) error) error {
	metadata := packet.Metadata()

	output := Output{
		Timestamp: time.Now(),
		PacketData: PacketData{
			Timestamp: metadata.Timestamp,
			Length:    metadata.Length,
		},
	}

	linkLayer := packet.LinkLayer()
	if linkLayer != nil {
		sourceMAC, destinationMAC := linkLayer.LinkFlow().Endpoints()

		output.PacketData.SourceMAC = sourceMAC.String()
		output.PacketData.DestinationMAC = destinationMAC.String()

		output.PacketData.Protocol = linkLayer.LayerType().String()
	}

	networkLayer := packet.NetworkLayer()
	if networkLayer != nil {
		sourceIP, destinationIP := networkLayer.NetworkFlow().Endpoints()

		output.PacketData.SourceIP = sourceIP.String()
		output.PacketData.DestinationIP = destinationIP.String()

		output.PacketData.Protocol = networkLayer.LayerType().String()
	}

	transportLayer := packet.TransportLayer()
	if transportLayer != nil {
		sourcePort, destinationPort := transportLayer.TransportFlow().Endpoints()

		sourcePortInt, err := strconv.Atoi(sourcePort.String())
		if err != nil {
			return err
		}
		output.PacketData.SourcePort = sourcePortInt

		destinationPortInt, err := strconv.Atoi(destinationPort.String())
		if err != nil {
			return err
		}
		output.PacketData.DestinationPort = destinationPortInt

		output.PacketData.Protocol = transportLayer.LayerType().String()
	}

	err := callback(output)
	if err != nil {
		return err
	}

	return nil
}

func Watch(interfaceName, filter string, callback func(output Output) error) error {
	handle, err := pcap.OpenLive(interfaceName, 1600, true, pcap.BlockForever)
	if err != nil {
		return err
	}

	err = handle.SetBPFFilter(filter)
	if err != nil {
		return err
	}

	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())

	for packet := range packetSource.Packets() {
		err = handlePacket(packet, callback)
		if err != nil {
			panic(err)
		}
	}

	return nil
}
