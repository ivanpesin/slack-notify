package slack

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

// Debug controls debugging output
var Debug = false

// Message is a top level structure for slack messages
type Message struct {
	Channel     string       `json:"channel"`
	Username    string       `json:"username"`
	Emoji       string       `json:"emoji"`
	Text        string       `json:"text"`
	Attachments []Attachment `json:"attachments"`
}

// Attachment for slack
type Attachment struct {
	Fallback string   `json:"fallback"`
	Color    string   `json:"color"`
	Text     string   `json:"text"`
	MrkdwnIn []string `json:"mrkdwn_in"`
	Fields   []Field  `json:"fields"`
}

// Field is a struct to describe fields in slack message
type Field struct {
	Title string `json:"title"`
	Value string `json:"value"`
	Short bool   `json:"short"`
}

func debug(show bool, s string) {
	if Debug {
		fmt.Printf("[slack] %s D: %s\n", time.Now().UTC().Format(time.RFC3339), s)
	}
}

// Send sends message to slack
func Send(url string, m *Message) error {
	// create JSON
	buf, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to create JSON payload: %v", err)
	}

	// post to Slack
	debug(Debug, fmt.Sprintf("Sending to %s, payload:\n%s", url, buf))
	resp, err := http.Post(url, "application/json", bytes.NewReader(buf))
	if err != nil {
		return fmt.Errorf("failed to send message to slack: %v", err)
	}

	// process response
	b, _ := ioutil.ReadAll(resp.Body)
	debug(Debug, fmt.Sprintf("Response received, status: %s\nBody:\n%s", resp.Status, b))

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("error status received: %v\nBody:\n%s", resp.StatusCode, b)
	}
	return nil
}
