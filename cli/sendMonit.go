package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/ivanpesin/slack-notify/slack"
	"gopkg.in/urfave/cli.v1"
)

var monit struct {
	service     string
	event       string
	description string
	host        string
	date        string
	color       string
}

func setMonitConfig() {
	if config.MonitWebhook != "" {
		config.Webhook = config.MonitWebhook
	}
	if config.MonitChannel != "" {
		config.Channel = config.MonitChannel
	}
	if config.MonitUsername != "" {
		config.Username = config.MonitUsername
	}
}

func sendMonit(c *cli.Context) error {

	setMonitConfig()
	readMonitData()

	// create payload for slack message
	payload := &slack.Message{}
	payload.Channel = c.GlobalString("channel")
	payload.Username = c.GlobalString("username")
	payload.Attachments = append(payload.Attachments, slack.Attachment{})
	payload.Attachments[0].Color = monit.color
	payload.Attachments[0].Fallback = fmt.Sprintf("%s: %s on %s\n%s", monit.service, monit.event, monit.host, monit.description)
	payload.Attachments[0].Text = fmt.Sprintf("`%s`: *%s*\n%s", monit.service, monit.event, monit.description)
	payload.Attachments[0].MrkdwnIn = []string{"text"}
	payload.Attachments[0].Fields = []slack.Field{
		slack.Field{Title: "Date", Value: monit.date, Short: true},
		slack.Field{Title: "Host", Value: monit.host, Short: true},
	}

	slack.Debug = config.Debug
	if err := slack.Send(c.GlobalString("webhook"), payload); err != nil {
		return cli.NewExitError(
			fmt.Sprintf("Error: %v", err), 2)
	}

	info(fmt.Sprintf("Monit event sent: %s %s", monit.service, monit.event))
	return nil

}

func readMonitData() {
	// Get monit information
	monit.service = os.Getenv("MONIT_SERVICE")
	monit.event = os.Getenv("MONIT_EVENT")
	monit.description = os.Getenv("MONIT_DESCRIPTION")
	monit.host = os.Getenv("MONIT_HOST")
	monit.date = os.Getenv("MONIT_DATE")

	if len(monit.service) == 0 &&
		len(monit.event) == 0 &&
		len(monit.description) == 0 &&
		len(monit.host) == 0 &&
		len(monit.date) == 0 {

		monit.service = "ECHO"
		monit.event = "Test"
		monit.description = "If you read this message, your system can send messages to slack"
		monit.host, _ = os.Hostname()
		monit.date = time.Now().UTC().Format(time.RFC3339)
		monit.color = "good"
	}

	if monit.color == "" {
		monit.color = "danger"
		if strings.Contains(monit.event, "succe") || strings.Contains(monit.event, "Exists") {
			monit.color = "good"
		}
	}
}
