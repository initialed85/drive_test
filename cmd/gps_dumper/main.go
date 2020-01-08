package main

import (
	"flag"
	"github.com/initialed85/drive_test/pkg/file_writer"
	"github.com/initialed85/drive_test/pkg/gps_dumper"
	"github.com/stratoberry/go-gpsd"
	"log"
	"strconv"
	"strings"
)

type Args struct {
	Host       string
	Port       int
	OutputPath string
}

var args Args

func getArgs() (Args, error) {
	target := Args{}

	parts := strings.Split(gpsd.DefaultAddress, ":")

	flag.StringVar(&target.Host, "host", parts[0], "IP, host or FQDN to connect to")

	port, err := strconv.Atoi(parts[1])
	if err != nil {
		return Args{}, err
	}

	flag.IntVar(&target.Port, "port", port, "Port to use")

	flag.StringVar(&target.OutputPath, "output-path", "gps_output.jsonl", "Path to JSON Lines output file")

	flag.Parse()

	return target, nil
}

func callback(output gps_dumper.Output) error {
	return file_writer.WriteIndentedJSONToFile(output, args.OutputPath)
}

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile)

	var err error

	args, err = getArgs()
	if err != nil {
		log.Fatal(err)
	}

	dumper, err := gps_dumper.New(args.Host, args.Port, callback)
	if err != nil {
		log.Fatal(err)
	}

	done := dumper.Watch()

	<-done
}
