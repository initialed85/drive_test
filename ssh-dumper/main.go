package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
	"time"

	"./gossh_python"
)
import "flag"

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

type Config struct {
	RawPrompts    []string `json:"prompts"`
	Prompts       []regexp.Regexp
	SetupCommands []string `json:"setup_commands"`
	CycleCommands []string `json:"cycle_commands"`
}

type CommandOutput struct {
	Command string `json:"command"`
	Output  string `json:"output"`
}

type CommandOutputs struct {
	Timestamp      time.Time       `json:"timestamp"`
	CommandOutputs []CommandOutput `json:"command_outputs"`
}

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

func getConfig(path string) (Config, error) {
	basePrompt, err := regexp.Compile("\n.*[\\$|^|#] ")
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

func cleanUp(sessionID uint64) {
	err := gossh_python.RPCClose(sessionID)
	if err != nil {
		panic(err)
	}
}

func readUntil(sessionID uint64, args Args, config Config, size, timeout int) (string, error) {
	cutoff := time.Now().Add(time.Duration(timeout) * time.Second)

	buf := ""

	matchedExpression := regexp.Regexp{}

	for {
		if time.Now().After(cutoff) {
			break
		}

		data, err := gossh_python.RPCRead(sessionID, size)

		data = strings.Replace(data, "\x00", "", -1)

		data = strings.Replace(data, "\r\n", "\n", -1)

		buf += data

		if err != nil {
			return buf, err
		}

		for _, expression := range config.Prompts {
			if expression.MatchString(buf) {
				matchedExpression = expression
				goto Break
			}
		}

		time.Sleep(time.Duration(10) * time.Millisecond)
	}
Break:

	if args.RemovePromptEcho {
		buf = matchedExpression.ReplaceAllString(buf, "")
	}

	return buf, nil
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

	// tell gossh_python not to try and toggle the non-existent (in this case) Python GIL
	gossh_python.SetPyPy()

	sessionID := gossh_python.NewRPCSession(
		args.Host,
		args.Username,
		args.Password,
		args.Port,
		args.Timeout,
	)

	err = gossh_python.RPCConnect(sessionID)
	if err != nil {
		panic(err)
	}
	defer cleanUp(sessionID)

	err = gossh_python.RPCGetShell(sessionID, "xterm", 1024, 1024)
	if err != nil {
		err = gossh_python.RPCGetShell(sessionID, "xterm", 80, 24)
		if err != nil {
			panic(err)
		}
	}

	if args.DumbAuthentication {
		tempArgs := Args{
			RemovePromptEcho:  false,
			RemoveCommandEcho: false,
		}

		expression, err := regexp.Compile("^.*:")
		if err != nil {
			panic(err)
		}

		tempConfig := Config{
			Prompts: []regexp.Regexp{*expression},
		}

		// wait for what is probably a Username prompt
		_, err = readUntil(sessionID, tempArgs, tempConfig, 65536, args.Timeout)
		if err != nil {
			panic(err)
		}

		// send the username
		err = gossh_python.RPCWrite(sessionID, args.Username+"\n")
		if err != nil {
			panic(err)
		}

		// wait for what is probably a Password prompt
		_, err = readUntil(sessionID, tempArgs, tempConfig, 65536, args.Timeout)
		if err != nil {
			panic(err)
		}

		// send the password
		err = gossh_python.RPCWrite(sessionID, args.Password+"\n")
		if err != nil {
			panic(err)
		}
	}

	_, err = readUntil(sessionID, args, config, 65536, args.Timeout)
	if err != nil {
		panic(err)
	}

	for _, command := range config.SetupCommands {
		err := gossh_python.RPCWrite(sessionID, strings.TrimRight(command, "\n")+"\n")
		if err != nil {
			panic(err)
		}

		_, err = readUntil(sessionID, args, config, 65536, args.Timeout)
		if err != nil {
			panic(err)
		}
	}

	ticker := time.NewTicker(time.Duration(args.Period) * time.Second)

	for range ticker.C {
		commandOutputs := CommandOutputs{}

		commandOutputs.Timestamp = time.Now()

		for _, command := range config.CycleCommands {
			actualCommand := strings.TrimRight(command, "\n") + "\n"

			err := gossh_python.RPCWrite(sessionID, actualCommand)
			if err != nil {
				panic(err)
			}

			output, err := readUntil(sessionID, args, config, 65536, args.Timeout)
			if err != nil {
				panic(err)

			}

			if args.RemoveCommandEcho {
				parts := strings.Split(output, actualCommand)

				output = strings.Join(parts[1:], actualCommand)
			}

			commandOutputs.CommandOutputs = append(
				commandOutputs.CommandOutputs,
				CommandOutput{command, output},
			)

		}

		jsonCommandOutputs, err := json.MarshalIndent(commandOutputs, "", "  ")
		if err != nil {
			panic(err)
		}

		fmt.Println(string(jsonCommandOutputs) + "\n")

		f, err := os.OpenFile(args.OutputPath, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
		if err != nil {
			panic(err)
		}

		_, err = f.WriteString(string(jsonCommandOutputs) + "\n")
		if err != nil {
			panic(err)
		}

		f.Close()
	}
}
