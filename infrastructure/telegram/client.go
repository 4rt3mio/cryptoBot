package telegram

import (
	"net/http"
	"os"

	domain "github.com/4rt3mio/cryptoCore/domain/telegram"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Client struct {
	bot *tgbotapi.BotAPI
}

func NewClient() (*Client, error) {
	token := os.Getenv("TELEGRAM_APITOKEN")
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, err
	}
	bot.Debug = false
	return &Client{bot: bot}, nil
}

func NewClientWithHTTP(token string, httpClient *http.Client) (*Client, error) {
	apiEndpoint := tgbotapi.APIEndpoint
	bot, err := tgbotapi.NewBotAPIWithClient(token, apiEndpoint, httpClient)
	if err != nil {
		return nil, err
	}
	return &Client{bot: bot}, nil
}

func (c *Client) GetUpdatesChan(offset int, timeoutSeconds int) (<-chan domain.Update, error) {
	cfg := tgbotapi.NewUpdate(offset)
	cfg.Timeout = timeoutSeconds
	raw := c.bot.GetUpdatesChan(cfg)
	out := make(chan domain.Update)
	go func() {
		defer close(out)
		for u := range raw {
			out <- domain.Update{UpdateID: u.UpdateID, Message: u.Message, Callback: u.CallbackQuery}
		}
	}()
	return out, nil
}

func (c *Client) SendMessage(chatID int64, text string, markup interface{}) error {
	msg := tgbotapi.NewMessage(chatID, text)
	if markup != nil {
		msg.ReplyMarkup = markup
	}
	_, err := c.bot.Send(msg)
	return err
}
