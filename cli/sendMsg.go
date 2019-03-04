package main

import (
	"fmt"

	"github.com/ivanpesin/slack-notify/slack"
	"gopkg.in/urfave/cli.v1"
)

func sendMsg(c *cli.Context) error {

	msg := c.Args().First()
	if msg == "" {
		msg = "This is a test message."
	}

	// create payload for slack message
	payload := &slack.Message{}
	payload.Channel = config.Channel
	payload.Username = config.Username
	payload.Text = msg

	slack.Debug = config.Debug
	if err := slack.Send(config.Webhook, payload); err != nil {
		return cli.NewExitError(
			fmt.Sprintf("Error: %v", err), 2)
	}
	info("Message sent: " + msg)
	return nil
}
