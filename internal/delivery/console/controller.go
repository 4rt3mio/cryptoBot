package console

// import (
// 	"bufio"
// 	"context"
// 	"fmt"
// 	"os"
// 	"strconv"
// 	"strings"
// 	"sync"
// 	"time"

// 	"CourseWork/internal/domain/model"
// 	subUsecase "CourseWork/internal/usecase/subscription"
// 	userUsecase "CourseWork/internal/usecase/user"
// )

// type ConsoleController struct {
// 	UserManager         *userUsecase.Manager
// 	SubscriptionManager *subUsecase.Manager
// 	MonitorService      *subUsecase.MonitorService
// 	monitors            map[int]context.CancelFunc
// 	mutex               sync.Mutex
// 	notifCh             chan string
// }

// func NewConsoleController(userMgr *userUsecase.Manager, subMgr *subUsecase.Manager, monSvc *subUsecase.MonitorService) *ConsoleController {
// 	return &ConsoleController{
// 		UserManager:         userMgr,
// 		SubscriptionManager: subMgr,
// 		MonitorService:      monSvc,
// 		monitors:            make(map[int]context.CancelFunc),
// 		notifCh:             make(chan string, 100),
// 	}
// }

// func (c *ConsoleController) Start() {
// 	reader := bufio.NewReader(os.Stdin)
// 	fmt.Println("Добро пожаловать в систему оповещений по криптовалютам!")
// 	fmt.Print("Введите ваше имя: ")
// 	userName, _ := reader.ReadString('\n')
// 	userName = strings.TrimSpace(userName)
// 	if userName == "" {
// 		userName = "anonymous"
// 	}
// 	userObj := c.UserManager.Register(userName)
// 	fmt.Printf("Привет, %s! Ваш ID: %s\n", userObj.Name, userObj.ID)
// 	fmt.Println("Доступные команды:")
// 	fmt.Println("/start — помощь")
// 	fmt.Println("/subscribe <Название> <Символ> <Целевая_цена>")
// 	fmt.Println("/unsubscribe <ID>")
// 	fmt.Println("/change <ID> <Новая_цена>")
// 	fmt.Println("/list — показать подписки")
// 	fmt.Println("/exit — выход")

// 	go func() {
// 		for msg := range c.notifCh {
// 			fmt.Print("\033[s")
// 			fmt.Print("\033[1A")
// 			fmt.Print("\033[1000D")
// 			fmt.Print("\033[2K")
// 			fmt.Println(msg)
// 			fmt.Print("\033[u")
// 		}
// 	}()

// 	for {
// 		fmt.Print("> ")
// 		input, err := reader.ReadString('\n')
// 		if err != nil {
// 			fmt.Println("Ошибка чтения:", err)
// 			continue
// 		}
// 		input = strings.TrimSpace(input)
// 		if input == "" {
// 			continue
// 		}
// 		parts := strings.Split(input, " ")
// 		cmd := parts[0]
// 		switch cmd {
// 		case "/start":
// 			fmt.Println("Команды:")
// 			fmt.Println("/subscribe <Название> <Символ> <Целевая_цена> — подписка")
// 			fmt.Println("/unsubscribe <ID> — удаление подписки")
// 			fmt.Println("/change <ID> <Новая_цена> — изменение цены подписки")
// 			fmt.Println("/list — список подписок")
// 			fmt.Println("/exit — выход")
// 		case "/subscribe":
// 			// Пример: /subscribe Bitcoin BTC 30000
// 			if len(parts) < 4 {
// 				fmt.Println("Использование: /subscribe <Название> <Символ> <Целевая_цена>")
// 				continue
// 			}
// 			tokenName := parts[1]
// 			tokenSymbol := strings.ToUpper(parts[2])
// 			targetPrice, err := strconv.ParseFloat(parts[3], 64)
// 			if err != nil {
// 				fmt.Println("Неверный формат целевой цены:", err)
// 				continue
// 			}
// 			err = c.SubscriptionManager.Subscribe(userObj.ID, tokenName, tokenSymbol, targetPrice)
// 			if err != nil {
// 				fmt.Println("Ошибка подписки:", err)
// 				continue
// 			}
// 			fmt.Printf("Подписка на %s (%s) с порогом %.2f создана.\n", tokenName, tokenSymbol, targetPrice)
// 			subID := c.SubscriptionManager.NextID() - 1
// 			c.startMonitoring(subID, model.Token{
// 				Name:      tokenName,
// 				Symbol:    tokenSymbol,
// 				Threshold: targetPrice,
// 			})
// 		case "/unsubscribe":
// 			if len(parts) < 2 {
// 				fmt.Println("Использование: /unsubscribe <ID>")
// 				continue
// 			}
// 			id, err := strconv.Atoi(parts[1])
// 			if err != nil {
// 				fmt.Println("Неверный формат ID:", err)
// 				continue
// 			}
// 			err = c.SubscriptionManager.Unsubscribe(userObj.ID, id)
// 			if err != nil {
// 				fmt.Println("Ошибка при удалении подписки:", err)
// 			} else {
// 				fmt.Printf("Подписка [%d] удалена.\n", id)
// 				c.cancelMonitoring(id)
// 			}
// 		case "/change":
// 			if len(parts) < 3 {
// 				fmt.Println("Использование: /change <ID> <Новая_цена>")
// 				continue
// 			}
// 			id, err := strconv.Atoi(parts[1])
// 			if err != nil {
// 				fmt.Println("Неверный формат ID:", err)
// 				continue
// 			}
// 			newPrice, err := strconv.ParseFloat(parts[2], 64)
// 			if err != nil {
// 				fmt.Println("Неверный формат новой цены:", err)
// 				continue
// 			}
// 			err = c.SubscriptionManager.UpdateSubscription(userObj.ID, id, newPrice)
// 			if err != nil {
// 				fmt.Println("Ошибка изменения подписки:", err)
// 			} else {
// 				fmt.Printf("Подписка [%d] обновлена до цены %.2f.\n", id, newPrice)
// 				c.cancelMonitoring(id)
// 				subs, err := c.SubscriptionManager.ListSubscriptions(userObj.ID)
// 				if err != nil {
// 					fmt.Println("Ошибка получения обновлённой подписки:", err)
// 					continue
// 				}
// 				var updatedSub model.Subscription
// 				for _, s := range subs {
// 					if s.ID == id {
// 						updatedSub = s
// 						break
// 					}
// 				}
// 				if updatedSub.ID == 0 {
// 					fmt.Println("Подписка не найдена после обновления.")
// 					continue
// 				}
// 				c.startMonitoring(id, updatedSub.Token)
// 			}
// 		case "/list":
// 			subs, err := c.SubscriptionManager.ListSubscriptions(userObj.ID)
// 			if err != nil {
// 				fmt.Println("Ошибка получения подписок:", err)
// 				continue
// 			}
// 			if len(subs) == 0 {
// 				fmt.Println("Нет активных подписок.")
// 			} else {
// 				fmt.Println("Ваши подписки:")
// 				for _, s := range subs {
// 					fmt.Printf("ID: %d | %s (%s) | Целевая цена: %.2f\n", s.ID, s.Token.Name, s.Token.Symbol, s.Token.Threshold)
// 				}
// 			}
// 		case "/exit":
// 			fmt.Println("Выход из программы.")
// 			return
// 		default:
// 			fmt.Println("Неизвестная команда. Введите /start для справки.")
// 		}
// 		time.Sleep(100 * time.Millisecond)
// 	}
// }

// func (c *ConsoleController) startMonitoring(subID int, token model.Token) {
// 	ctx, cancel := context.WithCancel(context.Background())
// 	c.mutex.Lock()
// 	c.monitors[subID] = cancel
// 	c.mutex.Unlock()

// 	go func() {
// 		c.MonitorService.MonitorToken(ctx, token, func(msg string) {
// 			c.notifCh <- fmt.Sprintf("[Подписка %d] %s", subID, msg)
// 		})
// 	}()
// }

// func (c *ConsoleController) cancelMonitoring(subID int) {
// 	c.mutex.Lock()
// 	if cancel, ok := c.monitors[subID]; ok {
// 		cancel()
// 		delete(c.monitors, subID)
// 	}
// 	c.mutex.Unlock()
// }
