package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	tgapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var (
	token = os.Getenv("TOKEN")
	pg    = NewPGStorage()
)

func main() {
	go func() {
		log.Fatal(http.ListenAndServe(":"+os.Getenv("PORT"), nil))
	}()

	ot, err := tgapi.NewBotAPI(token)
	if err != nil {
		log.Panic(err)
	}
	ot.Debug = true
	log.Printf("Authorized on account %s", ot.Self.UserName)

	defer pg.conn.Close(context.Background())

	updCfg := tgapi.NewUpdate(0)
	updCfg.Timeout = 60
	updates := ot.GetUpdatesChan(updCfg)

	go func() {
		for u := range updates {
			if u.Message != nil {
				userID := u.Message.From.ID
				userName := u.Message.From.UserName

				if u.Message.Text == "/start" {
					if err := pg.AddUser(userID, userName); err != nil {
						panic(err)
					}

					log.Printf("New user added %s", userName)
				}

				if u.Message.Text == "/get_user" {
					user, err := pg.User(userID)
					if err != nil {
						panic(err)
					}

					msg := fmt.Sprintf(
						"you are %s\nand your fucking days without weed %d",
						user.name, user.daysWithoutWeed)

					ot.Send(tgapi.NewMessage(userID, msg))
				}
			}

			if u.CallbackQuery != nil {
				id := u.CallbackQuery.From.ID
				switch u.CallbackData() {
				case "+":
					ot.Send(tgapi.NewMessage(id, "fuck you"))
					if err := pg.UpdateUser(id, true); err != nil {
						panic(err)
					}
				case "-":
					ot.Send(tgapi.NewMessage(id, "good for you"))
					if err := pg.UpdateUser(id, false); err != nil {
						panic(err)
					}
				}
			}
		}
	}()

	for {
		users, err := pg.Users()
		if err != nil {
			panic(err)
		}

		for _, user := range users {
			msg := tgapi.NewMessage(user.id, "пыхал вчера?")
			msg.ReplyMarkup = tgapi.NewInlineKeyboardMarkup(
				tgapi.NewInlineKeyboardRow(
					tgapi.NewInlineKeyboardButtonData("Да :)", "+"),
					tgapi.NewInlineKeyboardButtonData("Нет :)", "-"),
				),
			)

			ot.Send(msg)
		}

		time.Sleep(time.Minute)
	}
}
