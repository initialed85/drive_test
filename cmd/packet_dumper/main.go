package main

import (
	"encoding/json"
	"flag"
	"github.com/initialed85/drive_test/pkg/file_writer"
	"github.com/initialed85/drive_test/pkg/packet_dumper"
	"io/ioutil"
	"log"
)

type Args struct {
	Interface  string
	ConfigPath string
	OutputPath string
}

var args Args

func getArgs() (Args, error) {
	target := Args{}

	flag.StringVar(&target.Interface, "interface", "", "Interface to capture on")
	flag.StringVar(&target.ConfigPath, "config-path", "config.json", "Path to JSON config file")
	flag.StringVar(&target.OutputPath, "output-path", "packet_output.jsonl", "Path to JSON Lines output file")

	flag.Parse()

	return target, nil
}

type Config struct {
	Filter string `json:"filter"`
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

func callback(output packet_dumper.Output) error {
	return file_writer.WriteIndentedJSONToFile(output, args.OutputPath)
}

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile)

	var err error

	args, err = getArgs()
	if err != nil {
		panic(err)
	}

	config, err := getConfig(args.ConfigPath)
	if err != nil {
		panic(err)
	}

	err = packet_dumper.Watch(args.Interface, config.Filter, callback)
	if err != nil {
		log.Fatal(err)
	}
}
