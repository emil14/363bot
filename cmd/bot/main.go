package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"math"
	"os"
	"time"

	tgapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var pg = MustNewPostgres(os.Getenv("DATABASE_URL"))

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
	log.Printf("Bot %s activated", tg.Self.UserName)

	ctx := context.Background()

	defer func() {
		if err := pg.Close(ctx); err != nil {
			panic(err)
		}
	}()

	updCfg := tgapi.NewUpdate(0)
	updCfg.Timeout = 60
	updates := tg.GetUpdatesChan(updCfg)

	go func() {
		if err := startAskJob(ctx, tg); err != nil {
			panic(err)
		}
	}()

	if err := handleUpdates(updates, ctx, tg); err != nil {
		panic(err)
	}
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
				if err := pg.AddUser(ctx, userID, userName); err != nil {
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
				user, err := pg.User(ctx, userID)
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

			// s := u.Message.Sticker

			log.Printf("%+v", u.Message)

			f := tgapi.FilePath("../../assets/dukalis.jpg")

			log.Println(f)

			file, err := os.Open("../../assets/dukalis.jpg")
			if err != nil {
				panic(err)
			}

			reader := bufio.NewReader(file)

			file_ := tgapi.FileReader{
				Name:   "Дукалис",
				Reader: reader,
			}

			_, err = tg.Send(tgapi.NewSticker(userID, file_))
			if err != nil {
				return err
			}
		}

		if u.CallbackQuery != nil {
			id := u.CallbackQuery.From.ID
			switch u.CallbackData() {
			case "+":
				if err := pg.UpdateUser(ctx, id, true); err != nil {
					return err
				}
				_, err := tg.Send(tgapi.NewMessage(id, "fuck you"))
				if err != nil {
					return err
				}
			case "-":
				if err := pg.UpdateUser(ctx, id, false); err != nil {
					return err
				}
				_, err := tg.Send(tgapi.NewMessage(id, "good for you"))
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func startAskJob(ctx context.Context, tg *tgapi.BotAPI) error {
	for {
		<-wait()

		users, err := pg.Users(ctx)
		if err != nil {
			return err
		}

		log.Println(users)

		for _, u := range users {
			_, err := tg.Send(askAboutWeedMsg(u.id))
			if err != nil {
				return err
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
	log.Println("NOW", now)

	return time.After(time.Minute)
	// yyyy, mm, dd := now.Date()
	// nextMorning := time.Date(yyyy, mm, dd+1, 10, 0, 0, 0, now.Location())
	// log.Println("MORNING", nextMorning)

	// return time.After(nextMorning.Sub(now))
}
