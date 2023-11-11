package utils

import (
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/wxlai90/telehitch/models"
)

func FormatBookingText(booking *models.Booking, header string) string {
	text := fmt.Sprintf("*%s*\nLooking for drivers now!\n", header)
	text += fmt.Sprintf("No. of Pax: *%d*\n", booking.NumberOfPax)
	text += fmt.Sprintf("Fare: *%s*\n", booking.Fare)
	text += fmt.Sprintf("Distance: *%.2f*\n", booking.Distance)
	text += fmt.Sprintf("Google Maps: https://www.google.com/maps/dir/?api=1&origin=%f,%f&destination=%f,%f&travelmode=driving\n", booking.Pickup.Latitude, booking.Pickup.Longitude, booking.Dropoff.Latitude, booking.Dropoff.Longitude)

	return text
}

func FormatUserName(from *tgbotapi.User) string {
	if from.UserName != "" {
		return "@" + from.UserName
	}

	if from.FirstName != "" && from.LastName != "" {
		return fmt.Sprintf("%s %s", from.FirstName, from.LastName)
	}

	if from.FirstName != "" {
		return from.FirstName
	}

	if from.LastName != "" {
		return from.LastName
	}

	return "Anonymous"
}
