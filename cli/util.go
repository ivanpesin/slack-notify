package main

import (
	"bytes"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"gopkg.in/urfave/cli.v1"
)

// UTCLogger is a logger with UTC formatted time
type UTCLogger struct {
	Prefix string
}

func (l UTCLogger) Write(bytes []byte) (int, error) {
	var s string
	if l.Prefix != "" {
		s = time.Now().UTC().Format(time.RFC3339) + " [" + l.Prefix + "] " + string(bytes)
	} else {
		s = time.Now().UTC().Format(time.RFC3339) + " " + string(bytes)
	}
	return fmt.Println(s)
}

// Debug produces the output if in debug mode
func debug(s string) {
	if config.Debug {
		fmt.Printf("%s D: %s\n", time.Now().UTC().Format(time.RFC3339), s)
	}
}

func info(s string) {
	if config.Verbose || config.Debug {
		fmt.Printf("%s %s\n", time.Now().UTC().Format(time.RFC3339), s)
	}
}

func printConfig(c *cli.Context) {
	if config.Debug {
		debug("Active configuration (urfave): ")
		for _, flag := range c.GlobalFlagNames() {
			debug(" " + flag + ": " + c.GlobalString(flag))
		}

		debug("Active configuration (config): ")
		debug(fmt.Sprintf("%+v", config))
	}
}

func run(cmd string) (rc int, sout string, serr string) {
	// LookPath
	debug("running: " + cmd)
	c := exec.Command(strings.Fields(cmd)[0], strings.Fields(cmd)[1:]...)
	var stdout, stderr bytes.Buffer
	c.Stderr = &stderr
	c.Stdout = &stdout

	err := c.Run()
	if err != nil {
		stderr.WriteString(err.Error() + "\n")
	}
	debug(" rc = " + strconv.Itoa(c.ProcessState.ExitCode()))
	debug(" out:\n" + stdout.String())
	debug(" err:\n" + stderr.String())
	return c.ProcessState.ExitCode(),
		stdout.String(),
		stderr.String()
}

func runShell(cmd string) (rc int, sout string, serr string) {
	// LookPath
	debug("running in bash: " + cmd)
	c := exec.Command("/bin/bash", "-c", cmd)
	var stdout, stderr bytes.Buffer
	c.Stderr = &stderr
	c.Stdout = &stdout

	err := c.Run()
	if err != nil {
		stderr.WriteString(err.Error() + "\n")
	}
	debug(" rc = " + strconv.Itoa(c.ProcessState.ExitCode()))
	debug(" out:\n" + stdout.String())
	debug(" err:\n" + stderr.String())
	return c.ProcessState.ExitCode(),
		stdout.String(),
		stderr.String()
}
