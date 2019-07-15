package slack

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/speedtest-monitor/app/configuration"
	log "github.com/sirupsen/logrus"
)

// SlackClient contains basic info about the app to send messages to a Slack Webhook
type SlackClient struct {
	Endpoint string
	AppName  string
	Messages []Attachment `json:"attachments"`
}

// Attachment contains the information of each message to be sent to Slack
type Attachment struct {
	Fallback   string `json:"fallback"`
	Color      string `json:"color"`
	AuthorName string `json:"author_name"`
	Text       string `json:"text"`
}

type SlackMessage struct {
	Text     string       `json:"text"`
	Messages []Attachment `json:"attachments"`
}

// NewSlackClient instantiates a SlackClient struct to use for sending messages
func NewSlackClient(config *configuration.Configuration) *SlackClient {
	sc := &SlackClient{
		Endpoint: config.SlackEndpoint,
		AppName:  config.AppName,
	}
	return sc
}

// newAttachment creates an attachment to be added into the next event that sends messages
func newAttachment(message, appName string, isAlert bool) Attachment {
	var clr string
	if isAlert {
		clr = "#a6364f"
	} else {
		clr = "#36a64f"
	}
	return Attachment{
		Fallback:   message,
		Color:      clr,
		AuthorName: appName,
		Text:       message,
	}
}

// AddMessage creates an info message to be sent to Slack and adds it to the current stack of messages
func (sc *SlackClient) AddMessage(message string) {
	sc.Messages = append(sc.Messages, newAttachment(message, sc.AppName, false))
}

// AddAlert creates an alert message to be sent to Slack and adds it to the current stack of messages
func (sc *SlackClient) AddAlert(message string) {
	sc.Messages = append(sc.Messages, newAttachment(message, sc.AppName, true))
}

// SendMessages gets the messages in the stack and sends them to the Slack Webhook,
// and optionally empties the message stack
func (sc *SlackClient) SendMessages() {
	msg := SlackMessage{Messages: sc.Messages, Text: fmt.Sprintf("Message from %s", sc.AppName)}
	jsonData, _ := json.Marshal(msg)
	response, err := http.Post(sc.Endpoint, "application/json", bytes.NewBuffer(jsonData))

	sc.Messages = nil
	log.Debugf("Slack response status code: %d", response.StatusCode)
	if err != nil {
		log.Errorf("Error while sending Slack message: %s", err)
	} else {
		defer response.Body.Close()
		data, _ := ioutil.ReadAll(response.Body)
		log.Debug(string(data))
	}
}
