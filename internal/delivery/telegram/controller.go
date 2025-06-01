package telegram

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"

	model "github.com/4rt3mio/cryptoCore/domain/model"
	domain "github.com/4rt3mio/cryptoCore/domain/telegram"
	currencyUsecase "github.com/4rt3mio/cryptoCore/usecase/currency"
	notifUsecase "github.com/4rt3mio/cryptoCore/usecase/notification"
	subUsecase "github.com/4rt3mio/cryptoCore/usecase/subscription"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type TelegramController struct {
	client            domain.TelegramClient
	currencyAnalytics *currencyUsecase.Analytics
	subMgr            *subUsecase.Manager
	monitorSvc        *subUsecase.MonitorService
	notifier          notifUsecase.Notifier
	currencyMgr       *currencyUsecase.Manager
	keyboard          tgbotapi.ReplyKeyboardMarkup
	monitors          map[int]context.CancelFunc
	mu                sync.Mutex
}

func NewTelegramController(
	client domain.TelegramClient,
	currencyAnalytics *currencyUsecase.Analytics,
	subMgr *subUsecase.Manager,
	monitorSvc *subUsecase.MonitorService,
	notifier notifUsecase.Notifier,
	currencyMgr *currencyUsecase.Manager,
) *TelegramController {
	row1 := tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton("➕ Подписаться"),
		tgbotapi.NewKeyboardButton("➖ Отписаться"),
	)
	row2 := tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton("✏️ Изменить цену"),
		tgbotapi.NewKeyboardButton("📋 Список"),
	)
	row3 := tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton("🌐 Валюты"),
		tgbotapi.NewKeyboardButton("〽 Аналитика"),
	)
	km := tgbotapi.NewReplyKeyboard(row1, row2, row3)
	km.ResizeKeyboard = true
	km.OneTimeKeyboard = false

	return &TelegramController{
		client:            client,
		currencyAnalytics: currencyAnalytics,
		subMgr:            subMgr,
		monitorSvc:        monitorSvc,
		notifier:          notifier,
		currencyMgr:       currencyMgr,
		keyboard:          km,
		monitors:          make(map[int]context.CancelFunc),
	}
}

func (c *TelegramController) Start() {
	c.restoreAllSubscriptions()
	updates, _ := c.client.GetUpdatesChan(0, 60)
	for upd := range updates {
		if upd.Message == nil {
			continue
		}
		chatID := upd.Message.Chat.ID
		userID := fmt.Sprint(chatID)
		text := upd.Message.Text

		switch text {
		case "/start":
			c.client.SendMessage(chatID, "Добро пожаловать! Выберите действие:", c.keyboard)

		case "➕ Подписаться":
			c.client.SendMessage(chatID, "Введите через пробел: Название Символ Цена", nil)

		case "➖ Отписаться":
			c.client.SendMessage(chatID, "Введите ID подписки для удаления:", nil)

		case "✏️ Изменить цену":
			c.client.SendMessage(chatID, "Введите: ID Новая_цена", nil)

		case "📋 Список":
			subs, _ := c.subMgr.ListSubscriptions(userID)
			if len(subs) == 0 {
				c.client.SendMessage(chatID, "Нет активных подписок.", nil)
			} else {
				var lines []string
				for _, s := range subs {
					lines = append(lines, fmt.Sprintf(
						"%d: %s (%s) → %.2f",
						s.ID, s.Token.Name, s.Token.Symbol, s.Token.Threshold,
					))
				}
				c.client.SendMessage(chatID, strings.Join(lines, "\n"), nil)
			}

		case "🌐 Валюты":
			list, err := c.currencyMgr.List()
			if err != nil {
				c.client.SendMessage(chatID, "Ошибка получения списка валют", nil)
				break
			}
			for i := 0; i < len(list); i += 20 {
				end := i + 20
				if end > len(list) {
					end = len(list)
				}
				c.client.SendMessage(chatID, strings.Join(list[i:end], "\n"), nil)
			}

		case "〽 Аналитика":
			c.client.SendMessage(chatID, "Введите символ валюты (например: BTC):", nil)

		default:
			parts := strings.Fields(text)
			if len(parts) == 1 && isAlpha(text) {
				symbol := strings.ToUpper(text)
				trend, err := c.currencyAnalytics.GetDailyTrend(symbol)
				if err != nil {
					trend = "❌ Не удалось получить данные"
				}
				c.client.SendMessage(chatID, trend, nil)
				continue
			}

			if len(parts) == 3 {
				if price, err := strconv.ParseFloat(parts[2], 64); err == nil && isAlpha(parts[0]) && isAlpha(parts[1]) {
					fmt.Printf("Trying to subscribe: %s %s %.2f\n", parts[0], parts[1], price)

					exists, err := c.currencyExists(parts[1])
					if err != nil {
						fmt.Printf("Currency check error: %v\n", err)
						c.client.SendMessage(chatID, "❌ Ошибка проверки валюты", nil)
						continue
					}

					fmt.Printf("Currency exists: %v\n", exists)
					if !exists {
						currencies, _ := c.currencyMgr.List()
						fmt.Printf("Available currencies: %v\n", currencies)

						c.client.SendMessage(chatID, "❌ Валюта не найдена", nil)
						continue
					}

					err = c.subMgr.Subscribe(userID, parts[0], parts[1], price)
					if err != nil {
						c.client.SendMessage(chatID, "❌ Ошибка создания подписки", nil)
						continue
					}
					subs, _ := c.subMgr.ListSubscriptions(userID)
					if len(subs) == 0 {
						continue
					}
					newSub := subs[len(subs)-1]

					c.startMonitoring(newSub.ID, newSub.Token, userID)
					c.client.SendMessage(chatID, fmt.Sprintf(
						"✅ Подписка #%d создана",
						newSub.ID,
					), nil)
					continue
				}
			}
			if len(parts) == 1 {
				if id, err := strconv.Atoi(parts[0]); err == nil {
					err := c.subMgr.Unsubscribe(userID, id)
					if err != nil {
						if err.Error() == "subscription not found" {
							c.client.SendMessage(chatID, "❌ Подписка не найдена", nil)
						} else {
							c.client.SendMessage(chatID, "❌ Ошибка удаления", nil)
						}
					} else {
						c.client.SendMessage(chatID, "✅ Подписка удалена.", nil)
						c.cancelMonitoring(id)
					}
					continue
				}
			}
			if len(parts) == 2 {
				id, err1 := strconv.Atoi(parts[0])
				price, err2 := strconv.ParseFloat(parts[1], 64)
				if err1 == nil && err2 == nil {
					err := c.subMgr.UpdateSubscription(userID, id, price)
					if err != nil {
						if err.Error() == "subscription not found" {
							c.client.SendMessage(chatID, "❌ Подписка не найдена", nil)
						} else {
							c.client.SendMessage(chatID, "❌ Ошибка обновления", nil)
						}
					} else {
						c.client.SendMessage(chatID, "✅ Цена обновлена.", nil)
						c.cancelMonitoring(id)
						subs, _ := c.subMgr.ListSubscriptions(userID)
						for _, s := range subs {
							if s.ID == id {
								c.startMonitoring(id, s.Token, userID)
								break
							}
						}
					}
					continue
				}
			}

			c.client.SendMessage(chatID, "Неизвестная команда. Нажмите /start.", nil)
		}
	}
}

func (c *TelegramController) startMonitoring(subID int, token model.Token, userID string) {
	ctx, cancel := context.WithCancel(context.Background())
	c.mu.Lock()
	c.monitors[subID] = cancel
	c.mu.Unlock()

	go c.monitorSvc.MonitorToken(ctx, token, func(msg string) {
		c.notifier.Notify(userID, msg)
	})
}

func (c *TelegramController) cancelMonitoring(subID int) {
	c.mu.Lock()
	if cancel, ok := c.monitors[subID]; ok {
		cancel()
		delete(c.monitors, subID)
	}
	c.mu.Unlock()
}

func (c *TelegramController) restoreAllSubscriptions() {
	subs, err := c.subMgr.ListAllSubscriptions()
	if err != nil {
		fmt.Printf("Ошибка восстановления подписок: %v\n", err)
		return
	}

	for _, sub := range subs {
		c.startMonitoring(sub.ID, sub.Token, sub.UserID)
	}
}

func isAlpha(s string) bool {
	for _, r := range s {
		if !((r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z')) {
			return false
		}
	}
	return true
}

func (c *TelegramController) currencyExists(symbol string) (bool, error) {
    currencies, err := c.currencyMgr.List()
    if err != nil {
        return false, err
    }

    upperSymbol := strings.ToUpper(symbol)
    
    for _, currency := range currencies {
        if strings.EqualFold(currency, upperSymbol) {
            return true, nil
        }
    }
    return false, nil
}
