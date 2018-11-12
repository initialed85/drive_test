package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/google/gopacket"
	"github.com/google/gopacket/pcap"
	"io/ioutil"
	"os"
	"strconv"
	"time"
)

type Args struct {
	Interface  string
	ConfigPath string
	OutputPath string
}

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

type Config struct {
	Filter string `json:"filter"`
}

func getArgs() (Args, error) {
	target := Args{}

	flag.StringVar(&target.Interface, "interface", "", "Interface to capture on")
	flag.StringVar(&target.ConfigPath, "config-path", "config.json", "Path to JSON config file")
	flag.StringVar(&target.OutputPath, "output-path", "packet_output.jsonl", "Path to JSON Lines output file")

	flag.Parse()

	return target, nil
}

func getConfig(path string) (Config, error) {
	config := Config{}

	data, err := ioutil.ReadFile(path)
	if err != nil {
		return config, err
	}

	err = json.Unmarshal(data, &config)
	if err != nil {
		return config, err
	}

	return config, nil
}

func handlePacket(args Args, packet gopacket.Packet) error {
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

	jsonPacketDump, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return err
	}

	fmt.Println(string(jsonPacketDump) + "\n")

	f, err := os.OpenFile(args.OutputPath, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		return err
	}

	_, err = f.WriteString(string(jsonPacketDump) + "\n")
	if err != nil {
		return err
	}

	f.Close()

	return nil
}

func main() {
	args, err := getArgs()
	if err != nil {
		panic(err)
	}

	config, err := getConfig(args.ConfigPath)
	if err != nil {
		panic(err)
	}

	handle, err := pcap.OpenLive(args.Interface, 1600, true, pcap.BlockForever)
	if err != nil {
		panic(err)
	}

	err = handle.SetBPFFilter(config.Filter)
	if err != nil {
		panic(err)
	}

	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())

	for packet := range packetSource.Packets() {
		err = handlePacket(args, packet)
		if err != nil {
			panic(err)
		}
	}
}
