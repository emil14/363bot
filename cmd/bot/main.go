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

//go:embed assets/dukalis.jpg
var ducalis []byte

//go:embed assets/coop.jpg
var coop []byte

//go:embed assets/vin.jpg
var vin []byte

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

	go func() {
		if err := sendDairyMsg(tg); err != nil {
			panic(err)
		}
	}()

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
		
		Хочешь я буду крепить каждый вечер, скажем, в 22:00 по МСК, оставить маленькую запись об том, каков был день сегоднящний?

		Сейчас это будет просто в формате напоминалки, можешь писать сообщение боту, можешь в свой канальчки, можно назвать его, скажем, "Dairy
		Потом мы улучшим это. 
		
		Сообщение должно быть совсем короткое (пока опционально), мы это еше обдумаем, но почему-то мне кажется хорошей идей сделать его коротким.
		Пусть сообщения будут просто "Хороший день" или "Вот и день прошел ну и нахуй он пошел", конечно, они могут более развернутыми.
		Но не надо чтобы оно могло быть большое, тебя это может обламывать, поэтому лучше каждый день, но по чуть-чуть, чем вооще ничего
		
		Короче`

		msg := tgapi.NewMessage(u.id, m)
		msg.ReplyMarkup = tgapi.NewInlineKeyboardMarkup(
			tgapi.NewInlineKeyboardRow(
				tgapi.NewInlineKeyboardButtonData("От души движению! Крепить", "++"),
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

			_, err := tg.Send(tgapi.NewMessage(userID, "Много пиздиш"))
			if err != nil {
				return err
			}

			reader := bytes.NewReader(vin)

			_, err = tg.Send(tgapi.NewSticker(
				userID, tgapi.FileReader{
					Name:   "Vin",
					Reader: reader,
				}))
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
				if err := store.UpdateUser(ctx, userID, true); err != nil {
					return err
				}
				_, err := tg.Send(tgapi.NewMessage(userID, "fuck you"))
				if err != nil {
					return err
				}

				reader := bytes.NewReader(ducalis)

				_, err = tg.Send(tgapi.NewSticker(
					userID, tgapi.FileReader{
						Name:   "Ducalis",
						Reader: reader,
					}))
				if err != nil {
					return err
				}

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
			// Не пыхал
			case "-":
				if err := store.UpdateUser(ctx, userID, false); err != nil {
					return err
				}

				reader := bytes.NewReader(coop)

				_, err := tg.Send(tgapi.NewSticker(
					userID, tgapi.FileReader{
						Name:   "Cooper",
						Reader: reader,
					}))
				if err != nil {
					return err
				}
				_, err = tg.Send(tgapi.NewMessage(userID, "good for you"))
				if err != nil {
					return err
				}

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
			}
		}
	}

	return nil
}

func starDairyJob(ctx context.Context, tg *tgapi.BotAPI, id int64) {
	_, err := tg.Send(tgapi.NewMessage(id, "Все правильно. Что же каков был день сегоднящний?"))
	if err != nil {
		panic(err)
	}

	for {
		<-waitDairy()

		_, err := tg.Send(tgapi.NewMessage(id, "Каков был день сегоднящний?"))
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
