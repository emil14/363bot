package main

import (
	"context"
	"fmt"
	"log"
	"math"
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

					msg := fmt.Sprintf(
						"Салам алейкум, %s! Отныне я буду спрашивать, пыхал ли ты вчера и следить за твоей кармой",
						u.Message.From.FirstName,
					)

					ot.Send(tgapi.NewMessage(userID, msg))

					log.Printf("New user added %s", userName)

					send(userID, ot)
				}

				if u.Message.Text == "/get_user" {
					user, err := pg.User(userID)
					if err != nil {
						panic(err)
					}

					var msg string
					if user.daysWithoutWeed > 0 {
						msg = fmt.Sprintf("Ты не пыхал дней: %d", user.daysWithoutWeed)
					} else {
						msg = fmt.Sprintf("Ты пыхаешь дней: %v", math.Abs(float64(user.daysWithoutWeed)))
					}

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

}

func send(id int64, ot *tgapi.BotAPI) {
	msg := tgapi.NewMessage(id, "Йо, пыхал вчера?")
	msg.ReplyMarkup = tgapi.NewInlineKeyboardMarkup(
		tgapi.NewInlineKeyboardRow(
			tgapi.NewInlineKeyboardButtonData("Да :)", "+"),
			tgapi.NewInlineKeyboardButtonData("Нет !", "-"),
		),
	)

	ot.Send(msg)

	for {
		ot.Send(msg)
		time.Sleep(time.Hour * 24)
	}
}
