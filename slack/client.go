package slack

import (
	"fmt"

	slackClient "github.com/nlopes/slack"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// Client is a slack client
type Client struct {
	Slack  *slackClient.Client
	logger *logrus.Logger
}

// New returns a webhook client
func New(token string, logger *logrus.Logger) *Client {
	client := slackClient.New(token)
	return &Client{
		Slack:  client,
		logger: logger,
	}
}

//GetSlackChannelID returns the chanel id from an email
func (c *Client) GetSlackChannelID(email string) (string, error) {
	user, err := c.Slack.GetUserByEmail(email)
	if err != nil {
		return "", errors.Wrap(err, "could not find slack user for email")
	}
	if user == nil {
		return "", errors.New("email not found")
	}
	c.logger.Info(fmt.Sprintf("userID: %s", user.ID))
	_, _, channelID, err := c.Slack.OpenIMChannel(user.ID)
	return channelID, errors.Wrap(err, "could not open dm channel with user")
}

// PostMessage is DEPRECATED
func (c *Client) PostMessage(message Message) error {
	channelID, err := c.GetSlackChannelID(message.Email)
	if err != nil {
		return err
	}
	return c.postMessage(channelID, message)
}

// SendMessageToUserByEmail posts a message
func (c *Client) SendMessageToUserByEmail(email, message string, attachments []slackClient.Attachment) error {
	channelID, err := c.GetSlackChannelID(email)
	if err != nil {
		return errors.Wrapf(err, "could not find slack user with email %s", email)
	}
	return c.postMessage(channelID, Message{Text: message, Attachments: attachments})
}

// SendMessageToUser will send the given text to the specified userID.
func (c *Client) SendMessageToUser(userID, message string) error {
	_, _, channelID, err := c.Slack.OpenIMChannel(userID)
	if err != nil {
		return err
	}
	return c.postMessage(channelID, Message{Text: message})
}

func (c *Client) postMessage(channel string, message Message) error {
	options := []slackClient.MsgOption{
		slackClient.MsgOptionText(message.Text, true),
		slackClient.MsgOptionEnableLinkUnfurl(),
		slackClient.MsgOptionAttachments(message.Attachments...),
	}
	_, _, err := c.Slack.PostMessage(channel, options...)
	return errors.Wrap(err, "could not post message")
}
