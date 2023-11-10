package notifications

import (
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/wxlai90/telehitch/db"
	"github.com/wxlai90/telehitch/models"
	"github.com/wxlai90/telehitch/states"
	"github.com/wxlai90/telehitch/utils"
)

func BroadcastToDrivers(booking *models.Booking, bot *tgbotapi.BotAPI) {
	drivers := db.GetAllDrivers()

	for _, driver := range drivers {
		text := utils.FormatBookingText(booking, "New Booking - Accept?")
		msg := tgbotapi.NewMessage(driver.UserId, text)
		msg.ParseMode = "Markdown"

		kb := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("Yes", fmt.Sprintf("%s|%d", states.ACCEPT_BOOKING, booking.Passenger.UserId)),
			),
		)

		msg.ReplyMarkup = kb
		bot.Send(msg)
	}
}

func ReplyToPassenger(userId int64, booking *models.Booking, bot *tgbotapi.BotAPI) {
	text := fmt.Sprintf("%s is on the way to pick you up!\n", booking.Driver.Name)
	text += "You can chat with the driver by sending messages here"

	msg := tgbotapi.NewMessage(userId, text)
	kb := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Cancel Booking", fmt.Sprintf("%s|%d", states.PAX_CANCEL_BOOKING, booking.Passenger.UserId)),
			tgbotapi.NewInlineKeyboardButtonData("Completed", fmt.Sprintf("%s|%d", states.PAX_COMPLETED, booking.Passenger.UserId)),
		),
	)

	db.UpdateStateForUserId(userId, states.PENDING_PICKUP)
	msg.ReplyMarkup = kb
	bot.Send(msg)
}
