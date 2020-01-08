package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"github.com/initialed85/drive_test/pkg/file_writer"
	"github.com/initialed85/drive_test/pkg/ssh_dumper"
	"io/ioutil"
	"log"
	"regexp"
)

type Args struct {
	Host               string
	Port               int
	Timeout            int
	Username           string
	Password           string
	Period             float64
	ConfigPath         string
	OutputPath         string
	RemoveCommandEcho  bool
	RemovePromptEcho   bool
	TrimOutput         bool
	DumbAuthentication bool
}

var args Args

func getArgs() (Args, error) {
	target := Args{}

	flag.StringVar(&target.Host, "host", "localhost", "IP, host or FQDN to connect to")
	flag.IntVar(&target.Port, "port", 22, "Port to use")
	flag.IntVar(&target.Timeout, "timeout", 5, "Timeout in seconds")
	flag.StringVar(&target.Username, "username", "", "Username to use")
	flag.StringVar(&target.Password, "password", "", "Password to use")
	flag.Float64Var(&target.Period, "period", 8, "Period to cycle at in seconds")
	flag.StringVar(&target.ConfigPath, "config-path", "config.json", "Path to JSON config file")
	flag.StringVar(&target.OutputPath, "output-path", "ssh_output.jsonl", "Path to JSON Lines output file")
	flag.BoolVar(&target.RemoveCommandEcho, "remove-command-echo", true, "Remove command echo")
	flag.BoolVar(&target.RemovePromptEcho, "remove-prompt-echo", true, "Remove prompt echo")
	flag.BoolVar(&target.TrimOutput, "trim-output", true, "Trim leading and trailing whitespace from output")
	flag.BoolVar(&target.DumbAuthentication, "dumb-authentication", false, "Expect dumb text authentication (e.g. username/password prompt)")

	flag.Parse()

	if len(target.Username) == 0 {
		return target, errors.New("username flag missing or empty")
	}

	if len(target.Password) == 0 {
		return target, errors.New("password flag missing or empty")
	}

	return target, nil
}

type Config struct {
	RawPrompts    []string `json:"prompts"`
	Prompts       []regexp.Regexp
	SetupCommands []string `json:"setup_commands"`
	CycleCommands []string `json:"cycle_commands"`
}

func getConfig(path string) (Config, error) {
	basePrompt, err := regexp.Compile("\n.*[$|^|#] ")
	if err != nil {
		return Config{}, err
	}

	config := Config{
		Prompts: []regexp.Regexp{*basePrompt},
	}

	data, err := ioutil.ReadFile(path)
	if err != nil {
		return config, err
	}

	err = json.Unmarshal(data, &config)
	if err != nil {
		return config, err
	}

	for _, rawPrompt := range config.RawPrompts {
		prompt, err := regexp.Compile(rawPrompt)
		if err != nil {
			panic(err)
		}

		config.Prompts = append(config.Prompts, *prompt)
	}

	if len(config.CycleCommands) == 0 {
		return config, fmt.Errorf("missing or empty \"cycle_commands\" field in %s", path)
	}

	return config, nil
}

func callback(outputs ssh_dumper.CommandOutputs) error {
	return file_writer.WriteIndentedJSONToFile(outputs, args.OutputPath)
}

func main() {
	var err error

	args, err = getArgs()
	if err != nil {
		panic(err)
	}

	config, err := getConfig(args.ConfigPath)
	if err != nil {
		panic(err)
	}

	err = ssh_dumper.Watch(
		args.Host,
		args.Port,
		args.Username,
		args.Password,
		args.Timeout,
		config.Prompts,
		args.DumbAuthentication,
		args.RemovePromptEcho,
		config.SetupCommands,
		args.Period,
		config.CycleCommands,
		args.RemoveCommandEcho,
		callback,
	)
	if err != nil {
		log.Fatal(err)
	}
}
