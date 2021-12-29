package main

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"log"
	"math"
	"os"
	"time"

	tgapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var store = MustNewPostgres(os.Getenv("DATABASE_URL"))

func main() {
	log.Printf("Start 363Bot")

	token := os.Getenv("TOKEN")
	if token == "" {
		panic("no token")
	}

	tg, err := tgapi.NewBotAPI(token)
	if err != nil {
		panic(err)
	}

	ctx := context.Background()

	defer func() {
		if err := store.Close(ctx); err != nil {
			panic(err)
		}
	}()

	updCfg := tgapi.NewUpdate(0)
	updCfg.Timeout = 60
	updates := tg.GetUpdatesChan(updCfg)

	// go func() {
	// 	if err := sendDairyMsg(tg); err != nil {
	// 		panic(err)
	// 	}
	// }()

	go func() {
		if err := startAskJob(ctx, tg); err != nil {
			panic(err)
		}
	}()

	if err := handleUpdates(updates, ctx, tg); err != nil {
		panic(err)
	}
}

func sendDairyMsg(tg *tgapi.BotAPI) error {
	users, err := store.Users(context.Background())
	if err != nil {
		return fmt.Errorf("get users: %w", err)
	}

	for _, u := range users {
		log.Printf("send dairy %s", u.name)

		m := `Э-йоу, братишка, движению от всей души! 
		По поводу Dairy`

		msg := tgapi.NewMessage(u.id, m)
		msg.ReplyMarkup = tgapi.NewInlineKeyboardMarkup(
			tgapi.NewInlineKeyboardRow(
				tgapi.NewInlineKeyboardButtonData("Да", "++"),
				tgapi.NewInlineKeyboardButtonData("Нет", "--"),
			),
		)

		_, err := tg.Send(msg)
		if err != nil {
			panic(err)
		}

	}

	return nil
}

func handleUpdates(updates tgapi.UpdatesChannel, ctx context.Context, tg *tgapi.BotAPI) error {
	for u := range updates {
		if u.Message != nil {
			var (
				userID   = u.Message.From.ID
				userName = u.Message.From.UserName
				msgText  = u.Message.Text
			)

			if msgText == "/start" {
				if err := store.AddUser(ctx, userID, userName); err != nil {
					_, err := tg.Send(tgapi.NewMessage(userID, "Ты уже зареган, еблан!"))
					if err != nil {
						return err
					}
					continue
				}

				msg := fmt.Sprintf(
					"Салам алейкум, %s! Отныне я буду спрашивать, пыхал ли ты вчера и следить за твоей кармой.",
					u.Message.From.FirstName,
				)

				_, err := tg.Send(tgapi.NewMessage(userID, msg))
				if err != nil {
					return err
				}

				log.Printf("New user: %s", userName)

				continue
			}

			if msgText == "/get_user" {
				user, err := store.User(ctx, userID)
				if err != nil {
					return err
				}

				msg := fmt.Sprintf("Ты не пыхал дней: %d\nТвоя карма :%d", user.daysWithoutWeed, user.karma)
				if user.daysWithoutWeed < 0 {
					msg = fmt.Sprintf(
						"Ты пыхаешь дней: %v\nТвоя карма: %d",
						math.Abs(float64(user.daysWithoutWeed)), user.karma,
					)
				}

				_, err = tg.Send(tgapi.NewMessage(userID, msg))
				if err != nil {
					return err
				}

				continue
			}

			_, err := tg.Send(tgapi.NewMessage(userID, "Много пиздишь"))
			if err != nil {
				return err
			}

			x := newVin(userID)
			log.Println(x)
			_, err = tg.Send(x)
			if err != nil {
				return err
			}
		}

		if u.CallbackQuery != nil {
			userID := u.CallbackQuery.From.ID

			switch u.CallbackData() {
			case "--":
				_, err := tg.Send(tgapi.NewMessage(userID, "Ты долбоеб. Я никак это не обработаю"))
				if err != nil {
					panic(err)
				}
			case "++":
				go starDairyJob(ctx, tg, userID)
			// Пыхал
			case "+":
				user, err := store.User(ctx, userID)
				if err != nil {
					return err
				}

				karma, days := getKarma(user.daysWithoutWeed, user.karma, true)

				user.karma = karma
				user.daysWithoutWeed = days

				if err := store.UpdateUser(user); err != nil {
					return err
				}
				_, err = tg.Send(tgapi.NewMessage(userID, "fuck you"))
				if err != nil {
					return err
				}

				_, err = tg.Send(newDucalis(userID))
				if err != nil {
					return err
				}

				user, err = store.User(ctx, userID)
				if err != nil {
					return err
				}

				msg := fmt.Sprintf("Ты не пыхал дней: %d\nТвоя карма :%d", user.daysWithoutWeed, user.karma)
				if user.daysWithoutWeed < 0 {
					msg = fmt.Sprintf(
						"Ты пыхаешь дней: %v\nТвоя карма: %d",
						math.Abs(float64(user.daysWithoutWeed)), user.karma,
					)
				}

				_, err = tg.Send(tgapi.NewMessage(userID, msg))
				if err != nil {
					return err
				}
			// Не пыхал
			case "-":
				user, err := store.User(ctx, userID)
				if err != nil {
					return err
				}

				karma, daysWithoutWeed := getKarma(user.daysWithoutWeed, user.karma, false)

				user.karma = karma
				user.daysWithoutWeed = daysWithoutWeed

				if err := store.UpdateUser(user); err != nil {
					return err
				}

				_, err = tg.Send(newCoop(userID))
				if err != nil {
					return err
				}

				_, err = tg.Send(tgapi.NewMessage(userID, "good for you"))
				if err != nil {
					return err
				}

				msg := fmt.Sprintf("Ты не пыхал дней: %d\nТвоя карма :%d", user.daysWithoutWeed, user.karma)
				if user.daysWithoutWeed < 0 {
					msg = fmt.Sprintf(
						"Ты пыхаешь дней: %v\nТвоя карма: %d",
						math.Abs(float64(user.daysWithoutWeed)), user.karma,
					)
				}

				_, err = tg.Send(tgapi.NewMessage(userID, msg))
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func starDairyJob(ctx context.Context, tg *tgapi.BotAPI, id int64) {
	log.Printf("start dairy job for %d", id)

	_, err := tg.Send(tgapi.NewMessage(id, "Все правильно. Что же каков был день сегодняшний?"))
	if err != nil {
		panic(err)
	}

	for {
		<-waitDairy()

		_, err := tg.Send(tgapi.NewMessage(id, "Каков был день сегодняшний?"))
		if err != nil {
			panic(err)
		}
	}
}

func waitDairy() <-chan time.Time {
	loc, err := time.LoadLocation("Europe/Moscow")
	if err != nil {
		panic(err)
	}

	now := time.Now().In(loc)

	yyyy, mm, dd := now.Date()
	nextMorning := time.Date(yyyy, mm, dd+1, 22, 0, 0, 0, now.Location()) // <== work

	return time.After(nextMorning.Sub(now))
}

func startAskJob(ctx context.Context, tg *tgapi.BotAPI) error {
	log.Println("Start ask job")

	for {
		<-wait()

		users, err := store.Users(ctx)
		if err != nil {
			return fmt.Errorf("get users: %w", err)
		}

		for _, u := range users {
			_, err := tg.Send(askAboutWeedMsg(u.id))
			if err != nil {
				log.Printf("send tg msg to user %d: %v", u.id, err)
			}
		}
	}
}

func askAboutWeedMsg(id int64) tgapi.MessageConfig {
	msg := tgapi.NewMessage(id, "Пыхал вчера?")
	msg.ReplyMarkup = tgapi.NewInlineKeyboardMarkup(
		tgapi.NewInlineKeyboardRow(
			tgapi.NewInlineKeyboardButtonData("Да...", "+"),
			tgapi.NewInlineKeyboardButtonData("Нет!", "-"),
		),
	)
	return msg
}

func wait() <-chan time.Time {
	loc, err := time.LoadLocation("Europe/Moscow")
	if err != nil {
		panic(err)
	}

	now := time.Now().In(loc)

	yyyy, mm, dd := now.Date()
	nextMorning := time.Date(yyyy, mm, dd+1, 11, 0, 0, 0, now.Location()) // <== work

	return time.After(nextMorning.Sub(now))
}

/*
	=========
	STICKERS
	=========
*/

// Ducalis
var (
	//go:embed assets/ducalis.jpg
	ducalis []byte

	ducalisReader = tgapi.FileReader{
		Name:   "Ducalis",
		Reader: bytes.NewReader(ducalis),
	}
)

func newDucalis(id int64) tgapi.StickerConfig {
	return tgapi.NewSticker(id, ducalisReader)
}

// Cooper
var (
	//go:embed assets/coop.jpg
	coop []byte

	coopReader = tgapi.FileReader{
		Name:   "Cooper",
		Reader: bytes.NewReader(coop),
	}
)

func newCoop(id int64) tgapi.StickerConfig {
	return tgapi.NewSticker(id, coopReader)
}

// Vin
var (
	//go:embed assets/vin.jpg
	vin []byte

	vinReader = tgapi.FileReader{
		Name:   "Vin",
		Reader: bytes.NewReader(vin),
	}
)

func newVin(id int64) tgapi.StickerConfig {
	return tgapi.NewSticker(id, coopReader)
}
