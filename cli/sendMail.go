package main

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/ivanpesin/slack-notify/slack"
	"gopkg.in/urfave/cli.v1"
)

var mailxBin = "/usr/bin/mail"
var msgMaxSize = 3000

func setMailConfig() {
	if config.MailWebhook != "" {
		config.Webhook = config.MailWebhook
	}
	if config.MailChannel != "" {
		config.Channel = config.MailChannel
	}
	if config.MailUsername != "" {
		config.Username = config.MailUsername
	}
}

func sendMail(c *cli.Context) error {

	setMailConfig()

	rc, sout, serr := run(mailxBin + " -H")
	if rc != 0 {
		return cli.NewExitError(
			fmt.Sprintf("Error: mail command failed with rc = %d:\n stdout:\n%s\n stderr:\n%s", rc, sout, serr), 1)
	}

	mailCount := len(strings.Split(sout, "\n")) - 1

	if mailCount < 1 {
		info("No mail")
		return nil
	}

	for mailCount > 0 {
		rc, sout, serr = runShell("/bin/echo type 1 | " + mailxBin + " -N")
		if rc != 0 {
			return cli.NewExitError(
				fmt.Sprintf("Error: mail command failed with rc = %d:\n stdout:\n%s\n stderr:\n%s", rc, sout, serr), 1)
		}

		msg := strings.Split(sout, "\n")
		debug("Processing: " + msg[0])

		subj, from, to, part := "", "", "", "```"
		var text []string // array of messages to send respecting msgMaxSize
		var header string // header for each message to be able to trace the chain

		for i, line := range msg {
			// Skip "Message  N:" line
			if i == 0 {
				continue
			}

			// Skip "Held N messages in" lines
			if matched, _ := regexp.MatchString("^Held \\d+ messages? in ", line); matched {
				break
			}

			if matched, _ := regexp.MatchString("^Subject: ", line); matched {
				subj = line[strings.Index(line, ":")+2:]
			}
			if matched, _ := regexp.MatchString("^From: ", line); matched {
				from = line[strings.Index(line, ":")+2:]
			}
			if matched, _ := regexp.MatchString("^To: ", line); matched {
				to = line[strings.Index(line, ":")+2:]
			}

			part += line + "\n"
			if len(part) > msgMaxSize {
				if header == "" {
					header = "*" + to + "*: _" + subj + "_"
				}
				part = ":e-mail: " + header + part + "```"
				text = append(text, part)
				part = "```"
			}
		}

		if header == "" {
			header = "*" + to + "*: _" + subj + "_"
		}
		if len(part) > 3 {
			part = ":e-mail: " + header + part + "```"
			text = append(text, part)
		}

		debug(from + " -> " + to + "|" + subj)
		debug(" chunks: " + strconv.Itoa(len(text)))

		for _, line := range text {
			payload := &slack.Message{}
			payload.Text = line
			payload.Channel = config.Channel
			payload.Username = config.Username

			slack.Debug = config.Debug
			if err := slack.Send(config.Webhook, payload); err != nil {
				return cli.NewExitError(
					fmt.Sprintf("Error: %v", err), 2)
			}
		}

		debug("Message sent, deleting from mailbox")
		rc, sout, serr = runShell("/bin/echo d 1 | " + mailxBin + " -N")
		if rc != 0 {
			payload := &slack.Message{}
			payload.Text = ":exclamation: failed to remove message from inbox - " + header
			payload.Channel = config.Channel
			payload.Username = config.Username
			return cli.NewExitError(
				fmt.Sprintf("Error: mail command failed with rc = %d:\n stdout:\n%s\n stderr:\n%s", rc, sout, serr), 1)
		}

		info("Message processed: " + from + " -> " + to + " | " + subj)
		mailCount--
	}

	return nil
}
