package models

type Booking struct {
	NumberOfPax int
	Fare        string
	Pickup      Location
	Dropoff     Location
	Distance    float64
	Passenger   Passenger
	Driver      Driver
	Channel     chan bool
}

type Driver struct {
	UserId int64
	Name   string
}

type Passenger struct {
	UserId int64
	Name   string
}

type Location struct {
	Longitude float64
	Latitude  float64
}
