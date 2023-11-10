package db

import (
	"github.com/wxlai90/telehitch/config"
	"github.com/wxlai90/telehitch/models"
	"github.com/wxlai90/telehitch/states"
)

var driversAvailable map[int64]*models.Driver = map[int64]*models.Driver{}
var bookings map[int64]*models.Booking = map[int64]*models.Booking{}
var archivedBookings map[int64]*models.Booking = map[int64]*models.Booking{}
var statesMap map[int64]states.State = map[int64]states.State{}

func AddNewBooking(userId int64, name string) {
	bookings[userId] = &models.Booking{
		Passenger: models.Passenger{
			UserId: userId,
			Name:   name,
		},
	}
}

func InsertBookingForUserId(userId int64, booking *models.Booking) {
	bookings[userId] = booking
}

func GetStateForUserId(userId int64) states.State {
	if GetDriverByUserId(userId) != nil && !config.IsDev {
		return states.DRIVER_STATE
	}

	if state, ok := statesMap[userId]; ok {
		return state
	}

	statesMap[userId] = states.INIT
	return states.INIT
}

func UpdateStateForUserId(userId int64, state states.State) {
	statesMap[userId] = state
}

func GetBookingForUserId(userId int64) *models.Booking {
	return bookings[userId]
}

func ArchiveBookingForUserId(userId int64) {
	if booking, ok := bookings[userId]; ok {
		archivedBookings[userId] = booking
		delete(bookings, userId)
	}
}

func AddNewDriver(userId int64) {
	driversAvailable[userId] = &models.Driver{
		UserId: userId,
	}
	delete(statesMap, userId)
}

func RemoveDriver(userId int64) {
	delete(driversAvailable, userId)
}

func GetDriverByUserId(userId int64) *models.Driver {
	return driversAvailable[userId]
}

func GetAllDrivers() []*models.Driver {
	drivers := []*models.Driver{}

	for _, driver := range driversAvailable {
		drivers = append(drivers, driver)
	}

	return drivers
}

func GetLastBookingByUserId(userId int64) *models.Booking {
	return archivedBookings[userId]
}
