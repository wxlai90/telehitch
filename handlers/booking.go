package handlers

import (
	"fmt"
	"log"
	"strconv"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/wxlai90/telehitch/config"
	"github.com/wxlai90/telehitch/db"
	"github.com/wxlai90/telehitch/models"
	"github.com/wxlai90/telehitch/notifications"
	"github.com/wxlai90/telehitch/states"
	"github.com/wxlai90/telehitch/utils"
)

func HandleNewBooking(update tgbotapi.Update, bot *tgbotapi.BotAPI) {
	userId := update.Message.Chat.ID
	db.AddNewBooking(userId, utils.FormatUserName(update.Message.From))
	db.UpdateStateForUserId(userId, states.PASSENGER)

	msg := tgbotapi.NewMessage(userId, "How many passengers?")
	bot.Send(msg)
}

func HandlePassengerCount(update tgbotapi.Update, bot *tgbotapi.BotAPI) {
	userId := update.Message.Chat.ID
	msg := update.Message.Text

	noOfPax, err := strconv.Atoi(msg)
	if err != nil {
		reply := tgbotapi.NewMessage(update.Message.Chat.ID, "Number of passengers is invalid")
		bot.Send(reply)
		return
	}

	booking := db.GetBookingForUserId(userId)
	booking.NumberOfPax = noOfPax

	db.UpdateStateForUserId(userId, states.PICKUP)

	reply := tgbotapi.NewMessage(update.Message.Chat.ID, "Pickup from? (Please send a location message)")
	bot.Send(reply)
}

func HandlePickupLocation(update tgbotapi.Update, bot *tgbotapi.BotAPI) {
	userId := update.Message.Chat.ID
	location := update.Message.Location

	if config.IsDev {
		booking := db.GetBookingForUserId(userId)
		booking.Pickup = models.Location{
			Longitude: 1.23,
			Latitude:  4.56,
		}
		db.UpdateStateForUserId(userId, states.DROPOFF)

		reply := tgbotapi.NewMessage(update.Message.Chat.ID, "Drop-off at? (Please send a location message)")
		bot.Send(reply)
		return
	}

	if location == nil {
		reply := tgbotapi.NewMessage(update.Message.Chat.ID, "Pickup location is missing")
		bot.Send(reply)
		return
	}

	booking := db.GetBookingForUserId(userId)
	booking.Pickup = models.Location{
		Longitude: location.Longitude,
		Latitude:  location.Latitude,
	}
	db.UpdateStateForUserId(userId, states.DROPOFF)

	reply := tgbotapi.NewMessage(update.Message.Chat.ID, "Drop-off at? (Please send a location message)")
	bot.Send(reply)
}

func HandleDropoffLocation(update tgbotapi.Update, bot *tgbotapi.BotAPI) {
	userId := update.Message.Chat.ID
	location := update.Message.Location

	if config.IsDev {
		booking := db.GetBookingForUserId(userId)
		booking.Dropoff = models.Location{
			Longitude: 1.23,
			Latitude:  4.56,
		}
		db.UpdateStateForUserId(userId, states.FARE)

		reply := tgbotapi.NewMessage(update.Message.Chat.ID, "Fare range?")
		bot.Send(reply)
		return
	}

	if location == nil {
		reply := tgbotapi.NewMessage(update.Message.Chat.ID, "Drop-off location is missing")
		bot.Send(reply)
		return
	}

	booking := db.GetBookingForUserId(userId)
	booking.Dropoff = models.Location{
		Longitude: location.Longitude,
		Latitude:  location.Latitude,
	}
	booking.Distance = utils.DistanceWrapper(booking.Pickup.Latitude, booking.Pickup.Longitude, booking.Dropoff.Latitude, booking.Dropoff.Longitude)

	db.UpdateStateForUserId(userId, states.FARE)

	reply := tgbotapi.NewMessage(update.Message.Chat.ID, "Fare range?")
	bot.Send(reply)
}

func HandleFareAmount(update tgbotapi.Update, bot *tgbotapi.BotAPI) {
	userId := update.Message.Chat.ID
	msg := update.Message.Text

	booking := db.GetBookingForUserId(userId)
	booking.Fare = msg
	db.UpdateStateForUserId(userId, states.PENDING_DRIVER)

	text := utils.FormatBookingText(booking, "Booking Created")
	reply := tgbotapi.NewMessage(update.Message.Chat.ID, text)
	reply.ParseMode = "Markdown"

	bot.Send(reply)
	notifications.BroadcastToDrivers(booking, bot)

	c := make(chan bool)
	booking.Channel = c

	go func() {
		select {
		case <-c:
			log.Println("Booking accepted")
			return
		case <-time.After(config.BOOKING_TIMEOUT):
			paxReply := tgbotapi.NewMessage(update.Message.Chat.ID, "No drivers accepted the booking, feel free to make a new booking or re-use the last booking by selecting 'Create Last Booking'")
			kb := tgbotapi.NewInlineKeyboardMarkup(
				tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonData("Create Last Booking", fmt.Sprintf("%d|%d", states.RE_CREATE, update.Message.Chat.ID)),
				),
			)
			paxReply.ReplyMarkup = kb
			bot.Send(paxReply)
			db.ArchiveBookingForUserId(update.Message.Chat.ID)
		}
	}()
}

func HandleRelay(update tgbotapi.Update, bot *tgbotapi.BotAPI) {
	userId := update.Message.Chat.ID
	msgId := update.Message.MessageID

	booking := db.GetBookingForUserId(userId)
	toId := booking.Passenger.UserId
	fromId := userId
	if toId == fromId {
		toId = booking.Driver.UserId
	}

	msg := tgbotapi.NewForward(toId, fromId, msgId)
	bot.Send(msg)
}

func HandleInvalidRequest(update tgbotapi.Update, bot *tgbotapi.BotAPI) {
	userId := update.Message.Chat.ID

	msg := tgbotapi.NewMessage(userId, "You cannot create any booking since you are a driver. Send /driver to stop being a driver if you want to make a booking")
	bot.Send(msg)
}

func HandleDriverAcceptance(update tgbotapi.Update, bot *tgbotapi.BotAPI) {
	callback := tgbotapi.NewCallback(update.CallbackQuery.ID, "Successfully taken booking")
	if _, err := bot.Request(callback); err != nil {
		log.Printf("Couldn't display callback chattable")
	}

	msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "Please proceed to pickup location now\nYou can chat with the passenger by sending messages here")
	kb := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Cancel Pickup", fmt.Sprintf("%d|%d", states.CANCEL_PICKUP, update.CallbackQuery.From.ID)),
			tgbotapi.NewInlineKeyboardButtonData("Send Arrival Text", fmt.Sprintf("%d|%d", states.SEND_ARRIVAL, update.CallbackQuery.From.ID)),
		),
	)
	msg.ReplyMarkup = kb
	if _, err := bot.Send(msg); err != nil {
		log.Printf("Couldn't send callback message reply")
	}

	p, err := utils.ParseCallbackData(update.CallbackQuery.Data)
	if err != nil {
		log.Printf("Callback data is not userId, expected userId, gotten: %s\n", update.CallbackQuery.Data)
		return
	}

	booking := db.GetBookingForUserId(p.UserId)
	if booking == nil {
		return
	}
	booking.Driver = models.Driver{
		Name:   utils.FormatUserName(update.CallbackQuery.From),
		UserId: update.CallbackQuery.From.ID,
	}
	booking.Channel <- true
	// close(booking.Channel)
	notifications.ReplyToPassenger(p.UserId, booking, bot)
}

func HandleDriverSendArrival(update tgbotapi.Update, bot *tgbotapi.BotAPI) {
	callback := tgbotapi.NewCallback(update.CallbackQuery.ID, "Sent arrival notice")
	if _, err := bot.Request(callback); err != nil {
		log.Printf("Couldn't display callback chattable")
	}

	p, err := utils.ParseCallbackData(update.CallbackQuery.Data)
	if err != nil {
		log.Printf("Callback data is not userId, expected userId, gotten: %s\n", update.CallbackQuery.Data)
		return
	}

	booking := db.GetBookingForUserId(p.UserId)
	if booking == nil {
		return
	}
	reply := tgbotapi.NewMessage(booking.Passenger.UserId, "Driver is arriving! Get ready!")
	bot.Send(reply)
}

func HandleDriverCancellation(update tgbotapi.Update, bot *tgbotapi.BotAPI) {
	callback := tgbotapi.NewCallback(update.CallbackQuery.ID, "Cancelled pickup")
	if _, err := bot.Request(callback); err != nil {
		log.Printf("Couldn't display callback chattable")
	}

	p, err := utils.ParseCallbackData(update.CallbackQuery.Data)
	if err != nil {
		log.Printf("Callback data is not userId, expected userId, gotten: %s\n", update.CallbackQuery.Data)
		return
	}

	booking := db.GetBookingForUserId(p.UserId)
	if booking == nil {
		return
	}
	booking.Driver = models.Driver{}
	db.UpdateStateForUserId(booking.Passenger.UserId, states.PENDING_DRIVER)

	reply := tgbotapi.NewMessage(booking.Passenger.UserId, "The driver cancelled the booking, looking for drivers again!")
	bot.Send(reply)
	notifications.BroadcastToDrivers(booking, bot)
}

func HandlePaxCancellation(update tgbotapi.Update, bot *tgbotapi.BotAPI) {
	callback := tgbotapi.NewCallback(update.CallbackQuery.ID, "Cancelled booking")
	if _, err := bot.Request(callback); err != nil {
		log.Printf("Couldn't display callback chattable")
	}

	p, err := utils.ParseCallbackData(update.CallbackQuery.Data)
	if err != nil {
		log.Printf("Callback data is not userId, expected userId, gotten: %s\n", update.CallbackQuery.Data)
		return
	}

	booking := db.GetBookingForUserId(p.UserId)
	if booking == nil {
		return
	}
	db.UpdateStateForUserId(booking.Passenger.UserId, states.INIT)

	reply := tgbotapi.NewMessage(booking.Driver.UserId, "The passenger cancelled the booking, sorry!")
	bot.Send(reply)

	paxReply := tgbotapi.NewMessage(booking.Passenger.UserId, "You have cancelled the booking, feel free to make a new booking or re-use the last booking by selecting 'Create Last Booking'")
	kb := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Create Last Booking", fmt.Sprintf("%d|%d", states.RE_CREATE, update.CallbackQuery.From.ID)),
		),
	)
	paxReply.ReplyMarkup = kb
	bot.Send(paxReply)
	db.ArchiveBookingForUserId(booking.Passenger.UserId)
}

func HandlePaxRecreateLastBooking(update tgbotapi.Update, bot *tgbotapi.BotAPI) {
	callback := tgbotapi.NewCallback(update.CallbackQuery.ID, "Creating booking")
	if _, err := bot.Request(callback); err != nil {
		log.Printf("Couldn't display callback chattable")
	}

	p, err := utils.ParseCallbackData(update.CallbackQuery.Data)
	if err != nil {
		log.Printf("Callback data is not userId, expected userId, gotten: %s\n", update.CallbackQuery.Data)
		return
	}

	booking := db.GetLastBookingByUserId(p.UserId)
	if booking == nil {
		return
	}

	db.InsertBookingForUserId(p.UserId, booking)
	db.UpdateStateForUserId(p.UserId, states.PENDING_DRIVER)

	text := utils.FormatBookingText(booking, "Booking Created")
	reply := tgbotapi.NewMessage(p.UserId, text)
	reply.ParseMode = "Markdown"

	bot.Send(reply)
	notifications.BroadcastToDrivers(booking, bot)

	c := make(chan bool)
	booking.Channel = c

	go func() {
		select {
		case <-c:
			log.Println("Booking accepted")
			return
		case <-time.After(config.BOOKING_TIMEOUT):
			paxReply := tgbotapi.NewMessage(p.UserId, "No drivers accepted the booking, feel free to make a new booking or re-use the last booking by selecting 'Create Last Booking'")
			kb := tgbotapi.NewInlineKeyboardMarkup(
				tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonData("Create Last Booking", fmt.Sprintf("%d|%d", states.RE_CREATE, p.UserId)),
				),
			)
			paxReply.ReplyMarkup = kb
			bot.Send(paxReply)
			db.ArchiveBookingForUserId(p.UserId)
		}
	}()
}

func HandlePaxTripCompleted(update tgbotapi.Update, bot *tgbotapi.BotAPI) {
	callback := tgbotapi.NewCallback(update.CallbackQuery.ID, "Completed trip")
	if _, err := bot.Request(callback); err != nil {
		log.Printf("Couldn't display callback chattable")
	}

	p, err := utils.ParseCallbackData(update.CallbackQuery.Data)
	if err != nil {
		log.Printf("Callback data is not userId, expected userId, gotten: %s\n", update.CallbackQuery.Data)
		return
	}

	booking := db.GetBookingForUserId(p.UserId)
	if booking == nil {
		return
	}
	driverId := booking.Driver.UserId
	reply := tgbotapi.NewMessage(driverId, "The passenger marked the trip as completed!")
	bot.Send(reply)

	db.ArchiveBookingForUserId(p.UserId)
	db.UpdateStateForUserId(p.UserId, states.INIT)
}
