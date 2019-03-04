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
	payload.Channel = c.GlobalString("channel")
	payload.Username = c.GlobalString("username")
	payload.Text = msg

	slack.Debug = config.Debug
	if err := slack.Send(c.GlobalString("webhook"), payload); err != nil {
		return cli.NewExitError(
			fmt.Sprintf("Error: %v", err), 2)
	}
	info("Message sent: " + msg)
	return nil
}
