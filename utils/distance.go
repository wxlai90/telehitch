package utils

import (
	"math"
)

const (
	earthRadiusMi      = 3958 // radius of the earth in miles.
	earthRaidusKm      = 6371 // radius of the earth in kilometers.
	DistanceAdjustment = 1.3  // manual adjustment of distance to cater for non-linearity
)

// Coord represents a geographic coordinate.
type Coord struct {
	Lat float64
	Lon float64
}

// degreesToRadians converts from degrees to radians.
func degreesToRadians(d float64) float64 {
	return d * math.Pi / 180
}

// Distance calculates the shortest path between two coordinates on the surface
// of the Earth. This function returns two units of measure, the first is the
// distance in miles, the second is the distance in kilometers.
func Distance(p, q Coord) float64 {
	lat1 := degreesToRadians(p.Lat)
	lon1 := degreesToRadians(p.Lon)
	lat2 := degreesToRadians(q.Lat)
	lon2 := degreesToRadians(q.Lon)

	diffLat := lat2 - lat1
	diffLon := lon2 - lon1

	a := math.Pow(math.Sin(diffLat/2), 2) + math.Cos(lat1)*math.Cos(lat2)*
		math.Pow(math.Sin(diffLon/2), 2)

	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	km := c * earthRaidusKm

	return km
}

func DistanceWrapper(lat1, lon1, lat2, lon2 float64) float64 {
	coor1, coor2 := Coord{
		Lat: lat1,
		Lon: lon1,
	},
		Coord{
			Lat: lat2,
			Lon: lon2,
		}

	return Distance(coor1, coor2)
}
