package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/ghodss/yaml"
	"gopkg.in/urfave/cli.v1"
)

var buildTime string
var buildCommit string

var configFilename = "/etc/slack-notify.conf"
var config struct {
	Webhook  string
	Channel  string
	Username string

	Debug   bool
	Verbose bool

	MailWebhook  string `json:"mail_webhook"`
	MailChannel  string `json:"mail_channel"`
	MailUsername string `json:"mail_username"`

	MonitWebhook  string `json:"monit_webhook"`
	MonitChannel  string `json:"monit_channel"`
	MonitUsername string `json:"monit_username"`
}

func init() {
	log.SetFlags(0)
	log.SetOutput(new(UTCLogger))
}

func main() {
	app := cli.NewApp()
	app.Name = "slack-notify"
	app.Usage = "send information to slack via webhook"
	app.Version = "2.0.0/" + buildTime + "/" + buildCommit

	hn, _ := os.Hostname()

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "webhook",
			Usage:  "Slack webhook for sending messages",
			EnvVar: "SN_WEBHOOK",
		},
		cli.StringFlag{
			Name:   "channel",
			Value:  "#random",
			Usage:  "Slack channel where to send message",
			EnvVar: "SN_CHANNEL",
		},
		cli.StringFlag{
			Name:   "username",
			Value:  "slack-notify@" + hn,
			Usage:  "Slack username",
			EnvVar: "SN_USERNAME",
		},
		cli.StringFlag{
			Name:   "config, f",
			Usage:  "Read configuration from file.",
			EnvVar: "SN_CONFIG",
		},
		cli.BoolFlag{
			Name:  "debug, d",
			Usage: "Enable debugging output",
		},
		cli.BoolFlag{
			Name:  "verbose, i",
			Usage: "Enable verbose output",
		},
	}

	app.Commands = []cli.Command{
		{
			Name:   "msg",
			Usage:  "send a plain message to slack",
			Action: sendMsg,
		},
		{
			Name:   "monit",
			Usage:  "monit mode, read information from env variables",
			Action: sendMonit,
		},
		{
			Name:   "mail",
			Usage:  "relay user's email to slack",
			Action: sendMail,
		},
	}

	app.Before = validateState

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err.Error())
	}
}

func validateState(c *cli.Context) error {

	if c.GlobalBool("debug") {
		config.Debug = true
	}
	if c.GlobalBool("verbose") {
		config.Verbose = true
	}

	failOnMissingFile := false
	if c.GlobalString("config") != "" {
		configFilename = c.GlobalString("config")
		failOnMissingFile = true
	}
	debug("Config file: " + configFilename)

	_, err := os.Stat(configFilename)

	if os.IsNotExist(err) && failOnMissingFile {
		return cli.NewExitError(
			fmt.Sprintf("Error: Unable to read config file %s: %v", configFilename, err),
			1)
	}

	if !os.IsNotExist(err) {
		buf, err := ioutil.ReadFile(configFilename)
		if err != nil {
			return cli.NewExitError(
				fmt.Sprintf("Error: Unable to read config file %s: %v", configFilename, err),
				1)
		}

		err = yaml.Unmarshal(buf, &config)
		if err != nil {
			return cli.NewExitError(
				fmt.Sprintf("Error: Unable to parse config file %s: %v", configFilename, err),
				1)
		}

		// override config with command line parameters
		if c.GlobalBool("debug") {
			config.Debug = true
		}
		if c.GlobalBool("verbose") {
			config.Verbose = true
		}
		debug(fmt.Sprintf("Config read: %+v\n", config))

		if config.Webhook != "" && c.GlobalString("webhook") == "" {
			c.GlobalSet("webhook", config.Webhook)
		}
		if config.Channel != "" && c.GlobalString("channel") == "" {
			c.GlobalSet("channel", config.Channel)
		}
		if config.Username != "" && c.GlobalString("username") == "" {
			c.GlobalSet("username", config.Username)
		}

	}

	if c.GlobalString("webhook") == "" {
		return cli.NewExitError("Error: Slack webhook is not specified.", 1)
	}
	config.Webhook, config.Channel, config.Username =
		c.GlobalString("webhook"), c.GlobalString("channel"), c.GlobalString("username")

	printConfig(c)

	return nil
}
