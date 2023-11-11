package main

import (
	"log"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/wxlai90/telehitch/config"
	"github.com/wxlai90/telehitch/db"
	"github.com/wxlai90/telehitch/handlers"
	"github.com/wxlai90/telehitch/models"
	"github.com/wxlai90/telehitch/states"
	"github.com/wxlai90/telehitch/utils"
)

func main() {
	bot, err := tgbotapi.NewBotAPI(os.Getenv("TELEGRAM_BOT_TOKEN"))
	if err != nil {
		log.Panic(err)
	}

	log.Printf("Authorized on account %s", bot.Self.UserName)

	bot.Debug = config.IsDebugMode

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		msgType := determineMsgType(update)
		log.Printf("msgType: %d\n", msgType)

		switch msgType {
		case models.TextMessage:
			_, userState := getUserStateFromTextMessage(update)

			switch userState {
			case states.INIT:
				handlers.HandleNewBooking(update, bot)
			case states.PASSENGER:
				handlers.HandlePassengerCount(update, bot)
			case states.PICKUP:
				handlers.HandlePickupLocation(update, bot)
			case states.DROPOFF:
				handlers.HandleDropoffLocation(update, bot)
			case states.FARE:
				handlers.HandleFareAmount(update, bot)
			case states.PENDING_PICKUP:
				handlers.HandleRelay(update, bot)
			case states.DRIVER_STATE:
				handlers.HandleInvalidRequest(update, bot)
			}

		case models.CommandMessage:
			userId, userState := getUserStateFromTextMessage(update)

			switch update.Message.Command() {
			case "driver":
				driver := db.GetDriverByUserId(userId)
				if driver == nil {
					msg := tgbotapi.NewMessage(userId, "You are now a driver! Stay online to receive bookings!")
					bot.Send(msg)
					db.AddNewDriver(userId)
				} else {
					msg := tgbotapi.NewMessage(userId, "You are no longer a driver!")
					bot.Send(msg)
					db.RemoveDriver(userId)
				}
			case "new":
				switch userState {
				case states.INIT:
					handlers.HandleNewBooking(update, bot)
				default:
					msg := tgbotapi.NewMessage(userId, "You have an existing booking in progress. Please complete that first.")
					bot.Send(msg)
				}
			}

		case models.CallbackMessage:
			p, err := utils.ParseCallbackData(update.CallbackQuery.Data)
			if err != nil {
				log.Printf("Error parsing callback data: %s\n", err)
				continue
			}
			log.Printf("callback data: %v\n", p)

			switch p.Selection {
			case states.ACCEPT_BOOKING:
				handlers.HandleDriverAcceptance(update, bot)
			case states.SEND_ARRIVAL:
				handlers.HandleDriverSendArrival(update, bot)
			case states.CANCEL_PICKUP:
				handlers.HandleDriverCancellation(update, bot)
			case states.PAX_CANCEL_BOOKING:
				handlers.HandlePaxCancellation(update, bot)
			case states.RE_CREATE:
				handlers.HandlePaxRecreateLastBooking(update, bot)
			case states.PAX_COMPLETED:
				handlers.HandlePaxTripCompleted(update, bot)
			}
		}
	}
}

func getUserStateFromTextMessage(update tgbotapi.Update) (int64, states.State) {
	userId := update.Message.From.ID
	userState := db.GetStateForUserId(userId)
	log.Printf("Current userState: %d\n", userState)

	return userId, userState
}

func determineMsgType(update tgbotapi.Update) models.MessageType {
	if update.Message != nil {
		if update.Message.IsCommand() {
			return models.CommandMessage
		}

		return models.TextMessage
	}

	if update.CallbackQuery != nil {
		return models.CallbackMessage
	}

	return models.UnknownMessageType
}
