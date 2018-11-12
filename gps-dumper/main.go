package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/stratoberry/go-gpsd"
	"os"
	"strconv"
	"strings"
	"time"
)

type Args struct {
	Host       string
	Port       int
	OutputPath string
}

type Output struct {
	Timestamp time.Time       `json:"timestamp"`
	Report    *gpsd.TPVReport `json:"report"`
}

func getArgs() (Args, error) {
	target := Args{}

	parts := strings.Split(gpsd.DefaultAddress, ":")

	flag.StringVar(&target.Host, "host", parts[0], "IP, host or FQDN to connect to")

	port, err := strconv.Atoi(parts[1])
	if err != nil {
		panic(err)
	}

	flag.IntVar(&target.Port, "port", port, "Port to use")

	flag.StringVar(&target.OutputPath, "output-path", "gps_output.jsonl", "Path to JSON Lines output file")

	flag.Parse()

	return target, nil
}

func main() {
	args, err := getArgs()
	if err != nil {
		panic(err)
	}

	gps, err := gpsd.Dial(fmt.Sprintf("%v:%v", args.Host, args.Port))
	if err != nil {
		panic(err)
	}

	tpvFilter := func(r interface{}) {
		report := r.(*gpsd.TPVReport)

		output := Output{
			time.Now(),
			report,
		}

		jsonOutput, err := json.MarshalIndent(output, "", "  ")
		if err != nil {
			panic(err)
		}

		fmt.Println(string(jsonOutput) + "\n")

		f, err := os.OpenFile(args.OutputPath, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
		if err != nil {
			panic(err)
		}

		_, err = f.WriteString(string(jsonOutput) + "\n")
		if err != nil {
			panic(err)
		}

		f.Close()

	}

	gps.AddFilter("TPV", tpvFilter)

	done := gps.Watch()

	<-done
}
