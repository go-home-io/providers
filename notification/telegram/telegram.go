package main

import (
	"net/url"
	"strconv"

	"github.com/kimrgrey/go-telegram"
	"github.com/pkg/errors"
	"go-home.io/x/server/plugins/common"
	"go-home.io/x/server/plugins/notification"
)

// Describes telegram API send message response.
type messageResponse struct {
	Success bool              `json:"ok"`
	Result  *telegram.Message `json:"result"`
}

// TelegramNotification implements notification system.
type TelegramNotification struct {
	Settings *Settings

	logger common.ILoggerProvider
	client *telegram.Client
}

// Init performs initial plugin initialization.
func (t *TelegramNotification) Init(data *notification.InitDataNotification) error {
	t.logger = data.Logger

	t.client = telegram.NewClient(t.Settings.Token)
	me := t.client.GetMe()

	if 0 == me.ID {
		return errors.New("failed to get bot details")
	}

	t.logger.Info("Telegram bot was loaded", "bot_id", strconv.Itoa(me.ID), "bot_name", me.FirstName)

	return nil
}

// Message tries to send a message through telegram API.
func (t *TelegramNotification) Message(msg string) error {
	resp := &messageResponse{}
	t.client.Call("sendMessage", url.Values{
		"chat_id": []string{t.Settings.ChatID},
		"text":    []string{msg},
	}, resp)

	if !resp.Success {
		err := errors.New("telegram api error")
		t.logger.Error("Failed to call telegram api", err)
		return err
	}

	return nil
}
