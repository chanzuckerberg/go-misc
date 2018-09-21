package slack

import slackClient "github.com/nlopes/slack"

// Message is a slack message
type Message struct {
	Text        string `json:"text,omitempty"`
	Email       string
	Attachments []slackClient.Attachment
}
