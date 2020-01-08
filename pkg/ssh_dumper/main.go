package ssh_dumper

import (
	"github.com/initialed85/drive_test/internal/gossh_python"
	"regexp"
	"strings"
	"time"
)

type CommandOutput struct {
	Command string `json:"command"`
	Output  string `json:"output"`
}

type CommandOutputs struct {
	Timestamp      time.Time       `json:"timestamp"`
	CommandOutputs []CommandOutput `json:"command_outputs"`
}

func cleanUp(sessionID uint64) {
	err := gossh_python.RPCClose(sessionID)
	if err != nil {
		panic(err)
	}
}

func readUntil(sessionID uint64, prompts []regexp.Regexp, removePromptEcho bool, size, timeout int) (string, error) {
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

		for _, expression := range prompts {
			if expression.MatchString(buf) {
				matchedExpression = expression
				goto Break
			}
		}

		time.Sleep(time.Duration(10) * time.Millisecond)
	}
Break:

	if removePromptEcho {
		buf = matchedExpression.ReplaceAllString(buf, "")
	}

	return buf, nil
}

func Watch(host string, port int, username, password string, timeout int, prompts []regexp.Regexp, dumbAuthentication, removePromptEcho bool, setupCommands []string, period float64, cycleCommands []string, removeCommandEcho bool, callback func(CommandOutputs) error) error {
	sessionID := gossh_python.NewRPCSession(
		host,
		username,
		password,
		port,
		timeout,
	)

	err := gossh_python.RPCConnect(sessionID)
	if err != nil {
		return err
	}

	defer cleanUp(sessionID)

	err = gossh_python.RPCGetShell(sessionID, "xterm", 1024, 1024)
	if err != nil {
		err = gossh_python.RPCGetShell(sessionID, "xterm", 80, 24)
		if err != nil {
			return err
		}
	}

	if dumbAuthentication {
		_, err := regexp.Compile("^.*:")
		if err != nil {
			return err
		}

		// wait for what is probably a Username prompt
		_, err = readUntil(sessionID, prompts, false, 65536, timeout)
		if err != nil {
			return err
		}

		// send the username
		err = gossh_python.RPCWrite(sessionID, username+"\n")
		if err != nil {
			return err
		}

		// wait for what is probably a Password prompt
		_, err = readUntil(sessionID, prompts, false, 65536, timeout)
		if err != nil {
			return err
		}

		// send the password
		err = gossh_python.RPCWrite(sessionID, password+"\n")
		if err != nil {
			return err
		}
	}

	_, err = readUntil(sessionID, prompts, removePromptEcho, 65536, timeout)
	if err != nil {
		return err
	}

	for _, command := range setupCommands {
		err := gossh_python.RPCWrite(sessionID, strings.TrimRight(command, "\n")+"\n")
		if err != nil {
			return err
		}

		_, err = readUntil(sessionID, prompts, removePromptEcho, 65536, timeout)
		if err != nil {
			return err
		}
	}

	ticker := time.NewTicker(time.Duration(period) * time.Second)

	for range ticker.C {
		commandOutputs := CommandOutputs{}

		commandOutputs.Timestamp = time.Now()

		for _, command := range cycleCommands {
			actualCommand := strings.TrimRight(command, "\n") + "\n"

			err := gossh_python.RPCWrite(sessionID, actualCommand)
			if err != nil {
				return err
			}

			output, err := readUntil(sessionID, prompts, removePromptEcho, 65536, timeout)
			if err != nil {
				return err
			}

			if removeCommandEcho {
				parts := strings.Split(output, actualCommand)

				output = strings.Join(parts[1:], actualCommand)
			}

			commandOutputs.CommandOutputs = append(
				commandOutputs.CommandOutputs,
				CommandOutput{command, output},
			)

		}

		err = callback(commandOutputs)
		if err != nil {
			return err
		}
	}

	return nil
}
